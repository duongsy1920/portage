package pricing

import (
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// Domain events of pricing. The Event suffix keeps them apart from the
// QuoteStatus constants (QuoteIssued the state vs QuoteIssuedEvent the fact).
// Wire names are hand-written and stable, as everywhere.

// QuoteIssuedEvent carries the two numbers a customer acts on and whether the
// parcel was guessed — enough for ordering to show the offer without loading
// the quote.
type QuoteIssuedEvent struct {
	ID        QuoteID
	Product   shared.ID
	Lane      LaneCode
	TotalVND  shared.Money
	Deposit   shared.Money
	Estimated bool
	ExpiresAt time.Time
	At        time.Time
}

func (e QuoteIssuedEvent) EventName() string {
	return "pricing.quote_issued"
}

func (e QuoteIssuedEvent) OccurredAt() time.Time {
	return e.At
}

// QuoteAcceptedEvent is what ordering listens for: a customer said yes at
// these numbers; open an order and ask for the deposit.
type QuoteAcceptedEvent struct {
	ID       QuoteID
	Product  shared.ID
	TotalVND shared.Money
	Deposit  shared.Money
	At       time.Time
}

func (e QuoteAcceptedEvent) EventName() string {
	return "pricing.quote_accepted"
}

func (e QuoteAcceptedEvent) OccurredAt() time.Time {
	return e.At
}

type QuoteExpiredEvent struct {
	ID      QuoteID
	Product shared.ID
	At      time.Time
}

func (e QuoteExpiredEvent) EventName() string {
	return "pricing.quote_expired"
}

func (e QuoteExpiredEvent) OccurredAt() time.Time {
	return e.At
}

// LaneDefinedEvent tells other contexts how a lane counts weight (logistics
// needs it to split freight) — never the rates: those are pricing's business.
type LaneDefinedEvent struct {
	Code     LaneCode
	Name     string
	Divisor  int64
	Step     shared.Weight
	Currency shared.Currency
	At       time.Time
}

func (LaneDefinedEvent) EventName() string {
	return "pricing.lane_defined"
}

func (e LaneDefinedEvent) OccurredAt() time.Time {
	return e.At
}
