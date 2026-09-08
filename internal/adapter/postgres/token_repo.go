package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/auth"
)

// TokenRepo implements auth.Verifier on the api_tokens table, and issues and
// revokes the rows it verifies.
//
// It is the production half of the port platform/auth declares; the tests and
// a dev run use auth.Static, which answers the same way from a map.
type TokenRepo struct {
	pool *pgxpool.Pool
}

var _ auth.Verifier = (*TokenRepo)(nil)

func NewTokenRepo(pool *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{pool: pool}
}

// Verify hashes what it is given and looks the hash up. The token itself never
// reaches the database, and no query in this file takes a token in the clear.
//
// Every failure — no row, revoked, a kind the code no longer knows — answers
// auth.ErrUnauthenticated. The caller must not learn which.
func (r *TokenRepo) Verify(ctx context.Context, token string) (auth.Principal, error) {
	if token == "" {
		return auth.Principal{}, auth.ErrUnauthenticated
	}
	const q = `SELECT kind, subject FROM api_tokens
	           WHERE token_hash = $1 AND revoked_at IS NULL`

	var kind string
	var subject string
	err := db(ctx, r.pool).QueryRow(ctx, q, auth.HashToken(token)).Scan(&kind, &subject)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return auth.Principal{}, auth.ErrUnauthenticated
	case err != nil:
		return auth.Principal{}, fmt.Errorf("verify token: %w", err)
	}

	id, err := shared.ParseID(subject)
	if err != nil {
		// A row we wrote ourselves has an unreadable subject: that is data
		// corruption, not a bad credential, so it does NOT become a 401. It
		// travels as a real error and surfaces as a 500 with a log line.
		return auth.Principal{}, fmt.Errorf("token subject %q: %w", subject, err)
	}
	p, err := auth.NewPrincipal(auth.Kind(kind), id)
	if err != nil {
		return auth.Principal{}, fmt.Errorf("token kind %q: %w", kind, err)
	}
	return p, nil
}

// Issue stores a new token for a principal and returns nothing: the caller
// already has the token in the clear, and this is the last moment anyone can
// read it. A lost token is re-issued, never recovered.
func (r *TokenRepo) Issue(ctx context.Context, token string, p auth.Principal, label string, now time.Time) error {
	if token == "" {
		return fmt.Errorf("issue token: %w", auth.ErrUnauthenticated)
	}
	if p.IsZero() {
		return fmt.Errorf("issue token: %w", auth.ErrNoSubject)
	}
	const q = `INSERT INTO api_tokens (token_hash, kind, subject, label, created_at)
	           VALUES ($1, $2, $3, $4, $5)`
	_, err := db(ctx, r.pool).Exec(ctx, q,
		auth.HashToken(token), string(p.Kind()), p.ID().String(), label, now)
	if err != nil {
		return fmt.Errorf("issue token: %w", err)
	}
	return nil
}

// Revoke stops a token working, keeping the row so an audit can still answer
// "when did this stop". Revoking an already-revoked token changes nothing and
// is not an error: the caller's intent is satisfied either way (idempotent,
// P9-PLAN §1 rule 3).
func (r *TokenRepo) Revoke(ctx context.Context, token string, now time.Time) error {
	const q = `UPDATE api_tokens SET revoked_at = $2
	           WHERE token_hash = $1 AND revoked_at IS NULL`
	if _, err := db(ctx, r.pool).Exec(ctx, q, auth.HashToken(token), now); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}

// IsEmpty reports whether any token exists at all. cmd/api uses it to decide
// whether -bootstrap-operator-token may create the first operator: seeding a
// fixed token into a database that already has users would be a back door.
func (r *TokenRepo) IsEmpty(ctx context.Context) (bool, error) {
	var exists bool
	err := db(ctx, r.pool).QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM api_tokens)`).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("count tokens: %w", err)
	}
	return !exists, nil
}

// List is the operator's key screen. It never returns a token — the store does
// not have one — only the hash, which is the handle a revoke needs.
//
// Revoked keys stay in the list, greyed out by RevokedAt: "who had access and
// until when" is the question an audit actually asks.
func (r *TokenRepo) List(ctx context.Context) ([]auth.Key, error) {
	const q = `SELECT token_hash, kind, subject, label, created_at, revoked_at
	           FROM api_tokens ORDER BY created_at DESC, token_hash`
	rows, err := db(ctx, r.pool).Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list tokens: %w", err)
	}
	defer rows.Close()

	var out []auth.Key
	for rows.Next() {
		var (
			k       auth.Key
			kind    string
			subject string
			revoked *time.Time
		)
		if err := rows.Scan(&k.Hash, &kind, &subject, &k.Label, &k.CreatedAt, &revoked); err != nil {
			return nil, fmt.Errorf("list tokens: %w", err)
		}
		k.Kind, k.Subject = auth.Kind(kind), subject
		k.CreatedAt = k.CreatedAt.UTC()
		if revoked != nil {
			k.RevokedAt = revoked.UTC()
		}
		out = append(out, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tokens: %w", err)
	}
	return out, nil
}

// RevokeHash is Revoke by the handle an operator actually has. They saw the
// token once, when it was issued; the hash is what the list shows them.
func (r *TokenRepo) RevokeHash(ctx context.Context, hash string, now time.Time) error {
	const q = `UPDATE api_tokens SET revoked_at = $2
	           WHERE token_hash = $1 AND revoked_at IS NULL`
	if _, err := db(ctx, r.pool).Exec(ctx, q, hash, now); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}

var _ auth.Registry = (*TokenRepo)(nil)
