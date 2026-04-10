package notify_test

import (
	"testing"

	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/notify"
)

// mockSettingsRepository is a mock implementation of config.SettingsRepository for testing
type mockSettingsRepository struct {
	data map[string]string
}

func newMockSettingsRepository() *mockSettingsRepository {
	return &mockSettingsRepository{
		data: map[string]string{
			"timer_completion_sound": "bell.mp3",
			"chime_volume":           "0.5",
		},
	}
}

func (m *mockSettingsRepository) Get(key string) (string, error) {
	return m.data[key], nil
}

func (m *mockSettingsRepository) Set(key, value string) error {
	m.data[key] = value
	return nil
}

func TestNewService(t *testing.T) {
	mockRepo := newMockSettingsRepository()
	service := notify.NewService("test-app", nil, mockRepo)
	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if !service.IsEnabled() {
		t.Error("Expected service to be enabled by default")
	}
}

func TestSetEnabled(t *testing.T) {
	mockRepo := newMockSettingsRepository()
	service := notify.NewService("test-app", nil, mockRepo)

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
	mockRepo := newMockSettingsRepository()
	service := notify.NewService("test-app", nil, mockRepo)
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
	mockRepo := newMockSettingsRepository()
	service := notify.NewService("test-app", nil, mockRepo)
	service.SetEnabled(false) // Disable to avoid actual notifications during test

	// Should not error with formatting
	err := service.Notifyf("Title %s", "Message %s %d", "test", 123)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestTimerCompletion(t *testing.T) {
	mockRepo := newMockSettingsRepository()
	service := notify.NewService("test-app", nil, mockRepo)
	service.SetEnabled(false)

	err := service.TimerCompletion("Pomodoro Timer")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestStopwatchLap(t *testing.T) {
	mockRepo := newMockSettingsRepository()
	service := notify.NewService("test-app", nil, mockRepo)
	service.SetEnabled(false)

	err := service.StopwatchLap("Workout", 3, "01:45")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestDaemonStarted(t *testing.T) {
	mockRepo := newMockSettingsRepository()
	service := notify.NewService("test-app", nil, mockRepo)
	service.SetEnabled(false)

	err := service.DaemonStarted()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
