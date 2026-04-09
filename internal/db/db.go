// Package db provides database connection and migration utilities.
package db

import (
	"database/sql"
	"fmt"
	"time"

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
		citiesTable,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	// Seed cities data if table is empty
	if err := seedCities(db); err != nil {
		return fmt.Errorf("failed to seed cities: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// ClearAllData soft-deletes all data from all tables (use with caution!)
func (db *DB) ClearAllData() error {
	now := time.Now().Unix()
	queries := []string{
		"UPDATE timers SET deleted_at = ? WHERE deleted_at IS NULL",
		"UPDATE stopwatches SET deleted_at = ? WHERE deleted_at IS NULL",
		"DELETE FROM stopwatch_laps",
		"UPDATE journal_entries SET deleted_at = ? WHERE deleted_at IS NULL",
		"UPDATE scheduled_tasks SET deleted_at = ? WHERE deleted_at IS NULL",
		"DELETE FROM settings",
	}

	for _, query := range queries {
		if _, err := db.Exec(query, now); err != nil {
			return fmt.Errorf("failed to clear data: %w", err)
		}
	}
	return nil
}
