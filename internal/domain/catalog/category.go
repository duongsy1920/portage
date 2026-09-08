package catalog

import (
	"errors"
	"fmt"
	"github.com/duongsy/portage/internal/domain/shared"
	"slices"
	"strings"
	"time"
)

var (
	ErrInvalidCategoryCode = errors.New("invalid category code")
	ErrUnknownRestriction  = errors.New("unknown restriction")
)

// CategoryCode names a kind of goods: "footwear", "apparel", "electronics".
//
// It is the category's IDENTITY — a NATURAL key. Not every entity needs a
// generated id: a category is known by its name in every conversation the
// business has, so a UUID would only add a second name for the same thing
// and turn every row into something you must join to understand.
//
// Because it ends up in URLs, config files and a primary-key column, it is
// validated like Hostname rather than merely lowercased: ^[a-z][a-z0-9_]*$.
// Case and surrounding spaces carry no meaning and are normalised; anything
// else is refused — the domain does not guess (convention 3).
//
// [PHP] Vì sao là `struct{ code string }` thay vì `type CategoryCode string`?
// [PHP] Với kiểu string, ai cũng ép được: CategoryCode("Giày Dép!") — compile
// [PHP] bình thường, bỏ qua mọi kiểm tra. Struct có field private thì chỉ tạo
// [PHP] được qua ParseCategoryCode. Symfony: final class CategoryCode với
// [PHP] constructor private + static fromString(string). Cùng một ý.
type CategoryCode struct {
	code string
}

// ParseCategoryCode reads what an operator typed or a URL carried.
func ParseCategoryCode(s string) (CategoryCode, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" || !isSnakeCase(s) {
		return CategoryCode{}, fmt.Errorf("category code %q: want ^[a-z][a-z0-9_]*$: %w", s, ErrInvalidCategoryCode)
	}
	return CategoryCode{code: s}, nil
}

// MustParseCategoryCode is for tests and seed data; it panics on bad input.
func MustParseCategoryCode(s string) CategoryCode {
	c, err := ParseCategoryCode(s)
	if err != nil {
		panic(err)
	}
	return c
}

// isSnakeCase: a lowercase letter first, then lowercase letters, digits or
// underscores. Written out instead of a regexp because it is eight lines and
// the regexp would be the only one in the domain.
//
// [PHP] `for i, r := range s` duyệt chuỗi theo RUNE (ký tự Unicode), không
// [PHP] theo byte. 'à' là một rune, không phải hai byte — nên "giày" bị từ chối
// [PHP] đúng chỗ. PHP: preg_match('/^[a-z][a-z0-9_]*$/', $s).
func isSnakeCase(s string) bool {
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case i > 0 && (r == '_' || (r >= '0' && r <= '9')):
		default:
			return false
		}
	}
	return true
}

func (c CategoryCode) String() string {
	return c.code
}

func (c CategoryCode) IsZero() bool {
	return c.code == ""
}

// Restriction is a physical property that limits how goods may travel.
//
// These are not warnings — a lithium battery genuinely cannot ride every
// freight lane. A domain that does not know the concept will happily quote a
// route the parcel is not allowed to take, and the box stops in the warehouse.
//
// [PHP] Đây là kiểu "enum" của Go: một kiểu string riêng + hằng số + IsValid().
// [PHP] Khác CategoryCode, ở đây dùng `type Restriction string` được vì tập giá
// [PHP] trị đóng và IsValid() chặn mọi giá trị lạ ở cửa vào (NewCategoryPolicy).
// [PHP] Symfony: enum Restriction: string { case Battery = 'battery'; ... }
type Restriction string

const (
	RestrictionBattery    Restriction = "battery"    // lithium cells: air-freight limits
	RestrictionLiquid     Restriction = "liquid"     // perfume, skincare
	RestrictionAerosol    Restriction = "aerosol"    // pressurised cans
	RestrictionSupplement Restriction = "supplement" // food/health: import licensing
	RestrictionMagnet     Restriction = "magnet"     // headphones, speakers
)

func (r Restriction) IsValid() bool {
	switch r {
	case RestrictionBattery, RestrictionLiquid, RestrictionAerosol,
		RestrictionSupplement, RestrictionMagnet:
		return true
	}
	return false
}

func (r Restriction) String() string {
	return string(r)
}

// CategoryPolicy is everything the catalogue knows about a KIND of goods:
// what it usually weighs and measures (a shared.ParcelSpec, the default for the
// kind), and what it is physically
// not allowed to do (Restrictions).
//
// It is an immutable VALUE OBJECT keyed by a natural code — not an entity
// with a lifecycle. It is configuration, edited rarely by a person who
// replaces the whole thing: there is no method that changes it, and the
// repository saves a new value under the same code. Compare Merchant, which
// has a status that moves and events that record the moves.
//
// WHAT IS DELIBERATELY NOT HERE — both moved to pricing.ShippingLane:
//
//   - Import duty and HS tariff codes. Duty is a function of (goods, LANE):
//     the lane this business uses bundles clearance into the per-kg price and
//     never itemises it, so a rate stored here would look authoritative and be
//     ignored. Guarded by TestDecision_catalogKnowsNothingAboutTax.
//   - The volumetric divisor and billing step. Carrier data: the same for a
//     shoe and a jacket on one lane, different for the same shoe on two lanes.
//     Storing it per category created a second source of truth — exactly the
//     duty mistake again, one field over.
//
// [PHP] So với Symfony: đây là một Entity Doctrine có khoá tự nhiên
// [PHP] (#[ORM\Id] #[ORM\Column] private string $code) nhưng KHÔNG có setter
// [PHP] — muốn đổi thì tạo object mới và persist thay thế.
type CategoryPolicy struct {
	code         CategoryCode
	estimate     shared.ParcelSpec
	restrictions []Restriction
}

// NewCategoryPolicy validates operator input (convention 1). The estimate
// is required: a category the pricing context cannot quote from is a
// category that does not do its job.
func NewCategoryPolicy(code CategoryCode, estimate shared.ParcelSpec, restrictions []Restriction) (CategoryPolicy, error) {
	if code.IsZero() {
		return CategoryPolicy{}, ErrInvalidCategoryCode
	}
	if estimate.IsZero() {
		return CategoryPolicy{}, fmt.Errorf("category %q: %w", code, shared.ErrIncompleteParcelSpec)
	}
	rs, err := normaliseRestrictions(restrictions)
	if err != nil {
		return CategoryPolicy{}, fmt.Errorf("category %q: %w", code, err)
	}
	return CategoryPolicy{code: code, estimate: estimate, restrictions: rs}, nil
}

func normaliseRestrictions(in []Restriction) ([]Restriction, error) {
	out := make([]Restriction, 0, len(in))
	for _, r := range in {
		if !r.IsValid() {
			return nil, fmt.Errorf("restriction %q: %w", r, ErrUnknownRestriction)
		}
		// [PHP] `slices.Contains` = in_array(). Gói `slices` là thư viện chuẩn
		// [PHP] của Go từ 1.21, không phải thư viện ngoài.
		if slices.Contains(out, r) {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func (c CategoryPolicy) Code() CategoryCode {
	return c.code
}

// DefaultParcelSpec is what a product of this kind usually weighs and
// measures — what pricing quotes from until THIS product has been measured
// (Product.ParcelSpec). "Default" is the whole message: a fallback, not a
// fact about the item.
func (c CategoryPolicy) DefaultParcelSpec() shared.ParcelSpec {
	return c.estimate
}

func (c CategoryPolicy) IsZero() bool {
	return c.code.IsZero()
}

// Restrictions returns a copy, for the reason Merchant.Sourcing does.
func (c CategoryPolicy) Restrictions() []Restriction {
	return append([]Restriction(nil), c.restrictions...)
}

func (c CategoryPolicy) Restricted(r Restriction) bool {
	return slices.Contains(c.restrictions, r)
}

func (c CategoryPolicy) String() string {
	return fmt.Sprintf("%s (%s)", c.code, c.estimate)
}

// Defined is the fact that this policy is now in force, dated by the caller.
// CategoryPolicy is a value object with no event recorder of its own, so the
// use case that defines it asks the policy for its own announcement — the
// payload is composed here, in the domain, not assembled by hand in the app.
func (c CategoryPolicy) Defined(at time.Time) CategoryDefined {
	return CategoryDefined{Code: c.code, Estimate: c.estimate, Restrictions: c.Restrictions(), At: at}
}
