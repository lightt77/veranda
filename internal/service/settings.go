package service

import (
	"fmt"

	"github.com/lightt77/veranda/internal/repository"
)

// SettingsService handles settings business logic
type SettingsService struct {
	settingsRepo *repository.SettingsRepository
}

// NewSettingsService creates a new settings service
func NewSettingsService(settingsRepo *repository.SettingsRepository) *SettingsService {
	return &SettingsService{
		settingsRepo: settingsRepo,
	}
}

// Get retrieves a setting value by key
func (s *SettingsService) Get(key string) (string, error) {
	return s.settingsRepo.Get(key)
}

// GetOrDefault retrieves a setting with a fallback default value
func (s *SettingsService) GetOrDefault(key, defaultValue string) (string, error) {
	return s.settingsRepo.GetOrDefault(key, defaultValue)
}

// Set sets a setting value
func (s *SettingsService) Set(key, value string) error {
	if err := s.settingsRepo.Set(key, value); err != nil {
		return fmt.Errorf("failed to set setting: %w", err)
	}
	return nil
}

// Delete removes a setting
func (s *SettingsService) Delete(key string) error {
	return s.settingsRepo.Delete(key)
}

// GetAll retrieves all settings as a map
func (s *SettingsService) GetAll() (map[string]string, error) {
	return s.settingsRepo.GetAll()
}

// Common setting keys
const (
	SettingTUIRefreshMs    = "tui_refresh_ms"
	SettingAudioFadeInMs   = "audio_fade_in_ms"
	SettingAudioFadeOutMs  = "audio_fade_out_ms"
	SettingDefaultTimerMin = "default_timer_min"
	SettingSoundsDir       = "sounds_dir"
)

// GetTUIRefreshMs gets the TUI refresh interval in milliseconds
func (s *SettingsService) GetTUIRefreshMs() (int64, error) {
	value, err := s.GetOrDefault(SettingTUIRefreshMs, "500")
	if err != nil {
		return 500, err
	}
	var ms int64
	_, err = fmt.Sscanf(value, "%d", &ms)
	if err != nil {
		return 500, nil // fallback
	}
	return ms, nil
}

// GetDefaultTimerMin gets the default timer duration in minutes
func (s *SettingsService) GetDefaultTimerMin() (int, error) {
	value, err := s.GetOrDefault(SettingDefaultTimerMin, "25")
	if err != nil {
		return 25, err
	}
	var min int
	_, err = fmt.Sscanf(value, "%d", &min)
	if err != nil {
		return 25, nil // fallback
	}
	return min, nil
}
