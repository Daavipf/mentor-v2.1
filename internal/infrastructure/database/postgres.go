package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/000001_init_schema.sql
var DefaultMigrationSQL string

func NewPostgresDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sql.DB, migrationSQL string) error {
	if migrationSQL == "" {
		migrationSQL = DefaultMigrationSQL
	}
	_, err := db.Exec(migrationSQL)
	if err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}
	return nil
}

