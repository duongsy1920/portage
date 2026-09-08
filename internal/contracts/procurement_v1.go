package contracts

import "time"

// ── procurement → others ─────────────────────────────────────────────────────

// PurchaseConfirmedV1 moves an order past its point of no return (DDD.md §26).
type PurchaseConfirmedV1 struct {
	ID        string    `json:"id"`
	Order     string    `json:"order"`
	Reference string    `json:"reference"`
	Paid      MoneyV1   `json:"paid"`
	By        string    `json:"by"`
	At        time.Time `json:"at"`
}

// PurchaseFailedV1 asks ordering to compensate.
type PurchaseFailedV1 struct {
	ID     string    `json:"id"`
	Order  string    `json:"order"`
	Reason string    `json:"reason"`
	At     time.Time `json:"at"`
}
