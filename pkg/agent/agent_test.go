package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// envelope builds the structured-output payload the model is constrained to.
func envelope(t *testing.T, cfg string) string {
	t.Helper()
	raw, err := json.Marshal(configEnvelope{Config: cfg, Notes: "test notes"})
	require.NoError(t, err)
	return string(raw)
}

const validConfig = `{
  "item": {
    "connector_config": {
      "response_type": "json",
      "url": "https://api.example.com/price",
      "server_config": { "method": "GET" }
    },
    "model": {
      "object_config": {
        "fields": {
          "price": { "base_field": { "type": "float", "path": "usd" } }
        }
      }
    }
  }
}`

func TestParseResultValid(t *testing.T) {
	result, err := parseResult(envelope(t, validConfig))
	require.NoError(t, err)

	require.NotNil(t, result.Config.Item)
	assert.Equal(t, "https://api.example.com/price", result.Config.Item.ConnectorConfig.Url)
	assert.Equal(t, "test notes", result.Notes)
	// JSON is re-indented for display, so it should not match the raw string
	// byte for byte but must still parse.
	assert.Contains(t, result.JSON, "connector_config")
	assert.True(t, json.Valid([]byte(result.JSON)))
}

func TestParseResultRejectsBadPayloads(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"empty response", ""},
		{"whitespace only", "   \n  "},
		{"not an envelope", `{"item": {}}`},
		{"config is not json", envelope(t, "this is not json")},
		{"config is a json fragment", envelope(t, `{"item":`)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseResult(tc.text)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

// A plain config with no envelope used to be accepted after markdown stripping.
// It must now be rejected so the repair loop can correct it.
func TestParseResultRejectsMarkdownFencedConfig(t *testing.T) {
	_, err := parseResult("```json\n" + validConfig + "\n```")
	assert.Error(t, err)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name:    "missing item",
			config:  `{}`,
			wantErr: `missing "item"`,
		},
		{
			name:    "missing connector_config",
			config:  `{"item": {"model": {"base_field": {"type": "string"}}}}`,
			wantErr: `missing "connector_config"`,
		},
		{
			name:    "missing response_type",
			config:  `{"item": {"connector_config": {"url": "https://x.dev"}, "model": {"base_field": {"type": "string"}}}}`,
			wantErr: `missing "response_type"`,
		},
		{
			name:    "invalid response_type",
			config:  `{"item": {"connector_config": {"response_type": "yaml", "url": "https://x.dev"}, "model": {"base_field": {"type": "string"}}}}`,
			wantErr: `invalid "response_type"`,
		},
		{
			name:    "no connector source",
			config:  `{"item": {"connector_config": {"response_type": "json"}, "model": {"base_field": {"type": "string"}}}}`,
			wantErr: `needs a "url"`,
		},
		{
			name:    "missing model",
			config:  `{"item": {"connector_config": {"response_type": "json", "url": "https://x.dev"}}}`,
			wantErr: `missing "model"`,
		},
		{
			name:    "empty model",
			config:  `{"item": {"connector_config": {"response_type": "json", "url": "https://x.dev"}, "model": {}}}`,
			wantErr: `must define one of`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseResult(envelope(t, tc.config))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// static_config is a valid alternative to a url, so it must not be rejected.
func TestValidateConfigAcceptsStaticConnector(t *testing.T) {
	cfg := `{"item": {"connector_config": {"response_type": "json", "static_config": {"value": "{}"}}, "model": {"base_field": {"type": "string"}}}}`
	_, err := parseResult(envelope(t, cfg))
	assert.NoError(t, err)
}

func TestParseEffort(t *testing.T) {
	valid := map[string]anthropic.OutputConfigEffort{
		"":       anthropic.OutputConfigEffortHigh,
		"low":    anthropic.OutputConfigEffortLow,
		"medium": anthropic.OutputConfigEffortMedium,
		"high":   anthropic.OutputConfigEffortHigh,
		"xhigh":  anthropic.OutputConfigEffortXhigh,
		"max":    anthropic.OutputConfigEffortMax,
	}

	for input, want := range valid {
		got, err := parseEffort(input)
		require.NoError(t, err, "effort %q", input)
		assert.Equal(t, want, got)
	}

	for _, input := range []string{"HIGH", "extreme", "1"} {
		_, err := parseEffort(input)
		assert.Error(t, err, "effort %q should be rejected", input)
	}
}

func TestDefaultEffortIsValid(t *testing.T) {
	_, err := parseEffort(DefaultEffort)
	assert.NoError(t, err)
}

func TestStopReasonError(t *testing.T) {
	assert.NoError(t, stopReasonError(anthropic.StopReasonEndTurn))
	assert.Error(t, stopReasonError(anthropic.StopReasonMaxTokens))
	assert.Error(t, stopReasonError(anthropic.StopReasonRefusal))
}

// newTestAgent points an Agent at a local server so the request path can be
// exercised without credentials or network access.
func newTestAgent(t *testing.T, handler http.HandlerFunc) *Agent {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &Agent{
		client: anthropic.NewClient(
			option.WithBaseURL(server.URL),
			option.WithAPIKey("test-key"),
			option.WithMaxRetries(0),
		),
		model:  anthropic.ModelClaudeOpus4_8,
		effort: anthropic.OutputConfigEffortHigh,
		logger: logger.Null,
	}
}

// messageResponse renders the wire shape of a Messages API reply.
func messageResponse(text string) string {
	body, _ := json.Marshal(map[string]any{
		"id":          "msg_test",
		"type":        "message",
		"role":        "assistant",
		"model":       "claude-opus-4-8",
		"content":     []map[string]any{{"type": "text", "text": text}},
		"stop_reason": "end_turn",
		"usage":       map[string]any{"input_tokens": 1, "output_tokens": 1},
	})
	return string(body)
}

// An invalid first response should be repaired in a follow-up turn rather than
// surfacing to the user.
func TestGenerateConfigRepairsInvalidResponse(t *testing.T) {
	var requests [][]byte
	broken := envelope(t, `{"item": {"connector_config": {"response_type": "json"}}}`)

	a := newTestAgent(t, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		requests = append(requests, body)

		w.Header().Set("Content-Type", "application/json")
		if len(requests) == 1 {
			_, _ = w.Write([]byte(messageResponse(broken)))
			return
		}
		_, _ = w.Write([]byte(messageResponse(envelope(t, validConfig))))
	})

	result, err := a.GenerateConfig(context.Background(), "get the bitcoin price")
	require.NoError(t, err)
	require.NotNil(t, result.Config.Item)
	assert.Equal(t, "https://api.example.com/price", result.Config.Item.ConnectorConfig.Url)

	// The second request must carry the concrete validation failure so the model
	// can correct it instead of blindly regenerating.
	require.Len(t, requests, 2)
	assert.Contains(t, string(requests[1]), "not valid")
	assert.Contains(t, string(requests[1]), "connector_config")
}

// The repair loop must terminate rather than hammering the API forever.
func TestGenerateConfigGivesUpAfterMaxRepairAttempts(t *testing.T) {
	calls := 0

	a := newTestAgent(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(messageResponse(envelope(t, `{}`))))
	})

	_, err := a.GenerateConfig(context.Background(), "get something")
	require.Error(t, err)
	assert.Equal(t, MaxRepairAttempts, calls)
	// A failed turn leaves no residue behind.
	assert.False(t, a.HasHistory())
}

// A successful turn is retained so the next request can refine it.
func TestGenerateConfigKeepsHistoryForRefinement(t *testing.T) {
	a := newTestAgent(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(messageResponse(envelope(t, validConfig))))
	})

	_, err := a.GenerateConfig(context.Background(), "get the bitcoin price")
	require.NoError(t, err)
	assert.True(t, a.HasHistory())

	_, err = a.GenerateConfig(context.Background(), "only return 5 items")
	require.NoError(t, err)

	// user + assistant per turn.
	assert.Len(t, a.history, 4)
}

// An auth failure must be reported in terms the user can act on, and must not
// leave a half-finished turn in the conversation.
func TestGenerateConfigSurfacesAuthError(t *testing.T) {
	a := newTestAgent(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key"}}`))
	})

	_, err := a.GenerateConfig(context.Background(), "get the bitcoin price")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ANTHROPIC_API_KEY")
	assert.False(t, a.HasHistory())
}

// A failed turn must not discard the config the user already had.
func TestGenerateConfigPreservesHistoryAcrossFailure(t *testing.T) {
	fail := false

	a := newTestAgent(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if fail {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"type":"error","error":{"type":"authentication_error","message":"nope"}}`))
			return
		}
		_, _ = w.Write([]byte(messageResponse(envelope(t, validConfig))))
	})

	_, err := a.GenerateConfig(context.Background(), "get the bitcoin price")
	require.NoError(t, err)
	good := len(a.history)

	fail = true
	_, err = a.GenerateConfig(context.Background(), "now break it")
	require.Error(t, err)

	assert.Len(t, a.history, good)
	assert.Equal(t, anthropic.MessageParamRoleAssistant, a.history[len(a.history)-1].Role)
}

func TestReset(t *testing.T) {
	a := &Agent{}
	a.history = append(a.history, anthropic.NewUserMessage(anthropic.NewTextBlock("request")))
	require.True(t, a.HasHistory())

	a.Reset()
	assert.False(t, a.HasHistory())
}

func TestIndentJSONLeavesInvalidInputAlone(t *testing.T) {
	assert.Equal(t, "not json", indentJSON("not json"))
	assert.Contains(t, indentJSON(`{"a":1}`), "\n")
}
