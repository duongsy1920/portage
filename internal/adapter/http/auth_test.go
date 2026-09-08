package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Every route sits behind the middleware, which wraps the whole mux. A route
// added tomorrow is protected before anyone remembers to protect it, and this
// test is what says so.
func TestAuth_noTokenIsAlways401(t *testing.T) {
	a := newAPI()

	for _, c := range []struct {
		method string
		path   string
	}{
		{"POST", "/categories"},
		{"POST", "/merchants"},
		{"POST", "/products"},
		{"POST", "/quotes"},
		{"POST", "/orders"},
		{"GET", "/purchase-tasks"},
		{"GET", "/parcels"},
		{"POST", "/batches"},
	} {
		res := a.callAnon(t, c.method, c.path, `{}`, nil)
		expectError(t, res, http.StatusUnauthorized, "unauthenticated")
	}
}

// Unknown, empty and wrong-case tokens are all one answer. A caller must not
// be able to tell a real token from a fake one by the reply.
func TestAuth_unknownTokenIs401(t *testing.T) {
	a := newAPI()

	for _, token := range []string{"nope", "", "TEST-OPERATOR", " test-operator"} {
		res := a.callAnon(t, "POST", "/categories", `{}`, map[string]string{
			"Authorization": "Bearer " + token,
		})
		expectError(t, res, http.StatusUnauthorized, "unauthenticated")
	}
}

// A malformed Authorization header is not a token: a missing scheme, another
// scheme, or a scheme with nothing after it are refused the same way.
func TestAuth_malformedHeaderIs401(t *testing.T) {
	a := newAPI()

	for _, header := range []string{"test-operator", "Basic test-operator", "Bearer", "Bearertest-operator"} {
		res := a.callAnon(t, "POST", "/categories", `{}`, map[string]string{"Authorization": header})
		expectError(t, res, http.StatusUnauthorized, "unauthenticated")
	}
}

// 403 is a different conversation from 401: your token works, this route is
// not yours.
func TestAuth_wrongKindIs403(t *testing.T) {
	a := newAPI()

	for _, c := range []struct {
		method string
		path   string
	}{
		{"POST", "/categories"},
		{"POST", "/merchants"},
		{"POST", "/lanes"},
		{"GET", "/purchase-tasks"},
		{"GET", "/parcels"},
		{"POST", "/batches"},
	} {
		res := a.call(t, c.method, c.path, `{}`, a.asCustomer())
		expectError(t, res, http.StatusForbidden, "forbidden")
	}
}

// POST /orders belongs to the customer. With customer_id gone from the body
// there is no field left for staff to order in somebody else's name, so the
// route refuses an operator rather than guessing whose order it would be.
func TestAuth_customerOnlyRoutesRefuseOperator(t *testing.T) {
	a := newAPI()

	expectError(t, a.call(t, "POST", "/orders", `{}`, a.asOperator()),
		http.StatusForbidden, "forbidden")
	expectError(t, a.call(t, "POST", "/quotes/"+shared.NewID().String()+"/accept", `{}`, a.asOperator()),
		http.StatusForbidden, "forbidden")
}

// Asking for a price is not staff-only: both kinds get past the door, and
// whatever happens next is the route's own business.
func TestAuth_sharedRoutesAcceptBothKinds(t *testing.T) {
	a := newAPI()

	for _, headers := range []map[string]string{a.asOperator(), a.asCustomer()} {
		res := a.call(t, "POST", "/quotes", `{}`, headers)
		if res.Code == http.StatusUnauthorized || res.Code == http.StatusForbidden {
			t.Fatalf("POST /quotes must let both kinds through, got %d: %s", res.Code, res.Body)
		}
	}
}

// The operator's identity is the TOKEN, not a header the caller writes. A
// stranger's X-Operator-ID must change nothing about who vouched for a
// listing — that is the whole reason the header was removed.
func TestAuth_provenanceComesFromTheTokenNotAHeader(t *testing.T) {
	a := newAPI()
	m := a.seedMerchant(t)
	a.seedCategory(t)
	stranger := shared.NewOperatorID()

	headers := a.asOperator()
	headers["X-Operator-ID"] = stranger.String()

	rec := a.call(t, "POST", "/products", productJSON(m.ID().String()), headers)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	id, _ := catalog.ParseProductID(idOf(t, rec))
	p, _ := a.products.ByID(context.Background(), id)

	prov := p.ListingProvenance()
	if prov.By() == stranger {
		t.Fatal("the header decided who vouched for the listing — it must not be read at all")
	}
	if !prov.Verified() || prov.By() != a.operatorID {
		t.Errorf("provenance = %v, want verified by the token's operator %s", prov, a.operatorID)
	}
}

// A customer's paste is a draft nobody has checked: unverified provenance and
// no operator, decided by the token rather than a sourced_by field the caller
// could have set to "operator".
func TestAuth_customerPasteIsUnverified(t *testing.T) {
	a := newAPI()
	m := a.seedMerchant(t)
	a.seedCategory(t)

	rec := a.call(t, "POST", "/products", productJSON(m.ID().String()), a.asCustomer())
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	id, _ := catalog.ParseProductID(idOf(t, rec))
	p, _ := a.products.ByID(context.Background(), id)

	prov := p.ListingProvenance()
	if prov.Verified() {
		t.Error("a customer's paste must not be verified")
	}
	if prov.Source() != catalog.SourcedByCustomer {
		t.Errorf("source = %s, want %s", prov.Source(), catalog.SourcedByCustomer)
	}
}

// POST /tokens is the only route that hands back a secret, and only an
// operator may ask. A customer minting themselves an operator key would be
// the whole authorisation model undone in one request.
func TestTokens_operatorIssuesACustomerKey(t *testing.T) {
	a := newAPI()

	expectError(t, a.call(t, "POST", "/tokens", `{"kind":"customer"}`, a.asCustomer()),
		http.StatusForbidden, "forbidden")
	expectError(t, a.callAnon(t, "POST", "/tokens", `{"kind":"customer"}`, nil),
		http.StatusUnauthorized, "unauthenticated")

	rec := a.call(t, "POST", "/tokens", `{"kind":"customer","label":"phone"}`, a.asOperator())
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	var out struct {
		Token, Kind, Subject string
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Kind != "customer" || out.Token == "" || out.Subject == "" {
		t.Fatalf("response = %+v", out)
	}

	// The key works immediately, and it is a CUSTOMER key: it may place an
	// order and may not define a category.
	fresh := map[string]string{"Authorization": "Bearer " + out.Token}
	expectError(t, a.call(t, "POST", "/categories", `{}`, fresh), http.StatusForbidden, "forbidden")
	if res := a.call(t, "POST", "/orders", `{}`, fresh); res.Code == http.StatusForbidden || res.Code == http.StatusUnauthorized {
		t.Fatalf("a fresh customer key must pass the door for POST /orders, got %d", res.Code)
	}
}

// An unknown kind is refused by the domain's own constructor, not by a list
// the adapter keeps and forgets to update.
func TestTokens_rejectsUnknownKind(t *testing.T) {
	a := newAPI()

	for _, body := range []string{`{"kind":"admin"}`, `{"kind":""}`, `{"kind":"Operator"}`} {
		res := a.call(t, "POST", "/tokens", body, a.asOperator())
		if res.Code != http.StatusBadRequest {
			t.Errorf("%s = %d, want 400: %s", body, res.Code, res.Body)
		}
	}
}
