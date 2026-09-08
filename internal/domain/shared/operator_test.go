package shared_test

import (
	"errors"
	"testing"

	"github.com/duongsy/portage/internal/domain/shared"
)

// "Who confirmed this" is asked in every context — catalog (who measured),
// procurement (who bought), logistics (who packed) — so the staff identity
// lives in the shared kernel, as an id only. The staff member's name, role
// and login belong to a future identity context.
func TestOperatorID(t *testing.T) {
	id := shared.NewOperatorID()
	if id.IsZero() {
		t.Fatal("NewOperatorID() must not be zero")
	}
	back, err := shared.ParseOperatorID(id.String())
	if err != nil || back != id {
		t.Fatalf("ParseOperatorID(%s) = %v, %v", id, back, err)
	}
	if _, err := shared.ParseOperatorID("not-an-id"); !errors.Is(err, shared.ErrInvalidID) {
		t.Errorf("garbage: got %v, want ErrInvalidID", err)
	}
	if !(shared.OperatorID{}).IsZero() {
		t.Error("OperatorID{} must be zero")
	}
}
