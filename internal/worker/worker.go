// Package worker relays the outbox: it reads rows nobody has published yet,
// publishes them in order, and marks them sent — the fifth layer of DDD.md
// §30, and the second half of the outbox pattern (§27).
//
// Delivery is AT-LEAST-ONCE. A row is marked only after it was published, so
// a crash between the two re-publishes it on the next pass. Nothing is ever
// lost; something may be seen twice; every subscriber must therefore be
// idempotent. That trade is deliberate: the alternative — mark first, publish
// second — loses events, and a lost DepositPaid is an order nobody buys.
//
// The relay knows nothing about Postgres or memory: it talks to a Source and a
// Publisher (ports declared here, where they are needed) inside a UnitOfWork.
// On Postgres the pass runs in one transaction, so Pending's FOR UPDATE SKIP
// LOCKED lets several workers share the table without ever taking the same row.
//
// [PHP] Đây là `messenger:consume` — nhưng đọc từ bảng outbox thay vì queue,
// [PHP] và cố tình ~120 dòng để thấy hết vòng đời một message.
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/duongsy/portage/internal/app"
)

// Entry is one outbox row as the relay and its subscribers see it: the stable
// event name, when it happened, and the payload — the contract bytes written
// by eventcodec, never a Go struct.
type Entry struct {
	ID         int64
	Name       string
	OccurredAt time.Time
	Payload    json.RawMessage
}

// Source is the outbox, read side. Implemented by postgres.Outbox and
// memory.Outbox.
type Source interface {
	// Pending returns up to limit unsent entries, oldest first.
	Pending(ctx context.Context, limit int) ([]Entry, error)
	// MarkSent stamps entries the relay has published.
	MarkSent(ctx context.Context, ids []int64, at time.Time) error
}

// Publisher hands one entry to whoever listens. Implemented by Bus.
type Publisher interface {
	Publish(ctx context.Context, e Entry) error
}

// Deps wires a Relay (convention 10). Batch and Interval have defaults.
type Deps struct {
	Source    Source
	Publisher Publisher
	UoW       app.UnitOfWork
	Clock     app.Clock
	Batch     int           // rows per pass; default 100
	Interval  time.Duration // pause between passes; default 1s
	Log       *log.Logger   // default log.Default()
}

// Relay is the outbox relay. Build it with New.
type Relay struct {
	deps Deps
}

func New(d Deps) *Relay {
	for name, v := range map[string]any{"Source": d.Source, "Publisher": d.Publisher, "UoW": d.UoW, "Clock": d.Clock} {
		if v == nil {
			panic("worker: Relay wired without " + name)
		}
	}
	if d.Batch <= 0 {
		d.Batch = 100
	}
	if d.Interval <= 0 {
		d.Interval = time.Second
	}
	if d.Log == nil {
		d.Log = log.Default()
	}
	return &Relay{deps: d}
}

// RunOnce is one pass, in one transaction: read pending, publish in order,
// mark what was published, commit.
//
// A publish failure STOPS the pass but does not undo it: the entries already
// published are marked (the transaction still commits), the failed one and
// everything after stay pending for the next pass, and the error is returned
// so the loop can log it. Order within the outbox is preserved because the
// pass never skips a failed entry.
func (r *Relay) RunOnce(ctx context.Context) (sent int, err error) {
	var pubErr error
	txErr := r.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		entries, err := r.deps.Source.Pending(ctx, r.deps.Batch)
		if err != nil {
			return err
		}
		var done []int64
		for _, e := range entries {
			if err := r.deps.Publisher.Publish(ctx, e); err != nil {
				pubErr = fmt.Errorf("publish %s #%d: %w", e.Name, e.ID, err)
				break
			}
			done = append(done, e.ID)
		}
		if err := r.deps.Source.MarkSent(ctx, done, r.deps.Clock.Now()); err != nil {
			return err
		}
		sent = len(done)
		return nil
	})
	if txErr != nil {
		return 0, txErr
	}
	return sent, pubErr
}

// Run repeats RunOnce every Interval until ctx is cancelled, then returns
// ctx.Err(). Errors from a pass are logged, never fatal: the outbox keeps the
// rows, the next pass tries again.
func (r *Relay) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.deps.Interval)
	defer ticker.Stop()
	for {
		if _, err := r.RunOnce(ctx); err != nil && ctx.Err() == nil {
			r.deps.Log.Printf("worker: %v", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// Handler is what a subscriber gives the Bus.
type Handler func(ctx context.Context, e Entry) error

// Bus is an in-process Publisher: handlers subscribe by event name, "*" hears
// everything. A handler's error is the publish's error, so the relay holds
// the row back and retries — the reason handlers must be idempotent.
//
// Today the only subscriber is a log line in cmd/worker. The first real one
// is pricing listening for catalog.product_published.
//
// [PHP] EventDispatcher với listener đăng ký theo tên event; "*" ~ listener
// [PHP] cho KernelEvents::* — khác ở chỗ lỗi listener chặn cả pass.
type Bus struct {
	mu   sync.RWMutex
	subs map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{subs: map[string][]Handler{}}
}

func (b *Bus) Subscribe(name string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[name] = append(b.subs[name], h)
}

func (b *Bus) Publish(ctx context.Context, e Entry) error {
	b.mu.RLock()
	handlers := append(append([]Handler(nil), b.subs[e.Name]...), b.subs["*"]...)
	b.mu.RUnlock()
	for _, h := range handlers {
		if err := h(ctx, e); err != nil {
			return err
		}
	}
	return nil
}
