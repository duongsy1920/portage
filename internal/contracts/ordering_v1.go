package contracts

import "time"

// ── ordering → others ────────────────────────────────────────────────────────

// OrderPlacedV1: pricing learns which quote an order is on (Quote vs Actual).
type OrderPlacedV1 struct {
	ID       string    `json:"id"`
	Quote    string    `json:"quote"`
	Product  string    `json:"product"`
	Variant  string    `json:"variant"`
	Customer string    `json:"customer"`
	Total    MoneyV1   `json:"total"`
	Deposit  MoneyV1   `json:"deposit"`
	At       time.Time `json:"at"`
}

// DepositPaidV1 is procurement's signal to buy: which order, which product,
// which variant. The amount is the customer's deposit (home currency) — for
// the record, not for buying with.
type DepositPaidV1 struct {
	ID      string    `json:"id"`
	Quote   string    `json:"quote"`
	Product string    `json:"product"`
	Variant string    `json:"variant"`
	Amount  MoneyV1   `json:"amount"`
	At      time.Time `json:"at"`
}

// The rest of an order's life, for the reporting read model. Ordering itself
// needs none of these back — they exist because a screen does (DDD.md §24).

// OrderPurchasedV1: the buyer bought it. Past the point of no return.
type OrderPurchasedV1 struct {
	ID string    `json:"id"`
	At time.Time `json:"at"`
}

// OrderShippedV1: the box carrying this order left the warehouse.
type OrderShippedV1 struct {
	ID string    `json:"id"`
	At time.Time `json:"at"`
}

// BalancePaidV1: the other half of the money arrived.
type BalancePaidV1 struct {
	ID     string    `json:"id"`
	Amount MoneyV1   `json:"amount"`
	At     time.Time `json:"at"`
}

// OrderDeliveredV1: it reached the customer's door.
type OrderDeliveredV1 struct {
	ID string    `json:"id"`
	At time.Time `json:"at"`
}

// OrderCancelledV1 carries what the customer actually cares about: how much
// comes back, and whether the deposit was forfeited (bought already, §26).
type OrderCancelledV1 struct {
	ID        string    `json:"id"`
	Reason    string    `json:"reason"`
	Refund    MoneyV1   `json:"refund"`
	Forfeited bool      `json:"forfeited"`
	At        time.Time `json:"at"`
}
