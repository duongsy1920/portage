// Package openai is the ADAPTER behind catalogapp.ListingExtractor: it reads a
// shop's page with a language model and returns a draft listing.
//
// Three things about it are deliberate and worth reading before the code:
//
//  1. NO SDK. It is net/http and encoding/json, about a hundred lines. An SDK
//     would add a dependency, a second way to configure timeouts, and a set of
//     its own types that would want to leak upward. The port is four strings;
//     the wire format is JSON; there is nothing an SDK would do for us here.
//
//  2. NO SCRAPING. Portage does not fetch the merchant's page itself —
//     CATALOG.md §2: Nike's publisher terms forbid it, and most shops' do too.
//     The URL goes to the model, which answers from what it knows; a customer
//     or an operator remains free to type the fields in instead.
//
//  3. THE ANSWER IS A GUESS. Everything this package returns is provenance
//     "feed": recorded, not vouched for. catalog.Product.Publish refuses an
//     unverified listing, so nothing invented here can reach a customer's
//     price without a person confirming it first (DDD.md §22).
//
// [PHP] Một HttpClient của Symfony gói trong một service implement interface
// [PHP] của domain — chỉ khác là interface ở đây do tầng app khai báo, và
// [PHP] domain không biết package này tồn tại.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
)

// DefaultModel and DefaultBase are what wire uses when nothing overrides them.
const (
	DefaultModel = "gpt-4o-mini"
	DefaultBase  = "https://api.openai.com"
)

// Client calls the chat completions API and asks for a strict JSON object.
type Client struct {
	Base  string // "" → DefaultBase
	Key   string
	Model string // "" → DefaultModel
	HTTP  *http.Client
}

var _ catalogapp.ListingExtractor = (*Client)(nil)

// New builds a Client with a TIMEOUT, always. A call with no deadline is a
// request handler that can hang until the client gives up, holding a
// connection and a goroutine for as long as the other end feels like.
func New(key string) *Client {
	return &Client{Key: key, HTTP: &http.Client{Timeout: 15 * time.Second}}
}

// The shape we demand back. json_schema with strict:true makes the model
// answer these four fields or fail — better than parsing prose, and it is why
// this adapter has no regexes in it.
const responseSchema = `{
	"type": "object",
	"properties": {
		"name":          {"type": "string"},
		"price":         {"type": "string"},
		"currency":      {"type": "string"},
		"category_hint": {"type": "string"}
	},
	"required": ["name", "price", "currency", "category_hint"],
	"additionalProperties": false
}`

const systemPrompt = `You identify a single retail product from its page URL.
Answer with the product's name as the shop lists it, its list price as a plain
decimal number with no currency symbol and no thousands separators ("150.00"),
its currency as an ISO 4217 code ("USD"), and a one-word lowercase category
hint ("footwear", "apparel", "electronics"). If you are not sure of a value,
return an empty string for it rather than a guess.`

func (c *Client) Extract(ctx context.Context, url catalog.SourceURL) (catalogapp.ListingDraft, error) {
	if c.Key == "" {
		return catalogapp.ListingDraft{}, fmt.Errorf("openai: no api key: %w", catalogapp.ErrExtractorUnavailable)
	}
	body, err := json.Marshal(request{
		Model: firstNonEmpty(c.Model, DefaultModel),
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: url.String()},
		},
		ResponseFormat: responseFormat{
			Type: "json_schema",
			JSONSchema: jsonSchema{
				Name:   "listing_draft",
				Strict: true,
				Schema: json.RawMessage(responseSchema),
			},
		},
	})
	if err != nil {
		return catalogapp.ListingDraft{}, fmt.Errorf("openai: %w", err)
	}

	endpoint := strings.TrimSuffix(firstNonEmpty(c.Base, DefaultBase), "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return catalogapp.ListingDraft{}, fmt.Errorf("openai: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Key)

	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	res, err := client.Do(req)
	if err != nil {
		// A network failure is UNAVAILABLE, not a bad request: nothing the
		// caller sent is wrong, and 503 tells them to try again — 500 does not.
		return catalogapp.ListingDraft{}, fmt.Errorf("openai: %v: %w", err, catalogapp.ErrExtractorUnavailable)
	}
	defer res.Body.Close()

	// 4 KB is generous for four short strings and small enough that a wrong
	// endpoint answering with a web page cannot fill memory.
	raw, err := io.ReadAll(io.LimitReader(res.Body, 4<<10))
	if err != nil {
		return catalogapp.ListingDraft{}, fmt.Errorf("openai: %v: %w", err, catalogapp.ErrExtractorUnavailable)
	}
	if res.StatusCode != http.StatusOK {
		return catalogapp.ListingDraft{}, fmt.Errorf("openai: http %d: %s: %w",
			res.StatusCode, strings.TrimSpace(string(raw)), catalogapp.ErrExtractorUnavailable)
	}

	var out response
	if err := json.Unmarshal(raw, &out); err != nil || len(out.Choices) == 0 {
		return catalogapp.ListingDraft{}, fmt.Errorf("openai: unreadable answer: %w", catalogapp.ErrExtractorUnavailable)
	}
	var draft draftJSON
	if err := json.Unmarshal([]byte(out.Choices[0].Message.Content), &draft); err != nil {
		return catalogapp.ListingDraft{}, fmt.Errorf("openai: answer is not the schema we asked for: %w", catalogapp.ErrExtractorUnavailable)
	}
	// A name is the one field that must come back. The prompt says to
	// answer "" rather than guess, so an empty name means the model did not
	// recognise the page — the same outcome for the caller as the model
	// being down: nobody can read this link, type it in by hand. Letting it
	// through would surface as ErrEmptyName from the domain, a 400 that
	// blames the caller for something they did not send.
	if strings.TrimSpace(draft.Name) == "" {
		return catalogapp.ListingDraft{}, fmt.Errorf("openai: the model did not recognise %s: %w", url, catalogapp.ErrExtractorUnavailable)
	}
	return catalogapp.ListingDraft{
		Name:         strings.TrimSpace(draft.Name),
		Price:        strings.TrimSpace(draft.Price),
		Currency:     strings.ToUpper(strings.TrimSpace(draft.Currency)),
		CategoryHint: strings.ToLower(strings.TrimSpace(draft.CategoryHint)),
	}, nil
}

// ── the wire shapes, written by hand like every other contract here ──────────

type request struct {
	Model          string         `json:"model"`
	Messages       []message      `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type       string     `json:"type"`
	JSONSchema jsonSchema `json:"json_schema"`
}

type jsonSchema struct {
	Name   string          `json:"name"`
	Strict bool            `json:"strict"`
	Schema json.RawMessage `json:"schema"`
}

type response struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type draftJSON struct {
	Name         string `json:"name"`
	Price        string `json:"price"`
	Currency     string `json:"currency"`
	CategoryHint string `json:"category_hint"`
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
