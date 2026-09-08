package postgres_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/postgres"
	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The read model round trip. The interesting half is the FRAME: a row written
// by an event that arrived before order_placed has no customer, no money and
// no status, and every one of those must come back as a zero value rather than
// as a UUID of zeros or 0 VND — a fake fact on a screen is worse than a blank.
func TestOrderSummaryRepo_roundTripIncludingTheHalfEmptyFrame(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewOrderSummaryRepo(p)

	frame := reportingapp.OrderSummary{
		Order: ordering.NewOrderID(), Tracking: reportingapp.TrackingShipped, UpdatedAt: now,
	}
	if _, err := repo.ByOrder(ctx, frame.Order); !errors.Is(err, reportingapp.ErrSummaryNotFound) {
		t.Fatalf("missing summary: %v, want ErrSummaryNotFound", err)
	}
	if err := repo.Save(ctx, frame); err != nil {
		t.Fatalf("Save frame: %v", err)
	}
	got, err := repo.ByOrder(ctx, frame.Order)
	if err != nil {
		t.Fatalf("ByOrder: %v", err)
	}
	if !reflect.DeepEqual(got, frame) {
		t.Fatalf("frame round trip changed the row:\n got %+v\nwant %+v", got, frame)
	}
	if !got.Customer.IsZero() || got.Total.IsValid() || got.PlacedAt != (time.Time{}) {
		t.Errorf("empty columns came back filled: %+v", got)
	}

	full := frame
	full.Customer, full.Product, full.Variant, full.Quote = shared.NewID(), shared.NewID(), shared.NewID(), shared.NewID()
	full.ProductName, full.Status = "Air Trainer 90", ordering.StatusDelivered
	full.Total = shared.MustParseMoney("5393720", shared.VND)
	full.Deposit = shared.MustParseMoney("2696860", shared.VND)
	full.Refund = shared.MustParseMoney("0", shared.VND)
	full.DepositPaid, full.BalancePaid, full.Forfeited = true, true, true
	full.ShopReference = "NK-20260905-001"
	full.PlacedAt, full.DeliveredAt, full.CancelledAt = now, now.Add(time.Hour), now.Add(2*time.Hour)
	full.UpdatedAt = now.Add(2 * time.Hour)
	if err := repo.Save(ctx, full); err != nil {
		t.Fatalf("Save full: %v", err)
	}
	if got, err = repo.ByOrder(ctx, full.Order); err != nil || !reflect.DeepEqual(got, full) {
		t.Fatalf("full round trip:\n got %+v\nwant %+v\n(%v)", got, full, err)
	}
}

// The two screens, plus the backfill query. Ordering is newest first, which is
// what a "my orders" list shows and what both indexes in 0007 are built for.
func TestOrderSummaryRepo_theTwoScreens(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewOrderSummaryRepo(p)
	mine, product := shared.NewID(), shared.NewID()

	save := func(customer, product shared.ID, status ordering.OrderStatus, at time.Time) ordering.OrderID {
		t.Helper()
		id := ordering.NewOrderID()
		if err := repo.Save(ctx, reportingapp.OrderSummary{
			Order: id, Customer: customer, Product: product, Status: status,
			Tracking: reportingapp.TrackingNone, PlacedAt: at, UpdatedAt: at,
		}); err != nil {
			t.Fatal(err)
		}
		return id
	}
	older := save(mine, product, ordering.StatusAwaitingDeposit, now)
	newer := save(mine, product, ordering.StatusDeposited, now.Add(time.Hour))
	save(shared.NewID(), shared.NewID(), ordering.StatusDeposited, now.Add(2*time.Hour))

	rows, err := repo.ByCustomer(ctx, mine)
	if err != nil || len(rows) != 2 {
		t.Fatalf("ByCustomer = %d rows, %v", len(rows), err)
	}
	if rows[0].Order != newer || rows[1].Order != older {
		t.Errorf("not newest-first: %v", rows)
	}
	if rows, _ = repo.ByStatus(ctx, ordering.StatusDeposited); len(rows) != 2 {
		t.Errorf("ByStatus(deposited) = %d rows", len(rows))
	}
	if rows, _ = repo.ByProduct(ctx, product); len(rows) != 2 {
		t.Errorf("ByProduct = %d rows (the late-name backfill needs this)", len(rows))
	}
	if rows, _ = repo.All(ctx); len(rows) != 3 {
		t.Errorf("All = %d rows", len(rows))
	}
}

// A missing product name is not an error: the product may simply not be
// published yet, and the projector treats "not known" as a normal state.
func TestProductNameRepo_upsertAndMissIsNotAnError(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewProductNameRepo(p)
	product := shared.NewID()

	name, ok, err := repo.Name(ctx, product)
	if err != nil || ok || name != "" {
		t.Fatalf("miss = %q, %v, %v — want a clean not-found", name, ok, err)
	}
	if err := repo.Save(ctx, product, "Air Trainer 90"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, product, "Air Trainer 90 (2026)"); err != nil {
		t.Fatal(err) // a repriced/renamed product re-publishes: upsert, not insert
	}
	if name, ok, _ = repo.Name(ctx, product); !ok || name != "Air Trainer 90 (2026)" {
		t.Fatalf("name = %q, %v", name, ok)
	}
}

// The worklist read model on real columns: one nullable id, four booleans and
// two queries the screens live on.
func TestProductWorklistRepo_roundTripAndTheTwoQueries(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewProductWorklistRepo(p)

	waiting := shared.NewID()
	mine := reportingapp.WorklistItem{
		Product: shared.NewID(), Merchant: shared.NewID(), Category: "footwear", Name: "Air Trainer 90",
		Source: "https://www.example.com/t/air-trainer-90/abc",
		Price:  shared.MustParseMoney("150.00", shared.USD), SourcedBy: "customer", RequestedBy: waiting,
		Variants: []reportingapp.WorklistVariant{{ID: shared.NewID(), Label: "US 9 · black"}},
		AddedAt:  now, UpdatedAt: now,
	}
	// An operator added this one on spec: requested_by goes down as NULL, not
	// as a uuid of all zeros that would read back as a real customer.
	onSpec := reportingapp.WorklistItem{
		Product: shared.NewID(), Merchant: shared.NewID(), Category: "apparel", Name: "On spec",
		Source: "https://www.example.com/t/y", Price: shared.MustParseMoney("20.00", shared.USD),
		SourcedBy: "operator", Published: true,
		AddedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute),
	}

	if _, err := repo.ByProduct(ctx, mine.Product); !errors.Is(err, reportingapp.ErrWorklistItemNotFound) {
		t.Fatalf("missing row: %v", err)
	}
	for range 2 { // upsert, because the relay is at-least-once
		for _, item := range []reportingapp.WorklistItem{mine, onSpec} {
			if err := repo.Save(ctx, item); err != nil {
				t.Fatal(err)
			}
		}
	}
	got, err := repo.ByProduct(ctx, mine.Product)
	if err != nil || !reflect.DeepEqual(got, mine) {
		t.Fatalf("round trip:\n got %+v\nwant %+v\n%v", got, mine, err)
	}
	if back, err := repo.ByProduct(ctx, onSpec.Product); err != nil || !back.RequestedBy.IsZero() {
		t.Fatalf("NULL requested_by must read back as the zero id: %+v, %v", back, err)
	}

	// Open is unfinished work only, and the published row is finished.
	open, err := repo.Open(ctx)
	if err != nil || len(open) != 1 || open[0].Product != mine.Product {
		t.Fatalf("Open = %+v, %v", open, err)
	}
	// ByRequester is the customer's own list, and the zero id owns nothing.
	ours, err := repo.ByRequester(ctx, waiting)
	if err != nil || len(ours) != 1 || ours[0].Product != mine.Product {
		t.Fatalf("ByRequester = %+v, %v", ours, err)
	}
	if none, err := repo.ByRequester(ctx, shared.ID{}); err != nil || len(none) != 0 {
		t.Fatalf("the zero customer = %+v, %v", none, err)
	}
}
