package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PxyUp/fitter/lib"
	"github.com/PxyUp/fitter/pkg/agent"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/http_client"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/PxyUp/fitter/pkg/plugins/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

const referenceURI = "fitter://config-reference"

// overridden at release time via -ldflags "-X main.version=..."
var version = "dev"

type runArgs struct {
	Config string `json:"config" jsonschema:"Fitter CliItem config as a JSON or YAML string. Top-level keys: item (required), limits, references."`
	Input  string `json:"input,omitempty" jsonschema:"Optional input value (plain string or JSON), available in the config via {{{FromInput=.}}} or {{{FromInput=json.path}}} placeholders."`
}

type runFileArgs struct {
	Path  string `json:"path" jsonschema:"Absolute path to a Fitter config file (.json, .yaml or .yml) with top-level keys: item (required), limits, references."`
	Input string `json:"input,omitempty" jsonschema:"Optional input value (plain string or JSON), available in the config via {{{FromInput=.}}} or {{{FromInput=json.path}}} placeholders."`
}

type runURLArgs struct {
	URL   string `json:"url" jsonschema:"HTTP(S) URL of a Fitter config (JSON or YAML) with top-level keys: item (required), limits, references."`
	Input string `json:"input,omitempty" jsonschema:"Optional input value (plain string or JSON), available in the config via {{{FromInput=.}}} or {{{FromInput=json.path}}} placeholders."`
}

type validateArgs struct {
	Config string `json:"config" jsonschema:"Fitter CliItem config as a JSON or YAML string to validate without executing it."`
}

func parseCliItem(content []byte) (*config.CliItem, error) {
	cfg := &config.CliItem{}
	jsonErr := json.Unmarshal(content, cfg)
	if jsonErr == nil {
		return cfg, nil
	}
	cfg = &config.CliItem{}
	if yamlErr := yaml.Unmarshal(content, cfg); yamlErr != nil {
		return nil, fmt.Errorf("config is neither valid JSON (%s) nor valid YAML (%s)", jsonErr, yamlErr)
	}
	return cfg, nil
}

func runConfig(ctx context.Context, cfg *config.CliItem, input string) (result string, err error) {
	if err := agent.ValidateConfig(cfg); err != nil {
		return "", err
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fitter panicked while processing config: %v", r)
		}
	}()
	res, err := lib.ParseCtx(ctx, cfg.Item, cfg.Limits, cfg.References, builder.PureString(input), logger.Null)
	if err != nil {
		return "", err
	}
	return res.ToJson(), nil
}

// runConfigCtx guards runConfig with the MCP request context. lib.ParseCtx
// propagates ctx into the connectors, so cancellation aborts in-flight
// fetches; the select is a backstop for the parsing work between fetches,
// which is not context-aware.
func runConfigCtx(ctx context.Context, cfg *config.CliItem, input string) (string, error) {
	type parseOut struct {
		result string
		err    error
	}
	ch := make(chan parseOut, 1)
	go func() {
		result, err := runConfig(ctx, cfg, input)
		ch <- parseOut{result: result, err: err}
	}()
	select {
	case out := <-ch:
		return out.result, out.err
	case <-ctx.Done():
		return "", fmt.Errorf("request cancelled: %w", ctx.Err())
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func newServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "fitter",
		Title:   "Fitter",
		Version: version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name: "fitter_run",
		Description: "Run a Fitter scraping/parsing config passed inline (JSON or YAML) and return the extracted data as JSON. " +
			"Fitter fetches data via a connector (HTTP request, headless browser, static value, file, ...) and extracts structured data " +
			"using json/HTML/XML/xpath selectors described by a declarative model. " +
			"Call fitter_config_reference first if you are unsure about the config format.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in runArgs) (*mcp.CallToolResult, any, error) {
		cfg, err := parseCliItem([]byte(in.Config))
		if err != nil {
			return nil, nil, err
		}
		res, err := runConfigCtx(ctx, cfg, in.Input)
		if err != nil {
			return nil, nil, err
		}
		return textResult(res), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "fitter_run_file",
		Description: "Run a Fitter scraping/parsing config from a local JSON or YAML file and return the extracted data as JSON. " +
			"Same as fitter_run but reads the config from disk.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in runFileArgs) (*mcp.CallToolResult, any, error) {
		content, err := os.ReadFile(in.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to read config file %s: %w", in.Path, err)
		}
		cfg, err := parseCliItem(content)
		if err != nil {
			return nil, nil, err
		}
		res, err := runConfigCtx(ctx, cfg, in.Input)
		if err != nil {
			return nil, nil, err
		}
		return textResult(res), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "fitter_run_url",
		Description: "Run a Fitter scraping/parsing config downloaded from an HTTP(S) URL (JSON or YAML) and return the extracted data as JSON. " +
			"Same as fitter_run but fetches the config from a remote location, e.g. a raw GitHub link.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in runURLArgs) (*mcp.CallToolResult, any, error) {
		resp, err := http_client.GetDefaultClient().Get(in.URL)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to fetch config from %s: %w", in.URL, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, nil, fmt.Errorf("unable to fetch config from %s: unexpected status %s", in.URL, resp.Status)
		}
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to read config from %s: %w", in.URL, err)
		}
		cfg, err := parseCliItem(content)
		if err != nil {
			return nil, nil, err
		}
		res, err := runConfigCtx(ctx, cfg, in.Input)
		if err != nil {
			return nil, nil, err
		}
		return textResult(res), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "fitter_validate_config",
		Description: "Validate a Fitter config (JSON or YAML) without executing it. Checks the structural rules: item/connector_config/model " +
			"presence, valid response_type, and that the connector has a data source. Returns \"valid\" or the validation error. " +
			"Cheap and safe — use it while iterating on a config before calling fitter_run.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in validateArgs) (*mcp.CallToolResult, any, error) {
		cfg, err := parseCliItem([]byte(in.Config))
		if err != nil {
			return nil, nil, err
		}
		if err := agent.ValidateConfig(cfg); err != nil {
			return nil, nil, err
		}
		return textResult("valid"), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "fitter_config_reference",
		Description: "Return a condensed reference of the Fitter config format (connectors, parsers, model/field schema, placeholders, " +
			"notifiers, references, limits) with working examples. Use it before authoring a config for fitter_run.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
		return textResult(configReference), nil, nil
	})

	server.AddResource(&mcp.Resource{
		URI:         referenceURI,
		Name:        "fitter-config-reference",
		Title:       "Fitter config reference",
		Description: "Condensed reference of the Fitter config format with working examples.",
		MIMEType:    "text/markdown",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      referenceURI,
				MIMEType: "text/markdown",
				Text:     configReference,
			}},
		}, nil
	})

	return server
}

// withBearerAuth requires "Authorization: Bearer <token>" on every request
// when token is non-empty.
func withBearerAuth(next http.Handler, token string) http.Handler {
	if token == "" {
		return next
	}
	want := []byte("Bearer " + token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := []byte(r.Header.Get("Authorization"))
		if subtle.ConstantTimeCompare(got, want) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func newHTTPHandler(server *mcp.Server, authToken string, stateless bool) http.Handler {
	streamable := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		Stateless: stateless,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/mcp", withBearerAuth(streamable, authToken))
	return mux
}

func runHTTP(server *mcp.Server, addr string, authToken string, stateless bool) {
	httpServer := &http.Server{
		Addr:    addr,
		Handler: newHTTPHandler(server, authToken, stateless),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()
	log.Printf("fitter mcp listening on %s (endpoint /mcp, health /healthz, auth %s, stateless %t)",
		addr, map[bool]string{true: "bearer", false: "off"}[authToken != ""], stateless)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("mcp http server stopped with error: %s", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %s", err)
		}
	}
}

func main() {
	httpAddr := flag.String("http", os.Getenv("FITTER_MCP_HTTP_ADDR"), "serve MCP over streamable HTTP on this address (e.g. :8080) instead of stdio; env FITTER_MCP_HTTP_ADDR")
	stateless := flag.Bool("stateless", os.Getenv("FITTER_MCP_STATELESS") == "true", "run the HTTP transport without per-session state, allows load-balancing without sticky sessions; env FITTER_MCP_STATELESS=true")
	flag.Parse()

	var realStdout *os.File
	if *httpAddr == "" {
		// The stdio transport owns stdout: anything else writing there (the
		// console notifier, plugins, stray prints) would corrupt the JSON-RPC
		// stream. Keep the real stdout for the transport and point os.Stdout
		// at stderr for everybody else.
		realStdout = os.Stdout
		os.Stdout = os.Stderr
	}

	if pluginsPath := os.Getenv("FITTER_PLUGINS"); pluginsPath != "" {
		if err := store.PluginInitialize(pluginsPath); err != nil {
			log.Fatalf("unable to initialize plugins from %s: %s", pluginsPath, err)
		}
	}

	server := newServer()

	if *httpAddr != "" {
		runHTTP(server, *httpAddr, os.Getenv("FITTER_MCP_AUTH_TOKEN"), *stateless)
		return
	}

	transport := &mcp.IOTransport{Reader: os.Stdin, Writer: realStdout}
	if err := server.Run(context.Background(), transport); err != nil {
		log.Fatalf("mcp server stopped with error: %s", err)
	}
}
