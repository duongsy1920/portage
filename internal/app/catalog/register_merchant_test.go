package catalogapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/memory"
	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/clock"
)

// The app layer owns the clock (convention 7). Tests pin it.
var now = time.Date(2026, 9, 4, 15, 0, 0, 0, time.UTC)

// world is everything a handler is wired to — all in memory, no Docker.
type world struct {
	clock      *clock.Fixed
	merchants  *memory.MerchantRepo
	categories *memory.CategoryRepo
	products   *memory.ProductRepo
	outbox     *memory.Outbox
	deps       catalogapp.Deps
}

func newWorld() *world {
	w := &world{
		clock:      clock.FixedAt(now),
		merchants:  memory.NewMerchantRepo(),
		categories: memory.NewCategoryRepo(),
		products:   memory.NewProductRepo(),
		outbox:     memory.NewOutbox(),
	}
	w.deps = catalogapp.Deps{
		Clock:      w.clock,
		UoW:        memory.UnitOfWork{},
		Merchants:  w.merchants,
		Categories: w.categories,
		Products:   w.products,
		Outbox:     w.outbox,
	}
	return w
}

func exampleMerchant() catalog.MerchantDetails {
	return catalog.MerchantDetails{
		Name:         "Example Sports",
		Site:         catalog.MustParseHostname("www.example.com"),
		Currency:     shared.USD,
		FreeShipping: catalog.MustFreeShippingOver(shared.MustParseMoney("50.00", shared.USD)),
		Sourcing:     []catalog.SourcingMode{catalog.SourcedByOperator},
	}
}

// The request lifecycle of DDD.md §30, end to end in memory: clock → domain →
// save → pull events → outbox.
func TestRegisterMerchant_savesThenPublishes(t *testing.T) {
	w := newWorld()
	h := catalogapp.NewRegisterMerchantHandler(w.deps)

	id, err := h.Handle(context.Background(), exampleMerchant())
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if id.IsZero() {
		t.Fatal("handler must return the new id")
	}

	saved, err := w.merchants.ByID(context.Background(), id)
	if err != nil || saved.Name() != "Example Sports" {
		t.Fatalf("ByID = %v, %v", saved, err)
	}
	evs := w.outbox.Drain()
	if len(evs) != 1 || evs[0].EventName() != "catalog.merchant_registered" {
		t.Fatalf("outbox = %v, want one merchant_registered", evs)
	}
	if !evs[0].OccurredAt().Equal(now) {
		t.Errorf("event time = %v, want the app clock's %v", evs[0].OccurredAt(), now)
	}
	if rest := saved.PullEvents(); len(rest) != 0 {
		t.Errorf("aggregate still holds %d events after the handler pulled them", len(rest))
	}
}

// A business refusal comes back untouched — the HTTP adapter maps it to a
// status code — and nothing is saved or published.
func TestRegisterMerchant_domainErrorLeavesNoTrace(t *testing.T) {
	w := newWorld()
	h := catalogapp.NewRegisterMerchantHandler(w.deps)

	d := exampleMerchant()
	d.Name = "   "
	if _, err := h.Handle(context.Background(), d); !errors.Is(err, catalog.ErrEmptyName) {
		t.Fatalf("got %v, want ErrEmptyName", err)
	}
	if n := w.merchants.Len(); n != 0 {
		t.Errorf("%d merchants saved after a refused command", n)
	}
	if evs := w.outbox.Drain(); len(evs) != 0 {
		t.Errorf("%d events published after a refused command", len(evs))
	}
}

// Save BEFORE pull is not a comment, it is a test: if the save fails, no event
// leaves the aggregate, so no worker goes shopping for a merchant that does
// not exist.
func TestRegisterMerchant_saveFailurePublishesNothing(t *testing.T) {
	w := newWorld()
	boom := errors.New("disk on fire")
	w.merchants.FailSaveWith = boom
	h := catalogapp.NewRegisterMerchantHandler(w.deps)

	if _, err := h.Handle(context.Background(), exampleMerchant()); !errors.Is(err, boom) {
		t.Fatalf("got %v, want the storage error wrapped", err)
	}
	if evs := w.outbox.Drain(); len(evs) != 0 {
		t.Fatalf("%d events reached the outbox although Save failed", len(evs))
	}
}

// Wiring is programmer input (convention 1): a missing dependency is a bug at
// start-up, not an error at request time.
func TestNewRegisterMerchantHandler_panicsOnMissingDeps(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic for nil Merchants")
		}
	}()
	d := newWorld().deps
	d.Merchants = nil
	catalogapp.NewRegisterMerchantHandler(d)
}

// ATOMICITY, the other direction: Save succeeds, then Outbox.Append fails.
// The transaction must roll the save back — a merchant that exists with no
// event announcing it is as bad as an event for a merchant that does not
// exist. Both halves of "same transaction" (DDD.md §27) need a test.
//
// memory.UnitOfWork has no transaction, so this cannot pass here — reproduced
// 2026-09-04 by the Windows review: outbox fails, merchants.Len() == 1. It is
// SKIPPED, not deleted: it is the first test the Postgres UnitOfWork must pass.
func TestRegisterMerchant_outboxFailureRollsBackTheSave(t *testing.T) {
	w := newWorld()
	if _, inMemory := w.deps.UoW.(memory.UnitOfWork); inMemory {
		t.Skip("memory.UnitOfWork has no transaction: a Save cannot be rolled back (KNOWN GAP, adapter/memory/memory.go)")
	}
	w.outbox.FailWith = errors.New("outbox table locked")
	h := catalogapp.NewRegisterMerchantHandler(w.deps)

	if _, err := h.Handle(context.Background(), exampleMerchant()); err == nil {
		t.Fatal("Handle must fail when the outbox does")
	}
	if n := w.merchants.Len(); n != 0 {
		t.Fatalf("Save was not rolled back: %d merchant(s) remain with no event announcing them", n)
	}
}
