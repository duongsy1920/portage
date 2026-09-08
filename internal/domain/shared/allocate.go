package shared

import (
	"errors"
	"fmt"
	"math/big"
)

var ErrNothingToAllocate = errors.New("nothing to allocate over")

// Allocate splits total across weights in proportion, in MINOR units, so the
// parts add up to total exactly — the one thing float division never gives.
// Largest-remainder: every part is floored, then the leftover minor units go,
// one each, to the parts that lost the most in flooring (ties: earlier first).
//
//	Allocate(100.00 USD, [1, 1, 1]) → 33.34, 33.33, 33.33
//
// Used to split one carrier invoice over the parcels in a batch: whatever the
// rule (by weight, by volume, by value) the weights express it, and this
// function guarantees the cents reconcile.
//
// [PHP] moneyphp: `$money->allocate([1, 1, 1])` — cùng thuật toán. Ở đây viết
// [PHP] ra 30 dòng để thấy vì sao không được chia rồi làm tròn từng phần.
func Allocate(total Money, weights []int64) ([]Money, error) {
	if !total.IsValid() {
		return nil, fmt.Errorf("allocate %s: %w", total, ErrCurrencyMismatch)
	}
	var sum int64
	for _, w := range weights {
		if w < 0 {
			return nil, fmt.Errorf("allocate: negative weight %d: %w", w, ErrNothingToAllocate)
		}
		sum = addExact(sum, w)
	}
	if len(weights) == 0 || sum == 0 {
		return nil, fmt.Errorf("allocate %s over %v: %w", total, weights, ErrNothingToAllocate)
	}

	parts := make([]Money, len(weights))
	remainders := make([]*big.Int, len(weights))
	var given int64
	bigTotal, bigSum := big.NewInt(total.minor), big.NewInt(sum)
	for i, w := range weights {
		num := new(big.Int).Mul(bigTotal, big.NewInt(w))
		q, r := new(big.Int).QuoRem(num, bigSum, new(big.Int))
		parts[i] = Money{minor: q.Int64(), currency: total.currency}
		remainders[i] = r
		given = addExact(given, parts[i].minor)
	}
	for left := total.minor - given; left != 0; {
		best := -1
		for i, r := range remainders {
			if r.Sign() == 0 {
				continue
			}
			if best == -1 || r.Cmp(remainders[best]) > 0 {
				best = i
			}
		}
		if best == -1 {
			break // cannot happen: leftover implies a non-zero remainder
		}
		step := int64(1)
		if left < 0 {
			step = -1
		}
		parts[best].minor += step
		remainders[best].SetInt64(0)
		left -= step
	}
	return parts, nil
}
