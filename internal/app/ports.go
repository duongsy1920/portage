// Package app holds what every use case needs from the outside world and does
// not want to know the shape of: a clock, a transaction, an outbox.
//
// These are PORTS in the hexagonal sense (DDD.md §20) — declared here, where
// they are needed, and implemented in internal/adapter and internal/platform.
// A use case handler in internal/app/<context> takes them as interfaces and
// is therefore testable with the in-memory adapters, in milliseconds.
//
// [PHP] Đây là chỗ Symfony tự làm cho bạn bằng autowiring: ClockInterface,
// [PHP] EntityManager::wrapInTransaction(), MessageBus. Go không có container
// [PHP] — interface khai báo ở đây, main() cắm implementation vào bằng tay.
package app

import (
	"context"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

// Clock is the only place the current time comes from. The domain never reads
// it (convention 7); a handler asks once and passes `now` down.
//
// [PHP] Symfony\Component\Clock\ClockInterface — cùng lý do: test được "3 ngày
// [PHP] sau khi cọc" mà không đợi 3 ngày.
type Clock interface {
	Now() time.Time
}

// Outbox receives the events an aggregate recorded, AFTER it was saved and
// inside the same transaction (DDD.md §27). A worker reads the outbox and
// publishes; the handler never publishes directly.
type Outbox interface {
	Append(ctx context.Context, events []shared.Event) error
}

// UnitOfWork runs fn inside one transaction: everything fn saves and appends
// commits together or not at all. The transaction travels in ctx so the
// repositories fn calls can find it.
//
// Go has no generic methods, so fn returns only error; a handler that needs a
// result captures it in a closure variable (see RegisterMerchantHandler).
//
// [PHP] EntityManager::wrapInTransaction(fn) — nhưng ở đây là interface, và
// [PHP] bản in-memory không có transaction thật (memory.UnitOfWork).
type UnitOfWork interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}
