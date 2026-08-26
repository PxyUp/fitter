package main

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// observations reads how many samples a histogram series has seen — the
// histogram equivalent of testutil.ToFloat64, which only reads counters.
func observations(t *testing.T, tool, outcome string) uint64 {
	t.Helper()
	h, ok := toolSeconds.WithLabelValues(tool, outcome).(prometheus.Histogram)
	require.True(t, ok)
	m := &dto.Metric{}
	require.NoError(t, h.Write(m))
	return m.GetHistogram().GetSampleCount()
}

// A tool call is counted under its own name, a failed one under outcome
// "error" — including failures reported in-band as IsError, which is how
// tools usually fail and which arrives with err == nil.
func TestMetricsMiddlewareCountsToolCalls(t *testing.T) {
	server := newServer()
	server.AddReceivingMiddleware(metricsMiddleware())
	ts := httptest.NewServer(newHTTPHandler(server, "", false))
	t.Cleanup(ts.Close)

	session := connect(t, ts.URL)

	okBefore := observations(t, "fitter_validate_config", "ok")
	errBefore := observations(t, "fitter_validate_config", "error")

	valid, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "fitter_validate_config",
		Arguments: map[string]any{"config": staticConfig},
	})
	require.NoError(t, err)
	require.False(t, valid.IsError)

	invalid, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "fitter_validate_config",
		Arguments: map[string]any{"config": "{not json"},
	})
	require.NoError(t, err, "an invalid config is an in-band IsError, not a protocol error")
	require.True(t, invalid.IsError)

	assert.Equal(t, okBefore+1, observations(t, "fitter_validate_config", "ok"),
		"the successful call was not counted as ok")
	assert.Equal(t, errBefore+1, observations(t, "fitter_validate_config", "error"),
		"the IsError result was not counted as error")

	assert.GreaterOrEqual(t,
		testutil.ToFloat64(mcpRequests.WithLabelValues("tools/call")), 2.0,
		"tools/call requests were not counted by method")
}

// Without the flag no middleware is registered, so a plain server records
// nothing — being unobserved must cost nothing.
func TestMetricsAreOptIn(t *testing.T) {
	ts := httptest.NewServer(newHTTPHandler(newServer(), "", false))
	t.Cleanup(ts.Close)

	session := connect(t, ts.URL)
	before := observations(t, "fitter_config_reference", "ok")

	_, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "fitter_config_reference",
	})
	require.NoError(t, err)

	assert.Equal(t, before, observations(t, "fitter_config_reference", "ok"),
		"a server without the middleware still recorded a tool call")
}
