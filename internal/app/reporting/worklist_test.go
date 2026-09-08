package reportingapp_test

import (
	"context"
	"testing"

	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/shared"
)

// NextStep is the one place that decides what is missing, so it gets a table
// rather than being re-derived in two screens that will drift apart.
func TestNextStep_namesTheFirstThingMissing(t *testing.T) {
	for name, c := range map[string]struct {
		item reportingapp.WorklistItem
		want reportingapp.Step
	}{
		"fresh paste": {reportingapp.WorklistItem{}, reportingapp.StepVariant},
		"has a size":  {reportingapp.WorklistItem{Variants: oneSize()}, reportingapp.StepListing},
		"vouched for": {reportingapp.WorklistItem{Variants: oneSize(), ListingConfirmed: true}, reportingapp.StepMeasure},
		"weighed":     {reportingapp.WorklistItem{Variants: oneSize(), ListingConfirmed: true, Measured: true}, reportingapp.StepPublish},
		"published":   {reportingapp.WorklistItem{Variants: oneSize(), ListingConfirmed: true, Measured: true, Published: true}, reportingapp.StepDone},
		// Published wins even if the flags look impossible: the row is a
		// projection, and arguing with it would only hide a relay problem.
		"published early": {reportingapp.WorklistItem{Published: true}, reportingapp.StepDone},
	} {
		if got := c.item.NextStep(); got != c.want {
			t.Errorf("%s: NextStep = %q, want %q", name, got, c.want)
		}
	}
}

func oneSize() []reportingapp.WorklistVariant {
	return []reportingapp.WorklistVariant{{ID: shared.NewID(), Label: "US 9"}}
}

var (
	size1 = shared.NewID()
	size2 = shared.NewID()
	prod  = shared.NewID()
	shop  = shared.NewID()
	who   = shared.NewID()
)

func added() contracts.ProductAddedV1 {
	return contracts.ProductAddedV1{
		ID: prod.String(), Merchant: shop.String(), Category: "footwear", Name: "Air Trainer 90",
		Source:    "https://www.example.com/t/air-trainer-90/abc",
		Price:     contracts.MoneyV1{Minor: 15000, Currency: "USD"},
		SourcedBy: "customer", RequestedBy: who.String(),
		RequestedVariant: "M 8 / W 9.5", At: now,
	}
}

// The four steps, delivered twice and OUT OF ORDER — which is all the relay
// ever promises. The measurement arrives before anybody has heard of the
// product, and the row still ends up correct.
func TestWorklist_isIdempotentAndOrderTolerant(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()

	// Backwards on purpose: measure, then variant, then the paste itself.
	for range 2 {
		if err := w.p.OnProductMeasured(ctx, contracts.ProductMeasuredV1{ID: prod.String(), Verified: true, At: now}); err != nil {
			t.Fatal(err)
		}
	}
	// A row exists already, built from an event about a product the read
	// model has never been introduced to. Ignoring it would lose the fact.
	if item, err := w.worklist.ByProduct(ctx, prod); err != nil || !item.Measured {
		t.Fatalf("measured before added = %+v, %v", item, err)
	}
	for range 2 {
		// The SAME variant twice: a re-delivery, not a second size.
		if err := w.p.OnVariantAdded(ctx, contracts.VariantAddedV1{
			Product: prod.String(), Variant: size1.String(), Size: "M 8 / W 9.5", Color: "black", At: now,
		}); err != nil {
			t.Fatal(err)
		}
		if err := w.p.OnProductAdded(ctx, added()); err != nil {
			t.Fatal(err)
		}
		if err := w.p.OnListingConfirmed(ctx, contracts.ListingConfirmedV1{
			ID: prod.String(), By: shared.NewOperatorID().String(), At: now,
		}); err != nil {
			t.Fatal(err)
		}
	}

	item, err := w.worklist.ByProduct(ctx, prod)
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "Air Trainer 90" || item.Category != "footwear" || item.Merchant != shop {
		t.Fatalf("the late paste must fill the frame: %+v", item)
	}
	if item.Source == "" || item.Price.Minor() != 15000 {
		t.Fatalf("a row without the page and the price is not actionable: %+v", item)
	}
	if !item.HasVariant() || !item.ListingConfirmed || !item.Measured || item.Published {
		t.Fatalf("flags = %+v", item)
	}
	if len(item.Variants) != 1 {
		t.Fatalf("a size delivered twice must not appear twice: %+v", item.Variants)
	}
	if item.FirstLabel() != "M 8 / W 9.5 · black" {
		t.Fatalf("label = %q", item.FirstLabel())
	}

	// A genuinely different size DOES appear, and the customer's dropdown then
	// has two options to pick from.
	if err := w.p.OnVariantAdded(ctx, contracts.VariantAddedV1{
		Product: prod.String(), Variant: size2.String(), Size: "US 10", At: now,
	}); err != nil {
		t.Fatal(err)
	}
	item, _ = w.worklist.ByProduct(ctx, prod)
	if len(item.Variants) != 2 || item.Variants[1].Label != "US 10" {
		t.Fatalf("two sizes = %+v", item.Variants)
	}
	if item.Variants[0].ID != size1 || item.Variants[1].ID != size2 {
		t.Fatalf("order must be the order they were announced: %+v", item.Variants)
	}
	if item.NextStep() != reportingapp.StepPublish {
		t.Fatalf("next step = %q", item.NextStep())
	}
	if item.RequestedBy != who {
		t.Fatalf("requested_by = %s, want %s", item.RequestedBy, who)
	}
	// The wish stays on the row after the real variant exists, so a person can
	// see that "asked for M 8 / W 9.5" and "we created M 8 / W 9.5 · black"
	// are not the same string and decide whether that matters.
	if item.RequestedVariant != "M 8 / W 9.5" {
		t.Fatalf("requested_variant = %q", item.RequestedVariant)
	}

	// Only publish takes it off the staff queue.
	if open, _ := w.worklist.Open(ctx); len(open) != 1 {
		t.Fatalf("open = %d, want the row to still be waiting", len(open))
	}
	if err := w.p.OnProductPublished(ctx, contracts.ProductPublishedV1{
		ID: prod.String(), Merchant: shop.String(), Category: "footwear", Name: "Air Trainer 90",
		Price: contracts.MoneyV1{Minor: 15000, Currency: "USD"}, At: now,
	}); err != nil {
		t.Fatal(err)
	}
	if open, _ := w.worklist.Open(ctx); len(open) != 0 {
		t.Fatalf("open after publish = %d", len(open))
	}
	// But the customer keeps seeing it: their question is "where is my thing",
	// not "is there work left".
	mine, err := w.worklist.ByRequester(ctx, who)
	if err != nil || len(mine) != 1 || !mine[0].Published {
		t.Fatalf("ByRequester = %+v, %v", mine, err)
	}
}

// An operator adding a product on spec leaves nobody waiting, and the zero id
// must never behave like a customer who owns every unrequested row.
func TestWorklist_theZeroCustomerOwnsNothing(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	onSpec := added()
	onSpec.ID = shared.NewID().String()
	onSpec.SourcedBy, onSpec.RequestedBy = "operator", ""
	if err := w.p.OnProductAdded(ctx, onSpec); err != nil {
		t.Fatal(err)
	}
	if mine, err := w.worklist.ByRequester(ctx, shared.ID{}); err != nil || len(mine) != 0 {
		t.Fatalf("the zero customer = %+v, %v; want nothing", mine, err)
	}
	// It is still work somebody has to finish.
	if open, _ := w.worklist.Open(ctx); len(open) != 1 {
		t.Fatalf("open = %d", len(open))
	}
}
