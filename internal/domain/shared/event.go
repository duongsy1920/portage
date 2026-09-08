package shared

import "time"

// Event is something that HAS HAPPENED in the business. It is named in the
// past tense (DepositPaid, PurchaseConfirmed), carries the facts a listener
// in another context needs, and never holds a pointer into the aggregate
// that raised it — it is a message, not a reference.
//
// EventName is the stable wire name ("ordering.deposit_paid"). It goes into
// the outbox table and must not change when the Go type is renamed.
//
// Symfony: the message class you dispatch through Messenger — except here
// it is part of the domain model, not a framework mechanism, and the domain
// decides when one is raised.
// [PHP] `type X interface {...}` = interface, giống PHP. NHƯNG khác một điểm
// [PHP] cực lớn: Go KHÔNG CÓ TỪ KHOÁ `implements`.
// [PHP]     PHP: final class DepositPaid implements Event { ... }
// [PHP]     Go : type DepositPaid struct { OrderID OrderID; At time.Time }
// [PHP]          func (e DepositPaid) EventName() string     { return "..." }
// [PHP]          func (e DepositPaid) OccurredAt() time.Time { return e.At }
// [PHP]          → TỰ ĐỘNG là Event, vì có đủ 2 method. Không khai báo gì thêm.
// [PHP] Gọi là "interface ngầm" (implicit). Hệ quả: viết interface SAU khi đã
// [PHP] có struct cũng được, và struct không cần biết interface tồn tại.
// [PHP] Bên trong chỉ liệt kê chữ ký method, không có `public`/`abstract`.
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// Events is the recorder an aggregate root EMBEDS to collect the events its
// methods raise:
//
//	type CustomerOrder struct {
//	    shared.Events
//	    id     OrderID
//	    status Status
//	}
//
//	func (o *CustomerOrder) PayDeposit(amount Money, now time.Time) error {
//	    // ... check the invariant, change state ...
//	    o.Record(DepositPaid{OrderID: o.id, Amount: amount, At: now})
//	    return nil
//	}
//
// The aggregate never publishes. It only RECORDS; the application layer
// pulls after the repository has saved, and writes the events into the
// outbox in the same transaction (see Outbox in docs/DDD.md). Publishing
// before the save succeeds is how a worker goes shopping for an order that
// was never persisted.
//
// This is the one mutable type in the package: it is bookkeeping owned by an
// aggregate, not a value object, so pointer receivers are correct here.
//
// Symfony: the `private array $domainEvents = []` + `pullDomainEvents()`
// pair you see in Doctrine entities, drained by a postFlush listener into
// the Messenger bus.
// [PHP] `[]Event` = slice (mảng động) chứa Event.
// [PHP]     PHP: /** @var Event[] */ private array $pending = [];
// [PHP]     Go : pending []Event
// [PHP] Go phân biệt rõ 3 thứ mà PHP gộp chung vào `array`:
// [PHP]     []Event          slice   — danh sách, thêm bằng append()
// [PHP]     [5]Event         array   — độ dài CỐ ĐỊNH, hiếm dùng
// [PHP]     map[string]int   map     — mảng kết hợp, ~ ['a' => 1]
type Events struct {
	pending []Event
}

// Record appends one event to be published after the aggregate is saved.
//
// [PHP] Chú ý dấu `*` trong `(e *Events)` — receiver dạng CON TRỎ.
// [PHP] Chỉ khi dùng `*` thì method mới SỬA ĐƯỢC bản gốc. Không có `*` thì Go
// [PHP] copy struct và mọi thay đổi mất khi hàm kết thúc.
// [PHP] PHP không có phân biệt này: object luôn truyền theo tham chiếu.
// [PHP]     (e *Events) ~ hành vi mặc định của object PHP  → sửa được
// [PHP]     (m Money)   ~ như thể clone $this mỗi lần gọi  → không sửa được
// [PHP] Quy tắc: value object dùng `(m Money)`, thứ có trạng thái dùng `(e *Events)`.
func (e *Events) Record(ev Event) {
	// [PHP] `append(slice, phần_tử)` = thêm vào cuối, TRẢ VỀ slice mới.
	// [PHP]     PHP: $this->pending[] = $ev;
	// [PHP]     Go : e.pending = append(e.pending, ev)     ← phải gán lại!
	// [PHP] Quên gán lại là mất dữ liệu — bẫy kinh điển của người mới.
	e.pending = append(e.pending, ev)
}

// PullEvents returns everything recorded so far, in order, and clears the
// recorder so a second pull is empty. The zero value pulls nothing.
func (e *Events) PullEvents() []Event {
	out := e.pending
	// [PHP] `nil` cho slice ~ `[]` rỗng. Slice nil vẫn `append` và `range` được
	// [PHP] bình thường — khác PHP ở chỗ `null` bên PHP thì `foreach` sẽ lỗi.
	e.pending = nil
	return out
}
