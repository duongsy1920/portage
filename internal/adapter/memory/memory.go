// Package memory implements the catalog repositories and the app ports with
// maps and slices. It is a real adapter — main() can wire it for a dev run —
// and the one every handler test uses: no Docker, no network, milliseconds.
//
// TWO THINGS THIS ADAPTER DOES NOT DO, on purpose — postgres does both:
//
//  1. It keeps aggregates by POINTER instead of going through
//     Snapshot()/FromSnapshot(). Postgres does the round trip for real and
//     tests it; here a copy would only slow tests down.
//
//  2. UnitOfWork has NO transaction, so nothing rolls back. "Save, then Append
//     to the outbox, in one transaction" is only half protected here: a failing
//     Save publishes nothing (tested), but a failing Append after a successful
//     Save leaves the aggregate stored. TestRegisterMerchant_outboxFailure-
//     RollsBackTheSave is SKIPPED against this adapter and RUNS against
//     postgres (TestUnitOfWork_rollsBackTheSaveWhenTheOutboxFails).
//
// [PHP] Tương đương một "InMemoryRepository" hay dùng trong test Symfony thay
// [PHP] Doctrine — nhưng ở đây nó nằm trong adapter/ như một cài đặt bình
// [PHP] đẳng, không phải trong tests/.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/duongsy/portage/internal/adapter/eventcodec"
	"github.com/duongsy/portage/internal/app"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/worker"
)

// UnitOfWork has no transaction to open: memory either works or panics.
// It exists so a handler is wired the same way in a test and in production.
type UnitOfWork struct{}

func (UnitOfWork) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// Outbox is the in-memory outbox: app.Outbox on the write side (Append) and
// worker.Source on the read side (Pending, MarkSent), so the relay can be
// tested without a database. Entries are encoded with the same codec Postgres
// uses — a payload bug shows up here first.
//
// Drain is for tests that want the TYPED events back (evs[0].(catalog.X)); it
// is independent of Pending/MarkSent, which see the encoded entries.
type Outbox struct {
	mu       sync.Mutex
	events   []shared.Event // typed, for Drain
	entries  []outboxRow    // encoded, for the relay
	next     int64
	FailWith error // set in a test to make Append fail
}

type outboxRow struct {
	entry worker.Entry
	sent  bool
}

var (
	_ app.Outbox    = (*Outbox)(nil)
	_ worker.Source = (*Outbox)(nil)
)

func NewOutbox() *Outbox {
	return &Outbox{}
}

func (o *Outbox) Append(ctx context.Context, events []shared.Event) error {
	if o.FailWith != nil {
		return o.FailWith
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, ev := range events {
		env, err := eventcodec.Encode(ev)
		if err != nil {
			return err
		}
		o.next++
		o.entries = append(o.entries, outboxRow{entry: worker.Entry{
			ID: o.next, Name: env.Name, OccurredAt: env.OccurredAt, Payload: env.Payload,
		}})
		o.events = append(o.events, ev)
	}
	return nil
}

// Drain hands out the typed events appended since the last Drain, and empties
// that list. It does not touch the relay's entries.
func (o *Outbox) Drain() []shared.Event {
	o.mu.Lock()
	defer o.mu.Unlock()
	out := o.events
	o.events = nil
	return out
}

func (o *Outbox) Pending(ctx context.Context, limit int) ([]worker.Entry, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	var out []worker.Entry
	for _, row := range o.entries {
		if len(out) == limit {
			break
		}
		if !row.sent {
			out = append(out, row.entry)
		}
	}
	return out, nil
}

func (o *Outbox) MarkSent(ctx context.Context, ids []int64, at time.Time) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	for i := range o.entries {
		for _, id := range ids {
			if o.entries[i].entry.ID == id {
				o.entries[i].sent = true
			}
		}
	}
	return nil
}
