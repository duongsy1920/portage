package pricing

import "github.com/duongsy/portage/internal/domain/shared"

// Listing is what pricing knows about a product — a PROJECTION built from
// catalog's events (product_published, product_measured, product_repriced,
// product_retired), never read from catalog's tables. Plain exported fields:
// this is a record pricing keeps, not an aggregate with invariants of its own;
// Calculate validates what it needs when it needs it.
//
// The product id is a shared.ID, not a catalog.ProductID: pricing cannot name
// catalog's types (guard 7). It is an opaque identity that came in over the wire.
//
// [PHP] Một "read model"/projection: bảng riêng của bundle Pricing, được
// [PHP] MessageHandler cập nhật khi nhận event từ Catalog. Không phải Entity
// [PHP] của Catalog và không join sang bảng của Catalog.
type Listing struct {
	Product  shared.ID
	Name     string
	Category string // catalog's category code, as text
	Price    shared.Money
	Parcel   shared.ParcelSpec // what the product measured, when Measured
	Measured bool
	Active   bool // false once catalog retires it
}

// CategoryProfile is what pricing knows about a category, from
// catalog.category_defined: the default parcel for products nobody weighed
// yet, and the restrictions the lane may surcharge. Class is assigned by
// pricing's own Classification — it is the forwarder's vocabulary, not catalog's.
type CategoryProfile struct {
	Code         string
	Class        GoodsClass
	Estimate     shared.ParcelSpec
	Restrictions []string
}
