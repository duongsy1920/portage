package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

type keyJSON struct {
	Hash    string `json:"hash"`
	Kind    string `json:"kind"`
	Subject string `json:"subject"`
	Label   string `json:"label"`
	Active  bool   `json:"active"`
}

func keysOf(t *testing.T, body []byte) []keyJSON {
	t.Helper()
	var out []keyJSON
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode keys: %v — %s", err, body)
	}
	return out
}

// GET /tokens is the list of everyone who can call the API. It never contains
// a token: the store keeps a hash, so the hash is the only handle a key has
// after the one time its token was shown.
func TestGETTokens_listsKeysWithoutEverShowingOne(t *testing.T) {
	a := newAPI()

	rec := a.call(t, "POST", "/tokens", `{"kind":"customer","label":"phone"}`, a.asOperator())
	if rec.Code != http.StatusCreated {
		t.Fatalf("issue = %d %s", rec.Code, rec.Body)
	}
	var issued struct{ Token, Subject string }
	_ = json.Unmarshal(rec.Body.Bytes(), &issued)

	body := a.call(t, "GET", "/tokens", "", nil).Body
	if raw := body.String(); issued.Token != "" && strings.Contains(raw, issued.Token) {
		t.Fatal("the token itself appeared in the listing")
	}
	keys := keysOf(t, body.Bytes())
	if len(keys) != 3 { // the two dev keys plus the one just issued
		t.Fatalf("got %d keys, want 3: %v", len(keys), keys)
	}
	var phone *keyJSON
	for i, k := range keys {
		if k.Label == "phone" {
			phone = &keys[i]
		}
		if len(k.Hash) != 64 {
			t.Errorf("hash %q is not a sha256 hex", k.Hash)
		}
	}
	if phone == nil || phone.Kind != "customer" || phone.Subject != issued.Subject || !phone.Active {
		t.Fatalf("the issued key is not in the list correctly: %+v", phone)
	}

	expectError(t, a.call(t, "GET", "/tokens", "", a.asCustomer()), http.StatusForbidden, "forbidden")
	expectError(t, a.callAnon(t, "GET", "/tokens", "", nil), http.StatusUnauthorized, "unauthenticated")
}

// Revoking is by HASH, and it takes effect immediately.
func TestDELETETokens_revokesByHashAndIsIdempotent(t *testing.T) {
	a := newAPI()
	rec := a.call(t, "POST", "/tokens", `{"kind":"customer","label":"phone"}`, a.asOperator())
	var issued struct{ Token string }
	_ = json.Unmarshal(rec.Body.Bytes(), &issued)
	fresh := map[string]string{"Authorization": "Bearer " + issued.Token}

	// It works before the revoke...
	if res := a.call(t, "POST", "/orders", `{}`, fresh); res.Code == http.StatusUnauthorized {
		t.Fatal("a fresh key must work")
	}
	var hash string
	for _, k := range keysOf(t, a.call(t, "GET", "/tokens", "", nil).Body.Bytes()) {
		if k.Label == "phone" {
			hash = k.Hash
		}
	}
	if hash == "" {
		t.Fatal("the issued key is not listed")
	}

	if res := a.call(t, "DELETE", "/tokens/"+hash, "", nil); res.Code != http.StatusNoContent {
		t.Fatalf("revoke = %d %s", res.Code, res.Body)
	}
	// ...and not after.
	expectError(t, a.callAnon(t, "POST", "/orders", `{}`, fresh), http.StatusUnauthorized, "unauthenticated")

	// Revoking again, and revoking a hash nobody has, are both 204. A 404
	// would turn this route into a way to ask "is this hash real".
	for _, h := range []string{hash, "0000000000000000000000000000000000000000000000000000000000000000"} {
		if res := a.call(t, "DELETE", "/tokens/"+h, "", nil); res.Code != http.StatusNoContent {
			t.Errorf("second revoke of %s = %d", h[:8], res.Code)
		}
	}

	expectError(t, a.call(t, "DELETE", "/tokens/"+hash, "", a.asCustomer()), http.StatusForbidden, "forbidden")
}

// An API nobody can administer is not a safer API. The last active operator
// key is the one thing this route refuses — issue the replacement first.
func TestDELETETokens_refusesTheLastOperatorKey(t *testing.T) {
	a := newAPI()
	var mine string
	for _, k := range keysOf(t, a.call(t, "GET", "/tokens", "", nil).Body.Bytes()) {
		if k.Kind == "operator" {
			mine = k.Hash
		}
	}
	expectError(t, a.call(t, "DELETE", "/tokens/"+mine, "", nil), http.StatusConflict, "last_operator_key")

	// With a second operator key it is allowed: the API stays administrable.
	if rec := a.call(t, "POST", "/tokens", `{"kind":"operator","label":"laptop"}`, a.asOperator()); rec.Code != http.StatusCreated {
		t.Fatalf("issue second operator = %d %s", rec.Code, rec.Body)
	}
	if res := a.call(t, "DELETE", "/tokens/"+mine, "", a.asOperator()); res.Code != http.StatusNoContent {
		t.Fatalf("revoke with a spare = %d %s", res.Code, res.Body)
	}
}
