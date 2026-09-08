package httpapi

import (
	"fmt"
	"net/http"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// registerMerchantRequest is the JSON shape of POST /merchants. Strings all
// the way down: the wire does not know Hostname or Money, the adapter does
// the translating.
//
// [PHP] DTO cho FormType/Serializer. Tag `json:"..."` = #[SerializedName].
type registerMerchantRequest struct {
	Name         string               `json:"name"`
	Site         string               `json:"site"`
	Currency     string               `json:"currency"`
	FreeShipping *freeShippingRequest `json:"free_shipping"` // absent → the shop always charges
	Sourcing     []string             `json:"sourcing"`
}

// freeShippingRequest: kind is "never", "always" or "over"; threshold is a
// human-written amount, read in the request's locale, used only for "over".
type freeShippingRequest struct {
	Kind      string `json:"kind"`
	Threshold string `json:"threshold"`
}

// POST /merchants → 201 {"id"}.
func (s *server) registerMerchant(w http.ResponseWriter, r *http.Request) {
	var req registerMerchantRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	site, err := catalog.ParseHostname(req.Site)
	if err != nil {
		writeError(w, err)
		return
	}
	currency, err := shared.CurrencyFromCode(req.Currency)
	if err != nil {
		writeError(w, err)
		return
	}
	freeShip, err := freeShippingFrom(req.FreeShipping, currency, language(r))
	if err != nil {
		writeError(w, err)
		return
	}
	modes := make([]catalog.SourcingMode, 0, len(req.Sourcing))
	for _, m := range req.Sourcing {
		modes = append(modes, catalog.SourcingMode(m)) // the domain validates the value
	}

	id, err := s.register.Handle(r.Context(), catalog.MerchantDetails{
		Name:         req.Name,
		Site:         site,
		Currency:     currency,
		FreeShipping: freeShip,
		Sourcing:     modes,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, idResponse{ID: id.String()})
}

// freeShippingFrom builds the value object from the wire shape. The threshold
// is normalised by locale FIRST, then parsed strictly by the domain.
func freeShippingFrom(req *freeShippingRequest, currency shared.Currency, lang string) (catalog.FreeShipping, error) {
	if req == nil {
		return catalog.NoFreeShipping(), nil
	}
	switch req.Kind {
	case "never":
		return catalog.NoFreeShipping(), nil
	case "always":
		return catalog.AlwaysFreeShipping(), nil
	case "over":
		threshold, err := shared.ParseMoney(normalizeAmount(req.Threshold, lang), currency)
		if err != nil {
			return catalog.FreeShipping{}, err
		}
		return catalog.FreeShippingOver(threshold)
	default:
		return catalog.FreeShipping{}, fmt.Errorf("free_shipping.kind %q: want never, always or over: %w", req.Kind, errInvalidRequest)
	}
}
