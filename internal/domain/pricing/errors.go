// Package pricing is the CORE bounded context: it turns "a product on a lane"
// into a number a customer can pay — and remembers every input that number was
// built from, so that when the real parcel is weighed weeks later the
// difference (Quote vs Actual, DDD.md §28) can be measured, not guessed.
//
// It never imports catalog. What it knows about products and categories it
// learned from events (Listing, CategoryProfile): the Published Language, not
// another context's model.
//
// [PHP] Một bundle riêng với model riêng; không `use App\Catalog\Entity\Product`
// [PHP] — chỉ có bản sao dữ liệu nó cần, cập nhật qua Messenger handler.
package pricing

import "errors"

var (
	ErrInvalidLaneCode   = errors.New("invalid lane code")
	ErrUnknownGoodsClass = errors.New("unknown goods class")
	ErrInvalidRate       = errors.New("rate must be positive")
	ErrInvalidLane       = errors.New("invalid shipping lane")
	ErrInvalidPolicy     = errors.New("invalid pricing policy")
	ErrListingInactive   = errors.New("listing is not active")
	ErrNothingToWeigh    = errors.New("no parcel to quote from: product unmeasured and category has no estimate")
	ErrQuoteExpired      = errors.New("quote has expired")
	ErrQuoteNotIssued    = errors.New("quote is not in the issued state")
	ErrQuoteStillValid   = errors.New("quote has not expired yet")
	ErrInvalidSnapshot   = errors.New("invalid snapshot")
)
