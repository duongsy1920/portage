package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/auth"
)

// Static is the dev and test adapter, so it has to answer the admin questions
// the same way the table does — otherwise a route tested here would behave
// differently the moment it ran on Postgres.
func TestStatic_listAndRevoke(t *testing.T) {
	ctx := context.Background()
	op := auth.MustPrincipal(auth.Operator, shared.NewID())
	s := auth.NewStatic(map[string]auth.Principal{"seed": op})

	at := time.Date(2026, 9, 6, 10, 0, 0, 0, time.UTC)
	cust := auth.MustPrincipal(auth.Customer, shared.NewID())
	if err := s.Issue(ctx, "phone", cust, "phone", at); err != nil {
		t.Fatal(err)
	}

	keys, err := s.List(ctx)
	if err != nil || len(keys) != 2 {
		t.Fatalf("List = %d keys, %v", len(keys), err)
	}
	// Newest first, and never the token itself — only its hash.
	if keys[0].Label != "phone" || keys[0].Kind != auth.Customer {
		t.Fatalf("first key = %+v", keys[0])
	}
	for _, k := range keys {
		if k.Hash == "phone" || k.Hash == "seed" {
			t.Fatal("List returned a token, not a hash")
		}
		if !k.Active() {
			t.Errorf("key %q is not active", k.Label)
		}
	}

	if err := s.RevokeHash(ctx, auth.HashToken("phone"), at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Verify(ctx, "phone"); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("a revoked token still verifies: %v", err)
	}
	if keys, _ = s.List(ctx); len(keys) != 1 {
		t.Errorf("List after revoke = %d keys", len(keys))
	}

	// Idempotent, and an unknown hash is not an error: the caller wanted the
	// key not to work, and it does not.
	if err := s.RevokeHash(ctx, auth.HashToken("phone"), at); err != nil {
		t.Errorf("second revoke = %v", err)
	}
	if err := s.RevokeHash(ctx, auth.HashToken("never-issued"), at); err != nil {
		t.Errorf("unknown hash = %v", err)
	}
	if err := s.RevokeHash(ctx, "", at); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Errorf("empty hash = %v, want a refusal", err)
	}
}
