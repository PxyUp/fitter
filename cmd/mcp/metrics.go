package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics for the MCP server, opt-in via -metrics-addr.
//
// They live on their own listener rather than on the MCP mux so they work
// with both transports (a stdio server has no mux at all) and so operators
// can bind them somewhere more private than the MCP endpoint — metrics are
// typically scraped without auth, and the MCP port may carry a bearer token.

var (
	mcpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "fitter_mcp_requests_total",
		Help: "MCP requests received, by JSON-RPC method.",
	}, []string{"method"})

	toolSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "fitter_mcp_tool_seconds",
		Help: "Tool call duration, by tool and outcome.",
		// Tool calls range from instant (the config reference) to a
		// browser-rendered fetch that takes tens of seconds.
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120},
	}, []string{"tool", "outcome"})
)

// metricsMiddleware records every request, and times tool calls by name.
func metricsMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			mcpRequests.WithLabelValues(method).Inc()

			call, ok := req.(*mcp.CallToolRequest)
			if !ok {
				return next(ctx, method, req)
			}

			start := time.Now()
			res, err := next(ctx, method, req)
			outcome := "ok"
			switch {
			case err != nil:
				outcome = "error"
			default:
				// Tool failures are usually reported in-band as IsError
				// rather than as a protocol error; count them as errors too,
				// or the histogram calls every failed fetch a success.
				if r, ok := res.(*mcp.CallToolResult); ok && r.IsError {
					outcome = "error"
				}
			}
			toolSeconds.WithLabelValues(call.Params.Name, outcome).
				Observe(time.Since(start).Seconds())
			return res, err
		}
	}
}

// serveMetrics exposes /metrics (and /healthz, for symmetry with the MCP
// listener) on its own address. Errors are logged, not fatal: a server that
// cannot be observed still serves.
func serveMetrics(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("fitter mcp metrics listening on %s (endpoint /metrics)", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("metrics server stopped with error: %s", err)
	}
}
