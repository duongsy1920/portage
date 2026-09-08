package memory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/memory"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

var now = time.Date(2026, 9, 4, 15, 0, 0, 0, time.UTC)

func aMerchant(t *testing.T) *catalog.Merchant {
	t.Helper()
	m, err := catalog.RegisterMerchant(catalog.MerchantDetails{
		Name:     "Example Sports",
		Site:     catalog.MustParseHostname("www.example.com"),
		Currency: shared.USD,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// A repository answers "not there" with the DOMAIN's sentinel, wrapped, so a
// handler can errors.Is it without knowing which adapter is wired.
func TestRepos_notFoundUsesDomainSentinels(t *testing.T) {
	ctx := context.Background()
	if _, err := memory.NewMerchantRepo().ByID(ctx, catalog.NewMerchantID()); !errors.Is(err, catalog.ErrMerchantNotFound) {
		t.Errorf("merchant: got %v", err)
	}
	if _, err := memory.NewMerchantRepo().BySite(ctx, catalog.MustParseHostname("nobody.example.com")); !errors.Is(err, catalog.ErrMerchantNotFound) {
		t.Errorf("merchant by site: got %v", err)
	}
	if _, err := memory.NewCategoryRepo().ByCode(ctx, catalog.MustParseCategoryCode("luggage")); !errors.Is(err, catalog.ErrCategoryNotFound) {
		t.Errorf("category: got %v", err)
	}
	if _, err := memory.NewProductRepo().ByID(ctx, catalog.NewProductID()); !errors.Is(err, catalog.ErrProductNotFound) {
		t.Errorf("product: got %v", err)
	}
}

func TestMerchantRepo_saveThenFindByIDAndSite(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewMerchantRepo()
	m := aMerchant(t)
	if err := repo.Save(ctx, m); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.ByID(ctx, m.ID()); err != nil || got.ID() != m.ID() {
		t.Errorf("ByID = %v, %v", got, err)
	}
	if got, err := repo.BySite(ctx, catalog.MustParseHostname("WWW.example.com")); err != nil || got.ID() != m.ID() {
		t.Errorf("BySite (case-insensitive host) = %v, %v", got, err)
	}
	if repo.Len() != 1 {
		t.Errorf("Len() = %d", repo.Len())
	}
}

// BySource is an exact match on the stored URL — the cheap first question of
// duplicate detection. Two different pages are two different products.
func TestProductRepo_bySourceIsExact(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewProductRepo()
	m := aMerchant(t)
	prov := catalog.MustProvenance(catalog.SourcedByCustomer, now, shared.OperatorID{})
	add := func(url string) catalog.ProductID {
		p, err := catalog.AddProduct(catalog.ProductDetails{
			Name: "Thing", Merchant: m.ID(), Category: catalog.MustParseCategoryCode("footwear"),
			Source: catalog.MustParseSourceURL(url), Price: shared.MustParseMoney("10.00", shared.USD),
			ListingProvenance: prov, PriceProvenance: prov,
		}, now)
		if err != nil {
			t.Fatal(err)
		}
		if err := repo.Save(ctx, p); err != nil {
			t.Fatal(err)
		}
		return p.ID()
	}
	a := add("https://www.example.com/a")
	add("https://www.example.com/b")

	got, err := repo.BySource(ctx, catalog.MustParseSourceURL("https://www.example.com/a"))
	if err != nil || len(got) != 1 || got[0].ID() != a {
		t.Fatalf("BySource = %v, %v", got, err)
	}
	if got, _ := repo.BySource(ctx, catalog.MustParseSourceURL("https://www.example.com/c")); len(got) != 0 {
		t.Errorf("unknown page returned %d products", len(got))
	}
}

func TestOutbox_drainEmptiesAndFailWithFails(t *testing.T) {
	ctx := context.Background()
	o := memory.NewOutbox()
	m := aMerchant(t)
	if err := o.Append(ctx, m.PullEvents()); err != nil {
		t.Fatal(err)
	}
	if evs := o.Drain(); len(evs) != 1 {
		t.Fatalf("Drain = %d events, want 1", len(evs))
	}
	if evs := o.Drain(); len(evs) != 0 {
		t.Fatalf("second Drain = %d events, want 0", len(evs))
	}
	boom := errors.New("outbox table locked")
	o.FailWith = boom
	if err := o.Append(ctx, nil); !errors.Is(err, boom) {
		t.Errorf("FailWith: got %v", err)
	}
}

// UnitOfWork has nothing to open; it must still run fn and return its error.
func TestUnitOfWork_runsFn(t *testing.T) {
	boom := errors.New("inside")
	err := memory.UnitOfWork{}.InTx(context.Background(), func(ctx context.Context) error {
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("got %v", err)
	}
}
