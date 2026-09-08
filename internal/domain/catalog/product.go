package catalog

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

var (
	ErrMerchantRequired   = errors.New("a merchant is required")
	ErrPriceRequired      = errors.New("a price is required")
	ErrNegativePrice      = errors.New("price cannot be negative")
	ErrProvenanceRequired = errors.New("provenance is required")
	ErrDuplicateVariant   = errors.New("variant already exists")
	ErrUnnamedVariant     = errors.New("a variant with no size or colour must be the product's only one")
	ErrNoVariants         = errors.New("product has no variants")
	ErrUnverified         = errors.New("not verified by an operator")
	ErrSuspectedDuplicate = errors.New("product is a suspected duplicate")
	ErrNotDraft           = errors.New("product is not a draft")
	ErrInvalidDuplicate   = errors.New("invalid duplicate reference")
)

// ProductID identifies a product.
type ProductID struct {
	shared.ID
}

func NewProductID() ProductID {
	return ProductID{shared.NewID()}
}

func ParseProductID(s string) (ProductID, error) {
	id, err := shared.ParseID(s)
	if err != nil {
		return ProductID{}, fmt.Errorf("product id: %w", err)
	}
	return ProductID{id}, nil
}

// ProductStatus is where a product is in its life.
//
//	draft      recorded, possibly from a customer's paste; not quotable with confidence
//	published  an operator stands behind listing and parcel; pricing may quote it
//	retired    the shop discontinued it; kept for the orders that reference it
type ProductStatus string

const (
	ProductStatusDraft     ProductStatus = "draft"
	ProductStatusPublished ProductStatus = "published"
	ProductStatusRetired   ProductStatus = "retired"
)

// ProductDetails is what it takes to record a product (convention 10). Two
// provenances are required even for a draft: data with no origin can be
// neither trusted nor distrusted later, and Publish needs to know.
type ProductDetails struct {
	Name              string
	Merchant          MerchantID
	Category          CategoryCode
	Source            SourceURL
	Price             shared.Money // in the merchant's currency — the app layer checks that
	ListingProvenance Provenance   // where name, merchant, category and source came from
	PriceProvenance   Provenance

	// RequestedBy is the customer who asked for this thing. Optional: an
	// operator may add a product before anybody wants it. It is NOT part of
	// Provenance on purpose — provenance answers "how did this data get here
	// and who vouched for it", this answers "who is waiting for it".
	RequestedBy shared.ID
}

// Product is the second AGGREGATE ROOT of the catalogue: one thing a customer
// can ask us to buy, in one or more Variants, from one Merchant.
//
// The catalogue is EARNED, not copied (docs/CATALOG.md §0): a product starts
// as a draft from whatever the customer pasted, and becomes published only
// when an operator has confirmed what it is and put it on a scale. Three
// groups of attributes carry their own Provenance, because they arrive by
// different roads with different trust (§4):
//
//	listing   name, merchant, category, source URL   — what it IS
//	price     the shop's price                       — what it COSTS
//	parcel    packed weight and box                  — what it BILLS
//
// Duty and tariff are not here (pricing.ShippingLane). Stock levels are not
// here either: availability changes hourly and is procurement's problem at
// the moment of buying.
//
// [PHP] Aggregate root: pointer receiver (*Product) vì có trạng thái đổi;
// [PHP] mọi field private, KHÔNG có setter — đổi qua method có tên nghiệp vụ
// [PHP] (Measure, Reprice, Publish…) và method có quyền nói "không".
type Product struct {
	shared.Events

	id       ProductID
	merchant MerchantID
	category CategoryCode
	name     string
	source   SourceURL

	listingProv Provenance

	price     shared.Money
	priceProv Provenance

	// requestedBy: the customer waiting for this product, when one asked.
	requestedBy shared.ID

	parcel     shared.ParcelSpec // zero until Measure
	parcelProv Provenance

	variants []Variant

	suspectedDuplicateOf ProductID   // zero unless flagged (option C, CATALOG.md §7)
	dismissedDuplicates  []ProductID // pairs an operator already judged: not flagged again
	status               ProductStatus
	addedAt              time.Time
}

// AddProduct records a new draft. Every required field is checked
// (conventions 1 and 10); the zero ProductDetails is refused.
func AddProduct(d ProductDetails, now time.Time) (*Product, error) {
	name := strings.TrimSpace(d.Name)
	if name == "" {
		return nil, ErrEmptyName
	}
	if d.Merchant.IsZero() {
		return nil, fmt.Errorf("product %q: %w", name, ErrMerchantRequired)
	}
	if d.Category.IsZero() {
		return nil, fmt.Errorf("product %q: %w", name, ErrInvalidCategoryCode)
	}
	if d.Source.IsZero() {
		return nil, fmt.Errorf("product %q: %w", name, ErrInvalidSourceURL)
	}
	if err := checkPrice(d.Price); err != nil {
		return nil, fmt.Errorf("product %q: %w", name, err)
	}
	if d.ListingProvenance.IsZero() || d.PriceProvenance.IsZero() {
		return nil, fmt.Errorf("product %q: %w", name, ErrProvenanceRequired)
	}

	p := &Product{
		id:          NewProductID(),
		merchant:    d.Merchant,
		category:    d.Category,
		name:        name,
		source:      d.Source,
		listingProv: d.ListingProvenance,
		price:       d.Price,
		priceProv:   d.PriceProvenance,
		requestedBy: d.RequestedBy,
		status:      ProductStatusDraft,
		addedAt:     now,
	}
	p.Record(ProductAdded{
		ID: p.id, Merchant: p.merchant, Category: p.category, Name: name,
		Source: p.source, Price: p.price,
		SourcedBy: d.ListingProvenance.Source(), RequestedBy: p.requestedBy,
		At: now,
	})
	return p, nil
}

func checkPrice(m shared.Money) error {
	if !m.IsValid() {
		return ErrPriceRequired
	}
	if m.IsNegative() {
		return fmt.Errorf("price %s: %w", m, ErrNegativePrice)
	}
	return nil
}

func (p *Product) ID() ProductID {
	return p.id
}

func (p *Product) Merchant() MerchantID {
	return p.merchant
}

func (p *Product) Category() CategoryCode {
	return p.category
}

func (p *Product) Name() string {
	return p.name
}

func (p *Product) Source() SourceURL {
	return p.source
}

// RequestedBy is the customer waiting for this product, or the zero id when
// an operator added it with nobody asking.
func (p *Product) RequestedBy() shared.ID {
	return p.requestedBy
}

func (p *Product) Price() shared.Money {
	return p.price
}

func (p *Product) ListingProvenance() Provenance {
	return p.listingProv
}

func (p *Product) PriceProvenance() Provenance {
	return p.priceProv
}

// ParcelSpec reports the measured spec and whether there is one. Until
// Measure is called there is nothing here — and the caller must NOT quietly
// fall back: pricing decides to use CategoryPolicy.DefaultParcelSpec, and
// knows it is guessing.
func (p *Product) ParcelSpec() (shared.ParcelSpec, bool) {
	return p.parcel, !p.parcel.IsZero()
}

func (p *Product) ParcelProvenance() (Provenance, bool) {
	return p.parcelProv, !p.parcelProv.IsZero()
}

// Variants returns a COPY, for the reason Merchant.Sourcing does.
func (p *Product) Variants() []Variant {
	return append([]Variant(nil), p.variants...)
}

func (p *Product) Status() ProductStatus {
	return p.status
}

func (p *Product) IsPublished() bool {
	return p.status == ProductStatusPublished
}

// SuspectedDuplicateOf reports the product this one may be a copy of, if an
// operator or the app layer flagged it.
func (p *Product) SuspectedDuplicateOf() (ProductID, bool) {
	return p.suspectedDuplicateOf, !p.suspectedDuplicateOf.IsZero()
}

func (p *Product) AddedAt() time.Time {
	return p.addedAt
}

// AddVariant adds a purchasable form. Two variants with the same size and
// colour are the same variant, and refused — see variantKey for what "same"
// means.
//
// It records catalog.variant_added, because the size the customer picked is
// the one thing the person who has to BUY it cannot do without, and no other
// context may read catalog's tables to find out (guard 7). The event is what
// puts "M 8 / W 9.5" on the buyer's screen instead of a UUID.
func (p *Product) AddVariant(d VariantDetails, now time.Time) (VariantID, error) {
	v := Variant{
		id:          NewVariantID(),
		size:        strings.TrimSpace(d.Size),
		color:       strings.TrimSpace(d.Color),
		merchantRef: strings.TrimSpace(d.MerchantRef),
		addedAt:     now,
	}
	for _, have := range p.variants {
		if have.key() == v.key() {
			return VariantID{}, fmt.Errorf("product %q: variant %q/%q: %w", p.name, v.size, v.color, ErrDuplicateVariant)
		}
	}
	// A variant with nothing to say is legal, but only ALONE: a one-size
	// product has exactly one (see VariantDetails). What is never legal is
	// mixing it with named ones — that is an operator who forgot to type the
	// size, and the buyer would be sent to a shop with no size to ask for.
	if len(p.variants) > 0 {
		if v.unnamed() {
			return VariantID{}, fmt.Errorf("product %q already has %d variant(s): %w", p.name, len(p.variants), ErrUnnamedVariant)
		}
		for _, have := range p.variants {
			if have.unnamed() {
				return VariantID{}, fmt.Errorf("product %q already has a variant with no size or colour: %w", p.name, ErrUnnamedVariant)
			}
		}
	}
	p.variants = append(p.variants, v)
	p.Record(VariantAdded{
		ID: p.id, Variant: v.id, Size: v.size, Color: v.color, MerchantRef: v.merchantRef, At: now,
	})
	return v.id, nil
}

// ConfirmListing records that an operator has checked what this product is.
// Only a verified provenance can confirm; a customer cannot vouch for their
// own paste. No event and no clock: the provenance carries its own time.
func (p *Product) ConfirmListing(prov Provenance, now time.Time) error {
	if !prov.Verified() {
		return fmt.Errorf("confirm listing of %q with %s: %w", p.name, prov, ErrUnverified)
	}
	p.listingProv = prov
	p.Record(ListingConfirmed{ID: p.id, By: prov.By(), At: now})
	return nil
}

// Measure records what this exact item weighs and how big its box is — the
// moment the catalogue earns its data (CATALOG.md §3). Any provenance is
// accepted here; Publish is where "verified" is demanded.
func (p *Product) Measure(spec shared.ParcelSpec, prov Provenance, now time.Time) error {
	if spec.IsZero() {
		return fmt.Errorf("measure %q: %w", p.name, shared.ErrIncompleteParcelSpec)
	}
	if prov.IsZero() {
		return fmt.Errorf("measure %q: %w", p.name, ErrProvenanceRequired)
	}
	p.parcel, p.parcelProv = spec, prov
	p.Record(ProductMeasured{ID: p.id, Parcel: spec, Verified: prov.Verified(), At: now})
	return nil
}

// Reprice follows the shop's price. The currency cannot change — that would
// be a different listing, not a new price. The provenance is always updated
// (an operator confirming the same price is new information); the event is
// raised only when the amount actually moved.
func (p *Product) Reprice(price shared.Money, prov Provenance, now time.Time) error {
	if err := checkPrice(price); err != nil {
		return fmt.Errorf("reprice %q: %w", p.name, err)
	}
	if price.Currency() != p.price.Currency() {
		return fmt.Errorf("reprice %q from %s to %s: %w", p.name, p.price.Currency(), price.Currency(), shared.ErrCurrencyMismatch)
	}
	if prov.IsZero() {
		return fmt.Errorf("reprice %q: %w", p.name, ErrProvenanceRequired)
	}
	old := p.price
	p.price, p.priceProv = price, prov
	if price == old {
		return nil
	}
	p.Record(ProductRepriced{ID: p.id, From: old, To: price, At: now})
	return nil
}

// FlagDuplicateOf marks this product as a suspected copy of another (option C,
// CATALOG.md §7). It stays a draft until an operator clears the flag or merges
// the two — merging is an application workflow, not a method here.
//
// The reason ("same page url", "same name and merchant") is what the operator
// reads in the queue. A pair an operator has already dismissed is NOT flagged
// again: the automated pass would otherwise flag, the operator clear, the
// pass flag — forever. The decision outranks the detector.
func (p *Product) FlagDuplicateOf(other ProductID, reason string, now time.Time) error {
	if other.IsZero() || other == p.id {
		return fmt.Errorf("product %q: duplicate of %s: %w", p.name, other, ErrInvalidDuplicate)
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("product %q: flag duplicate: %w", p.name, ErrEmptyReason)
	}
	if p.suspectedDuplicateOf == other || slices.Contains(p.dismissedDuplicates, other) {
		return nil
	}
	p.suspectedDuplicateOf = other
	p.Record(ProductFlaggedDuplicate{ID: p.id, Of: other, Reason: reason, At: now})
	return nil
}

// ClearDuplicateFlag records an operator's decision that this is its own
// product, with the reason — the audit trail for "why is this published when
// it looks like that one". The pair is remembered so it is not flagged again.
// Idempotent on an unflagged product.
func (p *Product) ClearDuplicateFlag(reason string, now time.Time) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("product %q: clear duplicate flag: %w", p.name, ErrEmptyReason)
	}
	if p.suspectedDuplicateOf.IsZero() {
		return nil
	}
	of := p.suspectedDuplicateOf
	p.suspectedDuplicateOf = ProductID{}
	p.dismissedDuplicates = append(p.dismissedDuplicates, of)
	p.Record(ProductDuplicateCleared{ID: p.id, Of: of, Reason: reason, At: now})
	return nil
}

// Publish makes the product quotable with confidence. THE invariant of this
// aggregate (CATALOG.md §4), checked in the order an operator would fix it:
//
//  1. it is still a draft;
//  2. there is something to order — at least one variant;
//  3. a person has confirmed what it is (listing provenance verified);
//  4. a person has put it on a scale (parcel present AND verified);
//  5. nobody suspects it is a copy of something we already have.
//
// Each refusal names its reason, so the operator screen can say exactly what
// is missing instead of "cannot publish".
func (p *Product) Publish(now time.Time) error {
	if p.status != ProductStatusDraft {
		return fmt.Errorf("publish %q: status %s: %w", p.name, p.status, ErrNotDraft)
	}
	if len(p.variants) == 0 {
		return fmt.Errorf("publish %q: %w", p.name, ErrNoVariants)
	}
	if !p.listingProv.Verified() {
		return fmt.Errorf("publish %q: listing is %s: %w", p.name, p.listingProv, ErrUnverified)
	}
	if p.parcel.IsZero() || !p.parcelProv.Verified() {
		return fmt.Errorf("publish %q: parcel is %s: %w", p.name, p.parcelProv, ErrUnverified)
	}
	if !p.suspectedDuplicateOf.IsZero() {
		return fmt.Errorf("publish %q: flagged as duplicate of %s: %w", p.name, p.suspectedDuplicateOf, ErrSuspectedDuplicate)
	}
	p.status = ProductStatusPublished
	p.Record(ProductPublished{
		ID: p.id, Merchant: p.merchant, Category: p.category,
		Name: p.name, Source: p.source, Price: p.price, Parcel: p.parcel, At: now,
	})
	return nil
}

// Retire takes a discontinued product out of circulation. It is kept — orders
// still reference it — but can never be published again. Idempotent.
func (p *Product) Retire(reason string, now time.Time) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("retire %q: %w", p.name, ErrEmptyReason)
	}
	if p.status == ProductStatusRetired {
		return nil
	}
	p.status = ProductStatusRetired
	p.Record(ProductRetired{ID: p.id, Reason: reason, At: now})
	return nil
}
