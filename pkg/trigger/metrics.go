package trigger

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// The value of this counter is mostly its absence: a scheduler that should
// fire hourly and has not moved for two hours is a wedged cron nobody would
// otherwise notice until a day of scrapes was gone. absent()/rate() on this
// is the alert.
var firesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "fitter_trigger_fires_total",
	Help: "Trigger firings, by trigger name and kind.",
}, []string{"name", "kind"})
