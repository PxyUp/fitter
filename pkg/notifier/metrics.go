package notifier

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// A notifier failing is the worst silent failure fitter has: the scrape
// succeeded and the result simply never reached anyone. Counted at the one
// dispatch point every notifier shares, by kind rather than destination —
// no chat ids, no URLs.
var notifyTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "fitter_notify_total",
	Help: "Notification deliveries, by notifier kind and outcome.",
}, []string{"kind", "outcome"})

func kindOf(n Notifier) string {
	switch n.(type) {
	case *console:
		return "console"
	case *telegramBot:
		return "telegram"
	case *httpNotifier:
		return "http"
	case *redisNotifier:
		return "redis"
	case *fileNotifier:
		return "file"
	default:
		return "custom"
	}
}

func countNotify(n Notifier, err error) {
	outcome := "ok"
	if err != nil {
		outcome = "error"
	}
	notifyTotal.WithLabelValues(kindOf(n), outcome).Inc()
}
