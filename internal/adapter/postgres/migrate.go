package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrations are plain SQL files, embedded in the binary, applied once each in
// file-name order and recorded in schema_migrations. Forty lines instead of a
// tool: the schema is small and every step is readable as SQL.
//
//go:embed migrations/*.sql
var migrations embed.FS

// Migrate brings the database to the current schema. Safe to call at every
// start-up AND from two processes at once: everything happens inside one
// transaction that first takes an advisory lock, so the second starter waits,
// then finds the files already recorded and applies nothing.
//
// The lock has to cover the bookkeeping table too, not just the loop below.
// CREATE TABLE IF NOT EXISTS is not atomic in PostgreSQL: two of them both
// pass the existence check, then one fails inserting the table's row type
// ("duplicate key value violates unique constraint pg_type_typname_nsp_index").
// cmd/api and cmd/worker start together on a first deploy, so this is a real
// path, not a theoretical one — see TestMigrate_survivesTwoProcessesOnAColdDatabase.
//
// [PHP] doctrine:migrations:migrate — bảng schema_migrations là bảng
// [PHP] doctrine_migration_versions, mỗi file .sql là một Version class.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after Commit

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(727468)`); err != nil { // "portage" on a phone keypad, roughly
		return fmt.Errorf("migrate lock: %w", err)
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    text        PRIMARY KEY,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	for _, e := range entries {
		var applied bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, e.Name()).Scan(&applied); err != nil {
			return fmt.Errorf("migrate %s: %w", e.Name(), err)
		}
		if applied {
			continue
		}
		sql, err := migrations.ReadFile("migrations/" + e.Name())
		if err != nil {
			return fmt.Errorf("migrate %s: %w", e.Name(), err)
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("migrate %s: %w", e.Name(), err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, e.Name()); err != nil {
			return fmt.Errorf("migrate %s: %w", e.Name(), err)
		}
	}
	return tx.Commit(ctx)
}
