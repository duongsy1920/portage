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
// start-up: applied files are skipped, and an advisory lock keeps two
// processes starting at once from applying the same file twice.
//
// [PHP] doctrine:migrations:migrate — bảng schema_migrations là bảng
// [PHP] doctrine_migration_versions, mỗi file .sql là một Version class.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    text        PRIMARY KEY,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

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
