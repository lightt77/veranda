// Package service provides business logic for all application features.
package service

import (
	"fmt"
	"time"
)

// GenerateID generates a Unix epoch timestamp in seconds for use as entity ID.
// This provides natural chronological sorting.
func GenerateID() int64 {
	return time.Now().Unix()
}

// GenerateTimestampMs generates a Unix epoch timestamp in milliseconds.
func GenerateTimestampMs() int64 {
	return time.Now().UnixMilli()
}

// ParseDurationMs converts duration string to milliseconds.
// Supported formats: "1h30m", "25m", "90s", "2h", etc.
func ParseDurationMs(duration string) (int64, error) {
	d, err := time.ParseDuration(duration)
	if err != nil {
		return 0, err
	}
	return d.Milliseconds(), nil
}

// FormatDurationMs formats milliseconds as a human-readable string.
// Examples: "25m 30s", "1h 15m", "45s"
func FormatDurationMs(ms int64) string {
	if ms < 0 {
		ms = 0
	}

	hours := ms / (60 * 60 * 1000)
	ms %= 60 * 60 * 1000

	minutes := ms / (60 * 1000)
	ms %= 60 * 1000

	seconds := ms / 1000

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	if minutes > 0 {
		if seconds > 0 {
			return fmt.Sprintf("%dm %ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	}

	return fmt.Sprintf("%ds", seconds)
}
