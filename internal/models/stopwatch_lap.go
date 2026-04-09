package models

import "time"

// StopwatchLap represents a single lap in a stopwatch session
type StopwatchLap struct {
	ID             int64 `json:"id"`               // Auto-increment ID
	StopwatchID    int64 `json:"stopwatch_id"`     // Parent stopwatch ID
	LapNumber      int   `json:"lap_number"`       // Sequential lap number
	LapDurationMs  int64 `json:"lap_duration_ms"`  // Duration of this lap
	TotalElapsedMs int64 `json:"total_elapsed_ms"` // Total elapsed at lap completion
	CreatedAtMs    int64 `json:"created_at_ms"`    // Creation timestamp
}

// NewStopwatchLap creates a new lap record
func NewStopwatchLap(stopwatchID int64, lapNumber int, lapDurationMs, totalElapsedMs int64, now time.Time) *StopwatchLap {
	return &StopwatchLap{
		StopwatchID:    stopwatchID,
		LapNumber:      lapNumber,
		LapDurationMs:  lapDurationMs,
		TotalElapsedMs: totalElapsedMs,
		CreatedAtMs:    now.UnixMilli(),
	}
}
