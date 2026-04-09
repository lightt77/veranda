// Package db provides database connection and migration utilities.
package db

import (
	"database/sql"
	"fmt"

	"github.com/lightt77/veranda/internal/config"
	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with application-specific methods
type DB struct {
	*sql.DB
}

// Open creates a new database connection and runs migrations
func Open(cfg *config.AppConfig) (*DB, error) {
	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(1) // SQLite only supports one writer at a time
	db.SetMaxIdleConns(1)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{db}, nil
}

// runMigrations creates all tables if they don't exist
func runMigrations(db *sql.DB) error {
	migrations := []string{
		timersTable,
		stopwatchesTable,
		stopwatchLapsTable,
		journalEntriesTable,
		scheduledTasksTable,
		settingsTable,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
