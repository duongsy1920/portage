package catalog_test

import (
	"errors"
	"testing"

	"github.com/duongsy/portage/internal/domain/catalog"
)

// The URL is a REFERENCE the customer or operator supplies — the system never
// fetches it (docs/CATALOG.md §2). What the domain needs from it is the host,
// to tie the product to the merchant we buy from.
func TestParseSourceURL(t *testing.T) {
	ok := map[string]string{ // input -> host
		"https://www.example.com/t/air-max-90/abc":   "www.example.com",
		"  HTTPS://Example.com/x  ":                  "example.com",
		"http://shop.example.com:8443/item?id=1#top": "shop.example.com",
	}
	for in, host := range ok {
		u, err := catalog.ParseSourceURL(in)
		if err != nil {
			t.Errorf("ParseSourceURL(%q): %v", in, err)
			continue
		}
		if got := u.Host().String(); got != host {
			t.Errorf("ParseSourceURL(%q).Host() = %q, want %q", in, got, host)
		}
		if u.IsZero() {
			t.Errorf("ParseSourceURL(%q) reported zero", in)
		}
	}
	// Stored as given (trimmed), not rewritten: the reference must stay
	// recognisable to the person who pasted it.
	u, _ := catalog.ParseSourceURL("  https://www.example.com/t/air-max-90/abc  ")
	if got := u.String(); got != "https://www.example.com/t/air-max-90/abc" {
		t.Errorf("String() = %q", got)
	}

	for _, in := range []string{
		"", "   ", "example.com/x", "ftp://example.com/x", "https://", "https:///x",
		"https://localhost/x", "not a url", "mailto:a@example.com",
	} {
		if _, err := catalog.ParseSourceURL(in); !errors.Is(err, catalog.ErrInvalidSourceURL) {
			t.Errorf("ParseSourceURL(%q): got %v, want ErrInvalidSourceURL", in, err)
		}
	}
	if !(catalog.SourceURL{}).IsZero() {
		t.Error("SourceURL{} must be zero")
	}
}
