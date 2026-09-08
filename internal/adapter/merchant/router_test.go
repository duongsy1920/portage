package merchant_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/memory"
	"github.com/duongsy/portage/internal/adapter/merchant"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

var now = time.Date(2026, 9, 6, 10, 0, 0, 0, time.UTC)

// spy is a shop with an API: it says yes and remembers being asked.
type spy struct {
	called  int
	receipt procurement.PurchaseReceipt
}

func (s *spy) Purchase(context.Context, *procurement.PurchaseTask) (procurement.PurchaseReceipt, bool, string, error) {
	s.called++
	return s.receipt, true, "", nil
}

func task(t *testing.T, product shared.ID) *procurement.PurchaseTask {
	t.Helper()
	task, err := procurement.OpenTask(procurement.TaskDetails{
		Order: shared.NewID(), Product: product, Variant: shared.NewID(), Currency: shared.USD,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	return task
}

// The routing itself: a shop with an adapter gets it, every other shop gets a
// person. Adding a shop is one map entry in wire — no branch in the use case.
func TestRouter_sendsEachShopToItsOwnAdapter(t *testing.T) {
	ctx := context.Background()
	items := memory.NewItemRepo()
	wired, byHand := shared.NewID(), shared.NewID()
	wiredProduct, handProduct := shared.NewID(), shared.NewID()
	if err := items.Save(ctx, procurement.Item{Product: wiredProduct, Merchant: wired, Name: "Air Trainer 90"}); err != nil {
		t.Fatal(err)
	}
	if err := items.Save(ctx, procurement.Item{Product: handProduct, Merchant: byHand, Name: "A bag"}); err != nil {
		t.Fatal(err)
	}

	shop := &spy{receipt: procurement.PurchaseReceipt{}}
	r := merchant.Router{Items: items, ByMerchant: map[shared.ID]procurement.MerchantACL{wired: shop}}

	if _, ok, _, err := r.Purchase(ctx, task(t, wiredProduct)); err != nil || !ok {
		t.Fatalf("wired shop = %v, %v", ok, err)
	}
	if shop.called != 1 {
		t.Fatalf("the shop's adapter was called %d times", shop.called)
	}

	if _, _, _, err := r.Purchase(ctx, task(t, handProduct)); !errors.Is(err, procurement.ErrManualPurchase) {
		t.Fatalf("unmapped shop = %v, want ErrManualPurchase (a person buys it)", err)
	}
	if shop.called != 1 {
		t.Errorf("the wired shop's adapter was called for somebody else's product")
	}
}

// The projection is eventual (§25): a task can exist before procurement has
// heard which shop the product is from. Not knowing must hand the task to a
// person, not fail it — failing it would cancel an order over a race.
func TestRouter_unknownProductFallsBackToAPerson(t *testing.T) {
	ctx := context.Background()
	shop := &spy{}
	r := merchant.Router{
		Items:      memory.NewItemRepo(), // empty: the relay has not caught up
		ByMerchant: map[shared.ID]procurement.MerchantACL{shared.NewID(): shop},
	}
	if _, _, _, err := r.Purchase(ctx, task(t, shared.NewID())); !errors.Is(err, procurement.ErrManualPurchase) {
		t.Fatalf("unknown product = %v, want ErrManualPurchase", err)
	}
	if shop.called != 0 {
		t.Error("a shop's adapter must not be called for a product we cannot place")
	}
}

// With no adapters wired at all, the Router is exactly Manual — and does not
// even query the projection. That is the state the system ships in today.
func TestRouter_withNoAdaptersIsManual(t *testing.T) {
	var r merchant.Router
	if _, _, _, err := r.Purchase(context.Background(), task(t, shared.NewID())); !errors.Is(err, procurement.ErrManualPurchase) {
		t.Fatalf("empty router = %v", err)
	}
}
