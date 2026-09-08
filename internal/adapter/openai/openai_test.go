package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/duongsy/portage/internal/adapter/openai"
	catalogapp "github.com/duongsy/portage/internal/app/catalog"
	"github.com/duongsy/portage/internal/domain/catalog"
)

func url(t *testing.T) catalog.SourceURL {
	t.Helper()
	u, err := catalog.ParseSourceURL("https://www.example.com/t/air-trainer-90/abc")
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// answer builds a server that replies the way the real API does: the draft is
// a JSON STRING inside choices[0].message.content, not a nested object.
func answer(t *testing.T, status int, content string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		body, _ := json.Marshal(map[string]any{
			"choices": []any{map[string]any{"message": map[string]any{"content": content}}},
		})
		if status != http.StatusOK {
			_, _ = w.Write([]byte(`{"error":{"message":"slow down"}}`))
			return
		}
		_, _ = w.Write(body)
	}))
}

func clientFor(srv *httptest.Server) *openai.Client {
	return &openai.Client{Base: srv.URL, Key: "test-key", HTTP: srv.Client()}
}

// The happy path, and the two bits of cleaning the adapter does: the currency
// is upper-cased and the hint lower-cased, so the layer above never has to
// wonder which case a model felt like using today.
func TestClient_extractsADraft(t *testing.T) {
	srv := answer(t, http.StatusOK, `{"name":" Air Trainer 90 ","price":"150.00","currency":"usd","category_hint":"Footwear"}`)
	defer srv.Close()

	got, err := clientFor(srv).Extract(context.Background(), url(t))
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	want := catalogapp.ListingDraft{Name: "Air Trainer 90", Price: "150.00", Currency: "USD", CategoryHint: "footwear"}
	if got != want {
		t.Fatalf("draft = %+v, want %+v", got, want)
	}
}

// Everything that can go wrong with a remote model is ONE error for the caller
// — ErrExtractorUnavailable, which the HTTP layer answers with 503. The
// alternative, a different code per failure, would push the decision "is this
// retryable" onto every caller.
func TestClient_everyFailureIsUnavailable(t *testing.T) {
	for _, c := range []struct {
		name    string
		status  int
		content string
	}{
		{"rate limited", http.StatusTooManyRequests, ""},
		{"server error", http.StatusInternalServerError, ""},
		{"answer is not json", http.StatusOK, "I think it is about $150?"},
		{"answer is json but not the schema", http.StatusOK, `{"product":"Air Trainer 90"}`},
	} {
		t.Run(c.name, func(t *testing.T) {
			srv := answer(t, c.status, c.content)
			defer srv.Close()
			_, err := clientFor(srv).Extract(context.Background(), url(t))
			if !errors.Is(err, catalogapp.ErrExtractorUnavailable) {
				t.Fatalf("err = %v, want ErrExtractorUnavailable", err)
			}
		})
	}

	// "json but not the schema" deserves a note: it decodes into the struct
	// with every field empty, so the adapter accepts it and the layer ABOVE
	// refuses — an empty name is ErrEmptyName from the domain. Either way
	// nothing half-real is written.
}

// A call with no deadline is a request handler that can hang for as long as
// the other end feels like. The client always has a timeout.
func TestClient_givesUpOnASlowModel(t *testing.T) {
	// The handler is released by the test, not by the client hanging up:
	// httptest.Server.Close waits for outstanding handlers, and a server
	// that has written nothing does not always notice a dropped connection.
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer func() { close(release); srv.Close() }()

	c := &openai.Client{Base: srv.URL, Key: "test-key", HTTP: &http.Client{Timeout: 50 * time.Millisecond}}
	done := make(chan error, 1)
	go func() {
		_, err := c.Extract(context.Background(), url(t))
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, catalogapp.ErrExtractorUnavailable) {
			t.Fatalf("err = %v, want ErrExtractorUnavailable", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Extract did not give up — a request with no timeout can hang forever")
	}
}

// No key is not a 500 and not a panic: it is the feature being off.
func TestClient_withoutAKeyIsUnavailable(t *testing.T) {
	_, err := (&openai.Client{}).Extract(context.Background(), url(t))
	if !errors.Is(err, catalogapp.ErrExtractorUnavailable) {
		t.Fatalf("err = %v", err)
	}
	if _, err := (openai.Unavailable{}).Extract(context.Background(), url(t)); !errors.Is(err, catalogapp.ErrExtractorUnavailable) {
		t.Fatalf("Unavailable.Extract = %v", err)
	}
}

// The prompt asks for a plain decimal with no symbol, and the request carries a
// strict json_schema. Both are contract, so both are checked: a change to the
// wire shape should break a test, not a production call.
func TestClient_asksForAStrictSchema(t *testing.T) {
	var sent map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &sent)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"name\":\"x\",\"price\":\"1.00\",\"currency\":\"USD\",\"category_hint\":\"\"}"}}]}`))
	}))
	defer srv.Close()

	if _, err := clientFor(srv).Extract(context.Background(), url(t)); err != nil {
		t.Fatal(err)
	}
	format, _ := sent["response_format"].(map[string]any)
	if format["type"] != "json_schema" {
		t.Fatalf("response_format = %v", format)
	}
	schema, _ := format["json_schema"].(map[string]any)
	if schema["strict"] != true {
		t.Errorf("strict = %v — without it the model may answer prose", schema["strict"])
	}
	messages, _ := sent["messages"].([]any)
	if len(messages) != 2 {
		t.Fatalf("messages = %v", messages)
	}
	system, _ := messages[0].(map[string]any)
	if !strings.Contains(system["content"].(string), "ISO 4217") {
		t.Error("the prompt must pin the currency format, or the model picks its own")
	}
	user, _ := messages[1].(map[string]any)
	// The URL is all we send. Portage does not fetch the page (CATALOG.md §2).
	if user["content"] != url(t).String() {
		t.Errorf("user message = %v", user["content"])
	}
}
