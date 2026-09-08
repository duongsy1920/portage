package pricing

import (
	"fmt"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// QuoteID identifies a quote.
type QuoteID struct {
	shared.ID
}

func NewQuoteID() QuoteID {
	return QuoteID{shared.NewID()}
}

func ParseQuoteID(s string) (QuoteID, error) {
	id, err := shared.ParseID(s)
	if err != nil {
		return QuoteID{}, fmt.Errorf("quote id: %w", err)
	}
	return QuoteID{id}, nil
}

// QuoteStatus: a quote is a promise (issued) until the customer takes it
// (accepted) or the deadline passes (expired).
type QuoteStatus string

const (
	QuoteIssued   QuoteStatus = "issued"
	QuoteAccepted QuoteStatus = "accepted"
	QuoteExpired  QuoteStatus = "expired"
)

// Quote is the AGGREGATE ROOT of pricing: one price, for one product, on one
// lane, valid until a deadline — with every input it was computed from frozen
// inside (Breakdown). It is the "photograph" DDD.md §28 talks about: the FX
// rate, the rate card, the parcel used, the policy — none of it can change
// under the customer's feet.
//
// [PHP] Aggregate với con trỏ; Breakdown là một #[ORM\Embeddable] to. Không
// [PHP] có setter — Accept/Expire là hai hành vi duy nhất.
type Quote struct {
	shared.Events

	id        QuoteID
	product   shared.ID
	lane      LaneCode
	breakdown Breakdown
	issuedAt  time.Time
	expiresAt time.Time
	status    QuoteStatus
}

// IssueQuote computes and freezes a quote. `now` comes from the app clock and
// dates both the issue and the deadline (policy TTL).
func IssueQuote(in QuoteInputs, now time.Time) (*Quote, error) {
	b, err := Calculate(in)
	if err != nil {
		return nil, err
	}
	q := &Quote{
		id:        NewQuoteID(),
		product:   in.Listing.Product,
		lane:      in.Lane.Code(),
		breakdown: b,
		issuedAt:  now,
		expiresAt: now.Add(in.Policy.TTL()),
		status:    QuoteIssued,
	}
	q.Record(QuoteIssuedEvent{
		ID: q.id, Product: q.product, Lane: q.lane,
		TotalVND: b.TotalVND, Deposit: b.Deposit, Estimated: b.Estimated, ExpiresAt: q.expiresAt, At: now,
	})
	return q, nil
}

func (q *Quote) ID() QuoteID {
	return q.id
}

func (q *Quote) Product() shared.ID {
	return q.product
}

func (q *Quote) Lane() LaneCode {
	return q.lane
}

func (q *Quote) Breakdown() Breakdown {
	return q.breakdown
}

func (q *Quote) IssuedAt() time.Time {
	return q.issuedAt
}

func (q *Quote) ExpiresAt() time.Time {
	return q.expiresAt
}

func (q *Quote) Status() QuoteStatus {
	return q.status
}

func (q *Quote) IsExpired(now time.Time) bool {
	return now.After(q.expiresAt)
}

// Accept is the customer saying yes. After the deadline the quote is not a
// quote any more — the rate and the parcel may both have moved — so a late
// Accept flips it to expired and refuses. Ordering reacts to QuoteAccepted.
func (q *Quote) Accept(now time.Time) error {
	if q.status != QuoteIssued {
		return fmt.Errorf("accept quote %s: status %s: %w", q.id, q.status, ErrQuoteNotIssued)
	}
	if q.IsExpired(now) {
		q.status = QuoteExpired
		q.Record(QuoteExpiredEvent{ID: q.id, Product: q.product, At: now})
		return fmt.Errorf("accept quote %s: expired %s: %w", q.id, q.expiresAt.Format(time.RFC3339), ErrQuoteExpired)
	}
	q.status = QuoteAccepted
	q.Record(QuoteAcceptedEvent{ID: q.id, Product: q.product, TotalVND: q.breakdown.TotalVND, Deposit: q.breakdown.Deposit, At: now})
	return nil
}

// Expire is the sweep for quotes nobody answered: only an issued quote, only
// once its deadline has passed.
func (q *Quote) Expire(now time.Time) error {
	if q.status != QuoteIssued {
		return fmt.Errorf("expire quote %s: status %s: %w", q.id, q.status, ErrQuoteNotIssued)
	}
	if !q.IsExpired(now) {
		return fmt.Errorf("expire quote %s: valid until %s: %w", q.id, q.expiresAt.Format(time.RFC3339), ErrQuoteStillValid)
	}
	q.status = QuoteExpired
	q.Record(QuoteExpiredEvent{ID: q.id, Product: q.product, At: now})
	return nil
}

// QuoteSnapshot is Quote, flat — for the repository only (see catalog/snapshot.go).
type QuoteSnapshot struct {
	ID        QuoteID
	Product   shared.ID
	Lane      LaneCode
	Breakdown Breakdown
	IssuedAt  time.Time
	ExpiresAt time.Time
	Status    QuoteStatus
}

func (q *Quote) Snapshot() QuoteSnapshot {
	return QuoteSnapshot{
		ID: q.id, Product: q.product, Lane: q.lane, Breakdown: q.breakdown,
		IssuedAt: q.issuedAt, ExpiresAt: q.expiresAt, Status: q.status,
	}
}

func QuoteFromSnapshot(s QuoteSnapshot) (*Quote, error) {
	bad := func(why string) error { return fmt.Errorf("quote snapshot %s: %s: %w", s.ID, why, ErrInvalidSnapshot) }
	switch {
	case s.ID.IsZero():
		return nil, fmt.Errorf("quote snapshot: no id: %w", ErrInvalidSnapshot)
	case s.Product.IsZero(), s.Lane.IsZero():
		return nil, bad("missing product or lane")
	case s.Status != QuoteIssued && s.Status != QuoteAccepted && s.Status != QuoteExpired:
		return nil, bad(fmt.Sprintf("status %q", s.Status))
	case !s.Breakdown.TotalVND.IsValid(), !s.Breakdown.Deposit.IsValid(), s.Breakdown.Chargeable.IsZero():
		return nil, bad("incomplete breakdown")
	case !s.ExpiresAt.After(s.IssuedAt):
		return nil, bad("expires before issued")
	}
	return &Quote{
		id: s.ID, product: s.Product, lane: s.Lane, breakdown: s.Breakdown,
		issuedAt: s.IssuedAt, expiresAt: s.ExpiresAt, status: s.Status,
	}, nil
}
