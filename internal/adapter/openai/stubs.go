package openai

import (
	"context"
	"fmt"

	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
)

// The two adapters that are not a network call. Both satisfy the same port,
// which is the entire point of having one (DDD.md §20).

// Fake answers with a fixed draft. It is what a MEMORY run and every test use,
// so `go run ./cmd/api` works with no API key, no network and no bill.
type Fake struct {
	Draft catalogapp.ListingDraft
	Err   error // set to make the extractor fail on purpose
}

var _ catalogapp.ListingExtractor = Fake{}

func (f Fake) Extract(context.Context, catalog.SourceURL) (catalogapp.ListingDraft, error) {
	if f.Err != nil {
		return catalogapp.ListingDraft{}, f.Err
	}
	return f.Draft, nil
}

// Unavailable always refuses. It is what a POSTGRES run gets when there is no
// OPENAI_API_KEY, and the choice is deliberate:
//
// Falling back to Fake in production would be far worse than a 503. The API
// would answer 201, a draft product would exist with an invented name and
// price, and the only sign anything was wrong is that the numbers are fiction.
// A feature that is off must LOOK off (P9-PLAN §4, câu 6).
type Unavailable struct{}

var _ catalogapp.ListingExtractor = Unavailable{}

func (Unavailable) Extract(context.Context, catalog.SourceURL) (catalogapp.ListingDraft, error) {
	return catalogapp.ListingDraft{}, fmt.Errorf("openai: not configured (set OPENAI_API_KEY): %w", catalogapp.ErrExtractorUnavailable)
}
