package pricing

import (
	"context"
	"errors"

	"github.com/duongsy/portage/internal/domain/shared"
)

var ErrReconciliationNotFound = errors.New("reconciliation not found")

// Reconciliation is Quote vs Actual for ONE order (DDD.md §28) — the record
// that says whether we made or lost money on it. It is a projection built
// from three contexts' events: ordering.order_placed (which quote), procurement
// .purchase_confirmed (what the shop really charged), logistics.batch_shipped
// (what the carrier really charged for this parcel). Pricing owns it because
// pricing owns the other half — the quote.
//
// Everything is in the LANE currency (USD): that is where the costs happen;
// the customer's VND total is the quote's business, not this comparison's.
type Reconciliation struct {
	Order shared.ID
	Quote shared.ID

	QuotedGoods      shared.Money  // item + sales tax + duty, as quoted
	QuotedFreight    shared.Money  // freight + surcharge, as quoted
	QuotedChargeable shared.Weight // grams we quoted on

	ActualGoods      shared.Money  // what procurement paid the shop (zero Money until known)
	ActualFreight    shared.Money  // this order's share of the carrier invoice (zero Money until known)
	ActualChargeable shared.Weight // grams the carrier billed on
}

// Complete: both actuals are in — the order has been bought AND shipped.
func (r Reconciliation) Complete() bool {
	return r.ActualGoods.IsValid() && r.ActualFreight.IsValid()
}

// Variance is quoted minus actual: positive = we quoted more than it cost
// (margin protected), negative = the estimate was low (we ate it). Valid only
// when Complete.
func (r Reconciliation) Variance() (shared.Money, error) {
	if !r.Complete() {
		return shared.Money{}, errors.New("reconciliation incomplete")
	}
	quoted, err := r.QuotedGoods.Add(r.QuotedFreight)
	if err != nil {
		return shared.Money{}, err
	}
	actual, err := r.ActualGoods.Add(r.ActualFreight)
	if err != nil {
		return shared.Money{}, err
	}
	return quoted.Sub(actual)
}

type ReconciliationRepository interface {
	ByOrder(ctx context.Context, order shared.ID) (Reconciliation, error) // ErrReconciliationNotFound
	Save(ctx context.Context, r Reconciliation) error
}
