package httpapi

import (
	"net/http"

	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// defineCategoryRequest is the JSON shape of POST /categories: grams and
// millimetres as integers — no unit guessing at the edge either.
type defineCategoryRequest struct {
	Code     string `json:"code"`
	Estimate struct {
		WeightG  int64 `json:"weight_g"`
		LengthMM int64 `json:"length_mm"`
		WidthMM  int64 `json:"width_mm"`
		HeightMM int64 `json:"height_mm"`
	} `json:"estimate"`
	Restrictions []string `json:"restrictions"`
}

// POST /categories → 204. Defining and redefining look the same to the client:
// the policy is a value replaced whole (PUT semantics on a POST, kept POST so
// all write endpoints read alike).
func (s *server) defineCategory(w http.ResponseWriter, r *http.Request) {
	var req defineCategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	weight, err := shared.NewWeight(req.Estimate.WeightG)
	if err != nil {
		writeError(w, err)
		return
	}
	if req.Estimate.LengthMM < 0 || req.Estimate.WidthMM < 0 || req.Estimate.HeightMM < 0 {
		writeError(w, errInvalidRequest)
		return
	}
	estimate, err := shared.NewParcelSpec(weight, shared.NewDimensionsMM(req.Estimate.LengthMM, req.Estimate.WidthMM, req.Estimate.HeightMM))
	if err != nil {
		writeError(w, err)
		return
	}
	restrictions := make([]catalog.Restriction, 0, len(req.Restrictions))
	for _, x := range req.Restrictions {
		restrictions = append(restrictions, catalog.Restriction(x)) // the domain validates the value
	}
	if err := s.define.Handle(r.Context(), catalogapp.DefineCategory{
		Code: req.Code, Estimate: estimate, Restrictions: restrictions,
	}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
