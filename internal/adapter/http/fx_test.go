package httpapi_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/duongsy/portage/internal/domain/shared"
)

// POST /fx publishes today's rate. Operator only, no GET twin, and — the part
// worth remembering — no event: nobody keeps a copy of the rate, because a
// quote froze the one it used (DDD.md §28).
func TestPOSTFx_publishesTodaysRate(t *testing.T) {
	a := newAPI()
	ctx := context.Background()
	a.outbox.Drain()

	if rec := a.call(t, "POST", "/fx", `{"from":"USD","to":"VND","rate":"27000.5"}`, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("set rate = %d %s", rec.Code, rec.Body)
	}
	got, err := a.rates.Current(ctx, shared.USD, shared.VND)
	if err != nil || got.Rate() != "27000.5" {
		t.Fatalf("rate = %v, %v", got, err)
	}
	if evs := a.outbox.Drain(); len(evs) != 0 {
		t.Fatalf("setting a rate announced %v — nobody keeps a copy of it", evs)
	}

	// Publishing again replaces: "today's rate" has one answer at a time.
	a.call(t, "POST", "/fx", `{"from":"USD","to":"VND","rate":"26000"}`, nil)
	if got, _ = a.rates.Current(ctx, shared.USD, shared.VND); got.Rate() != "26000" {
		t.Errorf("rate after the second publish = %s", got.Rate())
	}
}

// A rate is a ratio typed by a human at the edge, so it gets the same locale
// treatment every number does — and the same domain validation.
func TestPOSTFx_localeAndRefusals(t *testing.T) {
	a := newAPI()

	rec := a.call(t, "POST", "/fx", `{"from":"USD","to":"VND","rate":"26.000,5"}`,
		map[string]string{"Accept-Language": "vi", "Authorization": "Bearer " + devOperatorToken})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("vi rate = %d %s", rec.Code, rec.Body)
	}
	if got, _ := a.rates.Current(context.Background(), shared.USD, shared.VND); got.Rate() != "26000.5" {
		t.Errorf("vi rate stored as %s, want 26000.5", got.Rate())
	}

	for _, c := range []struct{ body, code string }{
		{`{"from":"USD","to":"USD","rate":"1"}`, "invalid_exchange_rate"},  // same currency
		{`{"from":"USD","to":"VND","rate":"0"}`, "invalid_exchange_rate"},  // not positive
		{`{"from":"USD","to":"VND","rate":"-1"}`, "invalid_exchange_rate"}, // not positive
		{`{"from":"XYZ","to":"VND","rate":"1"}`, "unknown_currency"},
		{`{"from":"USD","to":"VND","rate":"x"}`, "malformed_amount"},
	} {
		expectError(t, a.call(t, "POST", "/fx", c.body, nil), http.StatusBadRequest, c.code)
	}
	expectError(t, a.call(t, "POST", "/fx", `{"from":"USD","to":"VND","rate":"26000"}`, a.asCustomer()),
		http.StatusForbidden, "forbidden")
}
