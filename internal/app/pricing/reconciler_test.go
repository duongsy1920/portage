package pricingapp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/duongsy/portage/internal/adapter/memory"
	pricingapp "github.com/duongsy/portage/internal/app/pricing"
	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Quote vs Actual for one order: the quoted side comes from pricing's own
// quote, the actual side from procurement's receipt and logistics' invoice.
// Events may arrive in any order; the row completes when all three are in.
func TestReconciler_quoteVsActual(t *testing.T) {
	w := newWorld(t)
	w.seedLaneAndRate(t)
	recs := memory.NewReconciliationRepo()
	w.deps.Reconciliations = recs
	ctx := context.Background()
	p := pricingapp.NewProjector(w.deps)
	_ = p.OnCategoryDefined(ctx, footwearDefined())
	_ = p.OnProductPublished(ctx, published())
	quote, err := pricingapp.NewIssueQuoteHandler(w.deps).Handle(ctx, pricingapp.IssueQuote{Product: shoe, Lane: "us_forwarder"})
	if err != nil {
		t.Fatal(err)
	}
	order := shared.NewID()
	r := pricingapp.NewReconciler(w.deps)

	// the invoice arrives before we even know the order — kept, not lost
	if err := r.OnBatchShipped(ctx, contracts.BatchShippedV1{ID: shared.NewID().String(), Lane: "us_forwarder", Freight: contracts.MoneyV1{Minor: 2750, Currency: "USD"},
		Allocations: []contracts.AllocationV1{{Parcel: shared.NewID().String(), Order: order.String(), ChargeableG: 2500, Freight: contracts.MoneyV1{Minor: 2750, Currency: "USD"}}}, At: now}); err != nil {
		t.Fatal(err)
	}
	rec, err := recs.ByOrder(ctx, order)
	if err != nil || rec.Complete() || rec.ActualFreight != shared.MustParseMoney("27.50", shared.USD) {
		t.Fatalf("after freight only: %+v, %v", rec, err)
	}
	if err := r.OnOrderPlaced(ctx, contracts.OrderPlacedV1{ID: order.String(), Quote: shared.NewID().String(), At: now}); err == nil {
		t.Fatal("an order on a quote pricing never issued must be an error (retry + alert), not a silent row")
	}
	if err := r.OnOrderPlaced(ctx, contracts.OrderPlacedV1{ID: order.String(), Quote: quote.String(), Product: shoe.String(), At: now}); err != nil {
		t.Fatal(err)
	}
	for range 2 { // idempotent
		if err := r.OnPurchaseConfirmed(ctx, contracts.PurchaseConfirmedV1{ID: shared.NewID().String(), Order: order.String(), Reference: "NK-1", Paid: contracts.MoneyV1{Minor: 16322, Currency: "USD"}, At: now}); err != nil {
			t.Fatal(err)
		}
	}
	rec, _ = recs.ByOrder(ctx, order)
	if !rec.Complete() || rec.Quote != quote.ID {
		t.Fatalf("incomplete: %+v", rec)
	}
	// quoted: goods 150 + 13.22 = 163.22, freight 22.50 (standard class in this world) ; actual: 163.22 + 27.50 → variance -5.00
	variance, err := rec.Variance()
	if err != nil || variance != shared.MustParseMoney("-5.00", shared.USD) {
		t.Fatalf("variance = %s, %v (quoted %s + %s, actual %s + %s)", variance, err, rec.QuotedGoods, rec.QuotedFreight, rec.ActualGoods, rec.ActualFreight)
	}
	if rec.QuotedChargeable != shared.Grams(2500) || rec.ActualChargeable != shared.Grams(2500) {
		t.Fatalf("chargeable quoted %s actual %s", rec.QuotedChargeable, rec.ActualChargeable)
	}
	if _, err := (pricing.Reconciliation{}).Variance(); err == nil {
		t.Fatal("variance of an incomplete row must be an error")
	}
	if _, err := recs.ByOrder(ctx, shared.NewID()); !errors.Is(err, pricing.ErrReconciliationNotFound) {
		t.Fatalf("missing: %v", err)
	}
}
