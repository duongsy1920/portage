package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"sync"
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

// Two processes starting at once against a COLD database — exactly what
// cmd/api and cmd/worker do in scripts/smoke.sh on a first deploy.
//
// This used to fail. The advisory lock guarded the loop that applies the .sql
// files, but NOT the CREATE TABLE that makes schema_migrations itself, and in
// PostgreSQL two concurrent CREATE TABLE IF NOT EXISTS both pass the existence
// check and then collide inserting the table's row type:
//
//	ERROR: duplicate key value violates unique constraint "pg_type_typname_nsp_index"
//
// The test needs a throwaway database, because "cold" has to mean cold and the
// shared test database must keep its schema. A race is probabilistic, so this
// can pass on buggy code now and then — but it can never fail on correct code,
// which is the direction that matters.
func TestMigrate_survivesTwoProcessesOnAColdDatabase(t *testing.T) {
	admin := pool(t) // holds the pgtest lock, so no other package is migrating
	ctx := context.Background()

	// Digits only, so the name needs no quoting.
	name := fmt.Sprintf("portage_migrate_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, `CREATE DATABASE `+name+` OWNER portage`); err != nil {
		t.Fatalf("create database: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(context.Background(), `DROP DATABASE IF EXISTS `+name+` WITH (FORCE)`)
	})

	u, err := url.Parse(pgtest.DSN(t))
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	u.Path = "/" + name
	cold := u.String()

	const starters = 4
	var (
		wg    sync.WaitGroup
		start = make(chan struct{})
		errs  = make([]error, starters)
	)
	for i := range starters {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p, err := postgres.Connect(ctx, cold)
			if err != nil {
				errs[i] = err
				return
			}
			defer p.Close()
			<-start // released together, so they really do collide
			errs[i] = postgres.Migrate(ctx, p)
		}(i)
	}
	close(start)
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("starter %d: %v", i, err)
		}
	}

	// And the schema is applied exactly once, not four times.
	p, err := postgres.Connect(ctx, cold)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer p.Close()
	var versions, distinct int
	if err := p.QueryRow(ctx, `SELECT count(*), count(DISTINCT version) FROM schema_migrations`).Scan(&versions, &distinct); err != nil {
		t.Fatalf("count: %v", err)
	}
	if versions == 0 || versions != distinct {
		t.Fatalf("schema_migrations = %d rows, %d distinct", versions, distinct)
	}
}

// All is the read side of GET /merchants, and it has to come back in a stable
// order or a select box reshuffles under the person using it. Ids are UUIDv7,
// so ORDER BY id is order of creation — and the in-memory adapter sorts the
// same way, which is what keeps the two interchangeable.
func TestMerchantRepo_allInCreationOrder(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewMerchantRepo(p)

	if all, err := repo.All(ctx); err != nil || len(all) != 0 {
		t.Fatalf("empty database = %d rows, %v", len(all), err)
	}
	// Two DIFFERENT hostnames: merchants.site is UNIQUE, because one shop is
	// one place we buy from and two rows for one hostname would make
	// "who owns this page url" ambiguous.
	first := aMerchant(t)
	second, err := catalog.RegisterMerchant(catalog.MerchantDetails{
		Name: "Other Shop", Site: catalog.MustParseHostname("www.other-shop.com"), Currency: shared.USD,
		Sourcing: []catalog.SourcingMode{catalog.SourcedByOperator},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	second.PullEvents()
	for _, m := range []*catalog.Merchant{first, second} {
		if err := repo.Save(ctx, m); err != nil {
			t.Fatal(err)
		}
	}
	all, err := repo.All(ctx)
	if err != nil || len(all) != 2 {
		t.Fatalf("All = %d rows, %v", len(all), err)
	}
	if all[0].ID() != first.ID() || all[1].ID() != second.ID() {
		t.Fatalf("order = %s, %s; want %s, %s", all[0].ID(), all[1].ID(), first.ID(), second.ID())
	}
	// Same rebuild path as ByID: every column back through the domain's parsers.
	if !reflect.DeepEqual(all[0].Snapshot(), first.Snapshot()) {
		t.Fatal("All must rebuild the aggregate the same way ByID does")
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
	// RequestedBy is set here on purpose: it is the one nullable id on this
	// table, and a round trip that leaves it zero would pass while the column
	// did not exist at all.
	waiting := shared.NewID()
	prod, err := catalog.AddProduct(catalog.ProductDetails{
		Name: "Air Trainer 90", Merchant: m.ID(), Category: catalog.MustParseCategoryCode("footwear"),
		Source: catalog.MustParseSourceURL("https://www.example.com/t/air-trainer-90/abc"),
		Price:  shared.MustParseMoney("150.00", shared.USD), ListingProvenance: customer, PriceProvenance: customer,
		RequestedBy: waiting,
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
	if err := prod.ConfirmListing(verified, now); err != nil {
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
