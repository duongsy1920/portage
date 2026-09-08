// Package contracts is the PUBLISHED LANGUAGE (DDD.md §7): the versioned,
// plain-struct shapes that cross a context boundary. It is neither domain nor
// adapter — a domain package may not import another context, an adapter is
// too low to define meaning — so it sits on its own, imported by the codec
// that writes/reads the wire and by the app layer that consumes messages.
//
// Rules: primitives and other V1 structs only; a version in every type name;
// never a domain type. Growing a message = a new V1 field with a default the
// old readers ignore, or a V2 type — never a changed meaning under an old name.
//
// [PHP] Một package "Contracts"/"Messages" dùng chung giữa các bundle — DTO
// [PHP] đánh version, thay cho việc bundle này deserialize vào Entity bundle kia.
package contracts

import "time"

type MoneyV1 struct {
	Minor    int64  `json:"minor"`
	Currency string `json:"currency"`
}

type ParcelV1 struct {
	WeightG  int64 `json:"weight_g"`
	LengthMM int64 `json:"length_mm"`
	WidthMM  int64 `json:"width_mm"`
	HeightMM int64 `json:"height_mm"`
}

// ── catalog → others ─────────────────────────────────────────────────────────

// MerchantRegisteredV1: procurement keeps the shop's currency to check receipts.
type MerchantRegisteredV1 struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Site     string    `json:"site"`
	Currency string    `json:"currency"`
	At       time.Time `json:"at"`
}

type ProductPublishedV1 struct {
	ID       string    `json:"id"`
	Merchant string    `json:"merchant"`
	Category string    `json:"category"`
	Name     string    `json:"name"`
	Source   string    `json:"source"` // the page to buy from, for whoever has to buy it
	Price    MoneyV1   `json:"price"`
	Parcel   ParcelV1  `json:"parcel"`
	At       time.Time `json:"at"`
}

// VariantAddedV1 is the size and colour in the SHOP'S OWN WORDS. Free text on
// purpose: Nike unisex, EU/UK/JP numbering, apparel S/M/L and kids 4Y share no
// vocabulary, so a structured Size type would need a taxonomy per shop per
// category — and a wrong guess merges two real sizes into one. MerchantRef is
// the shop's own code, the field that makes an order unambiguous when it exists.
// ProductAddedV1 is a DRAFT product appearing. The reading side needs it to
// build the list of work still to do, which is why it carries the page and the
// price and not only the ids: a row a person cannot act on is not a worklist.
//
// RequestedBy is the customer waiting, and it is EMPTY when an operator added
// the product with nobody asking. Never treat it as required.
type ProductAddedV1 struct {
	ID          string    `json:"id"`
	Merchant    string    `json:"merchant"`
	Category    string    `json:"category"`
	Name        string    `json:"name"`
	Source      string    `json:"source"`
	Price       MoneyV1   `json:"price"`
	SourcedBy   string    `json:"sourced_by"`
	RequestedBy string    `json:"requested_by"`
	At          time.Time `json:"at"`
}

// ListingConfirmedV1 is an operator vouching for what a product is. The
// worklist consumes it to stop asking for a step somebody already did.
type ListingConfirmedV1 struct {
	ID string    `json:"id"`
	By string    `json:"by"`
	At time.Time `json:"at"`
}

type VariantAddedV1 struct {
	Product     string    `json:"product"`
	Variant     string    `json:"variant"`
	Size        string    `json:"size"`
	Color       string    `json:"color"`
	MerchantRef string    `json:"merchant_ref"`
	At          time.Time `json:"at"`
}

type ProductMeasuredV1 struct {
	ID       string    `json:"id"`
	Parcel   ParcelV1  `json:"parcel"`
	Verified bool      `json:"verified"`
	At       time.Time `json:"at"`
}

type ProductRepricedV1 struct {
	ID   string    `json:"id"`
	From MoneyV1   `json:"from"`
	To   MoneyV1   `json:"to"`
	At   time.Time `json:"at"`
}

type ProductRetiredV1 struct {
	ID     string    `json:"id"`
	Reason string    `json:"reason"`
	At     time.Time `json:"at"`
}

type CategoryDefinedV1 struct {
	Code         string    `json:"code"`
	Estimate     ParcelV1  `json:"estimate"`
	Restrictions []string  `json:"restrictions"`
	At           time.Time `json:"at"`
}
