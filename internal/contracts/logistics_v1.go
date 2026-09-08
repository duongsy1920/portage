package contracts

import "time"

// ── logistics → others ───────────────────────────────────────────────────────

// AllocationV1 is one order's share of a shipped batch's freight.
type AllocationV1 struct {
	Parcel      string  `json:"parcel"`
	Order       string  `json:"order"`
	ChargeableG int64   `json:"chargeable_g"`
	Freight     MoneyV1 `json:"freight"`
}

// BatchShippedV1: ordering moves each order to in_transit; pricing records
// each order's ACTUAL freight for reconciliation.
type BatchShippedV1 struct {
	ID          string         `json:"id"`
	Lane        string         `json:"lane"`
	Freight     MoneyV1        `json:"freight"`
	Allocations []AllocationV1 `json:"allocations"`
	At          time.Time      `json:"at"`
}

// ParcelExpectedV1 / ParcelReceivedV1 feed the reporting read model's
// "where is my box" column. Ordering does not listen: an order's STATUS moves
// on purchase and on shipment, not on the warehouse's internal steps.

type ParcelExpectedV1 struct {
	ID        string    `json:"id"`
	Order     string    `json:"order"`
	Reference string    `json:"reference"`
	At        time.Time `json:"at"`
}

type ParcelReceivedV1 struct {
	ID     string    `json:"id"`
	Order  string    `json:"order"`
	Actual ParcelV1  `json:"actual"`
	By     string    `json:"by"`
	At     time.Time `json:"at"`
}
