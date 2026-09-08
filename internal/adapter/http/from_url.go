package httpapi

import (
	"context"
	"fmt"
	"net/http"

	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
)

// draftFromURLRequest: a link and the category it belongs to.
//
// The category is REQUIRED even though the model returns a hint, because it
// decides the duty rate and the estimated weight — real money. A hint with no
// confidence attached is not a basis for that, so it comes back in the
// response for a human to act on and never picks anything by itself.
type draftFromURLRequest struct {
	MerchantID string `json:"merchant_id"`
	URL        string `json:"url"`
	Category   string `json:"category"`
}

type draftFromURLResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Price        string `json:"price,omitempty"`
	Currency     string `json:"currency,omitempty"`
	CategoryHint string `json:"category_hint,omitempty"`
}

// POST /products/from-url → 201. A machine reads the page and a DRAFT product
// is recorded — provenance "feed", nobody has vouched for it. The operator
// still has to confirm-listing, measure and publish before it can be quoted,
// exactly as with a customer's paste (WALKTHROUGH §14h).
//
// Without an extractor configured this is 503 extractor_unavailable, not a
// silent fallback to a fake: a feature that is off must look off.
func (s *server) draftFromURL(w http.ResponseWriter, r *http.Request) {
	var req draftFromURLRequest
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
	source, err := catalog.ParseSourceURL(req.URL)
	if err != nil {
		writeError(w, err)
		return
	}
	out, err := s.fromURL.Handle(r.Context(), catalogapp.DraftFromURL{
		Merchant: merchant, Category: category, Source: source,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, draftFromURLResponse{
		ID: out.Product.String(), Name: out.Draft.Name, Price: out.Draft.Price,
		Currency: out.Draft.Currency, CategoryHint: out.CategoryHint,
	})
}

// unavailableExtractor is the adapter's default when Deps.Extractor is nil.
// It duplicates openai.Unavailable on purpose: httpapi must not import an
// adapter package (that is the composition root's job, internal/platform/wire),
// and four lines here are cheaper than an import that inverts the dependency
// rule.
type unavailableExtractor struct{}

func (unavailableExtractor) Extract(context.Context, catalog.SourceURL) (catalogapp.ListingDraft, error) {
	return catalogapp.ListingDraft{}, fmt.Errorf("no listing extractor configured: %w", catalogapp.ErrExtractorUnavailable)
}
