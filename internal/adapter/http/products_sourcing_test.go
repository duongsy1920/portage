package httpapi_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Merchant.Sourcing says WHO may bring a product in from this shop, and it was
// recorded, announced through its own events, and never checked. This is the
// test that makes it a rule instead of a field.
//
// The screen filters the shop list too, but a filtered list is not enforcement:
// curl does not read the screen.
func TestPOSTProducts_refusesASourceTheShopDoesNotAccept(t *testing.T) {
	a := newAPI()
	a.seedCategory(t)

	// A shop staff buy from by hand: customers may not paste links for it.
	staffOnly, err := catalog.RegisterMerchant(catalog.MerchantDetails{
		Name: "Trade Only Supply", Site: catalog.MustParseHostname("www.trade-only.com"),
		Currency: shared.USD,
		Sourcing: []catalog.SourcingMode{catalog.SourcedByOperator},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	staffOnly.PullEvents()
	if err := a.merchants.Save(context.Background(), staffOnly); err != nil {
		t.Fatal(err)
	}

	body := func(id string) string {
		return `{"name":"Air Trainer 90","merchant_id":"` + id + `","category":"footwear",
			"source_url":"https://www.trade-only.com/t/x","price":"150.00","currency":"USD"}`
	}

	// The customer is refused by STATE, not by the door: their token is fine,
	// this shop just is not one they may paste for.
	expectError(t, a.call(t, "POST", "/products", body(staffOnly.ID().String()), a.asCustomer()),
		409, "sourcing_not_allowed")

	// The same request from staff goes through.
	if rec := a.call(t, "POST", "/products", body(staffOnly.ID().String()), nil); rec.Code != http.StatusCreated {
		t.Fatalf("operator paste = %d %s", rec.Code, rec.Body)
	}

	// And the AI route is a third source, refused for the same reason.
	fromURL := `{"merchant_id":"` + staffOnly.ID().String() + `","url":"https://www.trade-only.com/t/y","category":"footwear"}`
	expectError(t, a.call(t, "POST", "/products/from-url", fromURL, nil), 409, "sourcing_not_allowed")
}
