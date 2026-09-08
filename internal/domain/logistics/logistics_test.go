package logistics_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/shared"
)

var now = time.Date(2026, 9, 6, 9, 0, 0, 0, time.UTC)

func usd(s string) shared.Money {
	return shared.MustParseMoney(s, shared.USD)
}

var rule = logistics.ByChargeableWeight{Divisor: 5000, Step: shared.Grams(500)}

func expected(t *testing.T) *logistics.Parcel {
	t.Helper()
	p, err := logistics.ExpectParcel(logistics.ParcelDetails{Order: shared.NewID(), Reference: " NK-1 "}, now)
	if err != nil {
		t.Fatal(err)
	}
	p.PullEvents()
	return p
}

func received(t *testing.T, spec shared.ParcelSpec) *logistics.Parcel {
	t.Helper()
	p := expected(t)
	if err := p.Receive(spec, shared.NewOperatorID(), now); err != nil {
		t.Fatal(err)
	}
	p.PullEvents()
	return p
}

var shoeBox = shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13))  // → 2500 g chargeable
var jacketBox = shared.MustParcelSpec(shared.Grams(900), shared.NewDimensionsCM(40, 32, 18)) // 23 040 cm³ / 5000 = 4608 → 5000 g

func TestParcel_lifecycle(t *testing.T) {
	if _, err := logistics.ExpectParcel(logistics.ParcelDetails{Reference: "x"}, now); !errors.Is(err, logistics.ErrInvalidParcel) {
		t.Fatalf("no order: %v", err)
	}
	if _, err := logistics.ExpectParcel(logistics.ParcelDetails{Order: shared.NewID()}, now); !errors.Is(err, logistics.ErrInvalidParcel) {
		t.Fatalf("no reference: %v", err)
	}
	p := expected(t)
	if p.Reference() != "NK-1" || p.Status() != logistics.ParcelExpected {
		t.Fatalf("parcel = %+v", p.Snapshot())
	}
	if err := p.AssignToBatch(logistics.NewBatchID()); !errors.Is(err, logistics.ErrParcelNotReceived) {
		t.Fatalf("batch before receipt: %v", err)
	}
	if err := p.Receive(shared.ParcelSpec{}, shared.NewOperatorID(), now); !errors.Is(err, shared.ErrIncompleteParcelSpec) {
		t.Fatalf("receive without a measurement: %v", err)
	}
	if err := p.Receive(shoeBox, shared.OperatorID{}, now); !errors.Is(err, shared.ErrOperatorRequired) {
		t.Fatalf("receive without an operator: %v", err)
	}
	op := shared.NewOperatorID()
	if err := p.Receive(shoeBox, op, now); err != nil {
		t.Fatal(err)
	}
	if err := p.Receive(shoeBox, op, now); !errors.Is(err, logistics.ErrParcelNotExpected) {
		t.Fatalf("receive twice: %v", err)
	}
	evs := p.PullEvents()
	if len(evs) != 1 || evs[0].EventName() != "logistics.parcel_received" || evs[0].(logistics.ParcelReceivedEvent).Actual != shoeBox {
		t.Fatalf("events = %v", evs)
	}
	if err := p.MarkShipped(); !errors.Is(err, logistics.ErrParcelNotBatched) {
		t.Fatalf("ship before batch: %v", err)
	}
	b := logistics.NewBatchID()
	if err := p.AssignToBatch(b); err != nil {
		t.Fatal(err)
	}
	if err := p.MarkShipped(); err != nil || p.Status() != logistics.ParcelShipped || p.Batch() != b {
		t.Fatalf("shipped: %v %+v", err, p.Snapshot())
	}
	back, err := logistics.ParcelFromSnapshot(p.Snapshot())
	if err != nil || !reflect.DeepEqual(back.Snapshot(), p.Snapshot()) {
		t.Fatalf("round trip: %v", err)
	}
	for name, mutate := range map[string]func(*logistics.ParcelSnapshot){
		"received, no scale": func(s *logistics.ParcelSnapshot) { s.Actual = shared.ParcelSpec{} },
		"batched, no batch":  func(s *logistics.ParcelSnapshot) { s.Batch = logistics.BatchID{} },
		"unknown status":     func(s *logistics.ParcelSnapshot) { s.Status = "lost" },
	} {
		s := p.Snapshot()
		mutate(&s)
		if _, err := logistics.ParcelFromSnapshot(s); !errors.Is(err, logistics.ErrInvalidSnapshot) {
			t.Errorf("%s: %v", name, err)
		}
	}
	e := expected(t).Snapshot()
	e.Actual = shoeBox
	if _, err := logistics.ParcelFromSnapshot(e); !errors.Is(err, logistics.ErrInvalidSnapshot) {
		t.Errorf("expected parcel with a measurement: %v", err)
	}
}

// One invoice, two parcels: the freight splits by CHARGEABLE weight (the
// carrier's number), and the cents add up to the invoice exactly.
func TestBatch_shipSplitsFreightByChargeableWeight(t *testing.T) {
	shoe, jacket := received(t, shoeBox), received(t, jacketBox)
	b, err := logistics.OpenBatch(" us_forwarder ", now)
	if err != nil || b.Lane() != "us_forwarder" {
		t.Fatal(err)
	}
	if _, err := logistics.OpenBatch("  ", now); !errors.Is(err, logistics.ErrInvalidBatch) {
		t.Fatalf("no lane: %v", err)
	}
	if err := b.Close(now); !errors.Is(err, logistics.ErrBatchEmpty) {
		t.Fatalf("close empty: %v", err)
	}
	if _, err := b.Ship(usd("87.50"), rule, now); !errors.Is(err, logistics.ErrBatchNotClosed) {
		t.Fatalf("ship while open: %v", err)
	}
	if err := b.AddParcel(expected(t).Item()); !errors.Is(err, logistics.ErrInvalidParcel) {
		t.Fatalf("an unweighed parcel: %v", err)
	}
	for _, p := range []*logistics.Parcel{shoe, jacket} {
		if err := b.AddParcel(p.Item()); err != nil {
			t.Fatal(err)
		}
	}
	if err := b.AddParcel(shoe.Item()); !errors.Is(err, logistics.ErrDuplicateParcel) {
		t.Fatalf("same parcel twice: %v", err)
	}
	if err := b.Close(now); err != nil {
		t.Fatal(err)
	}
	if err := b.AddParcel(received(t, shoeBox).Item()); !errors.Is(err, logistics.ErrBatchNotOpen) {
		t.Fatalf("add after close: %v", err)
	}
	if _, err := b.Ship(usd("0.00"), rule, now); !errors.Is(err, logistics.ErrInvalidFreight) {
		t.Fatalf("zero invoice: %v", err)
	}
	if _, err := b.Ship(usd("87.50"), logistics.ByChargeableWeight{}, now); !errors.Is(err, logistics.ErrLaneRuleNotFound) {
		t.Fatalf("no lane rule: %v", err)
	}

	// 2500 g : 5000 g of 87.50 → 29.17 : 58.33 (largest remainder gives the shoe the extra cent)
	allocs, err := b.Ship(usd("87.50"), rule, now)
	if err != nil {
		t.Fatal(err)
	}
	want := []logistics.Allocation{
		{Parcel: shoe.ID().ID, Order: shoe.Order(), Chargeable: shared.Grams(2500), Freight: usd("29.17")},
		{Parcel: jacket.ID().ID, Order: jacket.Order(), Chargeable: shared.Grams(5000), Freight: usd("58.33")},
	}
	if !reflect.DeepEqual(allocs, want) {
		t.Fatalf("allocations = %+v, want %+v", allocs, want)
	}
	evs := b.PullEvents()
	names := []string{}
	for _, e := range evs {
		names = append(names, e.EventName())
	}
	if !reflect.DeepEqual(names, []string{"logistics.batch_opened", "logistics.batch_closed", "logistics.batch_shipped"}) {
		t.Fatalf("events = %v", names)
	}
	shipped := evs[2].(logistics.BatchShippedEvent)
	if shipped.Freight != usd("87.50") || !reflect.DeepEqual(shipped.Allocations, want) {
		t.Fatalf("shipped event = %+v", shipped)
	}
	if _, err := b.Ship(usd("87.50"), rule, now); !errors.Is(err, logistics.ErrBatchNotClosed) {
		t.Fatalf("ship twice: %v", err)
	}

	back, err := logistics.BatchFromSnapshot(b.Snapshot())
	if err != nil || !reflect.DeepEqual(back.Snapshot(), b.Snapshot()) || len(back.PullEvents()) != 0 {
		t.Fatalf("round trip: %v", err)
	}
	for name, mutate := range map[string]func(*logistics.BatchSnapshot){
		"shipped, no invoice":     func(s *logistics.BatchSnapshot) { s.Freight = shared.Money{} },
		"shipped, half allocated": func(s *logistics.BatchSnapshot) { s.Allocations = s.Allocations[:1] },
		"closed, no parcels":      func(s *logistics.BatchSnapshot) { s.Status = logistics.BatchClosed; s.Items = nil },
		"open with allocations":   func(s *logistics.BatchSnapshot) { s.Status = logistics.BatchOpen },
	} {
		s := b.Snapshot()
		mutate(&s)
		if _, err := logistics.BatchFromSnapshot(s); !errors.Is(err, logistics.ErrInvalidSnapshot) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
