package notify_test

import (
	"testing"

	"github.com/lightt77/veranda/internal/notify"
)

func TestNewService(t *testing.T) {
	service := notify.NewService("test-app")
	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if !service.IsEnabled() {
		t.Error("Expected service to be enabled by default")
	}
}

func TestSetEnabled(t *testing.T) {
	service := notify.NewService("test-app")

	// Disable
	service.SetEnabled(false)
	if service.IsEnabled() {
		t.Error("Expected service to be disabled")
	}

	// Enable
	service.SetEnabled(true)
	if !service.IsEnabled() {
		t.Error("Expected service to be enabled")
	}
}

func TestNotifyDisabled(t *testing.T) {
	service := notify.NewService("test-app")
	service.SetEnabled(false)

	// Should not error when disabled
	err := service.Notify("Test", "Message")
	if err != nil {
		t.Errorf("Expected no error when disabled, got: %v", err)
	}
}

func TestPlatformDetection(t *testing.T) {
	platform := notify.GetPlatform()
	if platform == "" {
		t.Error("Expected platform to be detected")
	}

	// Log the platform for info
	t.Logf("Running on platform: %s", platform)
}

func TestIsSupported(t *testing.T) {
	supported := notify.IsSupported()
	platform := notify.GetPlatform()

	// Check that supported platforms are correctly identified
	switch platform {
	case "darwin", "linux", "windows":
		if !supported {
			t.Errorf("Expected %s to be supported", platform)
		}
	default:
		if supported {
			t.Errorf("Expected %s to not be supported", platform)
		}
	}
}

func TestNotifyf(t *testing.T) {
	service := notify.NewService("test-app")
	service.SetEnabled(false) // Disable to avoid actual notifications during test

	// Should not error with formatting
	err := service.Notifyf("Title %s", "Message %s %d", "test", 123)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestTimerCompletion(t *testing.T) {
	service := notify.NewService("test-app")
	service.SetEnabled(false)

	err := service.TimerCompletion("Pomodoro Timer")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestStopwatchLap(t *testing.T) {
	service := notify.NewService("test-app")
	service.SetEnabled(false)

	err := service.StopwatchLap("Workout", 3, "01:45")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestDaemonStarted(t *testing.T) {
	service := notify.NewService("test-app")
	service.SetEnabled(false)

	err := service.DaemonStarted()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
