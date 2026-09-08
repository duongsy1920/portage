package catalog

import (
	"errors"
	"fmt"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

var ErrInvalidSnapshot = errors.New("invalid snapshot")

// Snapshots are the aggregate's state with the door open — for ONE caller, the
// repository, which has to write the state down and build it back.
//
// Why not exported fields: the door would be open to everyone, and every
// invariant in this package would be optional. Why not a repository in the
// domain package that reaches private fields: the domain would know SQL.
// A snapshot is the narrow third way: a plain struct that carries state and
// nothing else — no behaviour, no invariants, no events.
//
// Two rules for FromSnapshot:
//
//   - it TRUSTS business values (they were validated on the way in) …
//   - … but REFUSES a shape Snapshot() could never have produced — a
//     published product with no variants, a parcel with no provenance. That is
//     corruption, and ErrInvalidSnapshot says so instead of building a lie.
//
// Loading raises no event: being read back is not a business fact.
//
// [PHP] Doctrine làm việc này ngầm bằng Reflection — ghi thẳng vào private
// [PHP] field. Ở đây tường minh: Snapshot() là "toArray()", FromSnapshot là
// [PHP] "fromArray()", và chỉ repository gọi chúng.

// MerchantSnapshot is Merchant, flat.
type MerchantSnapshot struct {
	ID           MerchantID
	Name         string
	Site         Hostname
	Currency     shared.Currency
	FreeShipping FreeShipping
	Sourcing     []SourcingMode
	Status       MerchantStatus
	AddedAt      time.Time
}

// Snapshot copies the state out. Slices are copied: the snapshot must not
// share memory with the aggregate.
func (m *Merchant) Snapshot() MerchantSnapshot {
	return MerchantSnapshot{
		ID:           m.id,
		Name:         m.name,
		Site:         m.site,
		Currency:     m.currency,
		FreeShipping: m.freeShip,
		Sourcing:     append([]SourcingMode(nil), m.sourcing...),
		Status:       m.status,
		AddedAt:      m.addedAt,
	}
}

// MerchantFromSnapshot rebuilds a merchant a repository read back.
func MerchantFromSnapshot(s MerchantSnapshot) (*Merchant, error) {
	switch {
	case s.ID.IsZero():
		return nil, fmt.Errorf("merchant snapshot: no id: %w", ErrInvalidSnapshot)
	case s.Name == "", s.Site.IsZero(), s.Currency.IsZero():
		return nil, fmt.Errorf("merchant snapshot %s: missing name, site or currency: %w", s.ID, ErrInvalidSnapshot)
	case s.Status != StatusActive && s.Status != StatusSuspended:
		return nil, fmt.Errorf("merchant snapshot %s: status %q: %w", s.ID, s.Status, ErrInvalidSnapshot)
	}
	if err := s.FreeShipping.validFor(s.Currency); err != nil {
		return nil, fmt.Errorf("merchant snapshot %s: %v: %w", s.ID, err, ErrInvalidSnapshot)
	}
	modes, err := normaliseSourcing(s.Sourcing)
	if err != nil {
		return nil, fmt.Errorf("merchant snapshot %s: %v: %w", s.ID, err, ErrInvalidSnapshot)
	}
	return &Merchant{
		id:       s.ID,
		name:     s.Name,
		site:     s.Site,
		currency: s.Currency,
		freeShip: s.FreeShipping,
		sourcing: modes,
		status:   s.Status,
		addedAt:  s.AddedAt,
	}, nil
}

// VariantSnapshot is Variant, flat.
type VariantSnapshot struct {
	ID          VariantID
	Size        string
	Color       string
	MerchantRef string
	AddedAt     time.Time
}

// ProductSnapshot is Product, flat. Parcel and ParcelProvenance are both zero
// until the product has been measured — never one without the other.
type ProductSnapshot struct {
	ID                   ProductID
	Merchant             MerchantID
	Category             CategoryCode
	Name                 string
	Source               SourceURL
	ListingProvenance    Provenance
	Price                shared.Money
	PriceProvenance      Provenance
	RequestedBy          shared.ID
	Parcel               shared.ParcelSpec
	ParcelProvenance     Provenance
	Variants             []VariantSnapshot
	SuspectedDuplicateOf ProductID
	DismissedDuplicates  []ProductID
	Status               ProductStatus
	AddedAt              time.Time
}

func (p *Product) Snapshot() ProductSnapshot {
	var variants []VariantSnapshot
	for _, v := range p.variants {
		variants = append(variants, VariantSnapshot{
			ID: v.id, Size: v.size, Color: v.color, MerchantRef: v.merchantRef, AddedAt: v.addedAt,
		})
	}
	return ProductSnapshot{
		ID:                   p.id,
		Merchant:             p.merchant,
		Category:             p.category,
		Name:                 p.name,
		Source:               p.source,
		ListingProvenance:    p.listingProv,
		Price:                p.price,
		PriceProvenance:      p.priceProv,
		RequestedBy:          p.requestedBy,
		Parcel:               p.parcel,
		ParcelProvenance:     p.parcelProv,
		Variants:             variants,
		SuspectedDuplicateOf: p.suspectedDuplicateOf,
		DismissedDuplicates:  append([]ProductID(nil), p.dismissedDuplicates...),
		Status:               p.status,
		AddedAt:              p.addedAt,
	}
}

// ProductFromSnapshot rebuilds a product a repository read back.
func ProductFromSnapshot(s ProductSnapshot) (*Product, error) {
	bad := func(why string) error {
		return fmt.Errorf("product snapshot %s: %s: %w", s.ID, why, ErrInvalidSnapshot)
	}
	switch {
	case s.ID.IsZero():
		return nil, fmt.Errorf("product snapshot: no id: %w", ErrInvalidSnapshot)
	case s.Merchant.IsZero(), s.Category.IsZero(), s.Source.IsZero(), s.Name == "":
		return nil, bad("missing merchant, category, source or name")
	case s.Status != ProductStatusDraft && s.Status != ProductStatusPublished && s.Status != ProductStatusRetired:
		return nil, bad(fmt.Sprintf("status %q", s.Status))
	case !s.Price.IsValid() || s.Price.IsNegative():
		return nil, bad("price")
	case s.ListingProvenance.IsZero(), s.PriceProvenance.IsZero():
		return nil, bad("missing listing or price provenance")
	case s.Parcel.IsZero() != s.ParcelProvenance.IsZero():
		return nil, bad("parcel and its provenance must come together")
	case s.SuspectedDuplicateOf == s.ID:
		return nil, bad("flagged as a duplicate of itself")
	}

	variants := make([]Variant, 0, len(s.Variants))
	seen := make(map[string]bool, len(s.Variants))
	for _, v := range s.Variants {
		if v.ID.IsZero() {
			return nil, bad("variant without id")
		}
		key := variantKey(v.Size, v.Color)
		if seen[key] {
			return nil, bad(fmt.Sprintf("duplicate variant %q/%q", v.Size, v.Color))
		}
		seen[key] = true
		variants = append(variants, Variant{
			id: v.ID, size: v.Size, color: v.Color, merchantRef: v.MerchantRef, addedAt: v.AddedAt,
		})
	}
	if len(variants) == 0 {
		variants = nil // Snapshot() of a product without variants yields nil; keep round trips equal
	}

	if s.Status == ProductStatusPublished {
		// Publish() could not have let these through; a store that says
		// otherwise is corrupt.
		if len(variants) == 0 || !s.ListingProvenance.Verified() || s.Parcel.IsZero() || !s.ParcelProvenance.Verified() {
			return nil, bad("published without variants or verification")
		}
	}

	var dismissed []ProductID
	for _, id := range s.DismissedDuplicates {
		if id.IsZero() || id == s.ID {
			return nil, bad("dismissed duplicate id")
		}
		dismissed = append(dismissed, id)
	}

	return &Product{
		id:                   s.ID,
		merchant:             s.Merchant,
		category:             s.Category,
		name:                 s.Name,
		source:               s.Source,
		listingProv:          s.ListingProvenance,
		price:                s.Price,
		priceProv:            s.PriceProvenance,
		requestedBy:          s.RequestedBy,
		parcel:               s.Parcel,
		parcelProv:           s.ParcelProvenance,
		variants:             variants,
		suspectedDuplicateOf: s.SuspectedDuplicateOf,
		dismissedDuplicates:  dismissed,
		status:               s.Status,
		addedAt:              s.AddedAt,
	}, nil
}
