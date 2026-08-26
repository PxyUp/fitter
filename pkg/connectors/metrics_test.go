package connectors

import (
	"context"
	"errors"
	"testing"

	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type scriptedConnector struct {
	responses []func() ([]byte, error)
	call      int
}

func (s *scriptedConnector) Get(_ context.Context, _ builder.Interfacable, _ *uint32, _ builder.Interfacable) ([]byte, error) {
	fn := s.responses[s.call]
	s.call++
	return fn()
}

// A fetch that needs three tries is a site drifting toward broken, and the
// counter is how that drift is seen before it costs results: two "retried",
// one "ok".
func TestAttemptsAreCountedByHowTheyEnded(t *testing.T) {
	retriedBefore := testutil.ToFloat64(attemptsTotal.WithLabelValues("retried"))
	okBefore := testutil.ToFloat64(attemptsTotal.WithLabelValues("ok"))
	exhaustedBefore := testutil.ToFloat64(attemptsTotal.WithLabelValues("exhausted"))

	flaky := &scriptedConnector{responses: []func() ([]byte, error){
		func() ([]byte, error) { return nil, errors.New("down") },
		func() ([]byte, error) { return nil, nil }, // empty body also retries
		func() ([]byte, error) { return []byte("ok"), nil },
	}}
	if _, err := WithAttempts(flaky, 5).Get(context.Background(), nil, nil, nil); err != nil {
		t.Fatal(err)
	}

	dead := &scriptedConnector{responses: []func() ([]byte, error){
		func() ([]byte, error) { return nil, errors.New("down") },
		func() ([]byte, error) { return nil, errors.New("down") },
	}}
	if _, err := WithAttempts(dead, 2).Get(context.Background(), nil, nil, nil); err == nil {
		t.Fatal("an exhausted connector reported success")
	}

	if got := testutil.ToFloat64(attemptsTotal.WithLabelValues("retried")) - retriedBefore; got != 4 {
		t.Errorf("retried = %v, want 4 (two flaky, two dead)", got)
	}
	if got := testutil.ToFloat64(attemptsTotal.WithLabelValues("ok")) - okBefore; got != 1 {
		t.Errorf("ok = %v, want 1", got)
	}
	if got := testutil.ToFloat64(attemptsTotal.WithLabelValues("exhausted")) - exhaustedBefore; got != 1 {
		t.Errorf("exhausted = %v, want 1", got)
	}
}
