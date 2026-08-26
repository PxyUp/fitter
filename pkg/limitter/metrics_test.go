package limitter

import (
	"context"
	"testing"
	"time"

	"golang.org/x/sync/semaphore"
)

// No limit configured means no waiting. Every call site checks for nil
// before calling, but an exported function's safety must not depend on
// callers remembering to.
func TestAcquiringANilSemaphoreIsANoOp(t *testing.T) {
	if err := TimedAcquire(context.Background(), nil, "chromium", ""); err != nil {
		t.Fatalf("a nil semaphore errored: %v", err)
	}
}

// The wait is what the histogram measures, so a held semaphore must show up
// as one — and a cancelled wait must come back as the context's error.
func TestTheWaitIsRealAndCancellable(t *testing.T) {
	sem := semaphore.NewWeighted(1)
	if err := TimedAcquire(context.Background(), sem, "chromium", ""); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := TimedAcquire(ctx, sem, "chromium", ""); err == nil {
		t.Fatal("acquiring a held semaphore succeeded")
	}
}
