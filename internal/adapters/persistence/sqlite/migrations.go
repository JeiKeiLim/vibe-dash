package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// Migration represents a database schema migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// migrations is the ordered list of all migrations to apply
var migrations = []Migration{
	{
		Version:     1,
		Description: "Initial schema with projects table",
		SQL: CreateSchemaVersionTableSQL + "\n" +
			CreateProjectsTableSQL + "\n" +
			CreateIndexPathSQL + "\n" +
			CreateIndexStateSQL,
	},
}

// RunMigrations applies all pending migrations to the database
func RunMigrations(ctx context.Context, db *sqlx.DB) error {
	// Ensure schema_version table exists first
	_, err := db.ExecContext(ctx, CreateSchemaVersionTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}

	currentVersion, err := getCurrentVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue
		}

		if err := applyMigration(ctx, db, m); err != nil {
			return fmt.Errorf("failed to apply migration v%d (%s): %w", m.Version, m.Description, err)
		}
	}

	return nil
}

// getCurrentVersion returns the current schema version from the database
func getCurrentVersion(ctx context.Context, db *sqlx.DB) (int, error) {
	var version int
	err := db.GetContext(ctx, &version, "SELECT COALESCE(MAX(version), 0) FROM schema_version")
	if err != nil {
		return 0, err
	}
	return version, nil
}

// applyMigration executes a single migration within a transaction
func applyMigration(ctx context.Context, db *sqlx.DB, m Migration) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // Rollback is no-op after successful commit

	if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_version (version, applied_at) VALUES (?, ?)",
		m.Version, time.Now().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("failed to record migration version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}
