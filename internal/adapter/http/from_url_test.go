package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	httpapi "github.com/duongsy/portage/internal/adapter/http"
	"net/http"
	"testing"

	"github.com/duongsy/portage/internal/adapter/openai"
	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
)

func fromURLBody(merchant string) string {
	return `{"merchant_id":"` + merchant + `","url":"https://www.example.com/t/air-trainer-90/xyz","category":"footwear"}`
}

// A machine may CREATE a draft; it may not vouch for one. The listing comes
// back with provenance "feed" and unverified, so Publish still refuses it
// until an operator has confirmed it by hand.
func TestPOSTProductsFromURL_recordsAnUnverifiedDraft(t *testing.T) {
	a := newAPI()
	m := a.seedMerchant(t)
	a.seedCategory(t)

	rec := a.call(t, "POST", "/products/from-url", fromURLBody(m.ID().String()), a.asCustomer())
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d %s", rec.Code, rec.Body)
	}
	var out struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Price        string `json:"price"`
		Currency     string `json:"currency"`
		CategoryHint string `json:"category_hint"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if out.Name != "Air Trainer 90" || out.Price != "150.00" || out.Currency != "USD" || out.CategoryHint != "footwear" {
		t.Fatalf("response = %+v", out)
	}

	id, _ := catalog.ParseProductID(out.ID)
	p, err := a.products.ByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	prov := p.ListingProvenance()
	if prov.Verified() {
		t.Error("a machine must not be able to vouch for a listing")
	}
	if prov.Source() != catalog.SourcedByFeed {
		t.Errorf("source = %s, want %s", prov.Source(), catalog.SourcedByFeed)
	}
	if p.Price().String() != "150.00 USD" || p.Name() != "Air Trainer 90" {
		t.Errorf("product = %s / %s", p.Name(), p.Price())
	}

	// The hint never picks the category: it decides duty and estimated weight,
	// so a body without one is refused rather than guessed.
	expectError(t, a.call(t, "POST", "/products/from-url",
		`{"merchant_id":"`+m.ID().String()+`","url":"https://www.example.com/t/x","category":""}`, a.asCustomer()),
		http.StatusBadRequest, "invalid_category_code")
}

// No extractor configured is 503, never a silent fallback: the API would
// otherwise answer 201 with an invented name and price, and the only sign
// anything was wrong is that the numbers are fiction.
func TestPOSTProductsFromURL_withoutAnExtractorIs503(t *testing.T) {
	a := newAPIWith(func(d *httpapi.Deps) { d.Extractor = nil })
	m := a.seedMerchant(t)
	a.seedCategory(t)

	expectError(t, a.call(t, "POST", "/products/from-url", fromURLBody(m.ID().String()), a.asCustomer()),
		http.StatusServiceUnavailable, "extractor_unavailable")
}

// A model that is down, rate-limited or talking nonsense is the same answer.
func TestPOSTProductsFromURL_failingExtractorIs503(t *testing.T) {
	a := newAPIWith(func(d *httpapi.Deps) {
		d.Extractor = openai.Fake{Err: errors.New("openai: http 429: " + catalogapp.ErrExtractorUnavailable.Error())}
	})
	m := a.seedMerchant(t)
	a.seedCategory(t)

	// A plain error with no sentinel is a 500 — that is the errorTable working
	// as designed, and the reason the adapter wraps every failure.
	res := a.call(t, "POST", "/products/from-url", fromURLBody(m.ID().String()), a.asCustomer())
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("an unwrapped error must be 500, got %d %s", res.Code, res.Body)
	}

	a2 := newAPIWith(func(d *httpapi.Deps) { d.Extractor = openai.Unavailable{} })
	m2 := a2.seedMerchant(t)
	a2.seedCategory(t)
	expectError(t, a2.call(t, "POST", "/products/from-url", fromURLBody(m2.ID().String()), a2.asCustomer()),
		http.StatusServiceUnavailable, "extractor_unavailable")
}

// The model's strings go through the DOMAIN's own Parse*, the same door a
// human's typing goes through — so "about $150" fails here, not three tables
// later.
func TestPOSTProductsFromURL_refusesAModelsBadNumbers(t *testing.T) {
	a := newAPIWith(func(d *httpapi.Deps) {
		d.Extractor = openai.Fake{Draft: catalogapp.ListingDraft{Name: "Air Trainer 90", Price: "about 150", Currency: "USD"}}
	})
	m := a.seedMerchant(t)
	a.seedCategory(t)

	expectError(t, a.call(t, "POST", "/products/from-url", fromURLBody(m.ID().String()), a.asCustomer()),
		http.StatusBadRequest, "malformed_amount")

	a2 := newAPIWith(func(d *httpapi.Deps) {
		d.Extractor = openai.Fake{Draft: catalogapp.ListingDraft{Name: "Air Trainer 90", Price: "150.00", Currency: "XYZ"}}
	})
	m2 := a2.seedMerchant(t)
	a2.seedCategory(t)
	expectError(t, a2.call(t, "POST", "/products/from-url", fromURLBody(m2.ID().String()), a2.asCustomer()),
		http.StatusBadRequest, "unknown_currency")
}
