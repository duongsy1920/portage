package shared_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// UUID version 7: the first 48 bits are a Unix-millisecond timestamp, so
// ids sort by creation time and B-tree indexes append instead of scatter.
// In the canonical text form the version nibble is character 14.
func TestNewID_isUUIDv7(t *testing.T) {
	s := shared.NewID().String()
	if len(s) != 36 || s[14] != '7' {
		t.Fatalf("NewID() = %q, want a 36-char UUID with version 7", s)
	}
}

func TestNewID_isUnique(t *testing.T) {
	seen := make(map[shared.ID]bool, 1000)
	for range 1000 {
		id := shared.NewID()
		if seen[id] {
			t.Fatalf("duplicate id %s", id)
		}
		seen[id] = true
	}
}

func TestNewID_sortsByCreationTime(t *testing.T) {
	first := shared.NewID()
	time.Sleep(2 * time.Millisecond)
	second := shared.NewID()
	if !(first.String() < second.String()) {
		t.Fatalf("ids must sort chronologically: %s !< %s", first, second)
	}
}

func TestParseID_roundTripsAndRejectsGarbage(t *testing.T) {
	id := shared.NewID()
	back, err := shared.ParseID(id.String())
	if err != nil || back != id {
		t.Fatalf("ParseID(%s) = %s, %v", id, back, err)
	}
	for _, in := range []string{"", "not-a-uuid", "123e4567-e89b-12d3-a456-426614174000" /* v1 */} {
		if _, err := shared.ParseID(in); !errors.Is(err, shared.ErrInvalidID) {
			t.Errorf("ParseID(%q): got %v, want ErrInvalidID", in, err)
		}
	}
}

func TestID_zeroValue(t *testing.T) {
	if !(shared.ID{}).IsZero() {
		t.Fatal("ID{} must be zero")
	}
	if shared.NewID().IsZero() {
		t.Fatal("NewID() must not be zero")
	}
}

// uuid.Parse also accepts uppercase, braces, the URN prefix and the
// dashless form. ID.String() writes exactly one of those, so the rest mean
// the value was rewritten somewhere — and two spellings of one id would
// index as two different database keys.
func TestParseID_rejectsNonCanonicalForms(t *testing.T) {
	canonical := shared.NewID().String()

	for name, in := range map[string]string{
		"uppercase": strings.ToUpper(canonical),
		"braces":    "{" + canonical + "}",
		"urn":       "urn:uuid:" + canonical,
		"no dashes": strings.ReplaceAll(canonical, "-", ""),
	} {
		if _, err := shared.ParseID(in); !errors.Is(err, shared.ErrInvalidID) {
			t.Errorf("ParseID(%s form %q): got %v, want ErrInvalidID", name, in, err)
		}
	}

	if _, err := shared.ParseID(canonical); err != nil {
		t.Fatalf("canonical form must still parse: %v", err)
	}
}
