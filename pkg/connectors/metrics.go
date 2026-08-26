package connectors

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// The retry wrapper swallows failures until the last attempt, so a source
// drifting toward broken looks exactly like a healthy one from outside.
// Counting attempts by how they ended shows the drift early: a rising
// "retried" rate is a site failing more often than it used to, long before
// "exhausted" starts costing results.
var attemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "fitter_connector_attempts_total",
	Help: "Connector fetch attempts inside the retry wrapper, by how each ended.",
}, []string{"result"})

// Browser fetches bypass the HTTP client entirely, so the RoundTripper
// histogram never sees the most expensive fetch type fitter has. Timed here
// at the one dispatch point all three engines share. Engine and outcome
// only — no URLs.
var browserSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "fitter_browser_fetch_seconds",
	Help:    "Browser-rendered fetch duration, by engine and outcome.",
	Buckets: []float64{0.5, 1, 2.5, 5, 10, 20, 40, 60, 120},
}, []string{"engine", "outcome"})

func observeBrowser(engine string, start time.Time, body []byte, err error) {
	outcome := "ok"
	if err != nil {
		outcome = "error"
	} else if len(body) == 0 {
		outcome = "empty"
	}
	browserSeconds.WithLabelValues(engine, outcome).Observe(time.Since(start).Seconds())
}
