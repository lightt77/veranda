package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lightt77/veranda/internal/db"
	"github.com/lightt77/veranda/internal/models"
)

// JournalRepository handles journal entry data access
type JournalRepository struct {
	db *db.DB
}

// NewJournalRepository creates a new journal repository
func NewJournalRepository(database *db.DB) *JournalRepository {
	return &JournalRepository{db: database}
}

// Create inserts a new journal entry
func (r *JournalRepository) Create(entry *models.JournalEntry) error {
	query := `
		INSERT INTO journal_entries (id, log, deleted_at, created_at_ms)
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.Exec(query,
		entry.ID,
		entry.Log,
		entry.DeletedAt,
		entry.CreatedAtMs,
	)
	return err
}

// GetByID retrieves a journal entry by ID (excluding soft-deleted)
func (r *JournalRepository) GetByID(id int64) (*models.JournalEntry, error) {
	query := `
		SELECT id, log, deleted_at, created_at_ms
		FROM journal_entries
		WHERE id = ? AND deleted_at IS NULL
	`
	row := r.db.QueryRow(query, id)
	return r.scanEntry(row)
}

// GetAll retrieves all non-deleted journal entries
func (r *JournalRepository) GetAll() ([]*models.JournalEntry, error) {
	query := `
		SELECT id, log, deleted_at, created_at_ms
		FROM journal_entries
		WHERE deleted_at IS NULL
		ORDER BY id DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEntries(rows)
}

// GetRecent retrieves recent journal entries with limit
func (r *JournalRepository) GetRecent(limit int) ([]*models.JournalEntry, error) {
	query := `
		SELECT id, log, deleted_at, created_at_ms
		FROM journal_entries
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT ?
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEntries(rows)
}

// Update updates a journal entry
func (r *JournalRepository) Update(entry *models.JournalEntry) error {
	query := `UPDATE journal_entries SET log = ? WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db.Exec(query, entry.Log, entry.ID)
	return err
}

// SoftDelete marks a journal entry as deleted
func (r *JournalRepository) SoftDelete(id int64) error {
	query := `UPDATE journal_entries SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`
	nowSec := time.Now().Unix()
	_, err := r.db.Exec(query, nowSec, id)
	return err
}

// Delete permanently deletes a journal entry
func (r *JournalRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM journal_entries WHERE id = ?", id)
	return err
}

func (r *JournalRepository) scanEntry(row *sql.Row) (*models.JournalEntry, error) {
	entry := &models.JournalEntry{}
	err := row.Scan(&entry.ID, &entry.Log, &entry.DeletedAt, &entry.CreatedAtMs)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("journal entry not found")
	}
	return entry, err
}

func (r *JournalRepository) scanEntries(rows *sql.Rows) ([]*models.JournalEntry, error) {
	var entries []*models.JournalEntry
	for rows.Next() {
		entry := &models.JournalEntry{}
		err := rows.Scan(&entry.ID, &entry.Log, &entry.DeletedAt, &entry.CreatedAtMs)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
