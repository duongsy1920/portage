package shared_test

import (
	"errors"
	"testing"

	"github.com/duongsy/portage/internal/domain/shared"
)

func TestAllocate_partsAlwaysSumToTheTotal(t *testing.T) {
	usd := func(s string) shared.Money { return shared.MustParseMoney(s, shared.USD) }
	cases := []struct {
		total   string
		weights []int64
		want    []string
	}{
		{"100.00", []int64{1, 1, 1}, []string{"33.34", "33.33", "33.33"}},
		{"0.05", []int64{1, 1, 1}, []string{"0.02", "0.02", "0.01"}},
		{"87.50", []int64{2500, 2000}, []string{"48.61", "38.89"}}, // freight over two parcels by chargeable grams
		{"10.00", []int64{0, 1}, []string{"0.00", "10.00"}},        // a zero weight gets nothing
		{"-3.00", []int64{1, 2}, []string{"-1.00", "-2.00"}},       // a credit splits the same way
	}
	for _, c := range cases {
		parts, err := shared.Allocate(usd(c.total), c.weights)
		if err != nil {
			t.Fatalf("Allocate(%s, %v): %v", c.total, c.weights, err)
		}
		sum, _ := shared.Sum(parts[0], parts[1:]...)
		if sum != usd(c.total) {
			t.Errorf("Allocate(%s, %v) sums to %s", c.total, c.weights, sum)
		}
		for i := range parts {
			if parts[i] != usd(c.want[i]) {
				t.Errorf("Allocate(%s, %v)[%d] = %s, want %s", c.total, c.weights, i, parts[i], c.want[i])
			}
		}
	}
	for _, bad := range [][]int64{nil, {}, {0, 0}, {-1, 2}} {
		if _, err := shared.Allocate(usd("1.00"), bad); !errors.Is(err, shared.ErrNothingToAllocate) {
			t.Errorf("weights %v: got %v", bad, err)
		}
	}
	if _, err := shared.Allocate(shared.Money{}, []int64{1}); err == nil {
		t.Error("zero-value money must be refused")
	}
}
