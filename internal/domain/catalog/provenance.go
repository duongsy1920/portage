package catalog

import (
	"errors"
	"fmt"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

var (
	ErrOperatorRequired = shared.ErrOperatorRequired // the shared kernel's; same value, so errors.Is works either way
	ErrProvenanceTime   = errors.New("provenance needs the time it was recorded")
)

// Provenance says where a GROUP of a product's attributes came from, when,
// and who confirmed it.
//
// Three ways data reaches the catalogue, three levels of trust:
//
//	operator   a person looked at the shop page (or the scale) and typed it   → verified
//	customer   the customer pasted the details themselves                      → unverified
//	feed       the merchant's structured feed                                  → unverified for OUR purposes
//
// A feed is authoritative about what the shop SAYS (price, name) but knows
// nothing about what WE need (box size, packed weight) — so nothing but an
// operator's own measurement counts as verified. See docs/CATALOG.md §4.
//
// It is per GROUP, not per product: a Product carries one Provenance for its
// listing, one for its price and one for its parcel, because the price may
// come from a feed while the box was measured by hand. One label for the
// whole product would have to lie about one of them.
//
// [PHP] Value object với time.Time bên trong → KHÔNG so sánh bằng `==`
// [PHP] (time.Time chứa con trỏ *Location; hai mốc bằng nhau có thể `!=`).
// [PHP] Dùng At().Equal(...) như test đang làm. IsZero() vì thế viết tay.
type Provenance struct {
	source SourcingMode
	at     time.Time
	by     shared.OperatorID
}

// NewProvenance validates a claim about where data came from (convention 1).
// An operator claim must name the operator; the other sources may.
func NewProvenance(source SourcingMode, at time.Time, by shared.OperatorID) (Provenance, error) {
	if !source.IsValid() {
		return Provenance{}, fmt.Errorf("provenance source %q: %w", source, ErrUnknownSourcing)
	}
	if at.IsZero() {
		return Provenance{}, fmt.Errorf("provenance from %s: %w", source, ErrProvenanceTime)
	}
	if source == SourcedByOperator && by.IsZero() {
		return Provenance{}, fmt.Errorf("provenance from %s: %w", source, ErrOperatorRequired)
	}
	return Provenance{source: source, at: at, by: by}, nil
}

// MustProvenance is for tests and seed data; it panics on bad input.
func MustProvenance(source SourcingMode, at time.Time, by shared.OperatorID) Provenance {
	p, err := NewProvenance(source, at, by)
	if err != nil {
		panic(err)
	}
	return p
}

func (p Provenance) Source() SourcingMode {
	return p.source
}

func (p Provenance) At() time.Time {
	return p.at
}

// By is the operator who confirmed it — zero unless a person was involved.
func (p Provenance) By() shared.OperatorID {
	return p.by
}

// Verified reports whether a person on our side stands behind the data.
// This is the one question Product.Publish asks of a Provenance.
func (p Provenance) Verified() bool {
	return p.source == SourcedByOperator
}

func (p Provenance) IsZero() bool {
	return p.source == "" && p.at.IsZero() && p.by.IsZero()
}

func (p Provenance) String() string {
	if p.by.IsZero() {
		return fmt.Sprintf("%s at %s", p.source, p.at.Format(time.RFC3339))
	}
	return fmt.Sprintf("%s at %s by %s", p.source, p.at.Format(time.RFC3339), p.by)
}
