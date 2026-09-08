package httpapi

import (
	"net/http"
	"time"

	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The READ side (DDD.md §24). These two routes are the payoff of everything
// the write side did with events: a screen that needs facts from five bounded
// contexts is one SELECT with no JOIN, because the projector already did the
// joining, once, as the facts arrived.
//
// Compare with GET /orders/{id} above it: that one reads ordering's own
// aggregate and can therefore only ever show what ordering knows — no product
// name, no shop reference, no idea where the box is.

// summaryView is one row of "my orders".
type summaryView struct {
	Order       string `json:"order_id"`
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name,omitempty"`
	VariantID   string `json:"variant_id,omitempty"`
	Status      string `json:"status"`
	Tracking    string `json:"tracking"`

	Total   moneyView  `json:"total"`
	Deposit moneyView  `json:"deposit"`
	Refund  *moneyView `json:"refund,omitempty"` // only once cancelled

	DepositPaid bool `json:"deposit_paid"`
	BalancePaid bool `json:"balance_paid"`
	Forfeited   bool `json:"forfeited"`

	// The shop's own order number. Operators only: it is an internal reference
	// a customer cannot use and would only invite them to contact the shop
	// directly about an order the shop thinks belongs to somebody else.
	ShopReference string `json:"shop_reference,omitempty"`

	PlacedAt    time.Time  `json:"placed_at"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
}

func summaryViewOf(s reportingapp.OrderSummary, staff bool) summaryView {
	v := summaryView{
		Order: s.Order.String(), ProductID: idText(s.Product), ProductName: s.ProductName, VariantID: idText(s.Variant),
		Status: string(s.Status), Tracking: string(s.Tracking),
		Total: viewOf(s.Total), Deposit: viewOf(s.Deposit),
		DepositPaid: s.DepositPaid, BalancePaid: s.BalancePaid, Forfeited: s.Forfeited,
		PlacedAt: s.PlacedAt,
	}
	if staff {
		v.ShopReference = s.ShopReference
	}
	if s.Refund.IsValid() && s.Status == ordering.StatusCancelled {
		refund := viewOf(s.Refund)
		v.Refund = &refund
	}
	if !s.DeliveredAt.IsZero() {
		t := s.DeliveredAt
		v.DeliveredAt = &t
	}
	if !s.CancelledAt.IsZero() {
		t := s.CancelledAt
		v.CancelledAt = &t
	}
	return v
}

// GET /me/orders → 200 [summaryView], newest first.
//
// "Me" is the token, not a path parameter: /orders/{customer_id} would be a
// route whose security depends on the handler remembering to compare two ids,
// and the day somebody forgets, every customer can read every other customer.
// A route that cannot NAME another customer cannot leak one.
func (s *server) myOrders(w http.ResponseWriter, r *http.Request) {
	customer, ok := customerOf(r)
	if !ok {
		writeError(w, errForbidden)
		return
	}
	rows, err := s.summaries.ByCustomer(r.Context(), customer)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, viewsOf(rows, false))
}

// GET /orders?status=deposited → 200 [summaryView]. Operator only: this is the
// work queue ("who is waiting for a deposit", "what is still in transit").
//
// An unknown status is 400, not an empty list: "?status=deposted" returning
// zero orders looks exactly like "nothing to do", which is the worst possible
// answer to a typo in an operations screen.
func (s *server) listOrders(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("status")
	var (
		rows []reportingapp.OrderSummary
		err  error
	)
	if raw == "" {
		rows, err = s.summaries.All(r.Context())
	} else {
		status := ordering.OrderStatus(raw)
		if !status.IsValid() {
			writeError(w, ordering.ErrInvalidOrder)
			return
		}
		rows, err = s.summaries.ByStatus(r.Context(), status)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, viewsOf(rows, true))
}

// viewsOf never returns nil: an empty list must serialise as [] and not null,
// or every client has to handle two shapes of "nothing".
func viewsOf(rows []reportingapp.OrderSummary, staff bool) []summaryView {
	out := make([]summaryView, 0, len(rows))
	for _, s := range rows {
		out = append(out, summaryViewOf(s, staff))
	}
	return out
}

// idText renders the zero ID as "" rather than a UUID of all zeros: a frame
// row has fields nothing has filled in yet, and a fake id on a screen is worse
// than a blank one.
func idText(id shared.ID) string {
	if id.IsZero() {
		return ""
	}
	return id.String()
}
