// Package auth answers one question for every request: WHO is calling.
//
// It lives in platform/, not domain/, because "who is calling" is not a
// business rule — it is something the edge establishes before any use case
// runs. The domain keeps taking a shared.OperatorID or a bare shared.ID; it
// never sees a Principal, a token or a header.
//
// The shape is the same port-and-adapter shape as Clock (DDD.md §20):
//
//	Verifier   the port — "turn this token into a principal"
//	Static     an adapter for dev and tests, tokens held in a map
//	postgres.TokenRepo   the adapter production uses
//
// [PHP] Tương đương Security component của Symfony, thu nhỏ còn phần cần
// [PHP] dùng: Principal ~ UserInterface + Passport, Verifier ~ Authenticator,
// [PHP] Static ~ InMemoryUserProvider. Không có role hierarchy, không có
// [PHP] voter — chỉ hai loại người gọi và bốn hàm require* ở tầng HTTP.
package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/duongsy/portage/internal/domain/shared"
)

var (
	// ErrUnauthenticated is the ONE answer to every failed verification: no
	// token, unknown token, revoked token. A caller must not be able to tell
	// which, or the API becomes an oracle for guessing valid tokens.
	ErrUnauthenticated = errors.New("unauthenticated")

	ErrUnknownKind = errors.New("unknown principal kind")
	ErrNoSubject   = errors.New("principal has no subject id")
)

// Kind is what sort of caller this is. Two, on purpose: there is no "admin"
// yet, and inventing one before a rule needs it would be a role nobody checks.
//
// [PHP] Lại là "enum" kiểu Go: type riêng + hằng, xem GO-CHO-PHP.md §2.
type Kind string

const (
	// Operator: staff. Runs the catalogue, the warehouse, buys at the shops,
	// confirms payments.
	Operator Kind = "operator"

	// Customer: the person the goods are for. Sees their own orders, accepts
	// their own quotes.
	Customer Kind = "customer"
)

func (k Kind) IsValid() bool {
	switch k {
	case Operator, Customer:
		return true
	}
	return false
}

func (k Kind) String() string {
	return string(k)
}

// Principal is the caller, established at the edge and carried in the request
// context. It is a value object: immutable, and a zero one is "nobody".
type Principal struct {
	kind Kind
	id   shared.ID
}

// NewPrincipal validates what an adapter read out of a token store.
func NewPrincipal(kind Kind, id shared.ID) (Principal, error) {
	if !kind.IsValid() {
		return Principal{}, fmt.Errorf("kind %q: %w", kind, ErrUnknownKind)
	}
	if id.IsZero() {
		return Principal{}, fmt.Errorf("kind %q: %w", kind, ErrNoSubject)
	}
	return Principal{kind: kind, id: id}, nil
}

// MustPrincipal is for wiring and tests, where the values are written by a
// programmer (SETUP.md convention 1).
func MustPrincipal(kind Kind, id shared.ID) Principal {
	p, err := NewPrincipal(kind, id)
	if err != nil {
		panic(err)
	}
	return p
}

func (p Principal) Kind() Kind {
	return p.kind
}

func (p Principal) ID() shared.ID {
	return p.id
}

func (p Principal) IsZero() bool {
	return p.kind == "" && p.id.IsZero()
}

// OperatorID is the bridge into the domain, which knows shared.OperatorID and
// nothing about principals.
//
// The bool is what stops a customer's id from being handed to a use case that
// wanted an operator: same bytes, different meaning, and the compiler cannot
// tell them apart once both are a shared.ID.
func (p Principal) OperatorID() (shared.OperatorID, bool) {
	if p.kind != Operator {
		return shared.OperatorID{}, false
	}
	return shared.OperatorID{ID: p.id}, true
}

func (p Principal) String() string {
	if p.IsZero() {
		return "nobody"
	}
	return fmt.Sprintf("%s %s", p.kind, p.id)
}

// Verifier turns a bearer token into a principal. It is the PORT; the HTTP
// adapter depends on this interface and never on a table or a map.
//
// It takes ctx because the production adapter reads a database.
type Verifier interface {
	Verify(ctx context.Context, token string) (Principal, error)
}

// HashToken is what a store keeps instead of the token itself. A leaked
// database then gives an attacker hashes, not credentials.
//
// SHA-256 with no salt on purpose: a token is 128+ bits of our own randomness,
// not a human-chosen password, so there is nothing to brute-force and nothing
// a rainbow table can precompute. Salting would only stop us from looking a
// token up by its hash, which is the whole operation.
//
// [PHP] hash('sha256', $token) — KHÔNG dùng password_hash() ở đây: bcrypt/argon
// [PHP] sinh salt ngẫu nhiên mỗi lần nên không tra ngược được, hợp cho mật khẩu
// [PHP] người đặt, không hợp cho token tra cứu theo hash.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Static verifies against a fixed map. It is the dev and test adapter, and
// the reason every handler test can name its caller in one line.
type Static struct {
	mu     sync.RWMutex
	byHash map[string]Principal
	meta   map[string]meta // label and timestamps, for List — never the token
}

// NewStatic takes tokens in the clear — they are written in wire.go and in
// tests — and keeps only their hashes, so this adapter and the Postgres one
// answer the same way and a dump of either reveals nothing.
//
// A zero Principal is a wiring bug, not bad input: it would create a token
// that authenticates nobody. Panic (convention 1).
func NewStatic(tokens map[string]Principal) *Static {
	byHash := make(map[string]Principal, len(tokens))
	for token, p := range tokens {
		if p.IsZero() {
			panic(fmt.Sprintf("auth: NewStatic called with a zero Principal for token %q", token))
		}
		if token == "" {
			panic("auth: NewStatic called with an empty token")
		}
		byHash[HashToken(token)] = p
	}
	return &Static{byHash: byHash, meta: map[string]meta{}}
}

// Verify answers ErrUnauthenticated for every failure, and compares in
// constant time so the answer's TIMING does not leak which prefix was right.
func (s *Static) Verify(ctx context.Context, token string) (Principal, error) {
	if token == "" {
		return Principal{}, ErrUnauthenticated
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	want := HashToken(token)
	for hash, p := range s.byHash {
		if subtle.ConstantTimeCompare([]byte(hash), []byte(want)) == 1 {
			return p, nil
		}
	}
	return Principal{}, ErrUnauthenticated
}
