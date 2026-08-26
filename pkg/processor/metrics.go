package processor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Every config run funnels through Process, and the processor already knows
// the config's name — a label bounded by whatever an operator actually runs.
// This is the number that answers "which of my configs is slow or failing",
// which per-host metrics cannot: one config may touch many hosts and one
// host may serve many configs.
var processSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "fitter_process_seconds",
	Help:    "Config run duration through the processor, by config name and outcome.",
	Buckets: []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60, 120, 300},
}, []string{"name", "outcome"})
