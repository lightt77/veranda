package repository

import (
	"database/sql"

	"github.com/lightt77/veranda/internal/db"
	"github.com/lightt77/veranda/internal/models"
)

// StopwatchLapRepository handles stopwatch lap data access
type StopwatchLapRepository struct {
	db *db.DB
}

// NewStopwatchLapRepository creates a new lap repository
func NewStopwatchLapRepository(database *db.DB) *StopwatchLapRepository {
	return &StopwatchLapRepository{db: database}
}

// Create inserts a new lap
func (r *StopwatchLapRepository) Create(lap *models.StopwatchLap) error {
	query := `
		INSERT INTO stopwatch_laps (stopwatch_id, lap_number, lap_duration_ms, total_elapsed_ms, created_at_ms)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		lap.StopwatchID,
		lap.LapNumber,
		lap.LapDurationMs,
		lap.TotalElapsedMs,
		lap.CreatedAtMs,
	)
	if err != nil {
		return err
	}
	lap.ID, _ = result.LastInsertId()
	return nil
}

// GetByStopwatchID retrieves all laps for a stopwatch
func (r *StopwatchLapRepository) GetByStopwatchID(stopwatchID int64) ([]*models.StopwatchLap, error) {
	query := `
		SELECT id, stopwatch_id, lap_number, lap_duration_ms, total_elapsed_ms, created_at_ms
		FROM stopwatch_laps
		WHERE stopwatch_id = ?
		ORDER BY lap_number ASC
	`
	rows, err := r.db.Query(query, stopwatchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLaps(rows)
}

// GetLastLapNumber gets the highest lap number for a stopwatch
func (r *StopwatchLapRepository) GetLastLapNumber(stopwatchID int64) (int, error) {
	query := `SELECT COALESCE(MAX(lap_number), 0) FROM stopwatch_laps WHERE stopwatch_id = ?`
	var lastLap int
	err := r.db.QueryRow(query, stopwatchID).Scan(&lastLap)
	return lastLap, err
}

// DeleteByStopwatchID deletes all laps for a stopwatch
func (r *StopwatchLapRepository) DeleteByStopwatchID(stopwatchID int64) error {
	_, err := r.db.Exec("DELETE FROM stopwatch_laps WHERE stopwatch_id = ?", stopwatchID)
	return err
}

func (r *StopwatchLapRepository) scanLaps(rows *sql.Rows) ([]*models.StopwatchLap, error) {
	var laps []*models.StopwatchLap
	for rows.Next() {
		lap := &models.StopwatchLap{}
		err := rows.Scan(
			&lap.ID, &lap.StopwatchID, &lap.LapNumber, &lap.LapDurationMs, &lap.TotalElapsedMs, &lap.CreatedAtMs,
		)
		if err != nil {
			return nil, err
		}
		laps = append(laps, lap)
	}
	return laps, rows.Err()
}
