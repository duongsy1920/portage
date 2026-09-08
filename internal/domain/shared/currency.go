package shared

import (
	"errors"
	"fmt"
)

var ErrUnknownCurrency = errors.New("unknown currency")

// Currency is a value object. Exponent is how many decimal digits the
// currency uses: USD has 2 (cents), VND has 0 (no subunit at all).
//
// A Money type that hardcodes 2 decimals silently breaks on VND, which is
// why the exponent travels with the currency instead.
//
// Symfony: moneyphp's Money\Currency carries only the code and looks the
// exponent up in an ISO table at runtime. Here it travels with the value so
// Money never has to consult anything.
type Currency struct {
	code     string
	exponent int32
}

// [PHP] `var (...)` = khai báo nhiều biến trong một khối cho gọn.
// [PHP] USD/VND ở đây đóng vai trò enum case của PHP 8.1:
// [PHP]     PHP: enum Currency: string { case USD = 'USD'; case VND = 'VND'; }
// [PHP] Go KHÔNG có enum. Cách thay thế là biến/hằng ở cấp package như dưới.
// [PHP]
// [PHP] `map[string]Currency{...}` = mảng kết hợp, key string, value Currency:
// [PHP]     PHP: ['USD' => $usd, 'VND' => $vnd]
// [PHP]     Go : map[string]Currency{"USD": USD, "VND": VND}
// [PHP] Đọc `map[KIỂU_KEY]KIỂU_VALUE` từ trái sang phải.
var (
	USD = Currency{code: "USD", exponent: 2}
	VND = Currency{code: "VND", exponent: 0}

	// currencies is the registry CurrencyFromCode reads. Adding a currency
	// is one line here — and nowhere else.
	currencies = map[string]Currency{USD.code: USD, VND.code: VND}
)

// CurrencyFromCode rebuilds a Currency from its persisted ISO code. It is
// case-sensitive on purpose: the only writer is Currency.Code(), so "usd"
// means corrupted data, not a user typo — the adapter normalises input
// before it gets here.
//
// Symfony: the convertToPHPValue() half of a Doctrine custom type.
func CurrencyFromCode(code string) (Currency, error) {
	// [PHP] `c, ok := m[key]` = "comma ok" — đọc map trả về HAI giá trị:
	// [PHP] giá trị, và `ok` (bool) cho biết key có tồn tại không.
	// [PHP]     PHP: if (!isset($currencies[$code])) { ... }
	// [PHP]          $c = $currencies[$code];
	// [PHP]     Go : c, ok := currencies[code]      ← gộp isset + đọc làm một
	// [PHP] Nếu key không có, `c` nhận giá trị rỗng (Currency{}) chứ KHÔNG lỗi
	// [PHP] như PHP notice "Undefined index" — nên phải kiểm `ok`.
	c, ok := currencies[code]
	if !ok {
		return Currency{}, fmt.Errorf("currency %q: %w", code, ErrUnknownCurrency)
	}
	return c, nil
}

func (c Currency) Code() string {
	return c.code
}
func (c Currency) Exponent() int32 {
	return c.exponent
}

// [PHP] Method tên `String() string` là QUY ƯỚC ĐẶC BIỆT của Go: fmt.Println
// [PHP] và `%s` sẽ tự gọi nó. Chính là `__toString()` bên PHP.
func (c Currency) String() string {
	return c.code
}

// IsZero reports the zero value Currency{} — a legal Go literal, but not a
// currency. Every struct has a zero value whether we like it or not; the
// constructors below refuse to build Money on it.
//
// [PHP] `c == Currency{}` — Go so sánh struct bằng `==`, THEO GIÁ TRỊ từng
// [PHP] field. PHP không có: `==` bên PHP so sánh lỏng, `===` so sánh xem có
// [PHP] phải cùng MỘT object không. Gần nhất là tự viết `equals()`.
// [PHP] Nhờ `==` này mà value object của Go so sánh rất tự nhiên:
// [PHP]     rateA == rateB   // true nếu mọi field bằng nhau
// [PHP] (Chỉ dùng được khi struct không chứa slice/map — hai thứ đó không so
// [PHP]  sánh được bằng `==`.)
func (c Currency) IsZero() bool {
	return c == Currency{}
}
