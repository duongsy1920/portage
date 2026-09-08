package httpapi_test

import (
	"net/http"
	"testing"

	"github.com/duongsy/portage/internal/domain/shared"
)

// The whole operator path over HTTP, the way a curl session goes: paste →
// variant → confirm → measure → publish. Every refusal on the way is a
// status code the client can act on, not a 500.
func TestOperatorFlow_overHTTP(t *testing.T) {
	a := newAPI()
	m := a.seedMerchant(t)
	a.seedCategory(t)
	id := idOf(t, a.call(t, "POST", "/products", productJSON(m.ID().String()), nil))
	operator := map[string]string{"X-Operator-ID": shared.NewOperatorID().String()}
	a.outbox.Drain()

	// variants
	expectError(t, a.call(t, "POST", "/products/nope/variants", `{"size":"US 9"}`, nil), 400, "invalid_id")
	expectError(t, a.call(t, "POST", "/products/"+id+"/variants", `{"size":"US 9","colour":"black"}`, nil), 400, "bad_json")
	vid := idOf(t, a.call(t, "POST", "/products/"+id+"/variants", `{"size":"US 9","color":"black"}`, nil))
	if vid == "" {
		t.Fatal("no variant id")
	}
	expectError(t, a.call(t, "POST", "/products/"+id+"/variants", `{"size":"us 9","color":"BLACK"}`, nil), 409, "duplicate_variant")
	expectError(t, a.call(t, "POST", "/products/"+id+"/publish", "", nil), 409, "unverified")

	// confirm the listing — operator only
	expectError(t, a.call(t, "POST", "/products/"+id+"/confirm-listing", "", a.asCustomer()), 403, "forbidden")
	if rec := a.call(t, "POST", "/products/"+id+"/confirm-listing", "", operator); rec.Code != http.StatusNoContent {
		t.Fatalf("confirm = %d %s", rec.Code, rec.Body)
	}

	// measure — operator only, integers in g/mm
	body := `{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}`
	expectError(t, a.call(t, "POST", "/products/"+id+"/measure", body, a.asCustomer()), 403, "forbidden")
	expectError(t, a.call(t, "POST", "/products/"+id+"/measure", `{"weight_g":-1,"length_mm":1,"width_mm":1,"height_mm":1}`, operator), 400, "negative_weight")
	expectError(t, a.call(t, "POST", "/products/"+id+"/measure", `{"weight_g":1250}`, operator), 400, "incomplete_parcel_spec")
	if rec := a.call(t, "POST", "/products/"+id+"/measure", body, operator); rec.Code != http.StatusNoContent {
		t.Fatalf("measure = %d %s", rec.Code, rec.Body)
	}
	// Two events, in the order they were decided: adding the size is a fact
	// ordering and procurement both need (it is what lets them refuse an
	// invented variant, and name a size on the buyer's screen), and the
	// measurement is the fact pricing needs.
	if evs := a.outbox.Drain(); len(evs) != 2 || evs[0].EventName() != "catalog.variant_added" || evs[1].EventName() != "catalog.product_measured" {
		t.Fatalf("outbox after measure = %v", evs)
	}

	if rec := a.call(t, "POST", "/products/"+id+"/publish", "", nil); rec.Code != http.StatusNoContent {
		t.Fatalf("publish after the operator steps = %d %s", rec.Code, rec.Body)
	}
	if evs := a.outbox.Drain(); len(evs) != 1 || evs[0].EventName() != "catalog.product_published" {
		t.Fatalf("outbox after publish = %v", evs)
	}
}
