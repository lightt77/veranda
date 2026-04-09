package models

import "time"

// ScheduleType represents the type of scheduling
type ScheduleType string

const (
	ScheduleTypeCron     ScheduleType = "cron"
	ScheduleTypeInterval ScheduleType = "interval"
	ScheduleTypeOnce     ScheduleType = "once"
)

// ScheduledTask represents a recurring or scheduled task
type ScheduledTask struct {
	ID                int64        `json:"id"`                   // Unix epoch seconds
	TaskName          string       `json:"task_name"`            // Task identifier
	ScheduleType      ScheduleType `json:"schedule_type"`        // Type of schedule
	ScheduleData      string       `json:"schedule_data"`        // JSON config
	NextDueAtMs       *int64       `json:"next_due_at_ms"`       // Next scheduled run
	LastCompletedAtMs *int64       `json:"last_completed_at_ms"` // Last completion
	DeletedAt         *int64       `json:"deleted_at"`           // Soft delete timestamp
	CreatedAtMs       int64        `json:"created_at_ms"`        // Creation timestamp
	UpdatedAtMs       int64        `json:"updated_at_ms"`        // Last update timestamp
}

// IsDeleted returns true if the task is soft-deleted
func (s *ScheduledTask) IsDeleted() bool {
	return s.DeletedAt != nil
}

// IsDue returns true if the task is due to run
func (s *ScheduledTask) IsDue(now time.Time) bool {
	if s.NextDueAtMs == nil {
		return false
	}
	return now.UnixMilli() >= *s.NextDueAtMs
}

// MarkCompleted marks the task as completed and updates next due time
func (s *ScheduledTask) MarkCompleted(now time.Time) {
	nowMs := now.UnixMilli()
	s.LastCompletedAtMs = &nowMs
	s.UpdatedAtMs = nowMs
	// NextDueAtMs would be calculated based on ScheduleType and ScheduleData
	// This is a placeholder - actual logic will be implemented in the service layer
}

// NewScheduledTask creates a new scheduled task
func NewScheduledTask(id int64, taskName string, scheduleType ScheduleType, scheduleData string, now time.Time) *ScheduledTask {
	nowMs := now.UnixMilli()
	return &ScheduledTask{
		ID:           id,
		TaskName:     taskName,
		ScheduleType: scheduleType,
		ScheduleData: scheduleData,
		CreatedAtMs:  nowMs,
		UpdatedAtMs:  nowMs,
	}
}
