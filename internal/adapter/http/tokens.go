package httpapi

import (
	"net/http"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/auth"
)

// POST /tokens — an operator cuts a key for somebody.
//
// This is the only route that hands back a secret, and it hands it back ONCE:
// the store keeps a hash, so a token that is lost is re-issued, never
// recovered. That is a property of the design, not an omission.

type issueTokenRequest struct {
	// "operator" or "customer". auth.NewPrincipal validates it; the adapter
	// does not keep its own list of kinds to fall out of step with.
	Kind string `json:"kind"`

	// What the key is for, so a human can revoke the right one later.
	Label string `json:"label"`

	// Subject is the person this key belongs to. Optional: left out, a NEW id
	// is minted, which is what "sign up a customer" means today. Given, it
	// adds a second key for someone who already exists — a replacement phone,
	// a second operator terminal.
	Subject string `json:"subject"`
}

// issueTokenResponse is deliberately not idResponse: the token is the point,
// and this is the only moment it exists outside the caller's hands.
type issueTokenResponse struct {
	Token   string `json:"token"`
	Kind    string `json:"kind"`
	Subject string `json:"subject"`
}

func (s *server) issueToken(w http.ResponseWriter, r *http.Request) {
	var req issueTokenRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	subject := shared.NewID()
	if req.Subject != "" {
		parsed, err := shared.ParseID(req.Subject)
		if err != nil {
			writeError(w, err)
			return
		}
		subject = parsed
	}
	principal, err := auth.NewPrincipal(auth.Kind(req.Kind), subject)
	if err != nil {
		writeError(w, err)
		return
	}

	token, err := auth.NewToken()
	if err != nil {
		writeError(w, err) // the OS random source failed: a 500, not a 400
		return
	}
	if err := s.tokens.Issue(r.Context(), token, principal, req.Label, s.now()); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, issueTokenResponse{
		Token:   token,
		Kind:    principal.Kind().String(),
		Subject: principal.ID().String(),
	})
}

// now is the adapter's only use of the clock: a token's created_at. Every
// other timestamp is the domain's, taken by a use case (convention 7).
func (s *server) now() time.Time {
	return s.clock.Now()
}

// keyView is one issued credential as an operator sees it. There is no token
// field and there never can be: the store keeps a hash.
type keyView struct {
	Hash      string     `json:"hash"`
	Kind      string     `json:"kind"`
	Subject   string     `json:"subject"`
	Label     string     `json:"label,omitempty"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// GET /tokens → 200 [keyView]. Operator only: this is the list of who can
// call the API at all.
func (s *server) listTokens(w http.ResponseWriter, r *http.Request) {
	keys, err := s.registry.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]keyView, 0, len(keys))
	for _, k := range keys {
		v := keyView{
			Hash: k.Hash, Kind: k.Kind.String(), Subject: k.Subject, Label: k.Label,
			Active: k.Active(), CreatedAt: k.CreatedAt,
		}
		if !k.RevokedAt.IsZero() {
			t := k.RevokedAt
			v.RevokedAt = &t
		}
		out = append(out, v)
	}
	writeJSON(w, http.StatusOK, out)
}

// DELETE /tokens/{hash} → 204. The hash, not the token: an operator reading
// the list has never seen the token and never will.
//
// Revoking a key that does not exist, or one already revoked, is also 204:
// the caller wanted it not to work, and it does not. Answering 404 would turn
// this route into a way to ask "is this hash real".
//
// The one refusal is the LAST ACTIVE OPERATOR KEY. An API nobody can
// administer is not a safer API — it is one where the only recovery is
// editing the database by hand. Issue the replacement first, then revoke.
func (s *server) revokeToken(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")
	if hash == "" {
		writeError(w, errInvalidRequest)
		return
	}
	keys, err := s.registry.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	operators, target := 0, (*auth.Key)(nil)
	for i, k := range keys {
		if !k.Active() {
			continue
		}
		if k.Kind == auth.Operator {
			operators++
		}
		if k.Hash == hash {
			target = &keys[i]
		}
	}
	if target != nil && target.Kind == auth.Operator && operators == 1 {
		writeError(w, errLastOperator)
		return
	}
	if err := s.registry.RevokeHash(r.Context(), hash, s.now()); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
