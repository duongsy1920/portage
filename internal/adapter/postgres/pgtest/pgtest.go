// Package pgtest gives integration tests a migrated, EMPTY database — and makes
// sure two test packages never use it at the same time.
//
// `go test ./...` runs packages in parallel. Two packages truncating and
// filling the same portage_test database race each other and fail one run in
// three. The fix is not `-p 1` (which hides it) but a Postgres advisory lock,
// held on a dedicated connection for the life of each test: whoever holds it
// owns the database, everyone else waits. One database, any parallelism.
//
// Skips the test when PORTAGE_TEST_DSN is unset, so unit tests never need Docker.
//
// [PHP] Tương đương DAMADoctrineTestBundle hoặc phpunit `--process-isolation`
// [PHP] cho test DB — ở đây ~50 dòng và một khoá của chính Postgres.
package pgtest

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/duongsy/portage/internal/adapter/postgres"
)

// lockKey is arbitrary but fixed: every package that wants the test database
// waits on the same number.
const lockKey = 727469

// DSN returns PORTAGE_TEST_DSN or skips the test.
func DSN(t testing.TB) string {
	t.Helper()
	dsn := os.Getenv("PORTAGE_TEST_DSN")
	if dsn == "" {
		t.Skip("PORTAGE_TEST_DSN not set — see docs/SETUP.md §5e")
	}
	return dsn
}

// Pool connects, takes the lock, migrates, empties every table, and returns
// the pool. Everything is released when the test ends.
func Pool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	pool, err := postgres.Connect(ctx, DSN(t))
	if err != nil {
		t.Fatalf("pgtest: connect: %v", err)
	}
	t.Cleanup(pool.Close)

	// A session-level advisory lock lives as long as its connection, so the
	// connection is held out of the pool until Cleanup.
	lock, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("pgtest: acquire: %v", err)
	}
	if _, err := lock.Exec(ctx, `SELECT pg_advisory_lock($1)`, lockKey); err != nil {
		t.Fatalf("pgtest: lock: %v", err)
	}
	t.Cleanup(func() {
		_, _ = lock.Exec(ctx, `SELECT pg_advisory_unlock($1)`, lockKey)
		lock.Release()
	})

	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("pgtest: migrate: %v", err)
	}
	if err := truncateAll(ctx, pool); err != nil {
		t.Fatalf("pgtest: truncate: %v", err)
	}
	return pool
}

// truncateAll empties every table the migrations created — FOUND, not listed.
//
// This was a hand-written list of table names, and it drifted the first time a
// migration added one: 0006 created api_tokens, the list did not know, and a
// test read a row another test had left behind. A list that must be updated by
// hand is a list that will be forgotten, so the database is asked what it has.
//
// schema_migrations is the one table kept: emptying it would make Migrate
// replay every file for the next test in the same run.
//
// [PHP] Tương đương `doctrine:schema:drop --force` rồi migrate lại — nhưng rẻ
// [PHP] hơn nhiều: TRUNCATE giữ nguyên schema, chỉ xoá dòng.
func truncateAll(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `SELECT string_agg(quote_ident(tablename), ', ')
	           FROM pg_tables
	           WHERE schemaname = 'public' AND tablename <> 'schema_migrations'`

	var tables *string
	if err := pool.QueryRow(ctx, q).Scan(&tables); err != nil {
		return fmt.Errorf("list tables: %w", err)
	}
	if tables == nil { // nothing but schema_migrations yet
		return nil
	}
	if _, err := pool.Exec(ctx, "TRUNCATE "+*tables+" RESTART IDENTITY CASCADE"); err != nil {
		return fmt.Errorf("truncate %s: %w", *tables, err)
	}
	return nil
}
