package httpapi

import (
	"net/http"

	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// defineLaneRequest is the carrier's price list as the operator types it.
type defineLaneRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Divisor  int64  `json:"divisor"`
	StepG    int64  `json:"step_g"`
	Currency string `json:"currency"`
	Rates    struct {
		Standard    string `json:"standard"`
		Branded     string `json:"branded"`
		Electronics string `json:"electronics"`
		Sensitive   string `json:"sensitive"`
	} `json:"rates_per_kg"`
	BatterySurcharge string `json:"battery_surcharge"` // "" = none
	DutyRatePercent  string `json:"duty_rate_percent"` // "" = bundled
}

// POST /lanes → 204: define or redefine a lane (PUT semantics, kept POST like
// every write). Quotes already issued keep their numbers — they are photographs.
func (s *server) defineLane(w http.ResponseWriter, r *http.Request) {
	var req defineLaneRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	code, err := pricing.ParseLaneCode(req.Code)
	if err != nil {
		writeError(w, err)
		return
	}
	cur, err := shared.CurrencyFromCode(req.Currency)
	if err != nil {
		writeError(w, err)
		return
	}
	lang := language(r)
	parse := func(s string) (shared.Money, error) {
		return shared.ParseMoney(normalizeAmount(s, lang), cur)
	}
	var perKg [4]shared.Money
	for i, raw := range []string{req.Rates.Standard, req.Rates.Branded, req.Rates.Electronics, req.Rates.Sensitive} {
		if perKg[i], err = parse(raw); err != nil {
			writeError(w, err)
			return
		}
	}
	rates, err := pricing.NewRateCard(perKg[0], perKg[1], perKg[2], perKg[3])
	if err != nil {
		writeError(w, err)
		return
	}
	step, err := shared.NewWeight(req.StepG)
	if err != nil {
		writeError(w, err)
		return
	}
	d := pricing.LaneDetails{Code: code, Name: req.Name, Divisor: req.Divisor, Step: step, Rates: rates}
	if req.BatterySurcharge != "" {
		if d.BatterySurcharge, err = parse(req.BatterySurcharge); err != nil {
			writeError(w, err)
			return
		}
	}
	if req.DutyRatePercent != "" {
		rate, err := shared.ParsePercent(req.DutyRatePercent)
		if err != nil {
			writeError(w, err)
			return
		}
		if d.Duty, err = pricing.DutyItemised(rate); err != nil {
			writeError(w, err)
			return
		}
	}
	if err := s.defineLane_.Handle(r.Context(), d); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// reconciliationView is Quote vs Actual for one order, in the lane currency.
type reconciliationView struct {
	OrderID  string `json:"order_id"`
	QuoteID  string `json:"quote_id,omitempty"`
	Complete bool   `json:"complete"`
	Quoted   struct {
		Goods       *moneyView `json:"goods,omitempty"`
		Freight     *moneyView `json:"freight,omitempty"`
		ChargeableG int64      `json:"chargeable_g,omitempty"`
	} `json:"quoted"`
	Actual struct {
		Goods       *moneyView `json:"goods,omitempty"`
		Freight     *moneyView `json:"freight,omitempty"`
		ChargeableG int64      `json:"chargeable_g,omitempty"`
	} `json:"actual"`
	Variance *moneyView `json:"variance,omitempty"` // quoted − actual; only when complete
}

func optMoney(m shared.Money) *moneyView {
	if !m.IsValid() {
		return nil
	}
	v := viewOf(m)
	return &v
}

// GET /reconciliations/{order} → 200 reconciliationView.
func (s *server) getReconciliation(w http.ResponseWriter, r *http.Request) {
	order, err := shared.ParseID(r.PathValue("order"))
	if err != nil {
		writeError(w, err)
		return
	}
	rec, err := s.reconciliations.ByOrder(r.Context(), order)
	if err != nil {
		writeError(w, err)
		return
	}
	v := reconciliationView{OrderID: rec.Order.String(), Complete: rec.Complete()}
	if !rec.Quote.IsZero() {
		v.QuoteID = rec.Quote.String()
	}
	v.Quoted.Goods, v.Quoted.Freight, v.Quoted.ChargeableG = optMoney(rec.QuotedGoods), optMoney(rec.QuotedFreight), rec.QuotedChargeable.Grams()
	v.Actual.Goods, v.Actual.Freight, v.Actual.ChargeableG = optMoney(rec.ActualGoods), optMoney(rec.ActualFreight), rec.ActualChargeable.Grams()
	if rec.Complete() {
		if variance, err := rec.Variance(); err == nil {
			v.Variance = optMoney(variance)
		}
	}
	writeJSON(w, http.StatusOK, v)
}
