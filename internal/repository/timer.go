// Package repository provides data access layer for all entities.
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lightt77/veranda/internal/db"
	"github.com/lightt77/veranda/internal/models"
)

// TimerRepository handles timer data access
type TimerRepository struct {
	db *db.DB
}

// NewTimerRepository creates a new timer repository
func NewTimerRepository(database *db.DB) *TimerRepository {
	return &TimerRepository{db: database}
}

// Create inserts a new timer
func (r *TimerRepository) Create(timer *models.Timer) error {
	query := `
		INSERT INTO timers (id, label, duration_ms, remaining_ms, status, started_at_ms, paused_at_ms, completed_at_ms, deleted_at, updated_at_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query,
		timer.ID,
		timer.Label,
		timer.DurationMs,
		timer.RemainingMs,
		timer.Status,
		timer.StartedAtMs,
		timer.PausedAtMs,
		timer.CompletedAtMs,
		timer.DeletedAt,
		timer.UpdatedAtMs,
	)
	return err
}

// GetByID retrieves a timer by ID (excluding soft-deleted)
func (r *TimerRepository) GetByID(id int64) (*models.Timer, error) {
	query := `
		SELECT id, label, duration_ms, remaining_ms, status, started_at_ms, paused_at_ms, completed_at_ms, deleted_at, updated_at_ms
		FROM timers
		WHERE id = ? AND deleted_at IS NULL
	`
	row := r.db.QueryRow(query, id)
	return r.scanTimer(row)
}

// GetAll retrieves all non-deleted timers
func (r *TimerRepository) GetAll() ([]*models.Timer, error) {
	query := `
		SELECT id, label, duration_ms, remaining_ms, status, started_at_ms, paused_at_ms, completed_at_ms, deleted_at, updated_at_ms
		FROM timers
		WHERE deleted_at IS NULL
		ORDER BY id DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTimers(rows)
}

// GetByStatus retrieves timers by status
func (r *TimerRepository) GetByStatus(status models.TimerStatus) ([]*models.Timer, error) {
	query := `
		SELECT id, label, duration_ms, remaining_ms, status, started_at_ms, paused_at_ms, completed_at_ms, deleted_at, updated_at_ms
		FROM timers
		WHERE status = ? AND deleted_at IS NULL
		ORDER BY id DESC
	`
	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTimers(rows)
}

// GetRunning retrieves all running timers
func (r *TimerRepository) GetRunning() ([]*models.Timer, error) {
	return r.GetByStatus(models.TimerRunning)
}

// Update updates a timer
func (r *TimerRepository) Update(timer *models.Timer) error {
	query := `
		UPDATE timers
		SET label = ?, duration_ms = ?, remaining_ms = ?, status = ?,
		    started_at_ms = ?, paused_at_ms = ?, completed_at_ms = ?, updated_at_ms = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query,
		timer.Label,
		timer.DurationMs,
		timer.RemainingMs,
		timer.Status,
		timer.StartedAtMs,
		timer.PausedAtMs,
		timer.CompletedAtMs,
		timer.UpdatedAtMs,
		timer.ID,
	)
	return err
}

// SoftDelete marks a timer as deleted
func (r *TimerRepository) SoftDelete(id int64) error {
	query := `UPDATE timers SET deleted_at = ?, updated_at_ms = ? WHERE id = ? AND deleted_at IS NULL`
	nowSec := time.Now().Unix()
	nowMs := time.Now().UnixMilli()
	_, err := r.db.Exec(query, nowSec, nowMs, id)
	return err
}

// Delete permanently deletes a timer
func (r *TimerRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM timers WHERE id = ?", id)
	return err
}

func (r *TimerRepository) scanTimer(row *sql.Row) (*models.Timer, error) {
	t := &models.Timer{}
	err := row.Scan(
		&t.ID, &t.Label, &t.DurationMs, &t.RemainingMs, &t.Status,
		&t.StartedAtMs, &t.PausedAtMs, &t.CompletedAtMs, &t.DeletedAt, &t.UpdatedAtMs,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("timer not found")
	}
	return t, err
}

func (r *TimerRepository) scanTimers(rows *sql.Rows) ([]*models.Timer, error) {
	var timers []*models.Timer
	for rows.Next() {
		t := &models.Timer{}
		err := rows.Scan(
			&t.ID, &t.Label, &t.DurationMs, &t.RemainingMs, &t.Status,
			&t.StartedAtMs, &t.PausedAtMs, &t.CompletedAtMs, &t.DeletedAt, &t.UpdatedAtMs,
		)
		if err != nil {
			return nil, err
		}
		timers = append(timers, t)
	}
	return timers, rows.Err()
}
