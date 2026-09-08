package shared

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrInvalidID = errors.New("invalid id")

// ID is the identity every entity and aggregate root in Portage carries.
//
// It is a UUID version 7 (RFC 9562): the first 48 bits are a Unix timestamp
// in milliseconds, the rest is random. That buys three things:
//
//   - it is generated IN THE DOMAIN, with no round trip to the database —
//     an aggregate has an identity from the first line of its constructor;
//   - it is time-ordered, so a B-tree index appends at the end instead of
//     scattering random v4 values across every page;
//   - it is opaque, unlike an auto-increment that leaks how many orders
//     you have taken.
//
// Symfony: symfony/uid Uuid::v7(), mapped with #[ORM\Column(type: 'uuid')]
// and NO #[ORM\GeneratedValue] — the entity assigns $this->id = Uuid::v7()
// itself. Compare with the usual `private ?int $id = null`, which leaves the
// object without an identity until flush(). In DDD that is an invalid state.
//
// github.com/google/uuid is the one third-party import in the domain. It is
// a utility in the sense symfony/uid is — not a framework, not an SDK — and
// the Dependency Rule is about those.
//
// Every aggregate wraps ID in its own type so the compiler keeps them apart:
//
//	type OrderID struct{ shared.ID }
//	func NewOrderID() OrderID { return OrderID{shared.NewID()} }
//
// Passing a MerchantID where an OrderID is expected is then a compile error,
// not a 3 a.m. incident. (Symfony: a `final class OrderId` wrapping a Uuid;
// Go gets the same effect from one line, by embedding.)
// [PHP] Ghi chú cú pháp cho đoạn `type OrderID struct{ shared.ID }` ở trên:
// [PHP] viết một KIỂU mà không đặt tên field = "embedding" (nhúng).
// [PHP]     type OrderID struct{ shared.ID }        ← nhúng
// [PHP]     type OrderID struct{ id shared.ID }     ← field bình thường
// [PHP] Nhúng thì OrderID DÙNG LUÔN được String(), IsZero() của shared.ID:
// [PHP]     orderID.String()          // không cần viết orderID.id.String()
// [PHP] Trông giống kế thừa `class OrderID extends ID` bên PHP, NHƯNG KHÔNG
// [PHP] PHẢI kế thừa: không có đa hình, không override, không gọi parent::.
// [PHP] Go gọi đây là "composition" — gần với `use SomeTrait;` của PHP hơn.
type ID struct {
	u uuid.UUID
}

// NewID mints a fresh, time-ordered identity.
func NewID() ID {
	u, err := uuid.NewV7()
	if err != nil {
		// Only if the OS random source is broken. Nothing sensible can
		// happen after that, so fail loudly instead of returning ID{}.
		panic("shared: cannot generate uuid v7: " + err.Error())
	}
	return ID{u: u}
}

// ParseID reads an ID back from its text form — a URL segment, a database
// column.
//
// Two shapes of input are refused, for the reason ParseMoney refuses
// "150,50": the domain does not guess.
//
//   - Anything but version 7. We never mint another version, so a v4 is
//     foreign or corrupted data.
//   - Anything but the canonical text. uuid.Parse is generous — it also
//     takes uppercase, "{...}", "urn:uuid:...", and 32 hex digits with no
//     dashes. But the only writer of an ID is ID.String(), which emits one
//     shape, so a different shape means the value took a detour through
//     something that rewrote it. That is worth an error rather than a shrug,
//     and it keeps the text form usable as a database key: two spellings of
//     one id would compare and index as two different keys.
//
// Normalising outside input is the adapter's job, before the domain sees it.
func ParseID(s string) (ID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ID{}, fmt.Errorf("parse id %q: %w", s, ErrInvalidID)
	}
	if u.Version() != 7 {
		return ID{}, fmt.Errorf("parse id %q: uuid version %d, want 7: %w", s, u.Version(), ErrInvalidID)
	}
	if s != u.String() {
		return ID{}, fmt.Errorf("parse id %q: not canonical, want %q: %w", s, u.String(), ErrInvalidID)
	}
	return ID{u: u}, nil
}

func (id ID) String() string {
	return id.u.String()
}

// IsZero reports the zero value ID{} — "no identity yet", which a saved
// aggregate must never have.
func (id ID) IsZero() bool {
	return id.u == uuid.Nil
}
