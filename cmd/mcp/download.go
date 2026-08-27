package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/PxyUp/fitter/pkg/connectors"
	"github.com/PxyUp/fitter/pkg/http_client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// The run tools parse — json, HTML, XML — and parsing is exactly wrong for
// an image or a PDF: the body would go through a text parser and a template
// before it ever reached disk, mangled twice. Downloading is a different
// job: fetch the bytes, verify there are bytes, write them untouched, say
// where they landed. This tool does only that.

// downloadDir is where fetched files are written. In the container image
// /downloads is the conventional mount; operators override with
// FITTER_MCP_DOWNLOAD_DIR. Empty disables the tool rather than guessing at
// a writable path.
func downloadDir() string {
	if dir := os.Getenv("FITTER_MCP_DOWNLOAD_DIR"); dir != "" {
		return dir
	}
	if info, err := os.Stat("/downloads"); err == nil && info.IsDir() {
		return "/downloads"
	}
	return ""
}

// maxDownload bounds one fetch. Fifty megabytes covers any picture or
// document an assistant plausibly wants; anything larger is a mirror job,
// not a tool call.
const maxDownload = 50 << 20

type downloadInput struct {
	URL string `json:"url" jsonschema:"the http(s) URL to download"`
	// Filename is optional; when empty the name is derived from the URL
	// path, falling back to the content type's extension.
	Filename string `json:"filename,omitempty" jsonschema:"optional file name to save as; derived from the URL when empty"`
}

func addDownloadTool(server *mcp.Server) {
	dir := downloadDir()
	if dir == "" {
		return
	}

	mcp.AddTool(server, &mcp.Tool{
		Name: "fitter_download",
		Description: "Download a file — an image, a PDF, any binary — from a URL and save it to " +
			"the downloads folder, bytes untouched. Returns the saved file name, size and " +
			"content type. Use this instead of fitter_run when the goal is the file itself " +
			"rather than data extracted from a page.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in downloadInput) (*mcp.CallToolResult, any, error) {
		name, size, ctype, err := download(ctx, in.URL, in.Filename, dir)
		if err != nil {
			return textResult("download failed: " + err.Error()), nil, nil
		}
		return textResult(fmt.Sprintf(
			"saved %s (%d bytes, %s) to the downloads folder", name, size, ctype)), nil, nil
	})
}

func download(ctx context.Context, rawURL, filename, dir string) (name string, size int64, ctype string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", 0, "", errors.New("the url must be http or https")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", 0, "", err
	}
	// The same identifiable agent the connectors send: Go's default
	// "Go-http-client" is refused outright by common image hosts.
	req.Header.Set("User-Agent", connectors.DefaultUserAgent)
	res, err := http_client.GetDefaultClient().Do(req)
	if err != nil {
		return "", 0, "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", 0, "", fmt.Errorf("the server answered %s", res.Status)
	}

	ctype = res.Header.Get("Content-Type")
	if i := strings.IndexByte(ctype, ';'); i >= 0 {
		ctype = strings.TrimSpace(ctype[:i])
	}
	if ctype == "" {
		ctype = "application/octet-stream"
	}

	name = safeName(filename, u, ctype)
	dst := filepath.Join(dir, name)

	f, err := os.Create(dst)
	if err != nil {
		return "", 0, "", err
	}
	size, err = io.Copy(f, io.LimitReader(res.Body, maxDownload+1))
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(dst)
		return "", 0, "", err
	}
	if size > maxDownload {
		_ = os.Remove(dst)
		return "", 0, "", fmt.Errorf("larger than the %d byte limit", int64(maxDownload))
	}
	if size == 0 {
		_ = os.Remove(dst)
		return "", 0, "", errors.New("the server sent no bytes")
	}
	return name, size, ctype, nil
}

// safeName produces a plain file name with no path in it — a URL is remote
// input, and "../" in one must never choose a directory here.
func safeName(explicit string, u *url.URL, ctype string) string {
	name := explicit
	if name == "" {
		name = path.Base(u.Path)
	}
	name = filepath.Base(strings.TrimSpace(name))
	if name == "" || name == "." || name == "/" || name == string(filepath.Separator) {
		name = "download"
	}
	if filepath.Ext(name) == "" {
		if exts, _ := mime.ExtensionsByType(ctype); len(exts) > 0 {
			name += exts[0]
		}
	}
	return name
}
