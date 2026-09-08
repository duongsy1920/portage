package postgres_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/duongsy/portage/internal/adapter/postgres"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

func TestOrderRepo_roundTripByIDAndByQuote(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewOrderRepo(p)
	o, err := ordering.PlaceOrder(ordering.OrderDetails{
		Quote: shared.NewID(), Product: shared.NewID(), Variant: shared.NewID(), Customer: shared.NewID(),
		Total: vnd("5393720"), Deposit: vnd("2696860"),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	o.PullEvents()

	if _, err := repo.ByID(ctx, o.ID()); !errors.Is(err, ordering.ErrOrderNotFound) {
		t.Fatalf("missing order: %v", err)
	}
	if _, err := repo.ByQuote(ctx, o.Quote()); !errors.Is(err, ordering.ErrOrderNotFound) {
		t.Fatalf("missing order by quote: %v", err)
	}
	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.ByQuote(ctx, o.Quote())
	if err != nil || !reflect.DeepEqual(got.Snapshot(), o.Snapshot()) {
		t.Fatalf("ByQuote round trip: %v\n got %+v\nwant %+v", err, got.Snapshot(), o.Snapshot())
	}

	// walk it to cancelled-with-forfeit: the nullable refund and the flag columns
	_ = o.PayDeposit(vnd("2696860"), now)
	_ = o.ConfirmPurchase(now)
	if _, err := o.Cancel("changed mind after purchase", now); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("Save after cancel: %v", err)
	}
	got, err = repo.ByID(ctx, o.ID())
	if err != nil || !reflect.DeepEqual(got.Snapshot(), o.Snapshot()) {
		t.Fatalf("cancelled round trip: %v\n got %+v\nwant %+v", err, got.Snapshot(), o.Snapshot())
	}
	if !got.Refund().Forfeited() || got.Refund().Amount() != vnd("0") {
		t.Fatalf("refund = %+v", got.Refund())
	}
}

// Ordering's copy is existence and owner, nothing else: it is what lets
// PlaceOrder refuse a variant id the customer did not get from us.
func TestOrderingVariantRepo_upsert(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewOrderingVariantRepo(p)
	v := ordering.Variant{Variant: shared.NewID(), Product: shared.NewID()}

	if _, err := repo.ByID(ctx, v.Variant); !errors.Is(err, ordering.ErrVariantNotFound) {
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
}

func TestAcceptedQuoteRepo_upsert(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewAcceptedQuoteRepo(p)
	q := ordering.AcceptedQuote{Quote: shared.NewID(), Product: shared.NewID(), Total: vnd("5393720"), Deposit: vnd("2696860")}

	if _, err := repo.ByID(ctx, q.Quote); !errors.Is(err, ordering.ErrAcceptedQuoteNotFound) {
		t.Fatalf("missing: %v", err)
	}
	for range 2 {
		if err := repo.Save(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.ByID(ctx, q.Quote)
	if err != nil || got != q {
		t.Fatalf("round trip = %+v, %v", got, err)
	}
}
