package httpapi

import (
	"net/http"

	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The three operator steps between a pasted product and a published one.
// Each is a POST on the product because each CHANGES it; none returns the
// product back (read models do that, DDD.md §24).

type addVariantRequest struct {
	Size        string `json:"size"`
	Color       string `json:"color"`
	MerchantRef string `json:"merchant_ref"`
}

// POST /products/{id}/variants → 201 {"id"}: the variant's id.
func (s *server) addVariant(w http.ResponseWriter, r *http.Request) {
	id, err := catalog.ParseProductID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req addVariantRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	vid, err := s.variant.Handle(r.Context(), catalogapp.AddVariant{
		Product: id, Size: req.Size, Color: req.Color, MerchantRef: req.MerchantRef,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, idResponse{ID: vid.String()})
}

// POST /products/{id}/confirm-listing → 204. No body; the operator is the
// bearer token, so a caller cannot confirm a listing in somebody else's name.
func (s *server) confirmListing(w http.ResponseWriter, r *http.Request) {
	id, err := catalog.ParseProductID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	operator, ok := operatorOf(r)
	if !ok {
		// requireOperator already guaranteed this; the check stays so that
		// moving the route out from behind it fails loudly instead of
		// recording the zero operator.
		writeError(w, errForbidden)
		return
	}
	if err := s.confirm.Handle(r.Context(), catalogapp.ConfirmListing{Product: id, Operator: operator}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parcelRequest is grams and millimetres as integers — the same shape as a
// category's estimate in POST /categories.
type parcelRequest struct {
	WeightG  int64 `json:"weight_g"`
	LengthMM int64 `json:"length_mm"`
	WidthMM  int64 `json:"width_mm"`
	HeightMM int64 `json:"height_mm"`
}

func (p parcelRequest) spec() (shared.ParcelSpec, error) {
	weight, err := shared.NewWeight(p.WeightG)
	if err != nil {
		return shared.ParcelSpec{}, err
	}
	if p.LengthMM < 0 || p.WidthMM < 0 || p.HeightMM < 0 {
		return shared.ParcelSpec{}, errInvalidRequest
	}
	return shared.NewParcelSpec(weight, shared.NewDimensionsMM(p.LengthMM, p.WidthMM, p.HeightMM))
}

// POST /products/{id}/measure → 204: what the item weighed on our scale.
func (s *server) measureProduct(w http.ResponseWriter, r *http.Request) {
	id, err := catalog.ParseProductID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req parcelRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	spec, err := req.spec()
	if err != nil {
		writeError(w, err)
		return
	}
	operator, ok := operatorOf(r)
	if !ok {
		// requireOperator already guaranteed this; the check stays so that
		// moving the route out from behind it fails loudly instead of
		// recording the zero operator.
		writeError(w, errForbidden)
		return
	}
	if err := s.measure.Handle(r.Context(), catalogapp.MeasureProduct{Product: id, Spec: spec, Operator: operator}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// operatorOf now lives in auth.go and reads the BEARER TOKEN. The header
// version that stood here is deliberately gone: X-Operator-ID was a value the
// CALLER wrote, and "who vouched for this listing" must never be.
