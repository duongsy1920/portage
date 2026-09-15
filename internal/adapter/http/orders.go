package httpapi

import (
	"net/http"
	"time"

	orderingapp "github.com/duongsy/portage/internal/app/ordering"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The ordering routes: the customer's side of a purchase. Payments arrive as
// "we received X" — an operator today, a gateway adapter tomorrow, same command.

// customer_id is in the body for the ONE caller allowed to fill it: an
// operator placing an order for a customer who is not holding a token
// (P10-PLAN.md). A customer sending it is not naming themselves — they
// already are, via the token — they are trying to order in someone else's
// name, which is refused outright (see placeOrder).
type placeOrderRequest struct {
	QuoteID    string `json:"quote_id"`
	VariantID  string `json:"variant_id"`
	CustomerID string `json:"customer_id"`
}

// POST /orders → 201 {"id"}.
func (s *server) placeOrder(w http.ResponseWriter, r *http.Request) {
	var req placeOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	ids := make([]shared.ID, 2)
	for i, raw := range []string{req.QuoteID, req.VariantID} {
		id, err := shared.ParseID(raw)
		if err != nil {
			writeError(w, err)
			return
		}
		ids[i] = id
	}
	cmd := orderingapp.PlaceOrder{Quote: ids[0], Variant: ids[1]}
	if customer, isCustomer := customerOf(r); isCustomer {
		// A customer places their own order. Sending customer_id alongside is
		// an attempt to order in somebody else's name — refused outright, not
		// silently ignored: silence would let a client believe it worked.
		if req.CustomerID != "" {
			writeError(w, errForbidden)
			return
		}
		cmd.Customer = customer
	} else {
		operator, ok := operatorOf(r)
		if !ok {
			// requireAny already guaranteed one kind or the other; the check
			// stays so that moving the route fails loudly instead of placing
			// an ownerless order.
			writeError(w, errForbidden)
			return
		}
		if req.CustomerID == "" {
			writeError(w, ordering.ErrCustomerRequired)
			return
		}
		customer, err := shared.ParseID(req.CustomerID)
		if err != nil {
			writeError(w, err)
			return
		}
		cmd.Customer, cmd.PlacedBy = customer, operator
	}
	id, err := s.place.Handle(r.Context(), cmd)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, idResponse{ID: id.String()})
}

type paymentRequest struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

func (s *server) payment(r *http.Request) (orderingapp.Payment, error) {
	id, err := ordering.ParseOrderID(r.PathValue("id"))
	if err != nil {
		return orderingapp.Payment{}, err
	}
	var req paymentRequest
	if err := decodeJSON(r, &req); err != nil {
		return orderingapp.Payment{}, err
	}
	cur, err := shared.CurrencyFromCode(req.Currency)
	if err != nil {
		return orderingapp.Payment{}, err
	}
	amount, err := shared.ParseMoney(normalizeAmount(req.Amount, language(r)), cur)
	if err != nil {
		return orderingapp.Payment{}, err
	}
	return orderingapp.Payment{Order: id, Amount: amount}, nil
}

// POST /orders/{id}/deposit {"amount","currency"} → 204.
func (s *server) payDeposit(w http.ResponseWriter, r *http.Request) {
	cmd, err := s.payment(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.deposit.Handle(r.Context(), cmd); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /orders/{id}/balance {"amount","currency"} → 204.
func (s *server) payBalance(w http.ResponseWriter, r *http.Request) {
	cmd, err := s.payment(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.balance.Handle(r.Context(), cmd); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type cancelRequest struct {
	Reason string `json:"reason"`
}

// POST /orders/{id}/cancel {"reason"} → 204. What was refunded is on the
// order (GET /orders/{id} → refund) — a write answers "done", a read answers "what".
func (s *server) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := ordering.ParseOrderID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req cancelRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	// A customer may only cancel their own; an operator cancels for anyone.
	// OnBehalfOf comes from the TOKEN and can never come from the body, so a
	// caller cannot send the zero that skips the check (see CancelOrder).
	onBehalfOf, _ := customerOf(r)
	cmd := orderingapp.CancelOrder{Order: id, Reason: req.Reason, OnBehalfOf: onBehalfOf}
	if _, err := s.cancel.Handle(r.Context(), cmd); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /orders/{id}/deliver → 204 (operator: the courier handed it over).
func (s *server) deliverOrder(w http.ResponseWriter, r *http.Request) {
	id, err := ordering.ParseOrderID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.deliver.Handle(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// orderView is the read model of an order.
type orderView struct {
	ID           string     `json:"id"`
	QuoteID      string     `json:"quote_id"`
	ProductID    string     `json:"product_id"`
	VariantID    string     `json:"variant_id"`
	CustomerID   string     `json:"customer_id"`
	Status       string     `json:"status"`
	Total        moneyView  `json:"total"`
	Deposit      moneyView  `json:"deposit"`
	Balance      moneyView  `json:"balance"`
	BalancePaid  bool       `json:"balance_paid"`
	Refund       *moneyView `json:"refund,omitempty"` // only once cancelled
	Forfeited    bool       `json:"forfeited"`
	CancelReason string     `json:"cancel_reason,omitempty"`
	PlacedAt     time.Time  `json:"placed_at"`
}

// GET /orders/{id} → 200 orderView.
func (s *server) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := ordering.ParseOrderID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	o, err := s.orders.ByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	// A customer sees only their own. The refusal is ErrNotOwner, which the
	// table answers with 404 — telling a stranger "not yours" would confirm
	// the order exists (P9-PLAN §4, câu 1).
	if customer, isCustomer := customerOf(r); isCustomer && o.Customer() != customer {
		writeError(w, ordering.ErrNotOwner)
		return
	}
	v := orderView{
		ID: o.ID().String(), QuoteID: o.Quote().String(), ProductID: o.Product().String(), VariantID: o.Variant().String(),
		CustomerID: o.Customer().String(), Status: string(o.Status()),
		Total: viewOf(o.Total()), Deposit: viewOf(o.Deposit()), Balance: viewOf(o.Balance()), BalancePaid: o.BalancePaid(),
		Forfeited: o.Refund().Forfeited(), CancelReason: o.CancelReason(), PlacedAt: o.PlacedAt(),
	}
	if o.Status() == ordering.StatusCancelled {
		refund := viewOf(o.Refund().Amount())
		v.Refund = &refund
	}
	writeJSON(w, http.StatusOK, v)
}
