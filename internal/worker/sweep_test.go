package worker_test

import (
	"context"
	"errors"
	"io"
	"log"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/worker"
)

// A Sweeper runs its Pass until the context ends, and a failing pass is a log
// line, not the end of the process: the binary's job is to relay the outbox,
// and a sweep that cannot reach the database this second must not take that
// down with it.
func TestSweeper_keepsRunningAfterAFailedPass(t *testing.T) {
	var calls atomic.Int64
	boom := errors.New("database is having a moment")
	pass := func(context.Context) (int, error) {
		if calls.Add(1) == 1 {
			return 0, boom
		}
		return 1, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- worker.NewSweeper("test", pass, time.Millisecond, discard()).Run(ctx) }()

	deadline := time.Now().Add(2 * time.Second)
	for calls.Load() < 3 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run returned %v, want context.Canceled", err)
	}
	if n := calls.Load(); n < 3 {
		t.Fatalf("the sweep ran %d times; a failed pass must not stop the ticker", n)
	}
}

// Unlike Relay.Run, a Sweeper waits one interval before its first pass: it has
// no backlog waiting for it at start-up, only a clock.
func TestSweeper_doesNotSweepBeforeTheFirstTick(t *testing.T) {
	var calls atomic.Int64
	pass := func(context.Context) (int, error) { calls.Add(1); return 0, nil }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = worker.NewSweeper("test", pass, time.Hour, discard()).Run(ctx) }()

	time.Sleep(20 * time.Millisecond)
	if n := calls.Load(); n != 0 {
		t.Fatalf("swept %d times before the first tick", n)
	}
}

func TestNewSweeper_panicsWithoutAPass(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic for a nil Pass")
		}
	}()
	worker.NewSweeper("test", nil, time.Second, discard())
}

// discard keeps the test output readable: the Sweeper's logging is behaviour
// we want, just not on the terminal during a test run.
func discard() *log.Logger {
	return log.New(io.Discard, "", 0)
}
