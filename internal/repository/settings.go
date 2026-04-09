package repository

import (
	"database/sql"
	"time"

	"github.com/lightt77/veranda/internal/db"
)

// SettingsRepository handles settings data access
type SettingsRepository struct {
	db *db.DB
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(database *db.DB) *SettingsRepository {
	return &SettingsRepository{db: database}
}

// Get retrieves a setting by key
func (r *SettingsRepository) Get(key string) (string, error) {
	query := `SELECT value FROM settings WHERE key = ?`
	var value string
	err := r.db.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// GetOrDefault retrieves a setting by key with fallback default
func (r *SettingsRepository) GetOrDefault(key, defaultValue string) (string, error) {
	value, err := r.Get(key)
	if err != nil {
		return "", err
	}
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

// Set creates or updates a setting
func (r *SettingsRepository) Set(key, value string) error {
	query := `
		INSERT INTO settings (key, value, updated_at_ms)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at_ms = excluded.updated_at_ms
	`
	_, err := r.db.Exec(query, key, value, time.Now().UnixMilli())
	return err
}

// SetWithTime creates or updates a setting with explicit timestamp
func (r *SettingsRepository) SetWithTime(key, value string, updatedAtMs int64) error {
	query := `
		INSERT INTO settings (key, value, updated_at_ms)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at_ms = excluded.updated_at_ms
	`
	_, err := r.db.Exec(query, key, value, updatedAtMs)
	return err
}

// Delete removes a setting
func (r *SettingsRepository) Delete(key string) error {
	_, err := r.db.Exec("DELETE FROM settings WHERE key = ?", key)
	return err
}

// GetAll retrieves all settings
func (r *SettingsRepository) GetAll() (map[string]string, error) {
	query := `SELECT key, value FROM settings`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}
	return settings, rows.Err()
}
