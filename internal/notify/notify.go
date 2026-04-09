// Package notify provides desktop notification functionality.
package notify

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/gen2brain/beeep"
)

// Service handles desktop notifications
type Service struct {
	enabled bool
	mutex   sync.RWMutex
	appName string
}

// NewService creates a new notification service
func NewService(appName string) *Service {
	return &Service{
		enabled: true,
		appName: appName,
	}
}

// IsEnabled returns whether notifications are enabled
func (s *Service) IsEnabled() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.enabled
}

// SetEnabled enables or disables notifications
func (s *Service) SetEnabled(enabled bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.enabled = enabled
}

// Notify sends a desktop notification
func (s *Service) Notify(title, message string) error {
	s.mutex.RLock()
	if !s.enabled {
		s.mutex.RUnlock()
		return nil
	}
	s.mutex.RUnlock()

	// Use beeep for cross-platform notifications
	return beeep.Notify(title, message, "")
}

// Notifyf sends a formatted desktop notification
func (s *Service) Notifyf(title, format string, args ...interface{}) error {
	return s.Notify(title, fmt.Sprintf(format, args...))
}

// Alert sends a critical alert notification
func (s *Service) Alert(title, message string) error {
	s.mutex.RLock()
	if !s.enabled {
		s.mutex.RUnlock()
		return nil
	}
	s.mutex.RUnlock()

	return beeep.Alert(title, message, "")
}

// Beep plays a beep sound
func (s *Service) Beep() error {
	return beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
}

// TimerCompletion notifies that a timer has completed
func (s *Service) TimerCompletion(timerName string) error {
	return s.Notifyf(
		"⏱️ Timer Complete",
		"%s has finished!",
		timerName,
	)
}

// StopwatchLap notifies of a stopwatch lap time
func (s *Service) StopwatchLap(stopwatchName string, lapNumber int, lapTime string) error {
	return s.Notifyf(
		"⏱️ Lap Recorded",
		"%s - Lap %d: %s",
		stopwatchName,
		lapNumber,
		lapTime,
	)
}

// DaemonStarted notifies that the daemon has started
func (s *Service) DaemonStarted() error {
	return s.Notify(
		"🚀 Veranda Started",
		"The Veranda daemon is now running",
	)
}

// GetPlatform returns the current operating system
func GetPlatform() string {
	return runtime.GOOS
}

// IsSupported returns whether notifications are supported on this platform
func IsSupported() bool {
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		return true
	default:
		return false
	}
}
