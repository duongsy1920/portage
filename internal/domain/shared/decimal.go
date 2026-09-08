package shared

// Exact decimal arithmetic shared by Money, Rate, ExchangeRate and Weight.
// Nothing in this file is exported: it is the machinery under the value
// objects, kept apart so money.go reads as business rules, not as long
// division.
//
// Symfony: the role Brick\Math (BigDecimal, RoundingMode::HALF_UP) plays under
// Brick\Money — you never call it directly, the value objects do.

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

var ErrMalformedAmount = errors.New("malformed amount")

// parseDecimal turns "-123.45" into the integer 12345 scaled by 10^maxFrac,
// i.e. it shifts the decimal point right by maxFrac places. It is the one
// strict number reader shared by Money, Rate and ExchangeRate.
//
// Grammar: -?digits[.digits]. Nothing else — see ParseMoney for why.
// [PHP] Tên hàm chữ THƯỜNG (`parseDecimal`) = private, chỉ dùng trong package
// [PHP] `shared`. Chữ HOA (`ParseMoney`) = public, package khác gọi được.
// [PHP] Đây thay cho `private`/`public` của PHP — không có từ khoá nào cả.
// [PHP]
// [PHP] Bảng hàm chuỗi hay dùng, đối chiếu PHP:
// [PHP]     strings.TrimSpace(s)        ~ trim($s)
// [PHP]     strings.HasPrefix(s, "-")   ~ str_starts_with($s, '-')
// [PHP]     strings.TrimPrefix(s, "-")  ~ ltrim($s, '-')  (chỉ bỏ 1 lần)
// [PHP]     strings.Repeat("0", n)      ~ str_repeat('0', $n)
// [PHP]     strings.ToUpper(s)          ~ strtoupper($s)
// [PHP]     strings.ReplaceAll(s, a, b) ~ str_replace($a, $b, $s)
// [PHP]     len(s)                      ~ strlen($s)
// [PHP] Chú ý: Go gọi `strings.X(s, ...)` — chuỗi là THAM SỐ, không phải method.
func parseDecimal(s string, maxFrac int) (int64, error) {
	s = strings.TrimSpace(s)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")

	// [PHP] `strings.Cut` trả về BA giá trị: phần trước, phần sau, có tìm thấy.
	// [PHP]     PHP: [$whole, $frac] = explode('.', $s, 2) + kiểm tra str_contains
	// [PHP]     Go : whole, frac, hasDot := strings.Cut(s, ".")
	// [PHP] Trả nhiều giá trị là chuyện bình thường ở Go, PHP phải trả array.
	whole, frac, hasDot := strings.Cut(s, ".")
	if whole == "" || !allDigits(whole) || (hasDot && (frac == "" || !allDigits(frac))) {
		return 0, ErrMalformedAmount
	}
	if len(frac) > maxFrac {
		return 0, fmt.Errorf("at most %d decimals: %w", maxFrac, ErrMalformedAmount)
	}
	frac += strings.Repeat("0", maxFrac-len(frac)) // pad "5" -> "50" for USD

	v, err := strconv.ParseInt(whole+frac, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%v: %w", err, ErrMalformedAmount)
	}
	if neg {
		v = -v
	}
	return v, nil
}

// [PHP] `for _, r := range s` trên CHUỖI = duyệt từng ký tự (rune).
// [PHP]     PHP: for ($i = 0; $i < strlen($s); $i++) { $r = $s[$i]; ... }
// [PHP] `r` kiểu `rune` (số nguyên đại diện một ký tự Unicode), nên so sánh
// [PHP] được với `'0'` — nháy ĐƠN là ký tự, nháy KÉP là chuỗi. PHP không phân
// [PHP] biệt: `'a'` và `"a"` đều là string.
// [PHP]     Go : 'a'  = rune (số)        |  "a" = string
// [PHP]     PHP: 'a' === "a"  → true
func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// [PHP] `for range n` (không có biến) = lặp n lần, bỏ chỉ số. Go 1.22+.
// [PHP]     PHP: for ($i = 0; $i < $n; $i++)
// [PHP]     Go : for range n
// [PHP] Go không có `while`; `for điều_kiện {}` chính là while.
func pow10(n int) int64 {
	p := int64(1)
	for range n {
		p *= 10
	}
	return p
}

func bigPow10(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

var bigOne = big.NewInt(1)

// divRoundHalfUp divides n by d (d > 0), rounding half-up away from zero,
// and panics if the result does not fit in int64 minor units.
func divRoundHalfUp(n, d *big.Int) int64 {
	q, rem := new(big.Int).QuoRem(n, d, new(big.Int)) // truncates toward zero
	twice := new(big.Int).Abs(rem)
	twice.Lsh(twice, 1) // |rem| * 2
	if twice.Cmp(d) >= 0 {
		if n.Sign() < 0 {
			q.Sub(q, bigOne)
		} else {
			q.Add(q, bigOne)
		}
	}
	if !q.IsInt64() {
		panic("shared: money overflow: " + q.String() + " minor units does not fit int64")
	}
	return q.Int64()
}

// addExact, subExact and mulExact are int64 arithmetic that refuses to wrap.
//
// Money.Mul and Money.Convert already route through math/big and panic rather
// than hand back a wrapped value (divRoundHalfUp above); these keep Add, Sub
// and Weight consistent with that, because the alternative is worse than a
// crash. int64 wrapping is silent: MaxInt64 + 1 is a large NEGATIVE number, so
// an overflowing sum becomes a negative balance, and an overflowing Weight
// becomes a negative weight — a value NewWeight exists to make impossible.
//
// Overflow here is always a bug, never a business case: 9.2 quintillion minor
// units is 92 quadrillion dollars, or every đồng in circulation many times
// over. So these panic, following the package rule that a programming error
// stops the program instead of quietly corrupting money.
func addExact(a, b int64) int64 {
	sum := a + b
	// Wrapping is only possible when both operands share a sign and the
	// result does not.
	if (a > 0 && b > 0 && sum < 0) || (a < 0 && b < 0 && sum >= 0) {
		panic(fmt.Sprintf("shared: int64 overflow: %d + %d", a, b))
	}
	return sum
}

func subExact(a, b int64) int64 {
	diff := a - b
	// Mirror image: only a difference of signs can carry the result out of
	// range. b == MinInt64 is covered — negating it is what would overflow.
	if (a >= 0 && b < 0 && diff < 0) || (a < 0 && b > 0 && diff >= 0) {
		panic(fmt.Sprintf("shared: int64 overflow: %d - %d", a, b))
	}
	return diff
}

func mulExact(a, b int64) int64 {
	product := a * b
	if a != 0 && (product/a != b || (a == -1 && b == minInt64)) {
		panic(fmt.Sprintf("shared: int64 overflow: %d * %d", a, b))
	}
	return product
}

const minInt64 = -1 << 63
