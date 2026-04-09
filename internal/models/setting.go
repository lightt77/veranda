package models

import "time"

// Setting represents a key-value setting
type Setting struct {
	Key         string `json:"key"`           // Setting identifier
	Value       string `json:"value"`         // Setting value
	UpdatedAtMs int64  `json:"updated_at_ms"` // Last update timestamp
}

// NewSetting creates a new setting
func NewSetting(key, value string, now time.Time) *Setting {
	return &Setting{
		Key:         key,
		Value:       value,
		UpdatedAtMs: now.UnixMilli(),
	}
}
