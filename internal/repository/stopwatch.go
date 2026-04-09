package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lightt77/veranda/internal/db"
	"github.com/lightt77/veranda/internal/models"
)

// StopwatchRepository handles stopwatch data access
type StopwatchRepository struct {
	db *db.DB
}

// NewStopwatchRepository creates a new stopwatch repository
func NewStopwatchRepository(database *db.DB) *StopwatchRepository {
	return &StopwatchRepository{db: database}
}

// Create inserts a new stopwatch
func (r *StopwatchRepository) Create(stopwatch *models.Stopwatch) error {
	query := `
		INSERT INTO stopwatches (id, label, elapsed_ms, status, started_at_ms, stopped_at_ms, deleted_at, updated_at_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query,
		stopwatch.ID,
		stopwatch.Label,
		stopwatch.ElapsedMs,
		stopwatch.Status,
		stopwatch.StartedAtMs,
		stopwatch.StoppedAtMs,
		stopwatch.DeletedAt,
		stopwatch.UpdatedAtMs,
	)
	return err
}

// GetByID retrieves a stopwatch by ID (excluding soft-deleted)
func (r *StopwatchRepository) GetByID(id int64) (*models.Stopwatch, error) {
	query := `
		SELECT id, label, elapsed_ms, status, started_at_ms, stopped_at_ms, deleted_at, updated_at_ms
		FROM stopwatches
		WHERE id = ? AND deleted_at IS NULL
	`
	row := r.db.QueryRow(query, id)
	return r.scanStopwatch(row)
}

// GetAll retrieves all non-deleted stopwatches
func (r *StopwatchRepository) GetAll() ([]*models.Stopwatch, error) {
	query := `
		SELECT id, label, elapsed_ms, status, started_at_ms, stopped_at_ms, deleted_at, updated_at_ms
		FROM stopwatches
		WHERE deleted_at IS NULL
		ORDER BY id DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStopwatches(rows)
}

// GetRunning retrieves all running stopwatches
func (r *StopwatchRepository) GetRunning() ([]*models.Stopwatch, error) {
	query := `
		SELECT id, label, elapsed_ms, status, started_at_ms, stopped_at_ms, deleted_at, updated_at_ms
		FROM stopwatches
		WHERE status = ? AND deleted_at IS NULL
		ORDER BY id DESC
	`
	rows, err := r.db.Query(query, models.StopwatchRunning)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStopwatches(rows)
}

// Update updates a stopwatch
func (r *StopwatchRepository) Update(stopwatch *models.Stopwatch) error {
	query := `
		UPDATE stopwatches
		SET label = ?, elapsed_ms = ?, status = ?, started_at_ms = ?, stopped_at_ms = ?, updated_at_ms = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query,
		stopwatch.Label,
		stopwatch.ElapsedMs,
		stopwatch.Status,
		stopwatch.StartedAtMs,
		stopwatch.StoppedAtMs,
		stopwatch.UpdatedAtMs,
		stopwatch.ID,
	)
	return err
}

// SoftDelete marks a stopwatch as deleted
func (r *StopwatchRepository) SoftDelete(id int64) error {
	query := `UPDATE stopwatches SET deleted_at = ?, updated_at_ms = ? WHERE id = ? AND deleted_at IS NULL`
	nowSec := time.Now().Unix()
	nowMs := time.Now().UnixMilli()
	_, err := r.db.Exec(query, nowSec, nowMs, id)
	return err
}

// Delete permanently deletes a stopwatch (and cascades to laps)
func (r *StopwatchRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM stopwatches WHERE id = ?", id)
	return err
}

func (r *StopwatchRepository) scanStopwatch(row *sql.Row) (*models.Stopwatch, error) {
	s := &models.Stopwatch{}
	err := row.Scan(
		&s.ID, &s.Label, &s.ElapsedMs, &s.Status, &s.StartedAtMs, &s.StoppedAtMs, &s.DeletedAt, &s.UpdatedAtMs,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("stopwatch not found")
	}
	return s, err
}

func (r *StopwatchRepository) scanStopwatches(rows *sql.Rows) ([]*models.Stopwatch, error) {
	var stopwatches []*models.Stopwatch
	for rows.Next() {
		s := &models.Stopwatch{}
		err := rows.Scan(
			&s.ID, &s.Label, &s.ElapsedMs, &s.Status, &s.StartedAtMs, &s.StoppedAtMs, &s.DeletedAt, &s.UpdatedAtMs,
		)
		if err != nil {
			return nil, err
		}
		stopwatches = append(stopwatches, s)
	}
	return stopwatches, rows.Err()
}
