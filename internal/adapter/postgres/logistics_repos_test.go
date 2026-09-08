package postgres_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/duongsy/portage/internal/adapter/postgres"
	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

func TestLogisticsRepos_parcelBatchAndLaneRule(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	parcels, batches, lanes := postgres.NewParcelRepo(p), postgres.NewBatchRepo(p), postgres.NewLaneRuleRepo(p)

	if _, err := lanes.ByCode(ctx, "us_forwarder"); !errors.Is(err, logistics.ErrLaneRuleNotFound) {
		t.Fatalf("missing rule: %v", err)
	}
	rule := logistics.LaneRule{Code: "us_forwarder", Divisor: 5000, Step: shared.Grams(500)}
	for range 2 {
		if err := lanes.Save(ctx, rule); err != nil {
			t.Fatal(err)
		}
	}
	if got, err := lanes.ByCode(ctx, "us_forwarder"); err != nil || got != rule {
		t.Fatalf("rule = %+v, %v", got, err)
	}

	shoe, _ := logistics.ExpectParcel(logistics.ParcelDetails{Order: shared.NewID(), Reference: "NK-1"}, now)
	jacket, _ := logistics.ExpectParcel(logistics.ParcelDetails{Order: shared.NewID(), Reference: "TNF-9"}, now)
	for _, x := range []*logistics.Parcel{shoe, jacket} {
		x.PullEvents()
		if err := parcels.Save(ctx, x); err != nil {
			t.Fatal(err)
		}
	}
	if got, err := parcels.ByOrder(ctx, shoe.Order()); err != nil || !reflect.DeepEqual(got.Snapshot(), shoe.Snapshot()) {
		t.Fatalf("expected round trip: %v", err)
	}
	op := shared.NewOperatorID()
	_ = shoe.Receive(shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13)), op, now)
	_ = jacket.Receive(shared.MustParcelSpec(shared.Grams(900), shared.NewDimensionsCM(40, 32, 18)), op, now)

	b, _ := logistics.OpenBatch("us_forwarder", now)
	b.PullEvents()
	if err := batches.Save(ctx, b); err != nil {
		t.Fatal(err)
	}
	for _, x := range []*logistics.Parcel{shoe, jacket} {
		if err := b.AddParcel(x.Item()); err != nil {
			t.Fatal(err)
		}
		_ = x.AssignToBatch(b.ID())
		if err := parcels.Save(ctx, x); err != nil {
			t.Fatal(err)
		}
	}
	_ = b.Close(now)
	if _, err := b.Ship(usd("87.50"), logistics.ByChargeableWeight{Divisor: 5000, Step: shared.Grams(500)}, now); err != nil {
		t.Fatal(err)
	}
	if err := batches.Save(ctx, b); err != nil {
		t.Fatal(err)
	}
	got, err := batches.ByID(ctx, b.ID())
	if err != nil || !reflect.DeepEqual(got.Snapshot(), b.Snapshot()) {
		t.Fatalf("shipped batch round trip: %v\n got %+v\nwant %+v", err, got.Snapshot(), b.Snapshot())
	}
	if _, err := batches.ByID(ctx, logistics.NewBatchID()); !errors.Is(err, logistics.ErrBatchNotFound) {
		t.Fatalf("missing batch: %v", err)
	}
	pending, _ := parcels.Pending(ctx)
	if len(pending) != 2 {
		t.Fatalf("pending = %d (batched, not yet shipped)", len(pending))
	}
	_ = shoe.MarkShipped()
	if err := parcels.Save(ctx, shoe); err != nil {
		t.Fatal(err)
	}
	if got, _ := parcels.ByID(ctx, shoe.ID()); !reflect.DeepEqual(got.Snapshot(), shoe.Snapshot()) {
		t.Fatalf("shipped parcel round trip: %+v", got.Snapshot())
	}
	if pending, _ := parcels.Pending(ctx); len(pending) != 1 {
		t.Fatalf("pending after one shipped = %d", len(pending))
	}
}

func TestReconciliationRepo_growsAsEventsArrive(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewReconciliationRepo(p)
	order := shared.NewID()
	if _, err := repo.ByOrder(ctx, order); !errors.Is(err, pricing.ErrReconciliationNotFound) {
		t.Fatalf("missing: %v", err)
	}
	partial := pricing.Reconciliation{Order: order, ActualFreight: usd("27.50"), ActualChargeable: shared.Grams(2500)}
	if err := repo.Save(ctx, partial); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ByOrder(ctx, order)
	if err != nil || got != partial || got.Complete() {
		t.Fatalf("partial = %+v, %v", got, err)
	}
	full := partial
	full.Quote, full.QuotedGoods, full.QuotedFreight, full.QuotedChargeable, full.ActualGoods = shared.NewID(), usd("163.22"), usd("25.00"), shared.Grams(2500), usd("163.22")
	if err := repo.Save(ctx, full); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.ByOrder(ctx, order)
	if got != full || !got.Complete() {
		t.Fatalf("full = %+v", got)
	}
	if v, _ := got.Variance(); v != usd("-2.50") {
		t.Fatalf("variance = %s", v)
	}
}
