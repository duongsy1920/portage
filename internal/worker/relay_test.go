package worker_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/memory"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/clock"
	"github.com/duongsy/portage/internal/worker"
)

var now = time.Date(2026, 9, 4, 20, 0, 0, 0, time.UTC)

// recorder is a Publisher that remembers what it was given and can be told to
// fail on the Nth call — the two things a relay test needs.
type recorder struct {
	mu     sync.Mutex
	got    []worker.Entry
	failOn int // 1-based call number that fails; 0 = never
	calls  int
}

func (r *recorder) Publish(ctx context.Context, e worker.Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	if r.failOn != 0 && r.calls == r.failOn {
		return errors.New("subscriber down")
	}
	r.got = append(r.got, e)
	return nil
}

func (r *recorder) names() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, 0, len(r.got))
	for _, e := range r.got {
		out = append(out, e.Name)
	}
	return out
}

func rig(t *testing.T, pub worker.Publisher, batch int) (*memory.Outbox, *worker.Relay) {
	t.Helper()
	outbox := memory.NewOutbox()
	relay := worker.New(worker.Deps{
		Source: outbox, Publisher: pub, UoW: memory.UnitOfWork{}, Clock: clock.FixedAt(now),
		Batch: batch, Interval: 5 * time.Millisecond,
	})
	return outbox, relay
}

func appendEvents(t *testing.T, outbox *memory.Outbox, n int) {
	t.Helper()
	m, err := catalog.RegisterMerchant(catalog.MerchantDetails{
		Name: "Example Sports", Site: catalog.MustParseHostname("www.example.com"), Currency: shared.USD,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	m.PullEvents()
	var evs []shared.Event
	for i := 0; i < n; i++ {
		if err := m.Rename("Name "+string(rune('A'+i)), now); err != nil {
			t.Fatal(err)
		}
		evs = append(evs, m.PullEvents()...)
	}
	if err := outbox.Append(context.Background(), evs); err != nil {
		t.Fatal(err)
	}
}

// One pass: everything pending goes out in order and is marked, so the next
// pass finds nothing.
func TestRelay_runOncePublishesInOrderAndMarksSent(t *testing.T) {
	pub := &recorder{}
	outbox, relay := rig(t, pub, 100)
	appendEvents(t, outbox, 3)

	sent, err := relay.RunOnce(context.Background())
	if err != nil || sent != 3 {
		t.Fatalf("RunOnce = %d, %v", sent, err)
	}
	if got := pub.got; len(got) != 3 || got[0].ID >= got[1].ID || got[1].ID >= got[2].ID {
		t.Fatalf("published out of order or wrong count: %+v", got)
	}
	if sent, _ := relay.RunOnce(context.Background()); sent != 0 {
		t.Fatalf("second pass sent %d, want 0", sent)
	}
	if left, _ := outbox.Pending(context.Background(), 10); len(left) != 0 {
		t.Fatalf("%d entries still pending", len(left))
	}
}

// AT-LEAST-ONCE, the promise of DDD.md §27: a failing subscriber stops the
// pass, what was published stays marked, what was not stays pending — and
// the next pass picks up exactly there. Nothing is lost; something may be
// seen twice, which is why subscribers must be idempotent.
func TestRelay_failureKeepsTheRestPendingAndRetries(t *testing.T) {
	pub := &recorder{failOn: 2}
	outbox, relay := rig(t, pub, 100)
	appendEvents(t, outbox, 3)

	sent, err := relay.RunOnce(context.Background())
	if err == nil || sent != 1 {
		t.Fatalf("RunOnce = %d, %v; want 1 sent and an error", sent, err)
	}
	if left, _ := outbox.Pending(context.Background(), 10); len(left) != 2 {
		t.Fatalf("%d pending, want the 2 unpublished ones", len(left))
	}

	pub.failOn = 0 // subscriber recovers
	sent, err = relay.RunOnce(context.Background())
	if err != nil || sent != 2 {
		t.Fatalf("retry RunOnce = %d, %v", sent, err)
	}
	if names := pub.names(); len(names) != 3 {
		t.Fatalf("published %v, want all three exactly once here", names)
	}
}

func TestRelay_batchLimit(t *testing.T) {
	pub := &recorder{}
	outbox, relay := rig(t, pub, 2)
	appendEvents(t, outbox, 5)
	for _, want := range []int{2, 2, 1, 0} {
		if sent, err := relay.RunOnce(context.Background()); err != nil || sent != want {
			t.Fatalf("RunOnce = %d, %v; want %d", sent, err, want)
		}
	}
}

// Run polls until the context is cancelled; an event appended after start is
// published within a few ticks.
func TestRelay_runPollsUntilCancelled(t *testing.T) {
	pub := &recorder{}
	outbox, relay := rig(t, pub, 100)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- relay.Run(ctx) }()

	appendEvents(t, outbox, 1)
	deadline := time.Now().Add(2 * time.Second)
	for len(pub.names()) == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run returned %v, want context.Canceled", err)
	}
	if len(pub.names()) != 1 {
		t.Fatalf("published %v, want one", pub.names())
	}
}

func TestNew_panicsOnMissingDeps(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic for a nil Publisher")
		}
	}()
	worker.New(worker.Deps{Source: memory.NewOutbox(), UoW: memory.UnitOfWork{}, Clock: clock.FixedAt(now)})
}

// The Bus routes by event name, "*" hears everything, and a handler's error
// is the publish's error — the relay must see it to hold the row back.
func TestBus_routesByNameAndWildcard(t *testing.T) {
	bus := worker.NewBus()
	var published, all []string
	bus.Subscribe("catalog.merchant_renamed", func(ctx context.Context, e worker.Entry) error {
		published = append(published, e.Name)
		return nil
	})
	bus.Subscribe("*", func(ctx context.Context, e worker.Entry) error {
		all = append(all, e.Name)
		return nil
	})
	if err := bus.Publish(context.Background(), worker.Entry{ID: 1, Name: "catalog.merchant_renamed"}); err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(context.Background(), worker.Entry{ID: 2, Name: "catalog.product_added"}); err != nil {
		t.Fatal(err)
	}
	if len(published) != 1 || len(all) != 2 {
		t.Fatalf("routing: named=%v all=%v", published, all)
	}

	boom := errors.New("handler exploded")
	bus.Subscribe("catalog.product_added", func(context.Context, worker.Entry) error { return boom })
	if err := bus.Publish(context.Background(), worker.Entry{ID: 3, Name: "catalog.product_added"}); !errors.Is(err, boom) {
		t.Fatalf("got %v, want the handler's error", err)
	}
}
