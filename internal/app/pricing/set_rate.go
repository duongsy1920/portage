package pricingapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/domain/shared"
)

// SetExchangeRateHandler publishes today's rate. It is the shortest use case in
// the system, and the interesting part is what it does NOT do.
//
// It announces NOTHING. Every other write in Portage ends with an event,
// because somebody downstream keeps a copy — but nobody keeps a copy of the
// rate. A Quote does not READ the rate later; it FROZE the rate it used into
// its own breakdown the moment it was issued (DDD.md §28). So changing today's
// rate changes only what the NEXT quote will say, which is exactly the point
// of a rate: it is the current answer to a question, not a fact about the past.
//
// The rule that makes this safe is in the domain, not here: shared.ExchangeRate
// refuses a non-positive rate and a from == to pair, so an operator cannot
// publish "1 USD = 0 VND" and take every quote after it to zero.
//
// [PHP] Một handler ba dòng — nhưng lý do KHÔNG dispatch event mới là phần
// [PHP] đáng đọc, không phải ba dòng kia.
type SetExchangeRateHandler struct {
	deps Deps
}

func NewSetExchangeRateHandler(d Deps) *SetExchangeRateHandler {
	mustHave("SetExchangeRateHandler", map[string]any{"UoW": d.UoW, "Rates": d.Rates})
	return &SetExchangeRateHandler{deps: d}
}

func (h *SetExchangeRateHandler) Handle(ctx context.Context, rate shared.ExchangeRate) error {
	return h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		if err := h.deps.Rates.Set(ctx, rate); err != nil {
			return fmt.Errorf("set rate %s: %w", rate, err)
		}
		return nil
	})
}
