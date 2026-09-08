package catalogapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// ErrExtractorUnavailable: nobody can read the page right now — no API key
// configured, the model is down, or it answered nonsense. It is a 503, not a
// 400: the request was fine, the help was not there.
var ErrExtractorUnavailable = errors.New("listing extractor unavailable")

// ListingDraft is what a machine THINKS is on a page. Every field is a string
// because that is what a language model returns, and because the moment this
// package turned them into Money it would be deciding what counts as valid —
// which is the domain's job (Parse* below).
//
// CategoryHint is advice, not data: it never picks the category. The caller
// supplies the category, because getting it wrong changes the duty rate and
// the estimated weight — real money — and a guess with no confidence attached
// is not a basis for that.
type ListingDraft struct {
	Name         string
	Price        string
	Currency     string
	CategoryHint string
}

// ListingExtractor is an ANTI-CORRUPTION LAYER (DDD.md §22), the second in the
// system after procurement.MerchantACL. What it protects against is worth
// naming precisely:
//
//	MerchantACL     protects us from another company's API and its model
//	ListingExtractor protects us from a WEB PAGE and a language model's guess
//
// The port says "give me a draft"; nothing behind it — HTML, a prompt, an
// HTTP call to OpenAI, a person — reaches the domain. The domain still only
// ever sees catalog.ProductDetails with a Provenance attached, exactly as it
// does when a customer types the fields in by hand.
type ListingExtractor interface {
	Extract(ctx context.Context, url catalog.SourceURL) (ListingDraft, error)
}

// DraftFromURL: "here is a link, fill in what you can".
type DraftFromURL struct {
	Merchant catalog.MerchantID
	Category catalog.CategoryCode
	Source   catalog.SourceURL
}

// DraftFromURLResult carries the id of the product recorded plus what the
// machine guessed, so a UI can show the operator what to check.
type DraftFromURLResult struct {
	Product      catalog.ProductID
	Draft        ListingDraft
	CategoryHint string
}

// DraftFromURLHandler reads a page with the extractor and records the result
// as an ordinary draft product.
//
// The rule that keeps this honest is that it ADDS, never overwrites. A machine
// may create a draft nobody has checked; it may not touch a listing an
// operator has already vouched for. So this use case ends in exactly the same
// place a customer's paste does — AddProductHandler — and the product still
// needs confirm-listing, measure and publish from a human before it can be
// quoted (WALKTHROUGH §14h).
//
// Provenance is SourcedByFeed: a machine put it there, and no person has
// vouched for it. Publish refuses an unverified listing, so nothing the model
// invents can reach a customer's price without somebody looking at it first.
type DraftFromURLHandler struct {
	deps    Deps
	extract ListingExtractor
	add     *AddProductHandler
}

func NewDraftFromURLHandler(d Deps, extractor ListingExtractor) *DraftFromURLHandler {
	mustHave("DraftFromURLHandler", map[string]any{
		"Clock": d.Clock, "UoW": d.UoW, "Merchants": d.Merchants,
		"Categories": d.Categories, "Products": d.Products, "Outbox": d.Outbox,
		"Extractor": extractor,
	})
	return &DraftFromURLHandler{deps: d, extract: extractor, add: NewAddProductHandler(d)}
}

func (h *DraftFromURLHandler) Handle(ctx context.Context, cmd DraftFromURL) (DraftFromURLResult, error) {
	draft, err := h.extract.Extract(ctx, cmd.Source)
	if err != nil {
		return DraftFromURLResult{}, err
	}

	// The machine's strings become value objects HERE, through the domain's
	// own Parse* — the same door a human's typing goes through. A model that
	// answers "about $150" fails at ParseMoney, not three tables later.
	var price shared.Money
	if draft.Price != "" || draft.Currency != "" {
		cur, err := shared.CurrencyFromCode(draft.Currency)
		if err != nil {
			return DraftFromURLResult{}, fmt.Errorf("extracted currency %q: %w", draft.Currency, err)
		}
		if price, err = shared.ParseMoney(draft.Price, cur); err != nil {
			return DraftFromURLResult{}, fmt.Errorf("extracted price %q: %w", draft.Price, err)
		}
	}

	id, err := h.add.Handle(ctx, AddProduct{
		Name:      draft.Name,
		Merchant:  cmd.Merchant,
		Category:  cmd.Category,
		Source:    cmd.Source,
		Price:     price,
		SourcedBy: catalog.SourcedByFeed, // a machine, and nobody has checked it
	})
	if err != nil {
		return DraftFromURLResult{}, err
	}
	return DraftFromURLResult{Product: id, Draft: draft, CategoryHint: draft.CategoryHint}, nil
}
