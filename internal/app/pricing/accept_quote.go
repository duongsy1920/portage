package pricingapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/domain/pricing"
)

// AcceptQuoteHandler: the customer says yes. Ordering will call this when an
// order is placed (P6); today it is also an endpoint.
//
// One subtlety worth reading twice: a LATE accept makes the quote expire, and
// that expiry is a state change with an event — it must be SAVED even though
// the command is refused. So the transaction commits and the domain's refusal
// is returned afterwards, the way the worker commits a partial pass.
type AcceptQuoteHandler struct {
	deps Deps
}

func NewAcceptQuoteHandler(d Deps) *AcceptQuoteHandler {
	mustHave("AcceptQuoteHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Quotes": d.Quotes})
	return &AcceptQuoteHandler{deps: d}
}

func (h *AcceptQuoteHandler) Handle(ctx context.Context, id pricing.QuoteID) error {
	now := h.deps.Clock.Now()
	var refused error
	err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		q, err := h.deps.Quotes.ByID(ctx, id)
		if err != nil {
			return err
		}
		if err := q.Accept(now); err != nil {
			if !errors.Is(err, pricing.ErrQuoteExpired) {
				return err // nothing changed; roll back, report
			}
			refused = err // it changed (issued → expired): save that, then report
		}
		if err := h.deps.Quotes.Save(ctx, q); err != nil {
			return fmt.Errorf("save quote: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, q.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return refused
}
