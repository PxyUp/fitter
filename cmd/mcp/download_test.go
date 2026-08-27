package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Binary bytes arrive on disk untouched — the whole reason this tool exists
// beside the parsing tools, which would put an image through a text parser
// and a template and mangle it twice.
func TestDownloadWritesBinaryBytesUntouched(t *testing.T) {
	payload := []byte{0x89, 'P', 'N', 'G', 0x00, 0xFF, 0x1B, '\n', 0x00}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(payload)
	}))
	t.Cleanup(ts.Close)

	dir := t.TempDir()
	name, size, ctype, err := download(context.Background(), ts.URL+"/cat.png", "", dir)
	require.NoError(t, err)
	assert.Equal(t, "cat.png", name)
	assert.Equal(t, int64(len(payload)), size)
	assert.Equal(t, "image/png", ctype)

	got, err := os.ReadFile(filepath.Join(dir, "cat.png"))
	require.NoError(t, err)
	assert.Equal(t, payload, got, "the bytes changed on the way to disk")
}

// A URL is remote input; a path inside one must never choose the directory.
func TestAFilenameCannotEscapeTheDownloadsDir(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte("jpg"))
	}))
	t.Cleanup(ts.Close)

	dir := t.TempDir()
	name, _, _, err := download(context.Background(), ts.URL+"/a", "../../etc/evil", dir)
	require.NoError(t, err)
	// The extension comes from the platform's mime table (.jpg or .jfif —
	// either is fine); what matters is the name is bare and lands in dir.
	assert.True(t, strings.HasPrefix(name, "evil."), "name = %q", name)
	assert.NotContains(t, name, "/")
	_, err = os.Stat(filepath.Join(dir, name))
	assert.NoError(t, err, "the file must land inside the downloads dir")
	entries, _ := os.ReadDir(filepath.Dir(dir))
	for _, e := range entries {
		assert.NotEqual(t, "etc", e.Name())
	}
}

// Failure shapes come back as errors, and nothing half-written stays behind.
func TestFailuresLeaveNoFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/missing":
			w.WriteHeader(http.StatusNotFound)
		case "/empty":
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(ts.Close)

	dir := t.TempDir()
	for _, p := range []string{"/missing", "/empty"} {
		_, _, _, err := download(context.Background(), ts.URL+p, "x.bin", dir)
		require.Error(t, err, p)
	}
	if _, _, _, err := download(context.Background(), "ftp://example.com/f", "", dir); err == nil {
		t.Error("a non-http scheme was accepted")
	}
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries, "a failed download left a file behind")
}

// The tool only exists when a downloads directory does; a server without one
// must not advertise a capability it cannot honour.
func TestNoDownloadsDirNoTool(t *testing.T) {
	t.Setenv("FITTER_MCP_DOWNLOAD_DIR", "")
	if _, err := os.Stat("/downloads"); err == nil {
		t.Skip("/downloads exists on this machine; the fallback applies")
	}
	assert.Empty(t, downloadDir())
}
