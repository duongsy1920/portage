package catalog

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrInvalidSourceURL = errors.New("invalid source url")

// SourceURL is the product page a customer or operator pointed at.
//
// It is a REFERENCE. The system never fetches it — that would be scraping,
// which shops forbid (docs/CATALOG.md §2); a person reads the page and types
// the details in. What the domain needs from the URL is the HOST, to tie the
// product to the Merchant we buy from, and to find duplicates of the same
// page (ProductRepository.BySource).
//
// Kept as given, only trimmed, so the person who pasted it still recognises
// it. Only the host is normalised — through ParseHostname, so "Example.com"
// and "example.com" are one merchant — and a port, query or fragment on the
// URL does not change which merchant it belongs to.
//
// [PHP] Gói `net/url` là thư viện chuẩn, tương đương parse_url(). Chuỗi thô
// [PHP] và host tách sẵn lưu cùng nhau để không phải parse lại mỗi lần hỏi.
type SourceURL struct {
	raw  string
	host Hostname
}

// ParseSourceURL accepts an absolute http(s) URL with a real hostname and
// refuses everything else — a bare "example.com/x" is not a URL, it is a
// guess about one, and the domain does not guess (convention 3).
func ParseSourceURL(s string) (SourceURL, error) {
	s = strings.TrimSpace(s)
	u, err := url.Parse(s)
	if s == "" || err != nil {
		return SourceURL{}, fmt.Errorf("source url %q: %w", s, ErrInvalidSourceURL)
	}
	if scheme := strings.ToLower(u.Scheme); scheme != "http" && scheme != "https" {
		return SourceURL{}, fmt.Errorf("source url %q: want http or https: %w", s, ErrInvalidSourceURL)
	}
	host, err := ParseHostname(u.Hostname()) // Hostname() drops any port
	if err != nil {
		return SourceURL{}, fmt.Errorf("source url %q: %v: %w", s, err, ErrInvalidSourceURL)
	}
	return SourceURL{raw: s, host: host}, nil
}

// MustParseSourceURL is for tests and seed data; it panics on bad input.
func MustParseSourceURL(s string) SourceURL {
	u, err := ParseSourceURL(s)
	if err != nil {
		panic(err)
	}
	return u
}

// Host is the merchant's site as the URL names it — the link back to a
// Merchant, and the key duplicate detection groups by.
func (u SourceURL) Host() Hostname {
	return u.host
}

func (u SourceURL) String() string {
	return u.raw
}

func (u SourceURL) IsZero() bool {
	return u.raw == ""
}
