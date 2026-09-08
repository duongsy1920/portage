package httpapi

import (
	"net/http"

	"github.com/duongsy/portage/internal/domain/shared"
)

// setRateRequest is today's rate as the operator types it: how many major
// units of `to` one major unit of `from` buys.
type setRateRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
	Rate string `json:"rate"`
}

// POST /fx → 204. Operator only, and deliberately without a GET twin: the one
// place a rate is READ is inside a quote, and there it is already frozen into
// the breakdown (GET /quotes/{id} shows it as "fx").
//
// Publishing a new rate does NOT reprice anything. Quotes already issued keep
// the rate they froze — they are photographs (DDD.md §28) — so this route
// changes only what the next quote will say.
func (s *server) setRate(w http.ResponseWriter, r *http.Request) {
	var req setRateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	from, err := shared.CurrencyFromCode(req.From)
	if err != nil {
		writeError(w, err)
		return
	}
	to, err := shared.CurrencyFromCode(req.To)
	if err != nil {
		writeError(w, err)
		return
	}
	// normalizeAmount, not ParseMoney: a rate is a ratio, not money — it has
	// no currency of its own — but "26.000,5" typed by a Vietnamese operator
	// needs the same locale treatment every number at the edge gets.
	rate, err := shared.NewExchangeRate(from, to, normalizeAmount(req.Rate, language(r)))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.rates.Handle(r.Context(), rate); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
