package agent

import (
	"bytes"
	"context"
	"encoding/json"
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
	DefaultModel = "claude-sonnet-4-20250514"
	MaxRetries   = 3
)

type Agent struct {
	client  anthropic.Client
	model   anthropic.Model
	logger  logger.Logger
	limits  *config.Limits
	verbose bool
}

type AgentConfig struct {
	APIKey  string
	Model   string
	Logger  logger.Logger
	Limits  *config.Limits
	Verbose bool
}

func NewAgent(cfg AgentConfig) (*Agent, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	model := cfg.Model
	if model == "" {
		model = DefaultModel
	}

	log := cfg.Logger
	if log == nil {
		log = logger.Null
	}

	client := anthropic.NewClient(option.WithAPIKey(cfg.APIKey))

	return &Agent{
		client:  client,
		model:   anthropic.Model(model),
		logger:  log,
		limits:  cfg.Limits,
		verbose: cfg.Verbose,
	}, nil
}

func (a *Agent) GenerateConfig(ctx context.Context, request string) (*config.CliItem, string, error) {
	var lastErr error

	for attempt := 1; attempt <= MaxRetries; attempt++ {
		if a.verbose {
			a.logger.Infof("Generating config (attempt %d/%d)", attempt, MaxRetries)
		}

		message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     a.model,
			MaxTokens: 4096,
			System: []anthropic.TextBlockParam{
				{Text: SystemPrompt},
			},
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(request)),
			},
		})
		if err != nil {
			lastErr = fmt.Errorf("API error: %w", err)
			continue
		}

		if len(message.Content) == 0 {
			lastErr = fmt.Errorf("empty response from Claude")
			continue
		}

		responseText := ""
		for _, block := range message.Content {
			if block.Type == "text" {
				responseText += block.Text
			}
		}

		responseText = cleanJSONResponse(responseText)

		cfg := &config.CliItem{}
		if err := json.Unmarshal([]byte(responseText), cfg); err != nil {
			lastErr = fmt.Errorf("invalid JSON response: %w\nResponse: %s", err, responseText)
			continue
		}

		if cfg.Item == nil {
			lastErr = fmt.Errorf("missing 'item' in config")
			continue
		}

		return cfg, responseText, nil
	}

	return nil, "", fmt.Errorf("failed after %d attempts: %w", MaxRetries, lastErr)
}

func cleanJSONResponse(response string) string {
	response = strings.TrimSpace(response)

	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		if idx := strings.LastIndex(response, "```"); idx != -1 {
			response = response[:idx]
		}
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		if idx := strings.LastIndex(response, "```"); idx != -1 {
			response = response[:idx]
		}
	}

	return strings.TrimSpace(response)
}

func (a *Agent) Execute(cfg *config.CliItem) (string, error) {
	if cfg == nil || cfg.Item == nil {
		return "", fmt.Errorf("invalid config: item is nil")
	}

	limits := cfg.Limits
	if limits == nil {
		limits = a.limits
	}

	result, err := lib.Parse(cfg.Item, limits, cfg.References, builder.NullValue, a.logger)
	if err != nil {
		return "", fmt.Errorf("execution error: %w", err)
	}

	jsonResult := result.ToJson()

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(jsonResult), "", "  "); err != nil {
		return jsonResult, nil
	}

	return prettyJSON.String(), nil
}

func (a *Agent) SetLimits(limits *config.Limits) {
	a.limits = limits
}
