package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/auth"
)

// errForbidden is the second half of the authentication conversation. The two
// are different answers to different questions, and conflating them is how an
// API tells an attacker which tokens are real:
//
//	401 unauthenticated  I do not know who you are
//	403 forbidden        I know who you are; this is not yours
var errForbidden = errors.New("forbidden")

// principalKey is an unexported type so no other package can read or write
// the principal in a request context — not even by accident, since a string
// key of the same text would not match.
//
// [PHP] Symfony giữ user trong TokenStorage (một service). Go giữ trong
// [PHP] context của request; key là kiểu private nên chỉ package này chạm được.
type principalKeyType struct{}

var principalKey principalKeyType

// authenticate is the middleware every route sits behind. It FAILS CLOSED: a
// request without a valid bearer token never reaches a handler, so a route
// added tomorrow is protected before anyone remembers to protect it.
//
// The alternative — a list of public routes — is one forgotten line away from
// an open endpoint. There is no public route today; when there is, it gets an
// explicit exemption here, in one visible place.
func authenticate(v auth.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				writeError(w, auth.ErrUnauthenticated)
				return
			}
			p, err := v.Verify(r.Context(), token)
			if err != nil {
				// Whatever went wrong — unknown token, revoked, database
				// down — the caller is told the same thing. The detail is in
				// the log, through writeError's 500 path if it is a bug.
				if errors.Is(err, auth.ErrUnauthenticated) {
					writeError(w, auth.ErrUnauthenticated)
					return
				}
				writeError(w, err)
				return
			}
			ctx := context.WithValue(r.Context(), principalKey, p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// bearerToken pulls the token out of "Authorization: Bearer <token>".
//
// The scheme match is case-insensitive because RFC 7235 says it is; the token
// itself is not touched, because a token is bytes we issued, not text a human
// typed — trimming or lowercasing it would accept credentials we never made.
func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}
	return token, true
}

// principalOf returns who is calling. Behind authenticate it is never zero;
// the zero value is what a handler would see if someone wired a route outside
// the middleware, and every require* below treats that as forbidden.
func principalOf(r *http.Request) auth.Principal {
	p, _ := r.Context().Value(principalKey).(auth.Principal)
	return p
}

// requireOperator, requireCustomer and requireAny are the whole authorisation
// model: three wrappers, applied in the route table so the rule for a route is
// visible on the same line as the route.
//
// [PHP] Thay cho #[IsGranted('ROLE_OPERATOR')] trên Controller. Ít quyền lực
// [PHP] hơn voter của Symfony — cố ý: chưa có luật nào cần hơn hai loại người.
func requireOperator(h http.HandlerFunc) http.HandlerFunc {
	return requireKind(h, auth.Operator)
}

func requireCustomer(h http.HandlerFunc) http.HandlerFunc {
	return requireKind(h, auth.Customer)
}

// requireAny lets both kinds through — for routes that serve staff and
// customers alike, like asking for a price.
func requireAny(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if principalOf(r).IsZero() {
			writeError(w, auth.ErrUnauthenticated)
			return
		}
		h(w, r)
	}
}

func requireKind(h http.HandlerFunc, want auth.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := principalOf(r)
		if p.IsZero() {
			writeError(w, auth.ErrUnauthenticated)
			return
		}
		if p.Kind() != want {
			writeError(w, errForbidden)
			return
		}
		h(w, r)
	}
}

// operatorOf is the caller's identity as the domain wants it. It comes from
// the token now, never from a header: a header is something the caller writes,
// and "who confirmed this listing" must not be.
//
// The bool is false for a customer. On a route behind requireOperator it is
// always true — the check stays so that moving a route out from behind that
// wrapper fails visibly instead of recording the zero operator.
func operatorOf(r *http.Request) (shared.OperatorID, bool) {
	return principalOf(r).OperatorID()
}

// customerOf is the subject id for the routes a customer owns: placing an
// order, accepting a quote. Behind requireCustomer the bool is always true.
func customerOf(r *http.Request) (shared.ID, bool) {
	p := principalOf(r)
	if p.Kind() != auth.Customer {
		return shared.ID{}, false
	}
	return p.ID(), true
}

// sourcingOf reads the provenance of a product from WHO pasted it, not from
// the body. A customer cannot claim their guess was typed by an operator, and
// an operator cannot be blamed for a customer's paste (CATALOG.md §4).
//
// This is why `sourced_by` left the request body: it was a caller-supplied
// answer to a question only the token can answer.
func sourcingOf(r *http.Request) catalog.SourcingMode {
	if principalOf(r).Kind() == auth.Operator {
		return catalog.SourcedByOperator
	}
	return catalog.SourcedByCustomer
}
