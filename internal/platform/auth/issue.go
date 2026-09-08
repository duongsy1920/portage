package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Issuer is the write half of the port: cutting a new key.
//
// It is separate from Verifier because almost everything only needs to VERIFY.
// One route issues; thirty read. Splitting the interfaces keeps the thirty
// from depending on a capability they must not have.
//
// [PHP] Cùng ý với việc tách ReadModel khỏi Repository: interface nhỏ, người
// [PHP] gọi chỉ thấy đúng thứ mình cần.
type Issuer interface {
	Issue(ctx context.Context, token string, p Principal, label string, now time.Time) error
}

// tokenBytes is 32 bytes — 256 bits of randomness, printed as 64 hex chars.
//
// Long enough that guessing is not a strategy, so the token needs no rate
// limit of its own to be safe, and no salt when hashed (see HashToken).
const tokenBytes = 32

// NewToken mints a token nobody can predict.
//
// crypto/rand, not math/rand: math/rand is seeded and reproducible, which is
// exactly what a credential must not be. An error here means the OS random
// source failed, and issuing a guessable token would be worse than failing.
func NewToken() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("new token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// Issue adds a token to the static set, so a dev run and a handler test can
// exercise POST /tokens the same way production does.
//
// Static is the only adapter with a mutex: a map written while an HTTP handler
// reads it is a data race, and unlike the Postgres adapter there is no database
// underneath to serialise access (GO-CHO-PHP.md §9.4 — every request is a
// goroutine sharing what main() built).
func (s *Static) Issue(ctx context.Context, token string, p Principal, label string, now time.Time) error {
	if token == "" {
		return fmt.Errorf("issue token: empty token: %w", ErrUnauthenticated)
	}
	if p.IsZero() {
		return fmt.Errorf("issue token: %w", ErrNoSubject)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.byHash == nil {
		s.byHash = map[string]Principal{}
	}
	if s.meta == nil {
		s.meta = map[string]meta{}
	}
	hash := HashToken(token)
	s.byHash[hash] = p
	s.meta[hash] = meta{principal: p, label: label, createdAt: now}
	return nil
}
