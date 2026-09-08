package postgres_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/duongsy/portage/internal/adapter/postgres"
	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

func TestTaskRepo_openListAndRoundTrips(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewTaskRepo(p)
	open := func() *procurement.PurchaseTask {
		task, err := procurement.OpenTask(procurement.TaskDetails{
			Order:    shared.NewID(),
			Product:  shared.NewID(),
			Variant:  shared.NewID(),
			Currency: shared.USD,
			// The subject is four more columns; DeepEqual on the snapshot
			// below is what proves they all come back.
			Subject: procurement.Subject{
				ProductName:  "Air Trainer 90",
				VariantLabel: "M 8 / W 9.5 · black",
				VariantRef:   "EX-AT90-8-BLK",
				Source:       source,
			},
		}, now)
		if err != nil {
			t.Fatal(err)
		}
		task.PullEvents()
		return task
	}
	a, b := open(), open()
	if _, err := repo.ByOrder(ctx, a.Order()); !errors.Is(err, procurement.ErrTaskNotFound) {
		t.Fatalf("missing: %v", err)
	}
	for _, task := range []*procurement.PurchaseTask{a, b} {
		if err := repo.Save(ctx, task); err != nil {
			t.Fatal(err)
		}
	}
	list, err := repo.Open(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("Open = %d, %v", len(list), err)
	}
	got, err := repo.ByOrder(ctx, a.Order())
	if err != nil || !reflect.DeepEqual(got.Snapshot(), a.Snapshot()) {
		t.Fatalf("open round trip: %v", err)
	}

	op := shared.NewOperatorID()
	if err := a.Confirm(procurement.PurchaseReceipt{Reference: "NK-1", Paid: usd("163.22"), PaidBy: op}, now); err != nil {
		t.Fatal(err)
	}
	if err := b.Fail("sold out", now); err != nil {
		t.Fatal(err)
	}
	for _, task := range []*procurement.PurchaseTask{a, b} {
		if err := repo.Save(ctx, task); err != nil {
			t.Fatal(err)
		}
		got, err := repo.ByID(ctx, task.ID())
		if err != nil || !reflect.DeepEqual(got.Snapshot(), task.Snapshot()) {
			t.Fatalf("closed round trip %s: %v\n got %+v\nwant %+v", task.Status(), err, got.Snapshot(), task.Snapshot())
		}
	}
	if list, _ := repo.Open(ctx); len(list) != 0 {
		t.Fatalf("still open: %d", len(list))
	}
}

const source = "https://www.example.com/t/air-trainer-90/abc"

// The variant dictionary: procurement's answer to "what does this id mean in
// the shop". Upsert, because catalog.variant_added is delivered at least once.
func TestVariantRepo_upsert(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewVariantRepo(p)
	v := procurement.Variant{
		Variant:     shared.NewID(),
		Product:     shared.NewID(),
		Size:        "M 8 / W 9.5",
		Color:       "black",
		MerchantRef: "EX-AT90-8-BLK",
	}

	if _, err := repo.ByID(ctx, v.Variant); !errors.Is(err, procurement.ErrVariantNotFound) {
		t.Fatalf("missing: %v", err)
	}
	for range 2 {
		if err := repo.Save(ctx, v); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.ByID(ctx, v.Variant)
	if err != nil || got != v {
		t.Fatalf("round trip = %+v, %v", got, err)
	}
	if got.Label() != "M 8 / W 9.5 · black" {
		t.Fatalf("label = %q", got.Label())
	}

	// A one-size product: nothing to say, and that is not an error.
	plain := procurement.Variant{Variant: shared.NewID(), Product: v.Product}
	if err := repo.Save(ctx, plain); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.ByID(ctx, plain.Variant); err != nil || got != plain || got.Label() != "" {
		t.Fatalf("one-size round trip = %+v (label %q), %v", got, got.Label(), err)
	}
}

func TestShopAndItemRepos_upsert(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	shops, items := postgres.NewShopRepo(p), postgres.NewItemRepo(p)
	shop := procurement.Shop{Merchant: shared.NewID(), Name: "Example Sports", Site: "www.example.com", Currency: shared.USD}
	item := procurement.Item{Product: shared.NewID(), Merchant: shop.Merchant, Name: "Air Trainer 90", Source: source}

	if _, err := shops.ByID(ctx, shop.Merchant); !errors.Is(err, procurement.ErrShopNotFound) {
		t.Fatalf("missing shop: %v", err)
	}
	if _, err := items.ByProduct(ctx, item.Product); !errors.Is(err, procurement.ErrItemNotFound) {
		t.Fatalf("missing item: %v", err)
	}
	for range 2 {
		if err := shops.Save(ctx, shop); err != nil {
			t.Fatal(err)
		}
		if err := items.Save(ctx, item); err != nil {
			t.Fatal(err)
		}
	}
	if got, err := shops.ByID(ctx, shop.Merchant); err != nil || got != shop {
		t.Fatalf("shop = %+v, %v", got, err)
	}
	if got, err := items.ByProduct(ctx, item.Product); err != nil || got != item {
		t.Fatalf("item = %+v, %v", got, err)
	}
}
