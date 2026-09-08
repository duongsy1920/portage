package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// The two lists a form needs so a person picks instead of typing a uuid. Both
// are open to any authenticated caller, because a CUSTOMER pasting a link has
// to choose a shop and a category too.
func TestReference_listsTheTwoClosedSets(t *testing.T) {
	a := newAPI()
	m := a.seedMerchant(t)
	a.seedCategory(t)

	var cats []struct {
		Code         string   `json:"code"`
		WeightG      int64    `json:"weight_g"`
		LengthMM     int64    `json:"length_mm"`
		Restrictions []string `json:"restrictions"`
	}
	rec := a.call(t, "GET", "/categories", "", a.asCustomer())
	if err := json.Unmarshal(rec.Body.Bytes(), &cats); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("GET /categories = %d %s (%v)", rec.Code, rec.Body, err)
	}
	if len(cats) != 1 || cats[0].Code != "footwear" {
		t.Fatalf("categories = %+v", cats)
	}
	// The default box travels with the code: it is what prices a product
	// nobody has weighed yet, so a screen can show what it is quoting on.
	if cats[0].WeightG != 1200 || cats[0].LengthMM != 330 {
		t.Fatalf("default parcel spec = %+v", cats[0])
	}

	var shops []struct {
		ID       string   `json:"id"`
		Name     string   `json:"name"`
		Site     string   `json:"site"`
		Currency string   `json:"currency"`
		Sourcing []string `json:"sourcing"`
		Status   string   `json:"status"`
	}
	rec = a.call(t, "GET", "/merchants", "", a.asCustomer())
	if err := json.Unmarshal(rec.Body.Bytes(), &shops); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("GET /merchants = %d %s (%v)", rec.Code, rec.Body, err)
	}
	if len(shops) != 1 || shops[0].ID != m.ID().String() || shops[0].Site != "www.example.com" {
		t.Fatalf("merchants = %+v", shops)
	}
	// Currency is on the row so the form can pin it once a shop is chosen,
	// instead of letting somebody pick a mismatch and learn from a 409.
	if shops[0].Currency != "USD" || shops[0].Status != "active" {
		t.Fatalf("merchant row = %+v", shops[0])
	}
	// Sourcing says who may bring a product in from this shop, and the list
	// carries it so a form can offer only the shops this caller may paste for.
	// It is not a filter the server relies on: POST /products enforces the
	// same rule and answers 409 sourcing_not_allowed.
	if len(shops[0].Sourcing) != 3 {
		t.Fatalf("sourcing = %+v", shops[0].Sourcing)
	}
}

// Reference data is still behind the door: a list of the shops we buy from is
// not public just because it has no invariant to protect.
func TestReference_needsAToken(t *testing.T) {
	a := newAPI()
	for _, path := range []string{"/categories", "/merchants"} {
		if rec := a.callAnon(t, "GET", path, "", nil); rec.Code != http.StatusUnauthorized {
			t.Fatalf("GET %s with no token = %d, want 401", path, rec.Code)
		}
	}
}
