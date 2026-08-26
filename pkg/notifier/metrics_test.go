package notifier

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// The dispatch must never see a notifier it cannot name — a new notifier
// landing without a case here would silently count as "custom", which is
// reserved for plugin notifiers the repo cannot know about.
func TestEveryBuiltInNotifierHasAKind(t *testing.T) {
	for want, n := range map[string]Notifier{
		"console":  &console{},
		"telegram": &telegramBot{},
		"http":     &httpNotifier{},
		"redis":    &redisNotifier{},
		"file":     &fileNotifier{},
	} {
		if got := kindOf(n); got != want {
			t.Errorf("kindOf(%T) = %q, want %q", n, got, want)
		}
	}
}

func TestDeliveriesAreCountedByOutcome(t *testing.T) {
	okBefore := testutil.ToFloat64(notifyTotal.WithLabelValues("console", "ok"))
	errBefore := testutil.ToFloat64(notifyTotal.WithLabelValues("console", "error"))

	countNotify(&console{}, nil)
	countNotify(&console{}, errors.New("the sink is gone"))

	if got := testutil.ToFloat64(notifyTotal.WithLabelValues("console", "ok")) - okBefore; got != 1 {
		t.Errorf("ok = %v, want 1", got)
	}
	if got := testutil.ToFloat64(notifyTotal.WithLabelValues("console", "error")) - errBefore; got != 1 {
		t.Errorf("error = %v, want 1", got)
	}
}
