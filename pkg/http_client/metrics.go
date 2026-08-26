package http_client

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Every outbound request fitter makes goes through the client built here, so
// this is the one place a wrapped RoundTripper sees them all: connectors,
// notifiers, file downloads, reference lookups. Browser-rendered fetches go
// through a real browser, not this client, and are timed at the tool level
// instead.
//
// The host label is the request's host, which is bounded by the configs an
// operator actually runs — a scraper talks to the sites it was told to
// scrape. Recording is always on and costs nanoseconds; whether anything
// reads it is decided by whoever serves a /metrics endpoint (the MCP
// server's -metrics-addr, for one).
var requestSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "fitter_http_client_request_seconds",
	Help:    "Outbound HTTP request duration, by host, method and status.",
	Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
}, []string{"host", "method", "status"})

type metricsRoundTripper struct {
	next http.RoundTripper
}

func (m metricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := m.next.RoundTrip(req)
	status := "error"
	if err == nil && resp != nil {
		status = strconv.Itoa(resp.StatusCode)
	}
	// A malformed request reaches here too — the transport underneath
	// answers a nil URL with an error, and observing it must not turn that
	// error into a panic. The wrapper is never allowed to be less safe than
	// what it wraps.
	host := ""
	if req.URL != nil {
		host = req.URL.Host
	}
	requestSeconds.WithLabelValues(host, req.Method, status).
		Observe(time.Since(start).Seconds())
	return resp, err
}

// WrapTransport instruments a transport with the request metrics. nil wraps
// the default transport, mirroring how http.Client treats a nil Transport.
func WrapTransport(next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return metricsRoundTripper{next: next}
}
