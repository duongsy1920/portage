package orderingapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/ordering"
)

// Reactor is ordering's ear on procurement: the two outcomes of a purchase
// task move the order — past the point of no return, or into compensation.
// Each method adapts a wire message to the matching handler; wire.Subscribe
// points the events here.
//
// Idempotent by the aggregate's own rules: a second purchase_confirmed finds
// the order already purchased → ErrNotDeposited. That is NOT a retry-worthy
// failure, so it is swallowed here — the only place that knows "already done"
// from "cannot do".
type Reactor struct {
	confirm *ConfirmPurchaseHandler
	fail    *FailPurchaseHandler
	ship    *ShipOrderHandler
}

func NewReactor(d Deps) *Reactor {
	return &Reactor{confirm: NewConfirmPurchaseHandler(d), fail: NewFailPurchaseHandler(d), ship: NewShipOrderHandler(d)}
}

func (r *Reactor) OnPurchaseConfirmed(ctx context.Context, m contracts.PurchaseConfirmedV1) error {
	id, err := ordering.ParseOrderID(m.Order)
	if err != nil {
		return fmt.Errorf("purchase_confirmed: %w", err)
	}
	return alreadyDone(r.confirm.Handle(ctx, id), ordering.ErrNotDeposited)
}

func (r *Reactor) OnPurchaseFailed(ctx context.Context, m contracts.PurchaseFailedV1) error {
	id, err := ordering.ParseOrderID(m.Order)
	if err != nil {
		return fmt.Errorf("purchase_failed: %w", err)
	}
	return alreadyDone(r.fail.Handle(ctx, FailPurchase{Order: id, Reason: m.Reason}), ordering.ErrNotDeposited)
}

// OnBatchShipped: every order in the box is now in transit. An order that was
// cancelled after purchase (deposit forfeited) may still be in the box — its
// goods are ours to resell — so "cancelled" is not a failure here either.
func (r *Reactor) OnBatchShipped(ctx context.Context, m contracts.BatchShippedV1) error {
	for _, a := range m.Allocations {
		id, err := ordering.ParseOrderID(a.Order)
		if err != nil {
			return fmt.Errorf("batch_shipped %s: %w", m.ID, err)
		}
		if err := alreadyDone(r.ship.Handle(ctx, id), ordering.ErrNotPurchased, ordering.ErrOrderCancelled); err != nil {
			return err
		}
	}
	return nil
}

// alreadyDone turns "the order is past that state" into success: the event
// was delivered twice (at-least-once) and the first delivery did the work.
// Each reaction names the sentinels that mean "already done" for it.
func alreadyDone(err error, done ...error) error {
	if err == nil {
		return nil
	}
	for _, d := range done {
		if errors.Is(err, d) {
			return nil
		}
	}
	return err
}
