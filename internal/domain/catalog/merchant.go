// Package catalog is the bounded context that answers "what can a customer
// order, and from where".
//
// It owns Merchant (the shops we buy from), CategoryPolicy (what a kind of
// goods weighs, measures and may not do), and — later — Product. It does NOT own price quotes, parcels
// or orders — those belong to pricing, logistics and ordering, which refer to
// things here by id only.
//
// The word "Nike" does not appear in this package, and must not. A merchant
// is a row in a table; adding Adidas is adding data, not deploying code.
package catalog

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

var (
	ErrEmptyName        = errors.New("merchant name is empty")
	ErrInvalidHostname  = errors.New("invalid hostname")
	ErrUnknownSourcing  = errors.New("unknown sourcing mode")
	ErrCurrencyRequired = errors.New("currency is required")
	ErrEmptyReason      = errors.New("a reason is required")
)

// MerchantStatus is where a merchant is in its life: bought from, or not.
//
// [PHP] Cùng kiểu "enum" với SourcingMode. Symfony: enum MerchantStatus: string.
type MerchantStatus string

const (
	// StatusActive: procurement may buy here.
	StatusActive MerchantStatus = "active"

	// StatusSuspended: the shop banned our account, closed, or we decided to
	// stop. Nothing new is bought; open purchase tasks need a human decision.
	StatusSuspended MerchantStatus = "suspended"
)

// MerchantID identifies a merchant.
//
// [PHP] `struct{ shared.ID }` — nhúng (embedding) kiểu shared.ID mà không đặt
// [PHP] tên field. MerchantID dùng luôn String(), IsZero() của shared.ID.
// [PHP] Mục đích: compiler phân biệt MerchantID với OrderID, dù ruột giống hệt.
// [PHP]     PHP: final class MerchantId { public function __construct(private Uuid $id) {} }
// [PHP] Xem quy ước 5 trong docs/SETUP.md.
type MerchantID struct {
	shared.ID
}

func NewMerchantID() MerchantID {
	return MerchantID{shared.NewID()}
}

// ParseMerchantID reads an id back from a URL segment or a database column.
func ParseMerchantID(s string) (MerchantID, error) {
	id, err := shared.ParseID(s)
	if err != nil {
		return MerchantID{}, fmt.Errorf("merchant id: %w", err)
	}
	return MerchantID{id}, nil
}

// SourcingMode says how product data for a merchant can reach us.
//
// This is a CAPABILITY of the merchant, not the history of one product — a
// single product's history lives in Provenance. A merchant may support
// several modes; a product arrived by exactly one.
//
// [PHP] Go không có enum. Cách thay thế: một kiểu string riêng + hằng số.
// [PHP]     PHP: enum SourcingMode: string { case Operator = 'operator'; ... }
// [PHP] Kiểu riêng (`type SourcingMode string`) khiến hàm nhận SourcingMode
// [PHP] không nhận bừa một string bất kỳ — compiler chặn.
type SourcingMode string

const (
	// SourcedByOperator: a person opens the shop's page and types the data in.
	// This is the primary mode: it is ordinary browsing, and it works for
	// every merchant on earth without asking anyone's permission.
	SourcedByOperator SourcingMode = "operator"

	// SourcedByCustomer: the customer supplies the product details themselves.
	// Note the boundary this draws — the CUSTOMER provides data; our system
	// does not go and fetch the merchant's page. Automated retrieval is
	// scraping, and shops forbid it (see docs/CATALOG.md §2).
	SourcedByCustomer SourcingMode = "customer"

	// SourcedByFeed: the merchant hands us a structured product feed. Not used
	// today — kept because the domain must not care which mode is switched on.
	SourcedByFeed SourcingMode = "feed"
)

func (s SourcingMode) IsValid() bool {
	switch s {
	case SourcedByOperator, SourcedByCustomer, SourcedByFeed:
		return true
	}
	return false
}

func (s SourcingMode) String() string {
	return string(s)
}

// Hostname is a merchant's site, normalised: lowercase, no scheme, no path.
//
// It is a value object rather than a string so that "Nike.com",
// "https://nike.com/" and "nike.com" cannot become three different merchants.
type Hostname struct {
	host string
}

// ParseHostname accepts what an operator types, and refuses what is ambiguous.
// It lowercases and trims, because those differences carry no meaning — but it
// does NOT strip a scheme or a path, because a domain that guesses is a domain
// that guesses wrong. Cleaning input is the adapter's job (convention 3).
func ParseHostname(s string) (Hostname, error) {
	h := strings.ToLower(strings.TrimSpace(s))
	switch {
	case h == "":
		return Hostname{}, fmt.Errorf("hostname is empty: %w", ErrInvalidHostname)
	case strings.ContainsAny(h, " /:?#@"):
		return Hostname{}, fmt.Errorf("hostname %q: no scheme, port or path allowed: %w", s, ErrInvalidHostname)
	case !strings.Contains(h, "."), strings.HasPrefix(h, "."), strings.HasSuffix(h, "."):
		return Hostname{}, fmt.Errorf("hostname %q: want something like example.com: %w", s, ErrInvalidHostname)
	}
	return Hostname{host: h}, nil
}

func MustParseHostname(s string) Hostname {
	h, err := ParseHostname(s)
	if err != nil {
		panic(err)
	}
	return h
}

func (h Hostname) String() string {
	return h.host
}
func (h Hostname) IsZero() bool {
	return h.host == ""
}

// Merchant is a shop we buy from — an ENTITY: two merchants with identical
// details are still two merchants, and a merchant's details change over time
// while it stays the same merchant.
//
// [PHP] So sánh với Value Object (shared.Money): Money bằng nhau khi mọi field
// [PHP] bằng nhau; Merchant bằng nhau khi cùng `id`, dù tên có đổi.
type Merchant struct {
	// [PHP] Nhúng shared.Events: Merchant có sẵn Record() và PullEvents().
	// [PHP] Đây là lý do Merchant phải dùng con trỏ (*Merchant) ở method đổi
	// [PHP] trạng thái — Events có trạng thái, xem quy ước 6.
	shared.Events

	id       MerchantID
	name     string
	site     Hostname
	currency shared.Currency // what its prices are listed in
	freeShip FreeShipping
	sourcing []SourcingMode
	status   MerchantStatus
	addedAt  time.Time
}

// MerchantDetails is what an operator fills in to register a merchant. It is
// a plain parameter struct, not a domain object: it has no invariants of its
// own, and RegisterMerchant is where every field is checked.
//
// Why a struct and not six positional parameters (convention 10): the call
// site names each value, two parameters of similar shape cannot be swapped,
// and adding a field does not break every caller. The price is that a
// FORGOTTEN field compiles — so every required field is validated below, and
// TestRegisterMerchant_everyFieldIsValidated zeroes each field in turn and
// demands a refusal. Adding a field means either validating it or listing it
// there as "zero is a legitimate answer", with the reason. Two fields are:
// FreeShipping{} means "the shop always charges" and Sourcing nil means "no
// way to get data yet".
//
// [PHP] Go không có named arguments (`new Merchant(name: 'X', site: ...)`),
// [PHP] struct tham số là cách thay thế. Nó cũng là DTO đầu vào — như một
// [PHP] RegisterMerchantRequest/Command trong Symfony Messenger, nhưng không
// [PHP] có #[Assert]: validate nằm trong RegisterMerchant, nơi duy nhất.
type MerchantDetails struct {
	Name         string
	Site         Hostname
	Currency     shared.Currency // what its prices are listed in
	FreeShipping FreeShipping
	Sourcing     []SourcingMode
}

// RegisterMerchant is the only way to build one. Every field is checked,
// because it comes from an operator filling in a form (convention 1).
//
// now is passed in rather than read from the clock: the domain does not know
// what time it is (convention 7). It stays outside the struct because it is
// not a detail of the merchant — it is when the registration happened.
func RegisterMerchant(d MerchantDetails, now time.Time) (*Merchant, error) {
	name := strings.TrimSpace(d.Name)
	if name == "" {
		return nil, ErrEmptyName
	}
	if d.Site.IsZero() {
		return nil, fmt.Errorf("merchant %q: %w", name, ErrInvalidHostname)
	}
	if d.Currency.IsZero() {
		return nil, fmt.Errorf("merchant %q: %w", name, ErrCurrencyRequired)
	}
	if err := d.FreeShipping.validFor(d.Currency); err != nil {
		return nil, fmt.Errorf("merchant %q: %w", name, err)
	}
	modes, err := normaliseSourcing(d.Sourcing)
	if err != nil {
		return nil, fmt.Errorf("merchant %q: %w", name, err)
	}

	m := &Merchant{
		id:       NewMerchantID(),
		name:     name,
		site:     d.Site,
		currency: d.Currency,
		freeShip: d.FreeShipping,
		sourcing: modes,
		status:   StatusActive,
		addedAt:  now,
	}
	m.Record(MerchantRegistered{ID: m.id, Name: name, Site: d.Site, Currency: d.Currency, At: now})
	return m, nil
}

// normaliseSourcing rejects unknown modes and drops duplicates, so that
// Supports() never has to think about them. An empty list is legal and means
// "no way to get data yet" — a merchant we know about but cannot buy from.
func normaliseSourcing(in []SourcingMode) ([]SourcingMode, error) {
	// [PHP] `make([]T, 0, n)` = tạo slice rỗng nhưng đặt sẵn sức chứa n, tránh
	// [PHP] cấp phát lại nhiều lần. PHP không có khái niệm này (array tự lo).
	out := make([]SourcingMode, 0, len(in))
	seen := make(map[SourcingMode]bool, len(in))
	for _, s := range in {
		if !s.IsValid() {
			return nil, fmt.Errorf("sourcing %q: %w", s, ErrUnknownSourcing)
		}
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out, nil
}

func (m *Merchant) ID() MerchantID {
	return m.id
}
func (m *Merchant) Name() string {
	return m.name
}
func (m *Merchant) Site() Hostname {
	return m.site
}
func (m *Merchant) Currency() shared.Currency {
	return m.currency
}
func (m *Merchant) FreeShipping() FreeShipping {
	return m.freeShip
}
func (m *Merchant) AddedAt() time.Time {
	return m.addedAt
}

func (m *Merchant) Status() MerchantStatus {
	return m.status
}

func (m *Merchant) IsActive() bool {
	return m.status == StatusActive
}

// Sourcing returns a COPY of the modes. Returning m.sourcing directly would
// hand the caller a slice that shares memory with the merchant, letting
// outside code rewrite an entry and bypass every check in this file.
//
// [PHP] Bẫy riêng của Go: slice là "cửa sổ" nhìn vào một mảng, không phải bản
// [PHP] sao. Trả thẳng m.sourcing ra ngoài thì `modes[0] = "xxx"` của người gọi
// [PHP] sửa luôn ruột Merchant. PHP array là copy-on-write nên không có bẫy này.
func (m *Merchant) Sourcing() []SourcingMode {
	return append([]SourcingMode(nil), m.sourcing...)
}

func (m *Merchant) Supports(s SourcingMode) bool {
	for _, have := range m.sourcing {
		if have == s {
			return true
		}
	}
	return false
}

// Rename is the only way the name changes. There is no SetName.
func (m *Merchant) Rename(name string, now time.Time) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyName
	}
	if name == m.name {
		return nil // nothing happened, so no event
	}
	old := m.name
	m.name = name
	m.Record(MerchantRenamed{ID: m.id, From: old, To: name, At: now})
	return nil
}

// EnableSourcing turns a mode on. Idempotent: enabling twice raises one event.
func (m *Merchant) EnableSourcing(s SourcingMode, now time.Time) error {
	if !s.IsValid() {
		return fmt.Errorf("sourcing %q: %w", s, ErrUnknownSourcing)
	}
	if m.Supports(s) {
		return nil
	}
	m.sourcing = append(m.sourcing, s)
	m.Record(MerchantSourcingEnabled{ID: m.id, Mode: s, At: now})
	return nil
}

// DisableSourcing turns a mode off. Idempotent: disabling an absent mode
// changes nothing and raises nothing.
//
// [PHP] `slices.DeleteFunc` xoá tại chỗ mọi phần tử thoả điều kiện và trả về
// [PHP] slice đã rút ngắn — PHP: array_values(array_filter($a, fn($x) => $x !== $s)).
func (m *Merchant) DisableSourcing(s SourcingMode, now time.Time) error {
	if !s.IsValid() {
		return fmt.Errorf("sourcing %q: %w", s, ErrUnknownSourcing)
	}
	if !m.Supports(s) {
		return nil
	}
	m.sourcing = slices.DeleteFunc(m.sourcing, func(have SourcingMode) bool {
		return have == s
	})
	m.Record(MerchantSourcingDisabled{ID: m.id, Mode: s, At: now})
	return nil
}

// ChangeFreeShipping replaces the domestic delivery rule. Shops change these
// constantly; re-registering the merchant to follow would lose its identity.
// The rule must be priced in the merchant's currency, as at registration.
func (m *Merchant) ChangeFreeShipping(rule FreeShipping, now time.Time) error {
	if err := rule.validFor(m.currency); err != nil {
		return fmt.Errorf("merchant %q: %w", m.name, err)
	}
	if rule == m.freeShip {
		return nil // nothing happened, so no event
	}
	old := m.freeShip
	m.freeShip = rule
	m.Record(MerchantFreeShippingChanged{ID: m.id, From: old, To: rule, At: now})
	return nil
}

// Suspend stops purchases at this shop. It is the one catalog event another
// context truly needs: procurement must stop creating purchase tasks here,
// and an operator must look at the ones already open. The reason is what
// that operator reads six weeks later, so it is required.
//
// Idempotent: suspending a suspended merchant is not a second suspension.
func (m *Merchant) Suspend(reason string, now time.Time) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("suspend %q: %w", m.name, ErrEmptyReason)
	}
	if m.status == StatusSuspended {
		return nil
	}
	m.status = StatusSuspended
	m.Record(MerchantSuspended{ID: m.id, Reason: reason, At: now})
	return nil
}

// Reinstate reopens a suspended merchant for purchases. Idempotent.
func (m *Merchant) Reinstate(now time.Time) error {
	if m.status == StatusActive {
		return nil
	}
	m.status = StatusActive
	m.Record(MerchantReinstated{ID: m.id, At: now})
	return nil
}
