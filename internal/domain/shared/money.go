// Package shared holds the Shared Kernel: value objects every bounded
// context depends on.
//
// RULES FOR THIS PACKAGE (and every package under internal/domain):
//   - It may NOT import a database driver, an HTTP framework, or any SDK.
//   - It may NOT import another bounded context.
//   - Everything here is immutable: methods return new values, never mutate.
//     (The one exception is Events, which is bookkeeping, not a value.)
//
// Two kinds of constructor live here, and the split is deliberate:
//
//   - Validating constructors — ParseMoney, ParsePercent, NewExchangeRate,
//     NewWeight, CurrencyFromCode, ParseID — read UNTRUSTED input (a form
//     field, a CSV cell, a database row) and return an error.
//   - Literal constructors — NewMoney, Grams, Kilos, NewDimensionsCM, RatePPM
//     — and the Must* helpers trust the PROGRAMMER and panic on impossible
//     input, the way Go itself panics on a negative slice length. A panic
//     here is a bug in our code, never a user mistake.
//
// Symfony: think src/Shared/Domain/ValueObject/*.php — except nothing here
// knows Doctrine exists. There is no #[ORM\Embeddable] on these structs; the
// column mapping is written once, in adapter/postgres, and the domain never
// sees it. Error values (ErrCurrencyMismatch...) play the role your typed
// exceptions play, and errors.Is() is your `catch (CurrencyMismatch $e)`.
package shared

import (
	"errors"
	"fmt"
	"math/big"
)

// [PHP] `var X = ...` ở cấp package = hằng số toàn cục của namespace.
// [PHP] Đây gọi là "sentinel error" — đóng vai trò class exception riêng:
// [PHP]     PHP: final class CurrencyMismatch extends \DomainException {}
// [PHP]     Go : var ErrCurrencyMismatch = errors.New("currency mismatch")
// [PHP] Quy ước: biến lỗi luôn bắt đầu bằng `Err`. Chữ E HOA = public.
var ErrCurrencyMismatch = errors.New("currency mismatch")

// Money is an immutable amount in a single currency.
//
// The amount is held in MINOR UNITS as an integer (USD cents, VND đồng).
// Floating point is never used: 0.1 + 0.2 != 0.3 in binary floating point,
// and money that is off by a cent per row is money lost at scale.
//
// Symfony: Brick\Money\Money or moneyphp — both also keep minor units and
// refuse float. Doctrine's DECIMAL coming back as a string is the same
// instinct. Here the type system enforces it: there is no float64 anywhere
// in the package.
// [PHP] `type X struct {}` = class chỉ có property, KHÔNG có method bên trong.
// [PHP] Method viết TÁCH RỜI bên dưới (xem `func (m Money) Add(...)`).
// [PHP]     PHP: final class Money { private int $minor; private Currency $currency; }
// [PHP]     Go : type Money struct { minor int64; currency Currency }
// [PHP] Kiểu viết SAU tên biến: `minor int64` ~ `int $minor`. Ngược với PHP.
// [PHP] Chữ thường (minor, currency) = private trong package. Không có từ khoá
// [PHP] `private`/`public` — Go dùng chữ HOA/thường của ký tự đầu để quyết định.
type Money struct {
	minor    int64
	currency Currency
}

// NewMoney builds an amount directly from minor units — the constructor the
// postgres adapter uses when it reads an amount column. The currency must
// be real; passing Currency{} is a programming error and panics.
//
// [PHP] Go KHÔNG có `__construct`. Quy ước: hàm thường tên `NewXxx` trả về Xxx.
// [PHP]     PHP: public function __construct(int $minor, Currency $c)
// [PHP]     Go : func NewMoney(minor int64, c Currency) Money
// [PHP] `Money` cuối dòng là KIỂU TRẢ VỀ (PHP viết `: Money` sau ngoặc).
// [PHP] `panic(...)` ~ `throw new \LogicException(...)` — dừng chương trình.
func NewMoney(minor int64, c Currency) Money {
	if c.IsZero() {
		panic("shared: NewMoney called with the zero Currency")
	}
	// [PHP] `Money{minor: ..., currency: ...}` = tạo object + gán field ngay,
	// [PHP] tương đương `new Money(minor: ..., currency: ...)` (named arguments).
	// [PHP] KHÔNG cần từ khoá `new`.
	return Money{minor: minor, currency: c}
}

func Zero(c Currency) Money {
	return NewMoney(0, c)
}

// ParseMoney reads a human-written decimal amount ("150.00", "3900000")
// exactly, without ever going through a float.
//
// Only the canonical form -?digits[.digits] is accepted. No thousands
// separators, no "+", no bare "150." or ".5": a decimal comma means 1/100
// in Vietnam and 1000 in the US, and a domain that guesses will be wrong
// for one of them. Normalising user input is the adapter's job.
//
// Symfony: Brick\Math\BigDecimal::of($string) with a fixed scale — never
// (float) $string.
//
// [PHP] `(Money, error)` = TRẢ VỀ HAI GIÁ TRỊ. Đây là điểm khác PHP lớn nhất.
// [PHP] Go không có exception, nên hàm nào có thể hỏng đều trả kèm `error`.
// [PHP]     PHP: function parseMoney(string $s): Money  // throws
// [PHP]     Go : func ParseMoney(s string) (Money, error)
// [PHP] Người gọi BẮT BUỘC nhận cả hai (hoặc viết `_` để vứt), nên không thể
// [PHP] "quên try/catch" như PHP.
func ParseMoney(s string, c Currency) (Money, error) {
	// [PHP] `:=` = khai báo biến MỚI + gán, Go tự suy ra kiểu.
	// [PHP]     PHP: $minor = ...; $err = ...;
	// [PHP]     Go : minor, err := parseDecimal(...)     ← nhận 2 giá trị cùng lúc
	// [PHP] (`=` dùng khi biến ĐÃ khai báo rồi; `:=` khi tạo mới.)
	// [PHP] `int(c.exponent)` = ép kiểu, ~ `(int) $c->exponent`.
	minor, err := parseDecimal(s, int(c.exponent))
	if err != nil {
		// [PHP] Đây là "trả lỗi", KHÔNG phải throw. Hàm kết thúc tại đây.
		// [PHP] `Money{}` = giá trị rỗng của Money (mọi field = 0/nil).
		// [PHP] `%w` trong Errorf = bọc lỗi gốc lại, để sau này
		// [PHP] `errors.Is(err, ErrMalformedAmount)` vẫn nhận ra — giống
		// [PHP] `throw new X("...", previous: $e)` rồi `catch (X)` bên PHP.
		// [PHP] `%q` = in chuỗi kèm dấu nháy; `%s` = in thường.
		return Money{}, fmt.Errorf("parse %q as %s: %w", s, c.code, err)
	}
	// [PHP] `nil` ~ `null`. Trả `nil` ở vị trí error = "không có lỗi".
	return NewMoney(minor, c), nil
}

// MustParseMoney is for tests and constants only. It panics on bad input.
//
// [PHP] Quy ước Go: tiền tố `Must` = "nếu hỏng thì panic, đừng trả lỗi".
// [PHP] Dùng cho hằng số và test, nơi input do lập trình viên viết ra.
func MustParseMoney(s string, c Currency) Money {
	m, err := ParseMoney(s, c)
	if err != nil {
		panic(err)
	}
	return m
}

// [PHP] `func (m Money) Minor() int64` — phần `(m Money)` gọi là RECEIVER,
// [PHP] chính là `$this` của PHP nhưng ĐẶT TÊN TƯỜNG MINH.
// [PHP]     PHP: public function minor(): int { return $this->minor; }
// [PHP]     Go : func (m Money) Minor() int64 { return m.minor }
// [PHP] `m` đóng vai trò `$this`; đặt tên gì cũng được, quy ước là 1 chữ cái.
// [PHP] Method viết NGOÀI struct, không nằm trong dấu ngoặc như PHP.
//
// [PHP] `(m Money)` — receiver dạng GIÁ TRỊ: Go COPY cả struct khi gọi, nên
// [PHP] method không sửa được bản gốc → đúng tinh thần value object bất biến.
// [PHP] Nếu viết `(m *Money)` (con trỏ) thì mới sửa được bản gốc, giống PHP
// [PHP] object luôn truyền theo tham chiếu. Xem `Events` trong event.go.
func (m Money) Minor() int64 {
	return m.minor
}
func (m Money) Currency() Currency {
	return m.currency
}
func (m Money) IsZero() bool {
	return m.minor == 0
}
func (m Money) IsNegative() bool {
	return m.minor < 0
}

// IsPositive: strictly more than nothing — what a price, a deposit or a rate
// must be. Zero is not positive; the zero value Money{} is not either.
func (m Money) IsPositive() bool {
	return m.minor > 0
}

// IsValid reports whether m carries a currency. The zero value Money{} is
// "unset" — like a nullable Doctrine column before assignment — and Add, Sub
// and Sum reject it as a currency mismatch, including against another zero
// value. Convert rejects it too, since no ExchangeRate has a zero source.
func (m Money) IsValid() bool {
	return !m.currency.IsZero()
}

// combinable reports whether two amounts may be added or subtracted. Two
// things stop them, and both are currency mismatches:
//
//   - either side is the zero value Money{}, which carries no currency at
//     all. Without this check two zero values would "agree" with each other
//     and yield a currencyless result that prints as "0 " and spreads;
//   - the currencies are real but different.
func (m Money) combinable(o Money, verb string) error {
	if !m.IsValid() || !o.IsValid() {
		return fmt.Errorf("%s: operand has no currency (zero Money): %w",
			verb, ErrCurrencyMismatch)
	}
	if m.currency != o.currency {
		return fmt.Errorf("%s %s and %s: %w",
			verb, m.currency, o.currency, ErrCurrencyMismatch)
	}
	return nil
}

// Add refuses to add different currencies. USD and VND are not
// interchangeable, so the type system and the domain both say no.
//
// A currency mismatch is an error for the caller to handle. An int64
// overflow is not: it is a bug, and addExact panics rather than hand back a
// negative balance — the same choice Mul and Convert already make.
//
// Symfony: moneyphp throws InvalidArgumentException here. Go has no
// exceptions; the error is a return value the caller cannot ignore without
// writing `_`, which is visible in code review.
func (m Money) Add(o Money) (Money, error) {
	// [PHP] `if err := ...; err != nil` — khai báo biến NGAY TRONG if, rồi
	// [PHP] mới kiểm tra. `err` chỉ sống trong khối if này (scope hẹp).
	// [PHP]     PHP: if (($err = $this->combinable($o, 'add')) !== null) {
	// [PHP] Đây là cách viết chuẩn của Go, gặp ở khắp nơi. Dấu `;` ngăn
	// [PHP] phần "khai báo" với phần "điều kiện".
	if err := m.combinable(o, "add"); err != nil {
		return Money{}, err
	}
	// [PHP] Trả về Money MỚI, không sửa `m`. Bất biến (immutable):
	// [PHP]     $a->add($b) bên PHP thường sửa $a;  bên đây thì không đụng tới.
	return Money{minor: addExact(m.minor, o.minor), currency: m.currency}, nil
}

func (m Money) Sub(o Money) (Money, error) {
	if err := m.combinable(o, "subtract"); err != nil {
		return Money{}, err
	}
	return Money{minor: subExact(m.minor, o.minor), currency: m.currency}, nil
}

// Sum adds many amounts, failing on the first currency mismatch.
//
// [PHP] `rest ...Money` = tham số biến thiên, y hệt `...Money $rest` bên PHP.
// [PHP] Gọi: Sum(a, b, c)  hoặc  Sum(a, list...)  (`...` để bung slice ra).
func Sum(first Money, rest ...Money) (Money, error) {
	total := first
	// [PHP] `for _, m := range rest` = foreach.
	// [PHP]     PHP: foreach ($rest as $i => $m)
	// [PHP]     Go : for i, m := range rest
	// [PHP] `_` là "biến vứt đi" — ở đây không cần chỉ số nên dùng `_`.
	// [PHP] Go BẮT LỖI biến khai báo mà không dùng, nên phải viết `_`.
	// [PHP] Go chỉ có MỘT từ khoá lặp là `for` (thay cho for/foreach/while).
	for _, m := range rest {
		// [PHP] `var err error` = khai báo trước, chưa gán (giá trị = nil).
		// [PHP] Cần vì dòng dưới dùng `=` chứ không `:=` — `total` đã tồn tại.
		var err error
		if total, err = total.Add(m); err != nil {
			return Money{}, err
		}
	}
	return total, nil
}

// Mul applies a rate (a tax rate, a service fee, a margin) and rounds
// half-up to the currency's smallest unit — half-up AWAY FROM ZERO, so the
// refund of a taxed amount is exactly the tax charged with the sign flipped.
//
// The intermediate product goes through math/big: 1<<62 cents times a 100%
// rate used to wrap an int64 to exactly 0. If the ROUNDED RESULT itself does
// not fit int64 (92 quadrillion dollars) the method panics: that is a bug,
// not a business case, and a bug must not come back as 0.00 USD.
func (m Money) Mul(r Rate) Money {
	n := new(big.Int).Mul(big.NewInt(m.minor), big.NewInt(r.ppm))
	return Money{minor: divRoundHalfUp(n, big.NewInt(ppmScale)), currency: m.currency}
}

// Convert changes currency at the given rate. The result is a DIFFERENT
// value object; the rate is a value the caller (a Quote) must keep, because
// a quote has to remember the rate it was priced at.
//
//	to.minor = from.minor × rate × 10^to.exponent ÷ 10^from.exponent
//
// rate is exact decimal digits over 10^scale (see ExchangeRate), so the
// whole thing is one integer multiply and one rounded divide. No float.
func (m Money) Convert(rate ExchangeRate) (Money, error) {
	if m.currency != rate.from {
		return Money{}, fmt.Errorf("convert %s with rate %s: %w",
			m.currency, rate, ErrCurrencyMismatch)
	}
	num := new(big.Int).Mul(big.NewInt(m.minor), big.NewInt(rate.digits))
	num.Mul(num, bigPow10(int(rate.to.exponent)))
	den := new(big.Int).Mul(bigPow10(int(rate.scale)), bigPow10(int(rate.from.exponent)))
	return Money{minor: divRoundHalfUp(num, den), currency: rate.to}, nil
}

func (m Money) String() string {
	exp := int(m.currency.exponent)
	if exp == 0 {
		return fmt.Sprintf("%d %s", m.minor, m.currency.code)
	}
	sign, v := "", m.minor
	if v < 0 {
		sign, v = "-", -v
	}
	unit := pow10(exp)
	return fmt.Sprintf("%s%d.%0*d %s", sign, v/unit, exp, v%unit, m.currency.code)
}
