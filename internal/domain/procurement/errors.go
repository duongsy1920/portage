package procurement

import "errors"

var (
	ErrInvalidTask       = errors.New("invalid purchase task")
	ErrTaskNotOpen       = errors.New("purchase task is not open")
	ErrTaskNotFound      = errors.New("purchase task not found")
	ErrTaskAlreadyExists = errors.New("purchase task already exists for this order")
	ErrEmptyReason       = errors.New("a reason is required")
	ErrEmptyReference    = errors.New("the shop's order reference is required")
	ErrPaidCurrency      = errors.New("amount paid is not in the shop's currency")
	ErrInvalidSnapshot   = errors.New("invalid purchase task snapshot")
	ErrShopNotFound      = errors.New("shop not found")
	ErrItemNotFound      = errors.New("item not found")
	ErrVariantNotFound   = errors.New("variant not found")

	// ErrManualPurchase is a MerchantACL adapter saying "I cannot buy by myself
	// — a person must". The task then stays open for the buyer's to-do list.
	ErrManualPurchase = errors.New("this shop is bought from by hand")
)
