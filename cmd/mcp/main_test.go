package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const staticConfig = `{
  "item": {
    "connector_config": {
      "response_type": "json",
      "static_config": {"value": "{\"a\": 1}"}
    },
    "model": {
      "base_field": {"type": "int", "path": "a"}
    }
  }
}`

func connect(t *testing.T, url string) *mcp.ClientSession {
	t.Helper()
	client := mcp.NewClient(&mcp.Implementation{Name: "fitter-test", Version: "0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: url + "/mcp"}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func textContent(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	require.Len(t, res.Content, 1)
	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	return text.Text
}

func TestHTTPTransport(t *testing.T) {
	ts := httptest.NewServer(newHTTPHandler(newServer(), "", false))
	// registered before connect() registers session cleanup: LIFO order
	// closes the session first, otherwise Close waits on the open SSE stream
	t.Cleanup(ts.Close)

	session := connect(t, ts.URL)

	tools, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)
	names := make([]string, 0, len(tools.Tools))
	for _, tool := range tools.Tools {
		names = append(names, tool.Name)
	}
	assert.ElementsMatch(t, []string{"fitter_run", "fitter_run_file", "fitter_run_url", "fitter_validate_config", "fitter_config_reference"}, names)

	validate, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "fitter_validate_config",
		Arguments: map[string]any{"config": staticConfig},
	})
	require.NoError(t, err)
	assert.Equal(t, "valid", textContent(t, validate))

	run, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "fitter_run",
		Arguments: map[string]any{"config": staticConfig},
	})
	require.NoError(t, err)
	assert.Equal(t, "1", textContent(t, run))

	reference, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: referenceURI})
	require.NoError(t, err)
	require.Len(t, reference.Contents, 1)
	assert.True(t, strings.Contains(reference.Contents[0].Text, "connector_config"))
}

func TestHTTPTransportStateless(t *testing.T) {
	ts := httptest.NewServer(newHTTPHandler(newServer(), "", true))
	t.Cleanup(ts.Close)

	session := connect(t, ts.URL)
	run, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "fitter_run",
		Arguments: map[string]any{"config": staticConfig},
	})
	require.NoError(t, err)
	assert.Equal(t, "1", textContent(t, run))
}

func TestBearerAuth(t *testing.T) {
	ts := httptest.NewServer(newHTTPHandler(newServer(), "secret-token", false))
	t.Cleanup(ts.Close)

	resp, err := http.Post(ts.URL+"/mcp", "application/json", strings.NewReader("{}"))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/mcp", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer wrong")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// health endpoint stays open
	resp, err = http.Get(ts.URL + "/healthz")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// with the right token the full MCP handshake works
	client := mcp.NewClient(&mcp.Implementation{Name: "fitter-test", Version: "0"}, nil)
	authed, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint: ts.URL + "/mcp",
		HTTPClient: &http.Client{
			Transport: authRoundTripper{token: "secret-token"},
		},
	}, nil)
	require.NoError(t, err)
	defer authed.Close()

	res, err := authed.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "fitter_validate_config",
		Arguments: map[string]any{"config": staticConfig},
	})
	require.NoError(t, err)
	assert.Equal(t, "valid", textContent(t, res))
}

type authRoundTripper struct {
	token string
}

func (a authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+a.token)
	return http.DefaultTransport.RoundTrip(req)
}
