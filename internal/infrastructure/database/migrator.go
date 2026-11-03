package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// MigrationRecord represents an applied migration entry in the database.
type MigrationRecord struct {
	Version   string    `json:"version"`
	AppliedAt time.Time `json:"applied_at"`
}

// RunMigrations executes all unapplied migrations embedded in MigrationsFS.
func RunMigrations(db *sql.DB) error {
	return RunMigrationsFromFS(context.Background(), db, MigrationsFS, "migrations")
}

// RunMigrationsFromFS executes unapplied migrations from a given filesystem directory.
// It ensures version control by tracking executed migrations in schema_migrations.
func RunMigrationsFromFS(ctx context.Context, db *sql.DB, fsys fs.FS, dir string) error {
	if db == nil {
		return fmt.Errorf("database connection cannot be nil")
	}

	// 1. Ensure the schema_migrations version table exists
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	// 2. Fetch already applied migration versions
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[v] = true
	}

	// 3. Read migration files from the filesystem
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory '%s': %w", dir, err)
	}

	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	// Sort migration files to guarantee deterministic version order (e.g., 000001, 000002)
	sort.Strings(migrationFiles)

	appliedCount := 0
	for _, fileName := range migrationFiles {
		if applied[fileName] {
			continue
		}

		filePath := fileName
		if dir != "" && dir != "." {
			filePath = dir + "/" + fileName
		}

		content, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file '%s': %w", filePath, err)
		}

		log.Printf("[Database Migrator] Applying migration: %s", fileName)

		// Execute each migration within its own transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration '%s': %w", fileName, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed executing migration '%s': %w", fileName, err)
		}

		insertRecordSQL := `
			INSERT INTO schema_migrations (version, applied_at)
			VALUES ($1, $2);
		`
		if _, err := tx.ExecContext(ctx, insertRecordSQL, fileName, time.Now()); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed recording migration version for '%s': %w", fileName, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction for migration '%s': %w", fileName, err)
		}

		log.Printf("[Database Migrator] Successfully applied migration: %s", fileName)
		appliedCount++
	}

	if appliedCount == 0 {
		log.Println("[Database Migrator] Schema is up to date. No pending migrations.")
	} else {
		log.Printf("[Database Migrator] Finished applying %d migration(s) successfully.", appliedCount)
	}

	return nil
}

// GetAppliedMigrations returns the list of all applied migrations in chronological order.
func GetAppliedMigrations(ctx context.Context, db *sql.DB) ([]MigrationRecord, error) {
	query := `
		SELECT version, applied_at
		FROM schema_migrations
		ORDER BY version ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var records []MigrationRecord
	for rows.Next() {
		var r MigrationRecord
		if err := rows.Scan(&r.Version, &r.AppliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan applied migration record: %w", err)
		}
		records = append(records, r)
	}

	return records, nil
}
