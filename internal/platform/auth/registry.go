package auth

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// Key is one issued credential as an ADMINISTRATOR sees it — never the token
// itself, which no store has and nobody can get back.
//
// Hash is the handle: it is what the store is keyed by, it is safe to show
// (knowing a SHA-256 does not let anybody authenticate), and it is the only
// name a key has after the one time its token was shown.
type Key struct {
	Hash      string
	Kind      Kind
	Subject   string
	Label     string
	CreatedAt time.Time
	RevokedAt time.Time // zero while the key still works
}

// Active reports whether the key still authenticates.
func (k Key) Active() bool {
	return k.RevokedAt.IsZero()
}

// Registry is the ADMIN half of the port: seeing which keys exist and stopping
// one. It is a third interface rather than more methods on Verifier because of
// the same rule that split Verifier from Issuer — thirty routes verify, one
// issues, and only the key-management routes may revoke.
//
// RevokeHash takes the hash, not the token: an operator looking at a list has
// never seen the token and never will.
type Registry interface {
	List(ctx context.Context) ([]Key, error)
	RevokeHash(ctx context.Context, hash string, now time.Time) error
}

// ── Static implements it too, so dev and tests behave like production ────────

// meta is what Static remembers besides the principal. It exists only so the
// dev adapter can answer List with the same fields the table has.
type meta struct {
	principal Principal
	label     string
	createdAt time.Time
	revokedAt time.Time
}

var _ Registry = (*Static)(nil)

func (s *Static) List(ctx context.Context) ([]Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Key, 0, len(s.byHash))
	for hash, p := range s.byHash {
		m := s.meta[hash]
		out = append(out, Key{
			Hash: hash, Kind: p.Kind(), Subject: p.ID().String(),
			Label: m.label, CreatedAt: m.createdAt, RevokedAt: m.revokedAt,
		})
	}
	// Newest first, ties broken by hash: map iteration is randomised, and a
	// listing that changed order every call would be unreadable and untestable.
	sort.Slice(out, func(i, j int) bool {
		if a, b := out[i].CreatedAt, out[j].CreatedAt; !a.Equal(b) {
			return a.After(b)
		}
		return out[i].Hash < out[j].Hash
	})
	return out, nil
}

// RevokeHash removes the key. Revoking one that is already gone is not an
// error: the caller wanted it not to work, and it does not (idempotent).
func (s *Static) RevokeHash(ctx context.Context, hash string, now time.Time) error {
	if hash == "" {
		return fmt.Errorf("revoke: empty hash: %w", ErrUnauthenticated)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byHash[hash]; !ok {
		return nil
	}
	// Static really deletes, where Postgres keeps the row with revoked_at: the
	// dev adapter has no audit to answer to, and a map that grows with dead
	// entries for the life of a process is worse than one that does not.
	delete(s.byHash, hash)
	delete(s.meta, hash)
	return nil
}
