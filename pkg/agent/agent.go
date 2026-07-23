package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/PxyUp/fitter/lib"
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	DefaultModel = string(anthropic.ModelClaudeOpus4_8)
	DefaultEffort = string(anthropic.OutputConfigEffortHigh)

	// MaxTokens stays under the SDK HTTP timeout for non streaming requests.
	MaxTokens = 16000

	// MaxRepairAttempts bounds how many times we hand a validation error back to
	// the model. Transport errors are not counted here - the SDK already retries
	// 429/5xx/connection failures on its own.
	MaxRepairAttempts = 3
)

// configEnvelopeSchema constrains the response to a flat object we can always
// parse. The Fitter config itself is carried as a JSON string because a Fitter
// model nests into itself through "generated" and structured outputs reject
// recursive schemas.
var configEnvelopeSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"config": map[string]any{
			"type":        "string",
			"description": "The complete Fitter CliItem configuration, as a JSON string.",
		},
		"notes": map[string]any{
			"type":        "string",
			"description": "One or two sentences describing what the config extracts.",
		},
	},
	"required":             []string{"config", "notes"},
	"additionalProperties": false,
}

type configEnvelope struct {
	Config string `json:"config"`
	Notes  string `json:"notes"`
}

// Result is a validated config plus the material needed to display it.
type Result struct {
	Config *config.CliItem
	// JSON is the config, re-indented for display.
	JSON  string
	Notes string
}

type Agent struct {
	client  anthropic.Client
	model   anthropic.Model
	effort  anthropic.OutputConfigEffort
	logger  logger.Logger
	limits  *config.Limits
	verbose bool

	// history carries the conversation so follow up requests can refine the
	// previous config instead of restating it from scratch.
	history []anthropic.MessageParam
}

type AgentConfig struct {
	// APIKey is optional. When empty the SDK resolves credentials from the
	// environment (ANTHROPIC_API_KEY and friends).
	APIKey  string
	Model   string
	Effort  string
	Logger  logger.Logger
	Limits  *config.Limits
	Verbose bool
}

func NewAgent(cfg AgentConfig) (*Agent, error) {
	model := cfg.Model
	if model == "" {
		model = DefaultModel
	}

	effort, err := parseEffort(cfg.Effort)
	if err != nil {
		return nil, err
	}

	log := cfg.Logger
	if log == nil {
		log = logger.Null
	}

	var opts []option.RequestOption
	if cfg.APIKey != "" {
		opts = append(opts, option.WithAPIKey(cfg.APIKey))
	}

	return &Agent{
		client:  anthropic.NewClient(opts...),
		model:   anthropic.Model(model),
		effort:  effort,
		logger:  log,
		limits:  cfg.Limits,
		verbose: cfg.Verbose,
	}, nil
}

func parseEffort(value string) (anthropic.OutputConfigEffort, error) {
	if value == "" {
		return anthropic.OutputConfigEffortHigh, nil
	}
	switch anthropic.OutputConfigEffort(value) {
	case anthropic.OutputConfigEffortLow:
		return anthropic.OutputConfigEffortLow, nil
	case anthropic.OutputConfigEffortMedium:
		return anthropic.OutputConfigEffortMedium, nil
	case anthropic.OutputConfigEffortHigh:
		return anthropic.OutputConfigEffortHigh, nil
	case anthropic.OutputConfigEffortXhigh:
		return anthropic.OutputConfigEffortXhigh, nil
	case anthropic.OutputConfigEffortMax:
		return anthropic.OutputConfigEffortMax, nil
	}
	return "", fmt.Errorf("invalid effort %q (want low, medium, high, xhigh or max)", value)
}

// GenerateConfig turns a natural language request into a validated Fitter
// config. Follow up calls on the same Agent refine the previous result, so
// "make it return 20 items instead" works without repeating the original ask.
func (a *Agent) GenerateConfig(ctx context.Context, request string) (*Result, error) {
	// Remember where this turn starts. A turn can append several messages once
	// the repair loop kicks in, and all of them have to go if it fails.
	mark := len(a.history)

	a.history = append(a.history, anthropic.NewUserMessage(anthropic.NewTextBlock(request)))

	result, err := a.generate(ctx)
	if err != nil {
		// Drop the whole failed exchange so the next request starts from the
		// last state we know was good.
		a.history = a.history[:mark]
		return nil, err
	}

	return result, nil
}

// Reset clears the conversation. The next request starts a fresh session.
func (a *Agent) Reset() {
	a.history = nil
}

// HasHistory reports whether a previous config exists to refine.
func (a *Agent) HasHistory() bool {
	return len(a.history) > 0
}

func (a *Agent) generate(ctx context.Context) (*Result, error) {
	for attempt := 1; ; attempt++ {
		if a.verbose {
			a.logger.Infof("Generating config (attempt %d/%d)", attempt, MaxRepairAttempts)
		}

		message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     a.model,
			MaxTokens: MaxTokens,
			System: []anthropic.TextBlockParam{
				{Text: SystemPrompt},
			},
			Thinking: anthropic.ThinkingConfigParamUnion{
				OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{},
			},
			OutputConfig: anthropic.OutputConfigParam{
				Effort: a.effort,
				Format: anthropic.JSONOutputFormatParam{Schema: configEnvelopeSchema},
			},
			// Cache the conversation prefix so refinement turns re-read the
			// history instead of reprocessing it.
			CacheControl: anthropic.NewCacheControlEphemeralParam(),
			Messages:     a.history,
		})
		if err != nil {
			return nil, a.describeAPIError(err)
		}

		if err := stopReasonError(message.StopReason); err != nil {
			return nil, err
		}

		result, parseErr := parseResult(responseText(message))

		// Keep the assistant turn either way: on success it becomes the base for
		// the next refinement, on failure the model needs to see what it wrote.
		a.history = append(a.history, message.ToParam())

		if parseErr == nil {
			if a.verbose && result.Notes != "" {
				a.logger.Infof("Model notes: %s", result.Notes)
			}
			return result, nil
		}

		if attempt >= MaxRepairAttempts {
			return nil, fmt.Errorf("model did not produce a usable config after %d attempts: %w", MaxRepairAttempts, parseErr)
		}

		if a.verbose {
			a.logger.Infof("Invalid config, asking model to fix it: %s", parseErr)
		}

		// Hand the concrete failure back so the retry is a repair rather than a
		// blind re-roll of the same prompt.
		a.history = append(a.history, anthropic.NewUserMessage(anthropic.NewTextBlock(
			fmt.Sprintf("That config is not valid: %s\n\nReturn a corrected config that fixes exactly this problem.", parseErr),
		)))
	}
}

func responseText(message *anthropic.Message) string {
	var sb strings.Builder
	for _, block := range message.Content {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			sb.WriteString(text.Text)
		}
	}
	return sb.String()
}

func stopReasonError(reason anthropic.StopReason) error {
	switch reason {
	case anthropic.StopReasonMaxTokens:
		return fmt.Errorf("response hit the %d token limit before the config was complete; try a narrower request", MaxTokens)
	case anthropic.StopReasonRefusal:
		return errors.New("the model declined this request")
	}
	return nil
}

func parseResult(text string) (*Result, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("empty response from the model")
	}

	envelope := &configEnvelope{}
	if err := json.Unmarshal([]byte(text), envelope); err != nil {
		return nil, fmt.Errorf("response was not the expected JSON envelope: %w", err)
	}

	cfg := &config.CliItem{}
	if err := json.Unmarshal([]byte(envelope.Config), cfg); err != nil {
		return nil, fmt.Errorf("the \"config\" field is not valid JSON: %w", err)
	}

	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	return &Result{
		Config: cfg,
		JSON:   indentJSON(envelope.Config),
		Notes:  envelope.Notes,
	}, nil
}

// ValidateConfig catches the structural mistakes that would otherwise surface
// as a confusing nil dereference or an empty result at execution time.
func ValidateConfig(cfg *config.CliItem) error {
	if cfg == nil || cfg.Item == nil {
		return errors.New(`missing "item" object at the top level`)
	}

	connector := cfg.Item.ConnectorConfig
	if connector == nil {
		return errors.New(`"item" is missing "connector_config"`)
	}

	switch connector.ResponseType {
	case config.Json, config.HTML, config.XML, config.XPath:
	case "":
		return errors.New(`"connector_config" is missing "response_type"`)
	default:
		return fmt.Errorf(`"connector_config" has invalid "response_type" %q (want json, HTML, XML or xpath)`, connector.ResponseType)
	}

	if connector.Url == "" &&
		connector.StaticConfig == nil &&
		connector.FileConfig == nil &&
		connector.IntSequenceConfig == nil &&
		connector.ReferenceConfig == nil &&
		connector.PluginConnectorConfig == nil {
		return errors.New(`"connector_config" needs a "url" or one of static_config/file_config/int_sequence_config/reference_config/plugin_connector_config`)
	}

	model := cfg.Item.Model
	if model == nil {
		return errors.New(`"item" is missing "model"`)
	}
	if model.ObjectConfig == nil && model.ArrayConfig == nil && model.BaseField == nil {
		return errors.New(`"model" must define one of "object_config", "array_config" or "base_field"`)
	}

	return nil
}

// describeAPIError turns transport failures into something a CLI user can act
// on. The SDK has already retried anything retryable by the time we get here.
func (a *Agent) describeAPIError(err error) error {
	var apiErr *anthropic.Error
	if !errors.As(err, &apiErr) {
		return fmt.Errorf("could not reach the Anthropic API: %w", err)
	}

	switch {
	case apiErr.StatusCode == 401:
		return fmt.Errorf("authentication failed - set ANTHROPIC_API_KEY or pass --api-key: %w", err)
	case apiErr.StatusCode == 403:
		return fmt.Errorf("this API key is not allowed to use model %q: %w", a.model, err)
	case apiErr.StatusCode == 404:
		return fmt.Errorf("unknown model %q - check --model: %w", a.model, err)
	case apiErr.StatusCode == 429:
		return fmt.Errorf("rate limited, please retry in a moment: %w", err)
	case apiErr.StatusCode >= 500:
		return fmt.Errorf("the Anthropic API is unavailable right now: %w", err)
	}

	return fmt.Errorf("API request failed: %w", err)
}

func indentJSON(raw string) string {
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(raw), "", "  "); err != nil {
		return raw
	}
	return out.String()
}

func (a *Agent) Execute(cfg *config.CliItem) (string, error) {
	if cfg == nil || cfg.Item == nil {
		return "", errors.New("invalid config: item is nil")
	}

	limits := cfg.Limits
	if limits == nil {
		limits = a.limits
	}

	result, err := lib.Parse(cfg.Item, limits, cfg.References, builder.NullValue, a.logger)
	if err != nil {
		return "", fmt.Errorf("execution error: %w", err)
	}

	return indentJSON(result.ToJson()), nil
}

func (a *Agent) SetLimits(limits *config.Limits) {
	a.limits = limits
}
