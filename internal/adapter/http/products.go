package httpapi

import (
	"net/http"

	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// addProductRequest is the JSON shape of POST /products. currency travels
// with the price so the adapter can parse it; the use case then checks it
// against the merchant's (ErrPriceCurrency → 409).
// There is no sourced_by field: the provenance of a product is WHO pasted it,
// and that is the token's business, not the body's (auth.go, sourcingOf). A
// customer could otherwise claim their guess was typed by an operator, which
// is exactly the claim Product.Publish trusts.
type addProductRequest struct {
	Name       string `json:"name"`
	MerchantID string `json:"merchant_id"`
	Category   string `json:"category"`
	SourceURL  string `json:"source_url"`
	Price      string `json:"price"`
	Currency   string `json:"currency"`
}

// POST /products → 201 {"id"}.
func (s *server) addProduct(w http.ResponseWriter, r *http.Request) {
	var req addProductRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	merchant, err := catalog.ParseMerchantID(req.MerchantID)
	if err != nil {
		writeError(w, err)
		return
	}
	category, err := catalog.ParseCategoryCode(req.Category)
	if err != nil {
		writeError(w, err)
		return
	}
	source, err := catalog.ParseSourceURL(req.SourceURL)
	if err != nil {
		writeError(w, err)
		return
	}
	currency, err := shared.CurrencyFromCode(req.Currency)
	if err != nil {
		writeError(w, err)
		return
	}
	price, err := shared.ParseMoney(normalizeAmount(req.Price, language(r)), currency)
	if err != nil {
		writeError(w, err)
		return
	}
	// Both facts come from the token: an operator's paste is vouched for and
	// carries their id; a customer's is a draft nobody has checked yet.
	operator, _ := operatorOf(r) // zero for a customer, which is correct here

	id, err := s.add.Handle(r.Context(), catalogapp.AddProduct{
		Name:      req.Name,
		Merchant:  merchant,
		Category:  category,
		Source:    source,
		Price:     price,
		SourcedBy: sourcingOf(r),
		Operator:  operator,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, idResponse{ID: id.String()})
}

// POST /products/{id}/publish → 204. No body: the client asked for a state
// change and got it; there is nothing new to say.
func (s *server) publishProduct(w http.ResponseWriter, r *http.Request) {
	id, err := catalog.ParseProductID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.publish.Handle(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
