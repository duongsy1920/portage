package httpapi

import (
	"net/http"
	"time"

	reportingapp "github.com/duongsy/portage/internal/app/reporting"
)

// The two screens' own endpoints, and the reason the worklist read model
// exists: a person has to be told what is left to do, not handed a 409 and
// asked to guess which of four steps is missing.
//
// Both answer from ONE table with no join, because the projector already did
// the joining as the events arrived (DDD.md §24).

// worklistView is one pasted product and how far along it is.
type worklistView struct {
	ProductID string    `json:"product_id"`
	Merchant  string    `json:"merchant_id"`
	Category  string    `json:"category"`
	Name      string    `json:"name"`
	Source    string    `json:"source"`
	Price     moneyView `json:"price"`

	// How it arrived, and who is waiting. Requester is empty when an operator
	// added the product on spec: there is nobody to notify, and a screen must
	// not pretend otherwise.
	SourcedBy string `json:"sourced_by"`
	Requester string `json:"requested_by,omitempty"`

	// The four steps, so a screen can render a checklist instead of a row of
	// buttons that all look equally available.
	HasVariant       bool   `json:"has_variant"`
	ListingConfirmed bool   `json:"listing_confirmed"`
	Measured         bool   `json:"measured"`
	Published        bool   `json:"published"`
	VariantLabel     string `json:"variant_label,omitempty"` // the first one, for a one-line summary

	// Variants is what the customer picks from to order. The id has to be
	// here: POST /orders takes a variant_id, and ordering refuses one it has
	// never heard of.
	Variants []variantOptionView `json:"variants"`

	// NextStep is decided in the read model, not here and not in the browser:
	// two screens that each work it out will disagree the day a fifth step
	// appears, and the disagreement will be silent.
	NextStep string `json:"next_step"`

	AddedAt   time.Time `json:"added_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// variantOptionView is one line of the customer's size dropdown.
type variantOptionView struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

func worklistViewOf(w reportingapp.WorklistItem) worklistView {
	v := worklistView{
		ProductID: w.Product.String(),
		Merchant:  w.Merchant.String(),
		Category:  w.Category,
		Name:      w.Name,
		Source:    w.Source,
		Price:     viewOf(w.Price),

		SourcedBy: w.SourcedBy,

		HasVariant:       w.HasVariant(),
		ListingConfirmed: w.ListingConfirmed,
		Measured:         w.Measured,
		Published:        w.Published,
		VariantLabel:     w.FirstLabel(),
		NextStep:         string(w.NextStep()),

		AddedAt:   w.AddedAt,
		UpdatedAt: w.UpdatedAt,
	}
	if !w.RequestedBy.IsZero() {
		v.Requester = w.RequestedBy.String()
	}
	v.Variants = make([]variantOptionView, 0, len(w.Variants))
	for _, x := range w.Variants {
		v.Variants = append(v.Variants, variantOptionView{ID: x.ID.String(), Label: x.Label})
	}
	return v
}

// GET /product-queue → 200 [worklistView]: everything not published yet,
// oldest first. Operators only — it is the staff worklist.
func (s *server) listProductQueue(w http.ResponseWriter, r *http.Request) {
	items, err := s.worklist.Open(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]worklistView, 0, len(items))
	for _, item := range items {
		out = append(out, worklistViewOf(item))
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /me/products → 200 [worklistView]: the things THIS customer asked for,
// at whatever stage they are.
//
// Like GET /me/orders, the route cannot name anybody else: the id comes from
// the token, so there is no path that leaks another customer's list.
func (s *server) listMyProducts(w http.ResponseWriter, r *http.Request) {
	customer, ok := customerOf(r)
	if !ok {
		writeError(w, errForbidden)
		return
	}
	items, err := s.worklist.ByRequester(r.Context(), customer)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]worklistView, 0, len(items))
	for _, item := range items {
		out = append(out, worklistViewOf(item))
	}
	writeJSON(w, http.StatusOK, out)
}
