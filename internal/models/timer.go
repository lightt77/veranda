// Package models contains domain models for the application.
package models

import "time"

// TimerStatus represents the state of a timer
type TimerStatus string

const (
	TimerPending   TimerStatus = "pending"
	TimerRunning   TimerStatus = "running"
	TimerPaused    TimerStatus = "paused"
	TimerCompleted TimerStatus = "completed"
)

// Timer represents a countdown timer
type Timer struct {
	ID            int64       `json:"id"`              // Unix epoch seconds
	Label         string      `json:"label"`           // Timer name
	DurationMs    int64       `json:"duration_ms"`     // Total duration in ms
	RemainingMs   int64       `json:"remaining_ms"`    // Remaining time in ms
	Status        TimerStatus `json:"status"`          // Current status
	StartedAtMs   *int64      `json:"started_at_ms"`   // When started
	PausedAtMs    *int64      `json:"paused_at_ms"`    // When paused
	CompletedAtMs *int64      `json:"completed_at_ms"` // When completed
	DeletedAt     *int64      `json:"deleted_at"`      // Soft delete (epoch seconds)
	UpdatedAtMs   int64       `json:"updated_at_ms"`   // Last update (epoch ms)
}

// IsDeleted returns true if the timer is soft-deleted
func (t *Timer) IsDeleted() bool {
	return t.DeletedAt != nil
}

// Start marks the timer as running
func (t *Timer) Start(now time.Time) {
	nowMs := now.UnixMilli()
	t.Status = TimerRunning
	t.StartedAtMs = &nowMs
	t.UpdatedAtMs = nowMs
}

// Pause pauses a running timer
func (t *Timer) Pause(now time.Time) {
	nowMs := now.UnixMilli()
	if t.Status == TimerRunning && t.StartedAtMs != nil {
		elapsed := nowMs - *t.StartedAtMs
		t.RemainingMs -= elapsed
		if t.RemainingMs < 0 {
			t.RemainingMs = 0
		}
	}
	t.Status = TimerPaused
	t.PausedAtMs = &nowMs
	t.UpdatedAtMs = nowMs
}

// Resume resumes a paused timer
func (t *Timer) Resume(now time.Time) {
	nowMs := now.UnixMilli()
	t.Status = TimerRunning
	t.StartedAtMs = &nowMs
	t.PausedAtMs = nil
	t.UpdatedAtMs = nowMs
}

// Complete marks the timer as completed
func (t *Timer) Complete(now time.Time) {
	nowMs := now.UnixMilli()
	t.Status = TimerCompleted
	t.RemainingMs = 0
	t.CompletedAtMs = &nowMs
	t.UpdatedAtMs = nowMs
}

// GetCurrentRemaining returns the current remaining time accounting for elapsed time
func (t *Timer) GetCurrentRemaining(now time.Time) int64 {
	if t.Status != TimerRunning || t.StartedAtMs == nil {
		return t.RemainingMs
	}

	elapsed := now.UnixMilli() - *t.StartedAtMs
	remaining := t.RemainingMs - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsExpired returns true if the timer has reached zero
func (t *Timer) IsExpired(now time.Time) bool {
	return t.GetCurrentRemaining(now) <= 0
}

// NewTimer creates a new timer with the given parameters
func NewTimer(id int64, label string, durationMs int64, now time.Time) *Timer {
	nowMs := now.UnixMilli()
	return &Timer{
		ID:          id,
		Label:       label,
		DurationMs:  durationMs,
		RemainingMs: durationMs,
		Status:      TimerPending,
		UpdatedAtMs: nowMs,
	}
}
