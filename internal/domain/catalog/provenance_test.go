package catalog_test

import (
	"errors"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

var (
	provAt  = time.Date(2026, 9, 4, 9, 30, 0, 0, time.UTC)
	someone = shared.NewOperatorID()
)

// The same fact — "this weighs 1.2 kg" — is worth different amounts depending
// on who says so. Provenance is that "who", per group of attributes.
func TestNewProvenance_operatorMeansVerified(t *testing.T) {
	p, err := catalog.NewProvenance(catalog.SourcedByOperator, provAt, someone)
	if err != nil {
		t.Fatalf("NewProvenance: %v", err)
	}
	if !p.Verified() {
		t.Error("an operator's word is verified")
	}
	if p.Source() != catalog.SourcedByOperator || !p.At().Equal(provAt) || p.By() != someone {
		t.Errorf("getters: %v %v %v", p.Source(), p.At(), p.By())
	}

	customer, err := catalog.NewProvenance(catalog.SourcedByCustomer, provAt, shared.OperatorID{})
	if err != nil {
		t.Fatalf("customer provenance: %v", err)
	}
	if customer.Verified() {
		t.Error("what a customer pasted is not verified")
	}
	if feed, _ := catalog.NewProvenance(catalog.SourcedByFeed, provAt, shared.OperatorID{}); feed.Verified() {
		t.Error("a feed says what the shop says, not what we checked")
	}
}

// An operator claim without an operator is a claim nobody made.
func TestNewProvenance_rejectsBadInput(t *testing.T) {
	cases := map[string]struct {
		source catalog.SourcingMode
		at     time.Time
		by     shared.OperatorID
		want   error
	}{
		"operator without id": {catalog.SourcedByOperator, provAt, shared.OperatorID{}, catalog.ErrOperatorRequired},
		"unknown source":      {"telepathy", provAt, someone, catalog.ErrUnknownSourcing},
		"zero time":           {catalog.SourcedByCustomer, time.Time{}, shared.OperatorID{}, catalog.ErrProvenanceTime},
	}
	for name, c := range cases {
		if _, err := catalog.NewProvenance(c.source, c.at, c.by); !errors.Is(err, c.want) {
			t.Errorf("%s: got %v, want %v", name, err, c.want)
		}
	}
	if !(catalog.Provenance{}).IsZero() {
		t.Error("Provenance{} must be zero")
	}
}
