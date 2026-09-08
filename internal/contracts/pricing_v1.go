package contracts

import "time"

// ── pricing → others ─────────────────────────────────────────────────────────

// LaneDefinedV1: logistics keeps the divisor and the step to split freight.
type LaneDefinedV1 struct {
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	Divisor  int64     `json:"divisor"`
	StepG    int64     `json:"step_g"`
	Currency string    `json:"currency"`
	At       time.Time `json:"at"`
}

// QuoteAcceptedV1 is what ordering needs to place an order: the numbers the
// customer said yes to. Ordering never recomputes them.
type QuoteAcceptedV1 struct {
	ID      string    `json:"id"`
	Product string    `json:"product"`
	Total   MoneyV1   `json:"total"`
	Deposit MoneyV1   `json:"deposit"`
	At      time.Time `json:"at"`
}
