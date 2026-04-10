// Package tui provides the Terminal User Interface using Bubble Tea.
package tui

import (
	"fmt"
	"strings"
)

// formatDuration formats milliseconds as MM:SS or HH:MM:SS
func formatDuration(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	seconds := ms / 1000
	minutes := seconds / 60
	seconds = seconds % 60
	hours := minutes / 60
	minutes = minutes % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// formatStopwatch formats milliseconds as stopwatch time with milliseconds
func formatStopwatch(ms int64) string {
	if ms < 0 {
		ms = 0
	}

	hours := ms / (60 * 60 * 1000)
	remaining := ms % (60 * 60 * 1000)
	minutes := remaining / (60 * 1000)
	remaining = remaining % (60 * 1000)
	seconds := remaining / 1000
	milliseconds := remaining % 1000

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d:%02d.%03d", minutes, seconds, milliseconds)
	}
	return fmt.Sprintf("%d.%03d", seconds, milliseconds)
}

// renderProgressBar creates a liquid filling style progress bar
func renderProgressBar(progress float64, width int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	filledBar := strings.Repeat("█", filled)
	emptyBar := strings.Repeat("░", empty)

	return progressBarStyle.Render(filledBar) + progressBarEmptyStyle.Render(emptyBar)
}

// renderVolumeBar creates a visual volume bar
func renderVolumeBar(volume float64, width int) string {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	filled := int(volume * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	filledBar := strings.Repeat("█", filled)
	emptyBar := strings.Repeat("░", empty)

	return progressBarStyle.Render(filledBar) + progressBarEmptyStyle.Render(emptyBar)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
