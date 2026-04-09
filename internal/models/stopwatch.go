package models

import "time"

// StopwatchStatus represents the state of a stopwatch
type StopwatchStatus string

const (
	StopwatchStopped StopwatchStatus = "stopped"
	StopwatchRunning StopwatchStatus = "running"
)

// Stopwatch represents a stopwatch/timer
type Stopwatch struct {
	ID          int64           `json:"id"`            // Unix epoch seconds
	Label       string          `json:"label"`         // Stopwatch name
	ElapsedMs   int64           `json:"elapsed_ms"`    // Total elapsed time in ms
	Status      StopwatchStatus `json:"status"`        // Current status
	StartedAtMs *int64          `json:"started_at_ms"` // When started
	StoppedAtMs *int64          `json:"stopped_at_ms"` // When stopped
	DeletedAt   *int64          `json:"deleted_at"`    // Soft delete (epoch seconds)
	UpdatedAtMs int64           `json:"updated_at_ms"` // Last update (epoch ms)
}

// IsDeleted returns true if the stopwatch is soft-deleted
func (s *Stopwatch) IsDeleted() bool {
	return s.DeletedAt != nil
}

// Start starts the stopwatch
func (s *Stopwatch) Start(now time.Time) {
	nowMs := now.UnixMilli()
	s.Status = StopwatchRunning
	s.StartedAtMs = &nowMs
	s.UpdatedAtMs = nowMs
}

// Stop stops the stopwatch and updates elapsed time
func (s *Stopwatch) Stop(now time.Time) {
	nowMs := now.UnixMilli()
	if s.Status == StopwatchRunning && s.StartedAtMs != nil {
		s.ElapsedMs += nowMs - *s.StartedAtMs
	}
	s.Status = StopwatchStopped
	s.StoppedAtMs = &nowMs
	s.UpdatedAtMs = nowMs
}

// GetCurrentElapsed returns the current elapsed time accounting for running time
func (s *Stopwatch) GetCurrentElapsed(now time.Time) int64 {
	elapsed := s.ElapsedMs
	if s.Status == StopwatchRunning && s.StartedAtMs != nil {
		elapsed += now.UnixMilli() - *s.StartedAtMs
	}
	return elapsed
}

// Lap records a lap and returns the lap duration
func (s *Stopwatch) Lap(now time.Time) int64 {
	currentElapsed := s.GetCurrentElapsed(now)
	lapDuration := currentElapsed - s.ElapsedMs
	s.ElapsedMs = currentElapsed
	s.UpdatedAtMs = now.UnixMilli()
	return lapDuration
}

// Reset resets the stopwatch to zero
func (s *Stopwatch) Reset(now time.Time) {
	s.ElapsedMs = 0
	s.Status = StopwatchStopped
	s.StartedAtMs = nil
	s.StoppedAtMs = nil
	s.UpdatedAtMs = now.UnixMilli()
}

// IsRunning returns true if the stopwatch is currently running
func (s *Stopwatch) IsRunning() bool {
	return s.Status == StopwatchRunning
}

// NewStopwatch creates a new stopwatch
func NewStopwatch(id int64, label string, now time.Time) *Stopwatch {
	return &Stopwatch{
		ID:          id,
		Label:       label,
		ElapsedMs:   0,
		Status:      StopwatchStopped,
		UpdatedAtMs: now.UnixMilli(),
	}
}
