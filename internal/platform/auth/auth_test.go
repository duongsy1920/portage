package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/auth"
)

func TestKind_isValid(t *testing.T) {
	for _, k := range []auth.Kind{auth.Operator, auth.Customer} {
		if !k.IsValid() {
			t.Errorf("%s must be a valid kind", k)
		}
	}
	for _, k := range []auth.Kind{"", "admin", "Operator", "root"} {
		if k.IsValid() {
			t.Errorf("%q must not be a valid kind", k)
		}
	}
}

// A principal without an id is not a principal: it would let a route that
// only checks Kind through with nobody behind it.
func TestNewPrincipal_needsKindAndID(t *testing.T) {
	id := shared.NewID()

	p, err := auth.NewPrincipal(auth.Operator, id)
	if err != nil {
		t.Fatalf("NewPrincipal: %v", err)
	}
	if p.Kind() != auth.Operator || p.ID() != id {
		t.Fatalf("NewPrincipal = %v", p)
	}
	if p.IsZero() {
		t.Error("a real principal must not be zero")
	}

	if _, err := auth.NewPrincipal("admin", id); !errors.Is(err, auth.ErrUnknownKind) {
		t.Errorf("unknown kind: got %v", err)
	}
	if _, err := auth.NewPrincipal(auth.Operator, shared.ID{}); !errors.Is(err, auth.ErrNoSubject) {
		t.Errorf("zero id: got %v", err)
	}
	if !(auth.Principal{}).IsZero() {
		t.Error("Principal{} must be zero")
	}
}

// OperatorID is the bridge to the domain: only an operator has one, and a
// customer asking for it gets the zero value rather than their own id wearing
// the wrong type.
func TestPrincipal_operatorID(t *testing.T) {
	id := shared.NewID()

	op, err := auth.NewPrincipal(auth.Operator, id)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := op.OperatorID(); !ok || got.String() != id.String() {
		t.Fatalf("OperatorID() = %v, %v", got, ok)
	}

	cust, err := auth.NewPrincipal(auth.Customer, id)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := cust.OperatorID(); ok || !got.IsZero() {
		t.Fatalf("a customer has no OperatorID, got %v, %v", got, ok)
	}
}

// The token never travels further than this package in the clear: what a
// store keeps is the hash. Same input, same hash; different input, different.
func TestHashToken(t *testing.T) {
	a := auth.HashToken("dev-operator")
	b := auth.HashToken("dev-operator")
	c := auth.HashToken("dev-operatoR")

	if a != b {
		t.Fatal("hashing must be deterministic")
	}
	if a == c {
		t.Fatal("a different token must hash differently")
	}
	if len(a) != 64 {
		t.Fatalf("want 64 hex chars of sha-256, got %d: %q", len(a), a)
	}
	if a == "dev-operator" {
		t.Fatal("the hash must not be the token")
	}
}

func TestStatic_verify(t *testing.T) {
	operator := shared.NewID()
	customer := shared.NewID()
	v := auth.NewStatic(map[string]auth.Principal{
		"dev-operator": auth.MustPrincipal(auth.Operator, operator),
		"dev-customer": auth.MustPrincipal(auth.Customer, customer),
	})
	ctx := context.Background()

	p, err := v.Verify(ctx, "dev-operator")
	if err != nil || p.Kind() != auth.Operator || p.ID() != operator {
		t.Fatalf("Verify(dev-operator) = %v, %v", p, err)
	}
	p, err = v.Verify(ctx, "dev-customer")
	if err != nil || p.Kind() != auth.Customer || p.ID() != customer {
		t.Fatalf("Verify(dev-customer) = %v, %v", p, err)
	}

	// Every failure is the SAME error: a caller must not learn from the
	// answer whether a token exists, only that this one did not work.
	for _, bad := range []string{"", "nope", "DEV-OPERATOR", " dev-operator"} {
		if _, err := v.Verify(ctx, bad); !errors.Is(err, auth.ErrUnauthenticated) {
			t.Errorf("Verify(%q): got %v, want ErrUnauthenticated", bad, err)
		}
	}
}

// Static is a dev and test adapter; it must still refuse to be built with
// nonsense, or a typo in wire silently produces a system nobody can log in to.
func TestNewStatic_rejectsZeroPrincipal(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic on a zero principal")
		}
	}()
	auth.NewStatic(map[string]auth.Principal{"t": {}})
}
