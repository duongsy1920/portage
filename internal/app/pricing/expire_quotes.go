package pricingapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/domain/pricing"
)

// ExpireQuotesHandler is the sweep for quotes nobody answered. It is the first
// use case in the system with NO caller: no customer, no operator, no event —
// only the passage of time (cmd/worker -sweep). That makes it the clearest
// example of why the Clock is a port: "is this quote past its deadline" is a
// question about the domain's rules, and the answer must be testable without
// waiting 48 hours.
//
// Why the quote expires at all: pricing.Quote is a SNAPSHOT (DDD.md §28). The
// FX rate, the merchant's price and the parcel it froze are all facts about a
// moment. Letting a customer accept a two-week-old snapshot would mean selling
// at a price the business no longer has.
//
// [PHP] Một command chạy bằng cron (`bin/console app:quotes:expire`), nhưng ở
// [PHP] đây nó chạy trong worker sẵn có, và "bây giờ là mấy giờ" là port.
type ExpireQuotesHandler struct {
	deps  Deps
	limit int
}

// NewExpireQuotesHandler builds the sweep. limit caps ONE pass: a backlog is
// worked through over several passes rather than one transaction the size of
// the table, and a pass that always finds `limit` quotes is the signal to run
// the sweep more often.
func NewExpireQuotesHandler(d Deps, limit int) *ExpireQuotesHandler {
	mustHave("ExpireQuotesHandler", map[string]any{"Clock": d.Clock, "UoW": d.UoW, "Outbox": d.Outbox, "Quotes": d.Quotes})
	if limit <= 0 {
		limit = 100
	}
	return &ExpireQuotesHandler{deps: d, limit: limit}
}

// Handle runs one pass and reports how many quotes it expired.
//
// ONE TRANSACTION PER QUOTE, on purpose. A single transaction around the whole
// batch would be shorter code and worse behaviour: one corrupt row, or one
// quote a concurrent Accept has already moved on, would roll back the work
// done for every other quote in the pass. The batch is not an invariant —
// nothing in the business says these quotes expire together — so it must not
// be an atomic unit (DDD.md §14: transaction boundary = aggregate boundary).
func (h *ExpireQuotesHandler) Handle(ctx context.Context) (int, error) {
	now := h.deps.Clock.Now()

	// Read outside a transaction: the list is only a work list, and each item
	// is re-checked by the domain inside its own transaction below.
	due, err := h.deps.Quotes.IssuedBefore(ctx, now, h.limit)
	if err != nil {
		return 0, fmt.Errorf("sweep quotes: %w", err)
	}

	expired := 0
	for _, q := range due {
		id := q.ID()
		err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
			// Re-read INSIDE the transaction. Between the list above and this
			// line a customer may have accepted the quote; the copy in the
			// list would then be stale and saving it would undo their accept.
			fresh, err := h.deps.Quotes.ByID(ctx, id)
			if err != nil {
				return err
			}
			if err := fresh.Expire(now); err != nil {
				// The domain refused: somebody got there first (accepted, or
				// another sweep). Not an error — the outcome we wanted is
				// already the case. Anything else is a real failure.
				if errors.Is(err, pricing.ErrQuoteNotIssued) || errors.Is(err, pricing.ErrQuoteStillValid) {
					return errAlreadySettled
				}
				return err
			}
			if err := h.deps.Quotes.Save(ctx, fresh); err != nil {
				return fmt.Errorf("save quote: %w", err)
			}
			if err := h.deps.Outbox.Append(ctx, fresh.PullEvents()); err != nil {
				return fmt.Errorf("outbox: %w", err)
			}
			return nil
		})
		switch {
		case err == nil:
			expired++
		case errors.Is(err, errAlreadySettled):
			// counted as nothing done, and not a failure
		default:
			// One quote must not end the pass: the rest are unrelated. Report
			// how far we got alongside the error so the caller can log both.
			return expired, fmt.Errorf("expire quote %s: %w", id, err)
		}
	}
	return expired, nil
}

// errAlreadySettled rolls the transaction back without making the pass fail.
// It never leaves this file.
var errAlreadySettled = errors.New("quote already settled")
