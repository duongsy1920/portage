package catalog

import (
	"fmt"
	"strings"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// VariantID identifies one purchasable form of a product — a size and a colour.
type VariantID struct {
	shared.ID
}

func NewVariantID() VariantID {
	return VariantID{shared.NewID()}
}

func ParseVariantID(s string) (VariantID, error) {
	id, err := shared.ParseID(s)
	if err != nil {
		return VariantID{}, fmt.Errorf("variant id: %w", err)
	}
	return VariantID{id}, nil
}

// VariantDetails is what an operator or customer supplies for one variant.
// Everything is optional: a one-size product has a single variant with
// nothing to say, and that is a legal variant — but only as the product's
// ONLY variant (Product.AddVariant, ErrUnnamedVariant).
type VariantDetails struct {
	Size        string // as the shop writes it: "US 9", "M", "42 EU"
	Color       string // as the shop writes it: "black", "Sail/Gum"
	MerchantRef string // the shop's own code for this exact variant — opaque to us
}

// Variant is a CHILD ENTITY of Product.
//
// It has an identity, because a customer orders THIS variant and procurement
// buys THIS one — but it has no life outside its product. Nobody holds a
// *Variant: the product hands out copies (Product.Variants) and is the only
// thing that adds one (Product.AddVariant). That is rule 1 of an aggregate
// (DDD.md §14): the outside touches the root, never the insides.
//
// [PHP] Doctrine: OneToMany(Product → Variant) với orphanRemoval=true, và
// [PHP] KHÔNG có VariantRepository. Receiver là giá trị `(v Variant)` vì
// [PHP] variant không tự đổi — muốn đổi thì đi qua Product.
type Variant struct {
	id          VariantID
	size        string
	color       string
	merchantRef string
	addedAt     time.Time
}

func (v Variant) ID() VariantID {
	return v.id
}

// unnamed reports a variant with neither size nor colour: legal for a
// one-size product, never legal beside a named one.
func (v Variant) unnamed() bool {
	return v.size == "" && v.color == ""
}

func (v Variant) Size() string {
	return v.size
}

func (v Variant) Color() string {
	return v.color
}

func (v Variant) MerchantRef() string {
	return v.merchantRef
}

func (v Variant) AddedAt() time.Time {
	return v.addedAt
}

// key is what makes two variants "the same": size and colour, ignoring case
// and ALL spacing. "US 9"/"black", " us 9 "/"BLACK" and "US9"/"black" are one
// variant, and so are "M 8 / W 9.5" and "M8/W9.5" — shops write the same size
// both ways, and two rows for one physical shoe is a variant the buyer can
// pick by mistake.
//
// Spacing is as far as this goes. Reordering ("8 M / 9.5 W") stays a different
// variant, because guessing that two differently written sizes are the same
// needs a size taxonomy per shop per category, and a wrong guess merges two
// real sizes into one.
func (v Variant) key() string {
	return variantKey(v.size, v.color)
}

func variantKey(size, color string) string {
	// [PHP] strings.Fields tách theo mọi khoảng trắng rồi Join lại bằng một
	// [PHP] dấu cách = preg_replace('/\s+/', ' ', trim($s)). "\x00" là dấu
	// [PHP] phân cách không thể xuất hiện trong size/color thật.
	norm := func(s string) string {
		return strings.ToLower(strings.Join(strings.Fields(s), ""))
	}
	return norm(size) + "\x00" + norm(color)
}
