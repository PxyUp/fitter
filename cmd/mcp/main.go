package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

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

const (
	version      = "1.0.0"
	referenceURI = "fitter://config-reference"
)

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

func runConfig(cfg *config.CliItem, input string) (result string, err error) {
	if err := agent.ValidateConfig(cfg); err != nil {
		return "", err
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fitter panicked while processing config: %v", r)
		}
	}()
	res, err := lib.Parse(cfg.Item, cfg.Limits, cfg.References, builder.PureString(input), logger.Null)
	if err != nil {
		return "", err
	}
	return res.ToJson(), nil
}

// runConfigCtx makes runConfig cancellable from the MCP client's side.
// lib.Parse has no context support, so on cancellation the parse keeps
// running in a goroutine until it finishes on its own; we just stop waiting.
func runConfigCtx(ctx context.Context, cfg *config.CliItem, input string) (string, error) {
	type parseOut struct {
		result string
		err    error
	}
	ch := make(chan parseOut, 1)
	go func() {
		result, err := runConfig(cfg, input)
		ch <- parseOut{result: result, err: err}
	}()
	select {
	case out := <-ch:
		return out.result, out.err
	case <-ctx.Done():
		return "", fmt.Errorf("request cancelled; the running fitter config is abandoned and will finish in the background: %w", ctx.Err())
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func main() {
	// The stdio transport owns stdout: anything else writing there (the
	// console notifier, plugins, stray prints) would corrupt the JSON-RPC
	// stream. Keep the real stdout for the transport and point os.Stdout
	// at stderr for everybody else.
	realStdout := os.Stdout
	os.Stdout = os.Stderr

	if pluginsPath := os.Getenv("FITTER_PLUGINS"); pluginsPath != "" {
		if err := store.PluginInitialize(pluginsPath); err != nil {
			log.Fatalf("unable to initialize plugins from %s: %s", pluginsPath, err)
		}
	}

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

	transport := &mcp.IOTransport{Reader: os.Stdin, Writer: realStdout}
	if err := server.Run(context.Background(), transport); err != nil {
		log.Fatalf("mcp server stopped with error: %s", err)
	}
}
