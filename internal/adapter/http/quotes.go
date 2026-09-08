package httpapi

import (
	"net/http"
	"strings"
	"time"

	pricingapp "github.com/duongsy/portage/internal/app/pricing"
	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The pricing routes. A quote is the first thing a CUSTOMER touches: ask for
// one, read it, accept it. Note what is not here — no lane definition, no FX
// rate endpoint: those are operator/config paths that arrive with auth (P9).

type issueQuoteRequest struct {
	ProductID string `json:"product_id"`
	Lane      string `json:"lane"`
}

// POST /quotes → 201 {"id"}. The numbers come from GET /quotes/{id}: a
// write answers with an id, a read answers with a view (DDD.md §24).
//
// 404 listing_not_found here has two honest meanings: the product was never
// published, or it was and the outbox relay has not yet delivered
// product_published to pricing. Eventual consistency has a status code.
func (s *server) issueQuote(w http.ResponseWriter, r *http.Request) {
	var req issueQuoteRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	product, err := shared.ParseID(req.ProductID)
	if err != nil {
		writeError(w, err)
		return
	}
	id, err := s.issue.Handle(r.Context(), pricingapp.IssueQuote{Product: product, Lane: req.Lane})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, idResponse{ID: id.String()})
}

// POST /quotes/{id}/accept → 204. A late accept is 409 quote_expired — and
// the quote IS now expired (the handler saved that before refusing).
func (s *server) acceptQuote(w http.ResponseWriter, r *http.Request) {
	id, err := pricing.ParseQuoteID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.accept.Handle(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// moneyView is money for a screen: major units as text ("150.00"), never a
// float — the client shows it, it does not add it up.
type moneyView struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

func viewOf(m shared.Money) moneyView {
	return moneyView{Amount: strings.TrimSuffix(m.String(), " "+m.Currency().Code()), Currency: m.Currency().Code()}
}

// quoteView is the READ MODEL of a quote: the breakdown line by line, the way
// a customer (or the reconciliation sheet) reads it. It is built from the
// aggregate's snapshot because pricing has no separate read store yet — the
// simplest CQRS there is: one write path (POST), one read shape (this).
type quoteView struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	Lane        string    `json:"lane"`
	Status      string    `json:"status"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Class       string    `json:"class"`
	ChargeableG int64     `json:"chargeable_g"`
	Estimated   bool      `json:"estimated"`
	Lines       struct {
		Item      moneyView `json:"item"`
		SalesTax  moneyView `json:"sales_tax"`
		Freight   moneyView `json:"freight"`
		Surcharge moneyView `json:"surcharge"`
		Duty      moneyView `json:"duty"`
		Subtotal  moneyView `json:"subtotal"`
	} `json:"lines"`
	FX   string `json:"fx"` // lane currency → home currency, as frozen in the quote
	Home struct {
		Subtotal   moneyView `json:"subtotal"`
		ServiceFee moneyView `json:"service_fee"`
		Total      moneyView `json:"total"`
		Deposit    moneyView `json:"deposit"`
	} `json:"home"`
}

// GET /quotes/{id} → 200 quoteView.
func (s *server) getQuote(w http.ResponseWriter, r *http.Request) {
	id, err := pricing.ParseQuoteID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	q, err := s.quotes.ByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	b := q.Breakdown()
	v := quoteView{
		ID: q.ID().String(), ProductID: q.Product().String(), Lane: q.Lane().String(), Status: string(q.Status()),
		IssuedAt: q.IssuedAt(), ExpiresAt: q.ExpiresAt(),
		Class: string(b.Class), ChargeableG: b.Chargeable.Grams(), Estimated: b.Estimated, FX: b.FX.Rate(),
	}
	v.Lines.Item, v.Lines.SalesTax, v.Lines.Freight = viewOf(b.ItemPrice), viewOf(b.SalesTax), viewOf(b.Freight)
	v.Lines.Surcharge, v.Lines.Duty, v.Lines.Subtotal = viewOf(b.Surcharge), viewOf(b.Duty), viewOf(b.SubtotalUSD)
	v.Home.Subtotal, v.Home.ServiceFee = viewOf(b.SubtotalVND), viewOf(b.ServiceFee)
	v.Home.Total, v.Home.Deposit = viewOf(b.TotalVND), viewOf(b.Deposit)
	writeJSON(w, http.StatusOK, v)
}
