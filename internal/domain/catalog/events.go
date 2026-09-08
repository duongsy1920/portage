package catalog

import (
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// Domain events raised by this context.
//
// Every name is in the PAST TENSE: an event records something that HAS
// happened, and a listener cannot argue with it. `RegisterMerchant` is the
// command; `MerchantRegistered` is the fact it left behind.
//
// EventName() is the WIRE name — the string that goes into the outbox table
// and travels to other contexts. It must survive a Go type being renamed, so
// it is written out by hand rather than derived from the type. Renaming the
// struct is a refactor; renaming the wire name breaks every consumer and
// every row already stored.
//
// [PHP] Mỗi struct dưới đây tự động thoả interface shared.Event vì có đủ hai
// [PHP] method EventName() và OccurredAt(). Không có `implements` — xem
// [PHP] docs/GO-CHO-PHP.md §2.

type MerchantRegistered struct {
	ID       MerchantID
	Name     string
	Site     Hostname
	Currency shared.Currency // what the shop bills in — procurement checks receipts against it
	At       time.Time
}

func (e MerchantRegistered) EventName() string {
	return "catalog.merchant_registered"
}
func (e MerchantRegistered) OccurredAt() time.Time {
	return e.At
}

type MerchantRenamed struct {
	ID   MerchantID
	From string
	To   string
	At   time.Time
}

func (e MerchantRenamed) EventName() string {
	return "catalog.merchant_renamed"
}
func (e MerchantRenamed) OccurredAt() time.Time {
	return e.At
}

type MerchantSourcingEnabled struct {
	ID   MerchantID
	Mode SourcingMode
	At   time.Time
}

func (e MerchantSourcingEnabled) EventName() string {
	return "catalog.merchant_sourcing_enabled"
}
func (e MerchantSourcingEnabled) OccurredAt() time.Time {
	return e.At
}

type MerchantSourcingDisabled struct {
	ID   MerchantID
	Mode SourcingMode
	At   time.Time
}

func (e MerchantSourcingDisabled) EventName() string {
	return "catalog.merchant_sourcing_disabled"
}

func (e MerchantSourcingDisabled) OccurredAt() time.Time {
	return e.At
}

// MerchantFreeShippingChanged carries both rules so a pricing listener can
// re-quote without loading the merchant.
type MerchantFreeShippingChanged struct {
	ID   MerchantID
	From FreeShipping
	To   FreeShipping
	At   time.Time
}

func (e MerchantFreeShippingChanged) EventName() string {
	return "catalog.merchant_free_shipping_changed"
}

func (e MerchantFreeShippingChanged) OccurredAt() time.Time {
	return e.At
}

// MerchantSuspended is the event procurement listens for: stop buying here,
// and put a human in front of every purchase task still open.
type MerchantSuspended struct {
	ID     MerchantID
	Reason string
	At     time.Time
}

func (e MerchantSuspended) EventName() string {
	return "catalog.merchant_suspended"
}

func (e MerchantSuspended) OccurredAt() time.Time {
	return e.At
}

type MerchantReinstated struct {
	ID MerchantID
	At time.Time
}

func (e MerchantReinstated) EventName() string {
	return "catalog.merchant_reinstated"
}

func (e MerchantReinstated) OccurredAt() time.Time {
	return e.At
}

// ── Product ──────────────────────────────────────────────────────────────────

type ProductAdded struct {
	ID       ProductID
	Merchant MerchantID
	Category CategoryCode
	Name     string
	At       time.Time
}

func (e ProductAdded) EventName() string {
	return "catalog.product_added"
}

// VariantAdded says which purchasable form now exists, in the shop's own
// words. It carries the words rather than only the id for the reason every
// event here carries state: the consumer may not import catalog to ask
// (guard 7), so the payload is all it will ever have. Procurement is the
// consumer that needs it — somebody has to walk into the shop and ask for
// this size.
type VariantAdded struct {
	ID          ProductID
	Variant     VariantID
	Size        string
	Color       string
	MerchantRef string
	At          time.Time
}

func (e VariantAdded) EventName() string {
	return "catalog.variant_added"
}

func (e VariantAdded) OccurredAt() time.Time {
	return e.At
}

func (e ProductAdded) OccurredAt() time.Time {
	return e.At
}

// ProductMeasured is what pricing waits for: with a real parcel it can quote
// this product on its own numbers instead of the category default.
type ProductMeasured struct {
	ID       ProductID
	Parcel   shared.ParcelSpec
	Verified bool
	At       time.Time
}

func (e ProductMeasured) EventName() string {
	return "catalog.product_measured"
}

func (e ProductMeasured) OccurredAt() time.Time {
	return e.At
}

// ProductRepriced carries both amounts so a listener can re-quote without
// loading the product. Quotes already issued keep their own snapshot.
type ProductRepriced struct {
	ID   ProductID
	From shared.Money
	To   shared.Money
	At   time.Time
}

func (e ProductRepriced) EventName() string {
	return "catalog.product_repriced"
}

func (e ProductRepriced) OccurredAt() time.Time {
	return e.At
}

// ProductFlaggedDuplicate feeds the operator's review queue; Reason is what
// the detector saw ("same page url").
type ProductFlaggedDuplicate struct {
	ID     ProductID
	Of     ProductID
	Reason string
	At     time.Time
}

func (e ProductFlaggedDuplicate) EventName() string {
	return "catalog.product_flagged_duplicate"
}

func (e ProductFlaggedDuplicate) OccurredAt() time.Time {
	return e.At
}

// ProductDuplicateCleared is the operator's decision on record: this product
// and Of are different things, and here is why. It is what stops the same
// pair from being asked about twice.
type ProductDuplicateCleared struct {
	ID     ProductID
	Of     ProductID
	Reason string
	At     time.Time
}

func (e ProductDuplicateCleared) EventName() string {
	return "catalog.product_duplicate_cleared"
}

func (e ProductDuplicateCleared) OccurredAt() time.Time {
	return e.At
}

// ProductPublished is the catalogue's promise to pricing: this product's
// listing and parcel are vouched for, quote it with confidence.
//
// It carries the price and the measured parcel because the listener cannot
// come and get them: another context may not import this one (guard 7), so
// what is in the event is all pricing will ever know about the product.
type ProductPublished struct {
	ID       ProductID
	Merchant MerchantID
	Category CategoryCode
	Name     string
	Source   SourceURL // the page to buy from; procurement cannot ask catalog for it
	Price    shared.Money
	Parcel   shared.ParcelSpec
	At       time.Time
}

func (e ProductPublished) EventName() string {
	return "catalog.product_published"
}

func (e ProductPublished) OccurredAt() time.Time {
	return e.At
}

type ProductRetired struct {
	ID     ProductID
	Reason string
	At     time.Time
}

func (e ProductRetired) EventName() string {
	return "catalog.product_retired"
}

func (e ProductRetired) OccurredAt() time.Time {
	return e.At
}

// ── Category ─────────────────────────────────────────────────────────────────

// CategoryDefined announces a category's policy — new or replaced. Pricing
// keeps its own copy (goods class, default parcel, restrictions) from these;
// nothing else has a way to learn that a category changed, because the
// policy is a value object saved whole, not an aggregate with a lifecycle.
type CategoryDefined struct {
	Code         CategoryCode
	Estimate     shared.ParcelSpec
	Restrictions []Restriction
	At           time.Time
}

func (e CategoryDefined) EventName() string {
	return "catalog.category_defined"
}

func (e CategoryDefined) OccurredAt() time.Time {
	return e.At
}
