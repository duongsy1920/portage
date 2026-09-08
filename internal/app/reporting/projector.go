package reportingapp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Projector builds order_summaries from the events of FIVE contexts. It is the
// largest consumer in the system and the only one that listens to all of them —
// exactly the shape CQRS predicts: the write side is split into contexts that
// each guard their own invariants, and the read side is joined back together
// for a screen.
//
// Two properties every handler here has, and why:
//
//	IDEMPOTENT       the relay is at-least-once; the same row may arrive twice
//	ORDER-TOLERANT   nothing guarantees order_placed lands before batch_shipped
//
// Order-tolerance is bought with one trick — the FRAME. Any event may create
// the row: a batch_shipped for an order the read model has never heard of
// writes a skeleton with the tracking filled in, and order_placed fills the
// rest in later. The alternative ("ignore events for unknown orders") silently
// loses data the first day a relay retries something out of order.
type Projector struct {
	deps Deps
}

func NewProjector(d Deps) *Projector {
	mustHave("Projector", map[string]any{"UoW": d.UoW, "Summaries": d.Summaries, "Names": d.Names, "Worklist": d.Worklist})
	return &Projector{deps: d}
}

// ── catalog ──────────────────────────────────────────────────────────────────

// OnProductPublished remembers the product's name, and backfills any summary
// written before the name was known.
func (p *Projector) OnProductPublished(ctx context.Context, m contracts.ProductPublishedV1) error {
	product, err := shared.ParseID(m.ID)
	if err != nil {
		return fmt.Errorf("product_published: %w", err)
	}
	// Two read models move on this one event: the name every order summary
	// shows, and the worklist row that can stop asking for anything.
	if err := p.worklist(ctx, product, m.At, func(w *WorklistItem) {
		w.Published = true
		if w.Name == "" {
			w.Name = m.Name // published before added was relayed: still order-tolerant
		}
	}); err != nil {
		return err
	}
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		if err := p.deps.Names.Save(ctx, product, m.Name); err != nil {
			return fmt.Errorf("product name %s: %w", m.ID, err)
		}
		rows, err := p.deps.Summaries.ByProduct(ctx, product)
		if err != nil {
			return err
		}
		for _, s := range rows {
			if s.ProductName == m.Name {
				continue // already right: the second delivery changes nothing
			}
			s.ProductName, s.UpdatedAt = m.Name, m.At
			if err := p.deps.Summaries.Save(ctx, s); err != nil {
				return err
			}
		}
		return nil
	})
}

// ── ordering ─────────────────────────────────────────────────────────────────

func (p *Projector) OnOrderPlaced(ctx context.Context, m contracts.OrderPlacedV1) error {
	id, err := ordering.ParseOrderID(m.ID)
	if err != nil {
		return fmt.Errorf("order_placed: %w", err)
	}
	fail := func(err error) error { return fmt.Errorf("order_placed %s: %w", m.ID, err) }
	customer, err := shared.ParseID(m.Customer)
	if err != nil {
		return fail(err)
	}
	product, err := shared.ParseID(m.Product)
	if err != nil {
		return fail(err)
	}
	variant, err := shared.ParseID(m.Variant)
	if err != nil {
		return fail(err)
	}
	quote, err := shared.ParseID(m.Quote)
	if err != nil {
		return fail(err)
	}
	total, err := moneyFrom(m.Total)
	if err != nil {
		return fail(err)
	}
	deposit, err := moneyFrom(m.Deposit)
	if err != nil {
		return fail(err)
	}

	return p.update(ctx, id, m.At, func(ctx context.Context, s *OrderSummary) error {
		s.Customer, s.Product, s.Variant, s.Quote = customer, product, variant, quote
		s.Total, s.Deposit = total, deposit
		s.PlacedAt = m.At
		// Only order_placed may set the FIRST status. A frame written by a
		// later event has none yet, and guessing one would be a lie the screen
		// then shows to a customer.
		if s.Status == "" {
			s.Status = ordering.StatusAwaitingDeposit
		}
		if s.ProductName == "" {
			name, ok, err := p.deps.Names.Name(ctx, product)
			if err != nil {
				return err
			}
			if ok {
				s.ProductName = name
			}
		}
		return nil
	})
}

func (p *Projector) OnDepositPaid(ctx context.Context, m contracts.DepositPaidV1) error {
	return p.set(ctx, m.ID, m.At, "deposit_paid", func(s *OrderSummary) {
		s.DepositPaid = true
		s.Status = ordering.StatusDeposited
	})
}

func (p *Projector) OnOrderPurchased(ctx context.Context, m contracts.OrderPurchasedV1) error {
	return p.set(ctx, m.ID, m.At, "order_purchased", func(s *OrderSummary) {
		s.Status = ordering.StatusPurchased
	})
}

func (p *Projector) OnOrderShipped(ctx context.Context, m contracts.OrderShippedV1) error {
	return p.set(ctx, m.ID, m.At, "order_shipped", func(s *OrderSummary) {
		s.Status = ordering.StatusInTransit
	})
}

func (p *Projector) OnBalancePaid(ctx context.Context, m contracts.BalancePaidV1) error {
	return p.set(ctx, m.ID, m.At, "balance_paid", func(s *OrderSummary) {
		s.BalancePaid = true
	})
}

func (p *Projector) OnOrderDelivered(ctx context.Context, m contracts.OrderDeliveredV1) error {
	return p.set(ctx, m.ID, m.At, "order_delivered", func(s *OrderSummary) {
		s.Status = ordering.StatusDelivered
		s.DeliveredAt = m.At
	})
}

func (p *Projector) OnOrderCancelled(ctx context.Context, m contracts.OrderCancelledV1) error {
	refund, err := moneyFrom(m.Refund)
	if err != nil {
		return fmt.Errorf("order_cancelled %s: %w", m.ID, err)
	}
	return p.set(ctx, m.ID, m.At, "order_cancelled", func(s *OrderSummary) {
		s.Status = ordering.StatusCancelled
		s.Refund, s.Forfeited, s.CancelledAt = refund, m.Forfeited, m.At
	})
}

// ── procurement ──────────────────────────────────────────────────────────────

// OnPurchaseConfirmed copies the shop's own order number — the thing a
// customer asking "did you actually buy it?" wants to see.
func (p *Projector) OnPurchaseConfirmed(ctx context.Context, m contracts.PurchaseConfirmedV1) error {
	return p.set(ctx, m.Order, m.At, "purchase_confirmed", func(s *OrderSummary) {
		s.ShopReference = m.Reference
	})
}

// ── logistics ────────────────────────────────────────────────────────────────

// Tracking is a SECOND axis, deliberately not folded into Status: ordering owns
// the status and moves it on money and on purchase, while the warehouse's own
// steps (waiting → weighed → flown) belong to the box, not to the order.
func (p *Projector) OnParcelExpected(ctx context.Context, m contracts.ParcelExpectedV1) error {
	return p.set(ctx, m.Order, m.At, "parcel_expected", func(s *OrderSummary) {
		s.Tracking = TrackingExpected
	})
}

func (p *Projector) OnParcelReceived(ctx context.Context, m contracts.ParcelReceivedV1) error {
	return p.set(ctx, m.Order, m.At, "parcel_received", func(s *OrderSummary) {
		s.Tracking = TrackingReceived
	})
}

// OnBatchShipped touches every order in the batch: one event, many rows.
func (p *Projector) OnBatchShipped(ctx context.Context, m contracts.BatchShippedV1) error {
	for _, a := range m.Allocations {
		if err := p.set(ctx, a.Order, m.At, "batch_shipped", func(s *OrderSummary) {
			s.Tracking = TrackingShipped
		}); err != nil {
			return err
		}
	}
	return nil
}

// ── the two shapes every handler above is built from ─────────────────────────

// set is the common case: parse the order id, load-or-frame the row, change a
// field, save. apply cannot fail — a handler that must parse something does it
// BEFORE calling here, so a bad payload is refused before any row is touched.
func (p *Projector) set(ctx context.Context, rawID string, at time.Time, event string, apply func(*OrderSummary)) error {
	id, err := ordering.ParseOrderID(rawID)
	if err != nil {
		return fmt.Errorf("%s: %w", event, err)
	}
	return p.update(ctx, id, at, func(_ context.Context, s *OrderSummary) error {
		apply(s)
		return nil
	})
}

// update is where idempotence and order-tolerance actually live: one
// transaction, load or frame, apply, save.
func (p *Projector) update(ctx context.Context, id ordering.OrderID, at time.Time, apply func(context.Context, *OrderSummary) error) error {
	return p.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		s, err := p.deps.Summaries.ByOrder(ctx, id)
		if errors.Is(err, ErrSummaryNotFound) {
			s = OrderSummary{Order: id, Tracking: TrackingNone} // the FRAME
		} else if err != nil {
			return err
		}
		if err := apply(ctx, &s); err != nil {
			return err
		}
		// The freshest event wins the timestamp; a retried old one must not
		// drag it backwards.
		if at.After(s.UpdatedAt) {
			s.UpdatedAt = at
		}
		return p.deps.Summaries.Save(ctx, s)
	})
}

func moneyFrom(m contracts.MoneyV1) (shared.Money, error) {
	cur, err := shared.CurrencyFromCode(m.Currency)
	if err != nil {
		return shared.Money{}, err
	}
	return shared.NewMoney(m.Minor, cur), nil
}
