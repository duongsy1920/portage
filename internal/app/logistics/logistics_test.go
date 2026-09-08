package logisticsapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/memory"
	logisticsapp "github.com/duongsy/portage/internal/app/logistics"
	"github.com/duongsy/portage/internal/contracts"
	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/clock"
)

var now = time.Date(2026, 9, 6, 9, 0, 0, 0, time.UTC)

type world struct {
	parcels *memory.ParcelRepo
	batches *memory.BatchRepo
	outbox  *memory.Outbox
	deps    logisticsapp.Deps
}

func newWorld(t *testing.T) *world {
	t.Helper()
	w := &world{parcels: memory.NewParcelRepo(), batches: memory.NewBatchRepo(), outbox: memory.NewOutbox()}
	w.deps = logisticsapp.Deps{Clock: clock.FixedAt(now), UoW: memory.UnitOfWork{}, Outbox: w.outbox, Parcels: w.parcels, Batches: w.batches, Lanes: memory.NewLaneRuleRepo()}
	if err := logisticsapp.NewProjector(w.deps).OnLaneDefined(context.Background(), contracts.LaneDefinedV1{Code: "us_forwarder", Name: "US forwarder", Divisor: 5000, StepG: 500, Currency: "USD", At: now}); err != nil {
		t.Fatal(err)
	}
	return w
}

func names(evs []shared.Event) []string {
	out := []string{}
	for _, e := range evs {
		out = append(out, e.EventName())
	}
	return out
}

func usd(s string) shared.Money {
	return shared.MustParseMoney(s, shared.USD)
}

// The warehouse flow: two purchases → two expected parcels → weighed → boxed
// → shipped with one invoice split by chargeable weight.
func TestLogistics_purchaseToShippedBatch(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	expect := logisticsapp.NewExpectParcelHandler(w.deps)
	shoeOrder, jacketOrder := shared.NewID(), shared.NewID()
	for range 2 { // at-least-once
		if err := expect.OnPurchaseConfirmed(ctx, contracts.PurchaseConfirmedV1{ID: shared.NewID().String(), Order: shoeOrder.String(), Reference: "NK-1", Paid: contracts.MoneyV1{Minor: 16322, Currency: "USD"}, At: now}); err != nil {
			t.Fatal(err)
		}
	}
	if err := expect.OnPurchaseConfirmed(ctx, contracts.PurchaseConfirmedV1{ID: shared.NewID().String(), Order: jacketOrder.String(), Reference: "TNF-9", Paid: contracts.MoneyV1{Minor: 20000, Currency: "USD"}, At: now}); err != nil {
		t.Fatal(err)
	}
	pending, _ := w.parcels.Pending(ctx)
	if len(pending) != 2 {
		t.Fatalf("expected parcels = %d", len(pending))
	}
	if got := names(w.outbox.Drain()); len(got) != 2 || got[0] != "logistics.parcel_expected" {
		t.Fatalf("outbox = %v", got)
	}
	shoe, _ := w.parcels.ByOrder(ctx, shoeOrder)
	jacket, _ := w.parcels.ByOrder(ctx, jacketOrder)

	open := logisticsapp.NewOpenBatchHandler(w.deps)
	if _, err := open.Handle(ctx, "sea_freight"); !errors.Is(err, logistics.ErrLaneRuleNotFound) {
		t.Fatalf("unknown lane: %v", err)
	}
	batch, err := open.Handle(ctx, "us_forwarder")
	if err != nil {
		t.Fatal(err)
	}
	add := logisticsapp.NewAddParcelHandler(w.deps)
	if err := add.Handle(ctx, logisticsapp.AddParcelToBatch{Batch: batch, Parcel: shoe.ID()}); !errors.Is(err, logistics.ErrInvalidParcel) {
		t.Fatalf("boxing an unweighed parcel: %v", err)
	}

	receive := logisticsapp.NewReceiveParcelHandler(w.deps)
	op := shared.NewOperatorID()
	if err := receive.Handle(ctx, logisticsapp.ReceiveParcel{Parcel: shoe.ID(), Actual: shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13))}); !errors.Is(err, shared.ErrOperatorRequired) {
		t.Fatalf("receive without an operator: %v", err)
	}
	for _, c := range []struct {
		id   logistics.ParcelID
		spec shared.ParcelSpec
	}{
		{shoe.ID(), shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13))},
		{jacket.ID(), shared.MustParcelSpec(shared.Grams(900), shared.NewDimensionsCM(40, 32, 18))},
	} {
		if err := receive.Handle(ctx, logisticsapp.ReceiveParcel{Parcel: c.id, Actual: c.spec, Operator: op}); err != nil {
			t.Fatal(err)
		}
		if err := add.Handle(ctx, logisticsapp.AddParcelToBatch{Batch: batch, Parcel: c.id}); err != nil {
			t.Fatal(err)
		}
	}
	if err := add.Handle(ctx, logisticsapp.AddParcelToBatch{Batch: batch, Parcel: shoe.ID()}); !errors.Is(err, logistics.ErrDuplicateParcel) {
		t.Fatalf("boxing the same parcel twice: %v", err) // the batch is asked first and says duplicate; the parcel would say "already batched"
	}
	if _, err := logisticsapp.NewShipBatchHandler(w.deps).Handle(ctx, logisticsapp.ShipBatch{Batch: batch, Freight: usd("87.50")}); !errors.Is(err, logistics.ErrBatchNotClosed) {
		t.Fatalf("ship an open batch: %v", err)
	}
	if err := logisticsapp.NewCloseBatchHandler(w.deps).Handle(ctx, batch); err != nil {
		t.Fatal(err)
	}
	w.outbox.Drain()

	allocs, err := logisticsapp.NewShipBatchHandler(w.deps).Handle(ctx, logisticsapp.ShipBatch{Batch: batch, Freight: usd("87.50")})
	if err != nil || len(allocs) != 2 || allocs[0].Freight != usd("29.17") || allocs[1].Freight != usd("58.33") {
		t.Fatalf("allocations = %+v, %v", allocs, err)
	}
	if got := names(w.outbox.Drain()); len(got) != 1 || got[0] != "logistics.batch_shipped" {
		t.Fatalf("outbox = %v", got)
	}
	shoe, _ = w.parcels.ByID(ctx, shoe.ID())
	if shoe.Status() != logistics.ParcelShipped || shoe.Batch() != batch {
		t.Fatalf("parcel after ship = %+v", shoe.Snapshot())
	}
	if pending, _ := w.parcels.Pending(ctx); len(pending) != 0 {
		t.Fatalf("still pending: %d", len(pending))
	}
}
