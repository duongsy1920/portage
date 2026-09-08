package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/auth"
)

// A token is our own randomness, not a human's choice: long enough that
// guessing is hopeless, and never the same twice.
func TestNewToken(t *testing.T) {
	seen := map[string]bool{}
	for range 200 {
		tok, err := auth.NewToken()
		if err != nil {
			t.Fatalf("NewToken: %v", err)
		}
		if len(tok) < 32 {
			t.Fatalf("token %q is %d chars — too short to be unguessable", tok, len(tok))
		}
		if seen[tok] {
			t.Fatalf("NewToken repeated itself: %q", tok)
		}
		seen[tok] = true
	}
}

// Static is the dev and test adapter for BOTH halves of the port: verifying a
// token and issuing one. Without Issue, POST /tokens would work on Postgres
// and fail in memory — and a dev run could never make a customer.
func TestStatic_issueThenVerify(t *testing.T) {
	v := auth.NewStatic(nil)
	ctx := context.Background()
	customer := auth.MustPrincipal(auth.Customer, shared.NewID())

	if _, err := v.Verify(ctx, "brand-new"); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("an unissued token must not verify: %v", err)
	}
	if err := v.Issue(ctx, "brand-new", customer, "app", time.Now()); err != nil {
		t.Fatalf("Issue: %v", err)
	}

	got, err := v.Verify(ctx, "brand-new")
	if err != nil || got.Kind() != auth.Customer || got.ID() != customer.ID() {
		t.Fatalf("Verify after Issue = %v, %v", got, err)
	}
}

func TestStatic_issueRejectsNonsense(t *testing.T) {
	v := auth.NewStatic(nil)
	ctx := context.Background()

	if err := v.Issue(ctx, "", auth.MustPrincipal(auth.Operator, shared.NewID()), "", time.Now()); err == nil {
		t.Error("an empty token must be refused")
	}
	if err := v.Issue(ctx, "tok", auth.Principal{}, "", time.Now()); err == nil {
		t.Error("a zero principal must be refused")
	}
}
