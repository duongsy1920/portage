package catalog

import (
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/domain/shared"
)

var (
	ErrFreeShipCurrency  = errors.New("free-shipping threshold is in the wrong currency")
	ErrNegativeThreshold = errors.New("free-shipping threshold cannot be negative")
)

// FreeShipping is a merchant's domestic free-shipping rule — the US leg, from
// the shop to our warehouse.
//
// Three states exist in the real world and a bare Money cannot tell them
// apart, which is exactly why this is its own value object:
//
//	never      the shop always charges for delivery
//	over $50   free once the basket reaches a threshold
//	always     free regardless of basket size
//
// Modelling this as `freeShipOver Money` would make "always free" and "never
// free" both look like the zero value — and the domain would silently pick
// one. See convention 9: a zero value is "unset", never a business answer.
type FreeShipping struct {
	kind      FreeShippingKind // "" and FreeShipNever both mean never — see Kind
	threshold shared.Money     // meaningful only when kind == FreeShipOver
}

// FreeShippingKind is the rule's shape as a store writes it: "never", "over",
// "always". Text, not iota, so a database row is readable without the code.
//
// [PHP] Kiểu "enum" như SourcingMode; Kind() là cái Doctrine sẽ ghi vào cột.
type FreeShippingKind string

const (
	FreeShipNever  FreeShippingKind = "never"
	FreeShipOver   FreeShippingKind = "over"
	FreeShipAlways FreeShippingKind = "always"
)

// NoFreeShipping: the shop always charges. This is the zero value of
// FreeShipping, deliberately — if someone forgets to set the rule, the system
// assumes we PAY, and quotes a little high. The opposite mistake quotes low
// and loses money on every order.
func NoFreeShipping() FreeShipping {
	return FreeShipping{} // the zero value IS "never", so FreeShipping{} == NoFreeShipping()
}

// AlwaysFreeShipping: the shop never charges.
func AlwaysFreeShipping() FreeShipping {
	return FreeShipping{kind: FreeShipAlways}
}

// FreeShippingOver: free once the basket reaches threshold. Validating
// constructor — the threshold comes from an operator form (convention 1).
func FreeShippingOver(threshold shared.Money) (FreeShipping, error) {
	if !threshold.IsValid() {
		return FreeShipping{}, fmt.Errorf("threshold has no currency: %w", ErrCurrencyRequired)
	}
	if threshold.IsNegative() {
		return FreeShipping{}, fmt.Errorf("threshold %s: %w", threshold, ErrNegativeThreshold)
	}
	return FreeShipping{kind: FreeShipOver, threshold: threshold}, nil
}

func MustFreeShippingOver(threshold shared.Money) FreeShipping {
	f, err := FreeShippingOver(threshold)
	if err != nil {
		panic(err)
	}
	return f
}

// validFor rejects a rule priced in a currency the merchant does not use — a
// $50 threshold on a merchant selling in VND is a data-entry mistake, and it
// would surface much later as a currency mismatch deep inside a quote.
func (f FreeShipping) validFor(c shared.Currency) error {
	if f.kind != FreeShipOver {
		return nil
	}
	if f.threshold.Currency() != c {
		return fmt.Errorf("threshold %s but merchant sells in %s: %w",
			f.threshold, c, ErrFreeShipCurrency)
	}
	return nil
}

// AppliesTo reports whether a basket of this size ships free.
//
// It returns an error rather than a bare false on a currency mismatch: a
// wrong answer here is money, and silence would hide the bug until someone
// wondered why every order carried a shipping charge.
func (f FreeShipping) AppliesTo(subtotal shared.Money) (bool, error) {
	switch f.kind {
	case FreeShipAlways:
		return true, nil
	case FreeShipOver:
		// compare below
	default:
		return false, nil // never, including the zero value
	}
	if subtotal.Currency() != f.threshold.Currency() {
		return false, fmt.Errorf("basket %s against threshold %s: %w",
			subtotal.Currency(), f.threshold.Currency(), shared.ErrCurrencyMismatch)
	}
	// free once the basket REACHES the threshold: "free over $50" in shop
	// copy almost always means $50.00 qualifies.
	return subtotal.Minor() >= f.threshold.Minor(), nil
}

// Threshold reports the amount and whether there is one. The bool is what
// keeps a caller from reading a meaningless zero for the "never"/"always"
// cases — the same "comma ok" shape Go uses for maps.
func (f FreeShipping) Threshold() (shared.Money, bool) {
	if f.kind != FreeShipOver {
		return shared.Money{}, false
	}
	return f.threshold, true
}

// Kind is what a store persists, next to Threshold(): the two together rebuild
// the rule through the constructors (NoFreeShipping / AlwaysFreeShipping /
// FreeShippingOver). The zero value reports FreeShipNever.
func (f FreeShipping) Kind() FreeShippingKind {
	if f.kind == "" {
		return FreeShipNever
	}
	return f.kind
}

func (f FreeShipping) String() string {
	switch f.kind {
	case FreeShipAlways:
		return "always free"
	case FreeShipOver:
		return "free over " + f.threshold.String()
	default:
		return "never free"
	}
}
