package catalog_test

import (
	"errors"
	"testing"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// Three real-world states that a bare Money cannot tell apart. This is the
// whole reason FreeShipping exists as its own value object.
func TestFreeShipping_threeStates(t *testing.T) {
	fifty := shared.MustParseMoney("50.00", shared.USD)
	basket := func(s string) shared.Money { return shared.MustParseMoney(s, shared.USD) }

	cases := []struct {
		name string
		rule catalog.FreeShipping
		at   shared.Money
		want bool
	}{
		{"never, small basket", catalog.NoFreeShipping(), basket("10.00"), false},
		{"never, huge basket", catalog.NoFreeShipping(), basket("999.00"), false},
		{"always, empty basket", catalog.AlwaysFreeShipping(), basket("0.00"), true},
		{"over 50, under", catalog.MustFreeShippingOver(fifty), basket("49.99"), false},
		{"over 50, exactly", catalog.MustFreeShippingOver(fifty), basket("50.00"), true},
		{"over 50, above", catalog.MustFreeShippingOver(fifty), basket("50.01"), true},
	}
	for _, c := range cases {
		got, err := c.rule.AppliesTo(c.at)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: AppliesTo(%s) = %v, want %v", c.name, c.at, got, c.want)
		}
	}
}

// The zero value must be the SAFE one. If someone forgets to set the rule,
// the system assumes we pay for delivery and quotes a little high; the
// opposite default quotes low and loses money on every order.
func TestFreeShipping_zeroValueIsNeverFree(t *testing.T) {
	var unset catalog.FreeShipping

	got, err := unset.AppliesTo(shared.MustParseMoney("999.00", shared.USD))
	if err != nil {
		t.Fatalf("AppliesTo: %v", err)
	}
	if got {
		t.Fatal("the zero value must mean 'never free', not 'always free'")
	}
	if _, ok := unset.Threshold(); ok {
		t.Error("the zero value has no threshold")
	}
}

// A wrong answer here is money, so a currency mismatch is an error rather
// than a quiet false.
func TestFreeShipping_currencyMismatchIsAnError(t *testing.T) {
	rule := catalog.MustFreeShippingOver(shared.MustParseMoney("50.00", shared.USD))

	_, err := rule.AppliesTo(shared.MustParseMoney("3900000", shared.VND))
	if !errors.Is(err, shared.ErrCurrencyMismatch) {
		t.Fatalf("got %v, want ErrCurrencyMismatch", err)
	}

	// "never" and "always" do not compare amounts, so they answer either way.
	if _, err := catalog.NoFreeShipping().AppliesTo(shared.MustParseMoney("1", shared.VND)); err != nil {
		t.Errorf("NoFreeShipping should not care about currency: %v", err)
	}
}

func TestFreeShippingOver_rejectsBadThreshold(t *testing.T) {
	if _, err := catalog.FreeShippingOver(shared.Money{}); !errors.Is(err, catalog.ErrCurrencyRequired) {
		t.Errorf("zero Money: got %v", err)
	}
	if _, err := catalog.FreeShippingOver(shared.NewMoney(-1, shared.USD)); !errors.Is(err, catalog.ErrNegativeThreshold) {
		t.Errorf("negative threshold: got %v", err)
	}
}

func TestFreeShipping_threshold(t *testing.T) {
	fifty := shared.MustParseMoney("50.00", shared.USD)
	rule := catalog.MustFreeShippingOver(fifty)

	got, ok := rule.Threshold()
	if !ok || got != fifty {
		t.Fatalf("Threshold() = %v, %v", got, ok)
	}
	if got, want := rule.String(), "free over 50.00 USD"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	if got := catalog.AlwaysFreeShipping().String(); got != "always free" {
		t.Errorf("String() = %q", got)
	}
}
