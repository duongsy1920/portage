package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/duongsy/portage/internal/adapter/postgres"
	"github.com/duongsy/portage/internal/adapter/postgres/pgtest"
	"github.com/duongsy/portage/internal/app"
	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/clock"
)

var now = time.Date(2026, 9, 4, 18, 0, 0, 0, time.UTC)

// pool is the shared, locked, empty test database (see pgtest).
func pool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return pgtest.Pool(t)
}

func aMerchant(t *testing.T) *catalog.Merchant {
	t.Helper()
	m, err := catalog.RegisterMerchant(catalog.MerchantDetails{
		Name: "Example Sports", Site: catalog.MustParseHostname("www.example.com"), Currency: shared.USD,
		FreeShipping: catalog.MustFreeShippingOver(shared.MustParseMoney("50.00", shared.USD)),
		Sourcing:     []catalog.SourcingMode{catalog.SourcedByOperator, catalog.SourcedByCustomer},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	m.PullEvents()
	return m
}

func aCategory(t *testing.T) catalog.CategoryPolicy {
	t.Helper()
	c, err := catalog.NewCategoryPolicy(catalog.MustParseCategoryCode("footwear"),
		shared.MustParcelSpec(shared.Grams(1200), shared.NewDimensionsCM(33, 22, 13)),
		[]catalog.Restriction{catalog.RestrictionMagnet})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestMigrate_isIdempotent(t *testing.T) {
	p := pool(t)
	if err := postgres.Migrate(context.Background(), p); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	var n int
	if err := p.QueryRow(context.Background(), `SELECT count(*) FROM schema_migrations`).Scan(&n); err != nil || n == 0 {
		t.Fatalf("schema_migrations = %d, %v", n, err)
	}
}

// Save → ByID must give back the SAME state — the snapshot round trip, now
// through real columns — and Save again must update, not duplicate.
func TestMerchantRepo_roundTrip(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewMerchantRepo(p)
	m := aMerchant(t)

	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save: %v", err)
	}
	back, err := repo.ByID(ctx, m.ID())
	if err != nil {
		t.Fatalf("ByID: %v", err)
	}
	if !reflect.DeepEqual(back.Snapshot(), m.Snapshot()) {
		t.Fatalf("round trip changed the state:\n got  %+v\n want %+v", back.Snapshot(), m.Snapshot())
	}

	if err := m.Suspend("account banned", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := m.ChangeFreeShipping(catalog.AlwaysFreeShipping(), now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, m); err != nil {
		t.Fatalf("Save (update): %v", err)
	}
	back, _ = repo.ByID(ctx, m.ID())
	if back.IsActive() || back.FreeShipping().Kind() != catalog.FreeShipAlways {
		t.Fatalf("update not persisted: %+v", back.Snapshot())
	}

	bySite, err := repo.BySite(ctx, catalog.MustParseHostname("www.example.com"))
	if err != nil || bySite.ID() != m.ID() {
		t.Fatalf("BySite = %v, %v", bySite, err)
	}
	if _, err := repo.ByID(ctx, catalog.NewMerchantID()); !errors.Is(err, catalog.ErrMerchantNotFound) {
		t.Errorf("unknown id: got %v", err)
	}
	if _, err := repo.BySite(ctx, catalog.MustParseHostname("nobody.example.com")); !errors.Is(err, catalog.ErrMerchantNotFound) {
		t.Errorf("unknown site: got %v", err)
	}
}

func TestCategoryRepo_roundTrip(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewCategoryRepo(p)
	c := aCategory(t)

	if err := repo.Save(ctx, c); err != nil {
		t.Fatalf("Save: %v", err)
	}
	back, err := repo.ByCode(ctx, c.Code())
	if err != nil || !reflect.DeepEqual(back, c) {
		t.Fatalf("ByCode = %+v, %v; want %+v", back, err, c)
	}
	// A value object is replaced whole.
	c2, _ := catalog.NewCategoryPolicy(c.Code(), shared.MustParcelSpec(shared.Grams(1300), shared.NewDimensionsCM(34, 23, 13)), nil)
	if err := repo.Save(ctx, c2); err != nil {
		t.Fatal(err)
	}
	back, _ = repo.ByCode(ctx, c.Code())
	if !reflect.DeepEqual(back, c2) {
		t.Fatalf("Save did not replace: %+v", back)
	}
	all, err := repo.All(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("All = %d, %v", len(all), err)
	}
	if _, err := repo.ByCode(ctx, catalog.MustParseCategoryCode("luggage")); !errors.Is(err, catalog.ErrCategoryNotFound) {
		t.Errorf("unknown code: got %v", err)
	}
}

func TestProductRepo_roundTrip(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	merchants, categories, products := postgres.NewMerchantRepo(p), postgres.NewCategoryRepo(p), postgres.NewProductRepo(p)
	m := aMerchant(t)
	if err := merchants.Save(ctx, m); err != nil {
		t.Fatal(err)
	}
	if err := categories.Save(ctx, aCategory(t)); err != nil {
		t.Fatal(err)
	}
	operator := shared.NewOperatorID()
	customer := catalog.MustProvenance(catalog.SourcedByCustomer, now, shared.OperatorID{})
	verified := catalog.MustProvenance(catalog.SourcedByOperator, now.Add(time.Minute), operator)
	prod, err := catalog.AddProduct(catalog.ProductDetails{
		Name: "Air Trainer 90", Merchant: m.ID(), Category: catalog.MustParseCategoryCode("footwear"),
		Source: catalog.MustParseSourceURL("https://www.example.com/t/air-trainer-90/abc"),
		Price:  shared.MustParseMoney("150.00", shared.USD), ListingProvenance: customer, PriceProvenance: customer,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	// A bare draft first: nil parcel, no variants.
	if err := products.Save(ctx, prod); err != nil {
		t.Fatalf("Save draft: %v", err)
	}
	back, err := products.ByID(ctx, prod.ID())
	if err != nil || !reflect.DeepEqual(back.Snapshot(), prod.Snapshot()) {
		t.Fatalf("draft round trip:\n got  %+v\n want %+v (%v)", back.Snapshot(), prod.Snapshot(), err)
	}

	// Then the full life: variants, measured, repriced, flagged, cleared, published.
	if _, err := prod.AddVariant(catalog.VariantDetails{Size: "US 9", Color: "black", MerchantRef: "DM-001"}, now); err != nil {
		t.Fatal(err)
	}
	if _, err := prod.AddVariant(catalog.VariantDetails{Size: "US 10", Color: "black"}, now); err != nil {
		t.Fatal(err)
	}
	if err := prod.ConfirmListing(verified); err != nil {
		t.Fatal(err)
	}
	if err := prod.Measure(shared.MustParcelSpec(shared.Grams(1250), shared.NewDimensionsCM(34, 23, 13)), verified, now); err != nil {
		t.Fatal(err)
	}
	if err := prod.Reprice(shared.MustParseMoney("160.00", shared.USD), verified, now); err != nil {
		t.Fatal(err)
	}
	other := catalog.NewProductID()
	if err := prod.FlagDuplicateOf(other, "same page url", now); err != nil {
		t.Fatal(err)
	}
	if err := prod.ClearDuplicateFlag("checked", now); err != nil {
		t.Fatal(err)
	}
	if err := prod.Publish(now); err != nil {
		t.Fatal(err)
	}
	prod.PullEvents()
	if err := products.Save(ctx, prod); err != nil {
		t.Fatalf("Save (update): %v", err)
	}
	back, err = products.ByID(ctx, prod.ID())
	if err != nil || !reflect.DeepEqual(back.Snapshot(), prod.Snapshot()) {
		t.Fatalf("full round trip:\n got  %+v\n want %+v (%v)", back.Snapshot(), prod.Snapshot(), err)
	}
	if vs := back.Variants(); len(vs) != 2 || vs[0].Size() != "US 9" || vs[1].Size() != "US 10" {
		t.Errorf("variant order lost: %+v", vs)
	}

	same, err := products.BySource(ctx, prod.Source())
	if err != nil || len(same) != 1 || same[0].ID() != prod.ID() {
		t.Fatalf("BySource = %v, %v", same, err)
	}
	if _, err := products.ByID(ctx, catalog.NewProductID()); !errors.Is(err, catalog.ErrProductNotFound) {
		t.Errorf("unknown id: got %v", err)
	}
}

// THE test that was skipped against memory (app/catalog): Save succeeds, the
// outbox fails, and the transaction must undo the save. Here it runs for real.
type failingOutbox struct{}

func (failingOutbox) Append(context.Context, []shared.Event) error {
	return errors.New("outbox table locked")
}

func TestUnitOfWork_rollsBackTheSaveWhenTheOutboxFails(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	merchants := postgres.NewMerchantRepo(p)
	h := catalogapp.NewRegisterMerchantHandler(catalogapp.Deps{
		Clock: clock.FixedAt(now), UoW: postgres.NewUnitOfWork(p),
		Merchants: merchants, Categories: postgres.NewCategoryRepo(p), Products: postgres.NewProductRepo(p),
		Outbox: failingOutbox{},
	})
	_, err := h.Handle(ctx, catalog.MerchantDetails{
		Name: "Example Sports", Site: catalog.MustParseHostname("www.example.com"), Currency: shared.USD,
	})
	if err == nil {
		t.Fatal("Handle must fail when the outbox does")
	}
	var n int
	if err := p.QueryRow(ctx, `SELECT count(*) FROM merchants`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("Save was not rolled back: %d merchant(s) remain (%v)", n, err)
	}
}

// And the happy path in the same transaction: merchant row AND outbox row, or
// neither. Plus the worker's half of the outbox: Pending in order, MarkSent.
func TestOutbox_appendedInTheSameTransactionAndReadBack(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	outbox := postgres.NewOutbox(p)
	var uow app.UnitOfWork = postgres.NewUnitOfWork(p)
	h := catalogapp.NewRegisterMerchantHandler(catalogapp.Deps{
		Clock: clock.FixedAt(now), UoW: uow,
		Merchants: postgres.NewMerchantRepo(p), Categories: postgres.NewCategoryRepo(p), Products: postgres.NewProductRepo(p),
		Outbox: outbox,
	})
	id, err := h.Handle(ctx, catalog.MerchantDetails{
		Name: "Example Sports", Site: catalog.MustParseHostname("www.example.com"), Currency: shared.USD,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	pending, err := outbox.Pending(ctx, 10)
	if err != nil || len(pending) != 1 {
		t.Fatalf("Pending = %d, %v", len(pending), err)
	}
	e := pending[0]
	if e.Name != "catalog.merchant_registered" || !e.OccurredAt.Equal(now) || e.ID == 0 {
		t.Errorf("entry = %+v", e)
	}
	var body map[string]any
	if err := json.Unmarshal(e.Payload, &body); err != nil || body["id"] != id.String() {
		t.Errorf("payload %s: id = %v, %v", e.Payload, body["id"], err)
	}

	if err := outbox.MarkSent(ctx, []int64{e.ID}, now.Add(time.Second)); err != nil {
		t.Fatalf("MarkSent: %v", err)
	}
	if again, _ := outbox.Pending(ctx, 10); len(again) != 0 {
		t.Fatalf("still pending after MarkSent: %d", len(again))
	}
}
