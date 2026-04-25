package business

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestTriggerGuard(t *testing.T) {
	var tg triggerGuard

	if !tg.tryStart() {
		t.Fatal("expected tryStart to return true on first call")
	}
	if tg.tryStart() {
		t.Fatal("expected tryStart to return false when already running")
	}

	tg.markDone()
	if !tg.tryStart() {
		t.Fatal("expected tryStart to return true after markDone")
	}
}

func TestConcurrentWorkerSuccess(t *testing.T) {
	cw := newConcurrentWorker(10)
	var count atomic.Int32

	handler := func(ctx context.Context, item string) error {
		count.Add(1)
		return nil
	}

	errs := cw.run(context.Background(), []string{"a", "b", "c"}, handler)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errs))
	}
	if count.Load() != 3 {
		t.Fatalf("expected 3 items processed, got %d", count.Load())
	}
}

func TestConcurrentWorkerError(t *testing.T) {
	cw := newConcurrentWorker(10)
	testErr := errors.New("handler error")

	handler := func(ctx context.Context, item string) error {
		if item == "b" {
			return testErr
		}
		return nil
	}

	errs := cw.run(context.Background(), []string{"a", "b", "c"}, handler)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestConcurrentWorkerContextCancellation(t *testing.T) {
	cw := newConcurrentWorker(1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	handler := func(ctx context.Context, item string) error {
		return nil
	}

	errs := cw.run(ctx, []string{"a"}, handler)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error (context cancelled), got %d", len(errs))
	}
}

func TestConcurrentWorkerConcurrencyLimit(t *testing.T) {
	cw := newConcurrentWorker(2)
	var running atomic.Int32
	var maxRunning atomic.Int32

	handler := func(ctx context.Context, item string) error {
		curr := running.Add(1)
		for {
			m := maxRunning.Load()
			if curr > m && maxRunning.CompareAndSwap(m, curr) {
				break
			}
			if curr <= m {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		running.Add(-1)
		return nil
	}

	items := make([]string, 10)
	for i := range items {
		items[i] = "item"
	}
	cw.run(context.Background(), items, handler)

	if maxRunning.Load() > 2 {
		t.Fatalf("expected max concurrent <= 2, got %d", maxRunning.Load())
	}
}
