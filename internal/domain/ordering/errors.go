package ordering

import "errors"

// One sentinel per reason (convention: a refusal names its cause), so the app
// layer and the HTTP table can tell "not yet paid" from "already delivered".
var (
	ErrInvalidOrder     = errors.New("invalid order")
	ErrQuoteNotAccepted = errors.New("quote is not accepted")
	ErrQuoteAlreadyUsed = errors.New("quote already has an order")

	// ErrVariantUnknown: ordering has never heard of this variant. Two causes,
	// one answer: it does not exist, or catalog.variant_added has not been
	// relayed yet. The caller cannot tell them apart and does not need to —
	// both mean "try again, and if it keeps failing the id is wrong".
	ErrVariantUnknown = errors.New("variant is unknown here")

	// ErrVariantNotForProduct: the size belongs to a different product than
	// the quote. Ordering can only catch this because it keeps its own
	// projection of variants; before that, any uuid was accepted.
	ErrVariantNotForProduct = errors.New("variant belongs to another product")
	ErrWrongAmount          = errors.New("payment does not match the amount due")
	ErrNotAwaitingDeposit   = errors.New("order is not awaiting its deposit")
	ErrNotDeposited         = errors.New("order has no deposit yet")
	ErrNotPurchased         = errors.New("order has not been purchased")
	ErrNotInTransit         = errors.New("order is not in transit")
	ErrBalanceUnpaid        = errors.New("balance has not been paid")
	ErrAlreadyDelivered     = errors.New("order already delivered")
	ErrOrderCancelled       = errors.New("order is cancelled")
	ErrEmptyReason          = errors.New("a reason is required")
	ErrInvalidSnapshot      = errors.New("invalid order snapshot")

	// ErrNotOwner: this order belongs to another customer.
	//
	// It lives in the domain, not the HTTP adapter, because "only the
	// customer whose order it is may cancel it" is a business rule; an
	// adapter that checked it would be the one place the rule could be
	// forgotten when a second caller appears (P9-PLAN §4, câu 2).
	//
	// The HTTP table answers it with 404, not 403: telling a stranger "this
	// is not yours" confirms the order exists.
	ErrNotOwner = errors.New("order belongs to another customer")
)
