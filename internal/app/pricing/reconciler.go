package pricingapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Reconciler builds Quote vs Actual (pricing.Reconciliation) from three
// contexts' events. Order-tolerant in the one way that matters: an actual
// arriving before the order is known is kept on a row with no quoted side yet
// — order_placed fills it in when it comes. Idempotent: every write is an
// upsert of a full row rebuilt from the message.
type Reconciler struct {
	deps Deps
}

func NewReconciler(d Deps) *Reconciler {
	mustHave("Reconciler", map[string]any{"UoW": d.UoW, "Quotes": d.Quotes, "Reconciliations": d.Reconciliations})
	return &Reconciler{deps: d}
}

// OnOrderPlaced copies the quoted side out of pricing's own quote.
func (r *Reconciler) OnOrderPlaced(ctx context.Context, m contracts.OrderPlacedV1) error {
	order, err := shared.ParseID(m.ID)
	if err != nil {
		return fmt.Errorf("order_placed: %w", err)
	}
	quoteID, err := pricing.ParseQuoteID(m.Quote)
	if err != nil {
		return fmt.Errorf("order_placed %s: %w", m.ID, err)
	}
	return r.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		q, err := r.deps.Quotes.ByID(ctx, quoteID)
		if err != nil {
			return err // the quote is pricing's own; missing = corrupt, retry and alert
		}
		b := q.Breakdown()
		goods, err := shared.Sum(b.ItemPrice, b.SalesTax, b.Duty)
		if err != nil {
			return err
		}
		freight, err := b.Freight.Add(b.Surcharge)
		if err != nil {
			return err
		}
		return r.update(ctx, order, func(rec *pricing.Reconciliation) {
			rec.Quote, rec.QuotedGoods, rec.QuotedFreight, rec.QuotedChargeable = quoteID.ID, goods, freight, b.Chargeable
		})
	})
}

func (r *Reconciler) OnPurchaseConfirmed(ctx context.Context, m contracts.PurchaseConfirmedV1) error {
	order, err := shared.ParseID(m.Order)
	if err != nil {
		return fmt.Errorf("purchase_confirmed: %w", err)
	}
	paid, err := moneyFrom(m.Paid)
	if err != nil {
		return fmt.Errorf("purchase_confirmed %s: %w", m.Order, err)
	}
	return r.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		return r.update(ctx, order, func(rec *pricing.Reconciliation) {
			rec.ActualGoods = paid
		})
	})
}

func (r *Reconciler) OnBatchShipped(ctx context.Context, m contracts.BatchShippedV1) error {
	return r.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		for _, a := range m.Allocations {
			order, err := shared.ParseID(a.Order)
			if err != nil {
				return fmt.Errorf("batch_shipped %s: %w", m.ID, err)
			}
			freight, err := moneyFrom(a.Freight)
			if err != nil {
				return fmt.Errorf("batch_shipped %s: %w", m.ID, err)
			}
			chargeable, err := shared.NewWeight(a.ChargeableG)
			if err != nil {
				return fmt.Errorf("batch_shipped %s: %w", m.ID, err)
			}
			if err := r.update(ctx, order, func(rec *pricing.Reconciliation) {
				rec.ActualFreight, rec.ActualChargeable = freight, chargeable
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

// update loads the row (or starts one) and saves it back — the upsert shape.
func (r *Reconciler) update(ctx context.Context, order shared.ID, fn func(*pricing.Reconciliation)) error {
	rec, err := r.deps.Reconciliations.ByOrder(ctx, order)
	if errors.Is(err, pricing.ErrReconciliationNotFound) {
		rec = pricing.Reconciliation{Order: order}
	} else if err != nil {
		return err
	}
	fn(&rec)
	return r.deps.Reconciliations.Save(ctx, rec)
}
