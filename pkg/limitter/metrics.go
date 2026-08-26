package limitter

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sync/semaphore"
)

// When everything is slow, the first question is whether the site is slow or
// the work is queueing behind one of these semaphores. Timing the acquire
// answers it. The host label is bounded: HostLimiter only exists for hosts an
// operator listed in the limits config, and instance semaphores leave it
// empty. No URLs — hosts and resource names only.
var waitSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "fitter_limiter_wait_seconds",
	Help:    "Time spent waiting to acquire a concurrency limit.",
	Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 15, 60},
}, []string{"resource", "host"})

// TimedAcquire is Acquire with the wait recorded. Use it at every limiter
// call site so saturation is visible instead of just felt.
func TimedAcquire(ctx context.Context, sem *semaphore.Weighted, resource, host string) error {
	start := time.Now()
	err := sem.Acquire(ctx, 1)
	waitSeconds.WithLabelValues(resource, host).Observe(time.Since(start).Seconds())
	return err
}
