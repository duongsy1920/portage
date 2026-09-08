package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// decodeJSON reads one JSON object into dst. Unknown fields are an error: a
// misspelled "nmae" silently becoming an empty name is exactly the kind of
// mistake that should fail at the edge, loudly.
//
// [PHP] Tương đương Serializer::deserialize() với ALLOW_EXTRA_ATTRIBUTES = false.
func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20)) // 1 MiB is plenty for a command
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: %v", errBadJSON, err)
	}
	if dec.More() {
		return fmt.Errorf("%w: trailing data after the object", errBadJSON)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) // nothing useful to do if the client went away
}

// language picks the request's primary language from Accept-Language:
// "vi-VN,vi;q=0.9,en;q=0.8" → "vi". No header → "en".
//
// [PHP] $request->getPreferredLanguage(). Ở đây cố tình đơn giản: chỉ lấy
// [PHP] thẻ đầu, bỏ q-value — đủ cho việc duy nhất nó phục vụ (dấu thập phân).
func language(r *http.Request) string {
	first := strings.Split(r.Header.Get("Accept-Language"), ",")[0]
	first = strings.Split(strings.TrimSpace(first), ";")[0] // drop ";q=0.9"
	lang := strings.ToLower(strings.Split(first, "-")[0])   // "vi-VN" → "vi"
	if lang == "" {
		return "en"
	}
	return lang
}

// normalizeAmount turns a human-written amount into the ONE canonical form
// the domain accepts, according to the locale — the guess the domain refuses
// to make (DDD.md §11d) is made here, where the locale is known:
//
//	vi   "1.234,56" → "1234.56"    dot groups thousands, comma is the decimal mark
//	en   "1,234.56" → "1234.56"    the reverse
//
// The result still goes through ParseMoney, which is strict: whatever this
// function cannot make canonical ("fifty") is refused there.
//
// [PHP] Đây là việc NumberFormatter làm trong FormType (MoneyType, locale-aware)
// [PHP] trước khi giá trị tới Entity.
func normalizeAmount(s, lang string) string {
	s = strings.TrimSpace(s)
	switch lang {
	case "vi":
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	default:
		s = strings.ReplaceAll(s, ",", "")
	}
	return s
}
