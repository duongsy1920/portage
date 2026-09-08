package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/postgres"
	"github.com/duongsy/portage/internal/adapter/postgres/pgtest"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/auth"
)

// The production half of the auth.Verifier port must answer exactly like the
// Static one the tests and a dev run use, or a system that works in memory
// locks everybody out on Postgres.
func TestTokenRepo_issueVerifyRevoke(t *testing.T) {
	pool := pgtest.Pool(t)
	repo := postgres.NewTokenRepo(pool)
	ctx := context.Background()
	at := time.Date(2026, 9, 6, 10, 0, 0, 0, time.UTC)

	// An empty table is what lets cmd/api bootstrap the first operator.
	if empty, err := repo.IsEmpty(ctx); err != nil || !empty {
		t.Fatalf("IsEmpty on a fresh database = %v, %v", empty, err)
	}

	operator := auth.MustPrincipal(auth.Operator, shared.NewID())
	if err := repo.Issue(ctx, "s3cret-operator", operator, "first operator", at); err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if empty, err := repo.IsEmpty(ctx); err != nil || empty {
		t.Fatalf("IsEmpty after one token = %v, %v", empty, err)
	}

	got, err := repo.Verify(ctx, "s3cret-operator")
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got.Kind() != auth.Operator || got.ID() != operator.ID() {
		t.Fatalf("Verify = %v, want %v", got, operator)
	}

	// Every failure is the same answer, so the API cannot be used to guess
	// which tokens are real.
	for _, bad := range []string{"", "nope", "S3CRET-OPERATOR", " s3cret-operator"} {
		if _, err := repo.Verify(ctx, bad); !errors.Is(err, auth.ErrUnauthenticated) {
			t.Errorf("Verify(%q) = %v, want ErrUnauthenticated", bad, err)
		}
	}

	// Revoking keeps the row — "when did this stop working" is a question an
	// audit asks — but the token stops verifying.
	if err := repo.Revoke(ctx, "s3cret-operator", at.Add(time.Hour)); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if _, err := repo.Verify(ctx, "s3cret-operator"); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("a revoked token still verifies: %v", err)
	}
	if empty, _ := repo.IsEmpty(ctx); empty {
		t.Fatal("Revoke deleted the row; it must be kept for the audit trail")
	}

	// Revoking twice is not an error: the caller's intent is satisfied either
	// way (at-least-once delivery means every write is called twice, §1 rule 3).
	if err := repo.Revoke(ctx, "s3cret-operator", at.Add(2*time.Hour)); err != nil {
		t.Fatalf("second Revoke: %v", err)
	}
}

// The token itself must never reach a column. A leaked dump then gives an
// attacker hashes of random strings — nothing to reverse, nothing to replay.
func TestTokenRepo_storesTheHashNotTheToken(t *testing.T) {
	pool := pgtest.Pool(t)
	repo := postgres.NewTokenRepo(pool)
	ctx := context.Background()
	const token = "s3cret-customer"

	customer := auth.MustPrincipal(auth.Customer, shared.NewID())
	if err := repo.Issue(ctx, token, customer, "app login", time.Now()); err != nil {
		t.Fatal(err)
	}

	var stored string
	if err := pool.QueryRow(ctx, `SELECT token_hash FROM api_tokens`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == token {
		t.Fatal("the token is in the table in the clear")
	}
	if stored != auth.HashToken(token) {
		t.Fatalf("token_hash = %q, want the sha-256 of the token", stored)
	}
}

// Two customers, two tokens: the id that comes back must be the one that went
// in, or a request would act as the wrong person.
func TestTokenRepo_keepsSubjectsApart(t *testing.T) {
	pool := pgtest.Pool(t)
	repo := postgres.NewTokenRepo(pool)
	ctx := context.Background()
	at := time.Now()

	alice := auth.MustPrincipal(auth.Customer, shared.NewID())
	bob := auth.MustPrincipal(auth.Operator, shared.NewID())
	if err := repo.Issue(ctx, "tok-alice", alice, "alice", at); err != nil {
		t.Fatal(err)
	}
	if err := repo.Issue(ctx, "tok-bob", bob, "bob", at); err != nil {
		t.Fatal(err)
	}

	gotAlice, err := repo.Verify(ctx, "tok-alice")
	if err != nil || gotAlice.ID() != alice.ID() || gotAlice.Kind() != auth.Customer {
		t.Fatalf("alice = %v, %v", gotAlice, err)
	}
	gotBob, err := repo.Verify(ctx, "tok-bob")
	if err != nil || gotBob.ID() != bob.ID() || gotBob.Kind() != auth.Operator {
		t.Fatalf("bob = %v, %v", gotBob, err)
	}
}

// The admin half of the port. Revoked keys STAY in the list with a timestamp:
// "who had access and until when" is what an audit asks, and a deleted row
// cannot answer it.
func TestTokenRepo_listAndRevokeByHash(t *testing.T) {
	p := pool(t)
	ctx := context.Background()
	repo := postgres.NewTokenRepo(p)

	op := auth.MustPrincipal(auth.Operator, shared.NewID())
	cust := auth.MustPrincipal(auth.Customer, shared.NewID())
	if err := repo.Issue(ctx, "op-token", op, "laptop", now); err != nil {
		t.Fatal(err)
	}
	if err := repo.Issue(ctx, "cust-token", cust, "phone", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	keys, err := repo.List(ctx)
	if err != nil || len(keys) != 2 {
		t.Fatalf("List = %d keys, %v", len(keys), err)
	}
	if keys[0].Label != "phone" { // newest first
		t.Errorf("list is not newest-first: %+v", keys)
	}
	for _, k := range keys {
		if k.Hash == "op-token" || k.Hash == "cust-token" {
			t.Fatal("the list returned a token, not a hash")
		}
		if !k.Active() {
			t.Errorf("key %s is not active", k.Label)
		}
	}

	target := auth.HashToken("cust-token")
	if err := repo.RevokeHash(ctx, target, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Verify(ctx, "cust-token"); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("a revoked token still verifies: %v", err)
	}
	keys, _ = repo.List(ctx)
	if len(keys) != 2 {
		t.Fatalf("a revoked key must stay in the list, got %d", len(keys))
	}
	for _, k := range keys {
		if k.Hash == target && (k.Active() || !k.RevokedAt.Equal(now.Add(time.Hour))) {
			t.Errorf("revoked key = %+v", k)
		}
	}

	// Idempotent, and a hash nobody has is not an error.
	if err := repo.RevokeHash(ctx, target, now.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := repo.RevokeHash(ctx, auth.HashToken("never-issued"), now); err != nil {
		t.Fatal(err)
	}
	keys, _ = repo.List(ctx)
	for _, k := range keys {
		if k.Hash == target && !k.RevokedAt.Equal(now.Add(time.Hour)) {
			t.Errorf("the second revoke moved revoked_at: %+v", k)
		}
	}
}
