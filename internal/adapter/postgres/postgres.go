// Package postgres is the persistence ADAPTER: the catalog repositories, the
// outbox and the unit of work, on PostgreSQL through pgx.
//
// It is the second implementation of the same ports the in-memory adapter
// implements — the handlers and use cases do not change a line. The domain
// gives it a way in and out through Snapshot()/FromSnapshot(); the codec gives
// it the outbox payload. Nothing here decides anything about the business.
//
// Transactions travel in the context: UnitOfWork.InTx begins one and stores it
// in ctx; every repository asks db(ctx) for "the transaction if there is one,
// else the pool". A handler therefore saves and appends to the outbox in ONE
// transaction without knowing how — which is the whole point of the outbox
// pattern (DDD.md §27), and the test that was skipped against memory.
//
// [PHP] Đây là tầng Doctrine: DBAL connection (pgxpool), EntityManager::
// [PHP] wrapInTransaction (UnitOfWork.InTx), và mapping struct ↔ cột viết tay
// [PHP] thay cho #[ORM\Column]. Không có lazy loading, không có identity map:
// [PHP] ByID trả về một *Merchant mới mỗi lần gọi.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a pool and pings once, so a wrong DSN fails at start-up, not
// on the first request.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return pool, nil
}

// querier is the subset of pgx both a pool and a transaction implement.
// Repositories are written against it and never know which one they got.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// txKey is the context key under which InTx stores the open transaction.
// An unexported empty struct: no other package can forge or read it.
type txKey struct{}

// db returns the transaction carried in ctx, or the pool.
//
// [PHP] Không có tương đương trực tiếp: Doctrine giữ transaction trong
// [PHP] EntityManager (trạng thái toàn cục của request). Go không có trạng thái
// [PHP] toàn cục như vậy — thứ đi cùng request là ctx, nên transaction đi theo ctx.
func db(ctx context.Context, pool *pgxpool.Pool) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}

// UnitOfWork implements app.UnitOfWork with a real transaction.
type UnitOfWork struct {
	pool *pgxpool.Pool
}

func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{pool: pool}
}

// InTx runs fn inside one transaction and commits if fn returns nil, rolls
// back otherwise — including when fn panics. A call nested inside an open
// transaction joins it instead of opening a second one.
//
// [PHP] `(err error)` là named return: defer bên dưới đọc và GHI ĐÈ được err —
// [PHP] đó là cách commit lỗi vẫn báo ra ngoài. recover() ~ finally + rethrow.
func (u *UnitOfWork) InTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, nested := ctx.Value(txKey{}).(pgx.Tx); nested {
		return fn(ctx)
	}
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx) // the caller's error is the one worth reporting
			return
		}
		if cerr := tx.Commit(ctx); cerr != nil {
			err = fmt.Errorf("commit: %w", cerr)
		}
	}()
	return fn(context.WithValue(ctx, txKey{}, tx))
}
