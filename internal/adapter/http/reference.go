package httpapi

import (
	"net/http"
)

// Reference data: the two closed sets a person has to CHOOSE from, not type.
//
// Both are reads straight off a repository, like GET /quotes/{id}, because
// neither has an invariant to protect: they answer "what are my options".
// Without them a form has no way to render a select box, so it falls back to
// asking a human for a uuid — which is how the first console ended up
// unusable.
//
// Server-side validation does NOT go away because the UI offers a list. A
// client can still post anything, so POST /products keeps answering 404
// category_not_found. The list is a convenience; the closed set lives in the
// domain.
//
// [PHP] Đây là mấy endpoint trả "danh mục" để đổ vào <select>, tương đương
// [PHP] mấy repository->findAll() bạn gọi trong FormType.

type categoryView struct {
	Code string `json:"code"`

	// The default box for this kind of goods, used to quote a product nobody
	// has weighed yet. Grams and millimetres, integers, like everywhere else.
	WeightG  int64 `json:"weight_g"`
	LengthMM int64 `json:"length_mm"`
	WidthMM  int64 `json:"width_mm"`
	HeightMM int64 `json:"height_mm"`

	// Restrictions a lane may refuse, e.g. a battery. Empty for most goods.
	Restrictions []string `json:"restrictions,omitempty"`
}

// GET /categories → 200 [categoryView]. Any authenticated caller: a customer
// pasting a link has to pick one of these, so the list is not operator-only.
func (s *server) listCategories(w http.ResponseWriter, r *http.Request) {
	all, err := s.categories.All(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]categoryView, 0, len(all))
	for _, c := range all {
		spec := c.DefaultParcelSpec()
		v := categoryView{
			Code:     c.Code().String(),
			WeightG:  spec.Weight().Grams(),
			LengthMM: spec.Dimensions().LengthMM(),
			WidthMM:  spec.Dimensions().WidthMM(),
			HeightMM: spec.Dimensions().HeightMM(),
		}
		for _, x := range c.Restrictions() {
			v.Restrictions = append(v.Restrictions, x.String())
		}
		out = append(out, v)
	}
	writeJSON(w, http.StatusOK, out)
}

type merchantView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Site string `json:"site"`

	// The currency this shop bills in. A product's price has to match it, so
	// the form can pin the currency once the shop is chosen instead of
	// letting someone pick a mismatch and learn about it from a 409.
	Currency string `json:"currency"`

	// Who is allowed to bring a product in from this shop. A customer whose
	// mode is missing here cannot paste a link for it.
	Sourcing []string `json:"sourcing"`

	// Suspended shops stay in the list, greyed out rather than hidden: a
	// missing shop looks like our bug, a suspended one looks like a decision.
	Status string `json:"status"`
}

// GET /merchants → 200 [merchantView].
func (s *server) listMerchants(w http.ResponseWriter, r *http.Request) {
	all, err := s.merchants.All(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]merchantView, 0, len(all))
	for _, m := range all {
		v := merchantView{
			ID:       m.ID().String(),
			Name:     m.Name(),
			Site:     m.Site().String(),
			Currency: m.Currency().Code(),
			Status:   string(m.Status()),
		}
		for _, mode := range m.Sourcing() {
			v.Sourcing = append(v.Sourcing, string(mode))
		}
		out = append(out, v)
	}
	writeJSON(w, http.StatusOK, out)
}
