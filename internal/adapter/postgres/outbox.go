package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/duongsy/portage/internal/adapter/eventcodec"
	"github.com/duongsy/portage/internal/app"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/worker"
)

// Outbox implements app.Outbox on the outbox table — and the worker's side of
// it, Pending and MarkSent. Append runs through db(ctx), so inside a handler's
// InTx it lands in the SAME transaction as the aggregate's Save: both commit
// or neither does. That sentence is the outbox pattern (DDD.md §27).
type Outbox struct {
	pool *pgxpool.Pool
}

var (
	_ app.Outbox    = (*Outbox)(nil)
	_ worker.Source = (*Outbox)(nil)
)

func NewOutbox(pool *pgxpool.Pool) *Outbox {
	return &Outbox{pool: pool}
}

// Append encodes each event with the codec (the payload contract) and inserts
// one row per event, in order.
func (o *Outbox) Append(ctx context.Context, events []shared.Event) error {
	q := db(ctx, o.pool)
	for _, ev := range events {
		env, err := eventcodec.Encode(ev)
		if err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		if _, err := q.Exec(ctx, `INSERT INTO outbox (event_name, occurred_at, payload) VALUES ($1, $2, $3)`,
			env.Name, env.OccurredAt, []byte(env.Payload)); err != nil {
			return fmt.Errorf("outbox append %s: %w", env.Name, err)
		}
	}
	return nil
}

// Pending returns unsent rows, oldest first. Inside a transaction, SKIP LOCKED
// lets two workers share the table without ever picking the same row.
func (o *Outbox) Pending(ctx context.Context, limit int) ([]worker.Entry, error) {
	rows, err := db(ctx, o.pool).Query(ctx, `
		SELECT id, event_name, occurred_at, payload FROM outbox
		WHERE sent_at IS NULL ORDER BY id LIMIT $1 FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return nil, fmt.Errorf("outbox pending: %w", err)
	}
	defer rows.Close()
	var out []worker.Entry
	for rows.Next() {
		var e worker.Entry
		var payload []byte
		if err := rows.Scan(&e.ID, &e.Name, &e.OccurredAt, &payload); err != nil {
			return nil, fmt.Errorf("outbox scan: %w", err)
		}
		e.OccurredAt, e.Payload = e.OccurredAt.UTC(), payload
		out = append(out, e)
	}
	return out, rows.Err()
}

// MarkSent stamps the rows the worker has published. Called AFTER publishing:
// a crash in between re-sends, which is the at-least-once promise, and why
// subscribers must be idempotent.
func (o *Outbox) MarkSent(ctx context.Context, ids []int64, at time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	if _, err := db(ctx, o.pool).Exec(ctx, `UPDATE outbox SET sent_at = $2 WHERE id = ANY($1)`, ids, at); err != nil {
		return fmt.Errorf("outbox mark sent: %w", err)
	}
	return nil
}
