package models

import "time"

// JournalEntry represents a single journal log entry
type JournalEntry struct {
	ID          int64  `json:"id"`            // Unix epoch seconds
	Log         string `json:"log"`           // Entry content
	DeletedAt   *int64 `json:"deleted_at"`    // Soft delete timestamp
	CreatedAtMs int64  `json:"created_at_ms"` // Creation timestamp
}

// IsDeleted returns true if the entry is soft-deleted
func (j *JournalEntry) IsDeleted() bool {
	return j.DeletedAt != nil
}

// NewJournalEntry creates a new journal entry
func NewJournalEntry(id int64, log string, now time.Time) *JournalEntry {
	return &JournalEntry{
		ID:          id,
		Log:         log,
		CreatedAtMs: now.UnixMilli(),
	}
}
