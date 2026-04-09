// Package notify provides notification integration with services.
package notify

import (
	"fmt"
	"time"

	"github.com/lightt77/veranda/internal/service"
)

// Notifier handles notifications for service events
type Notifier struct {
	notifyService *Service
	timerService  *service.TimerService
	enabled       bool
}

// NewNotifier creates a new notifier
func NewNotifier(notifyService *Service, timerService *service.TimerService) *Notifier {
	return &Notifier{
		notifyService: notifyService,
		timerService:  timerService,
		enabled:       true,
	}
}

// Start starts the notification monitoring
func (n *Notifier) Start() {
	if n.notifyService == nil || !n.enabled {
		return
	}

	// Send daemon started notification
	n.notifyService.DaemonStarted()

	// Start monitoring for timer completions
	go n.monitorTimers()
}

// monitorTimers periodically checks for timer completions
func (n *Notifier) monitorTimers() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Track notified timers to avoid duplicate notifications
	notifiedTimers := make(map[int64]bool)

	for range ticker.C {
		if !n.enabled || !n.notifyService.IsEnabled() {
			continue
		}

		// Get running timers
		timers, err := n.timerService.GetRunning()
		if err != nil {
			continue
		}

		now := time.Now()
		for _, timer := range timers {
			if timer.IsExpired(now) {
				// Check if we already notified for this timer
				if !notifiedTimers[timer.ID] {
					n.notifyService.TimerCompletion(timer.Label)
					notifiedTimers[timer.ID] = true
				}

				// Mark timer as completed
				n.timerService.Complete(timer.ID)
			}
		}

		// Clean up old entries from notifiedTimers periodically
		if len(notifiedTimers) > 100 {
			notifiedTimers = make(map[int64]bool)
		}
	}
}

// StopLapNotification sends a notification for a stopwatch lap
func (n *Notifier) StopLapNotification(stopwatchName string, lapNumber int, lapDurationMs int64) {
	if !n.enabled || n.notifyService == nil {
		return
	}

	lapTime := formatDuration(lapDurationMs)
	n.notifyService.StopwatchLap(stopwatchName, lapNumber, lapTime)
}

// formatDuration formats milliseconds as a human-readable string
func formatDuration(ms int64) string {
	if ms < 0 {
		ms = 0
	}

	hours := ms / (60 * 60 * 1000)
	ms %= 60 * 60 * 1000

	minutes := ms / (60 * 1000)
	ms %= 60 * 1000

	seconds := ms / 1000

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
