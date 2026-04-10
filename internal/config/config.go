// Package config contains application-wide configuration constants and settings.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Application metadata
const (
	AppName        = "veranda"
	AppVersion     = "0.1.0"
	DaemonHTTPPort = 17342 // "VER" in phone keypad: 8=V, 3=E, 7=R... close enough
)

// Timing constants
const (
	DefaultTimerDuration = 25 * 60 * 1000 // 25 minutes in milliseconds
	DefaultTUIRefreshMs  = 500            // TUI refresh interval in milliseconds
	AudioFadeInDuration  = 2 * 1000       // 2 seconds fade in
	AudioFadeOutDuration = 2 * 1000       // 2 seconds fade out
)

// Default paths (relative to home directory)
const (
	DefaultDataDir     = ".veranda/data"
	DefaultLogsDir     = ".veranda/logs"
	DefaultSoundsDir   = ".veranda/sounds"
	DefaultChimesDir   = ".veranda/sounds/chimes"
	DefaultAmbienceDir = ".veranda/sounds/ambience"
)

// Database
const (
	DatabaseFileName = "veranda.db"
)

// SettingsRepository defines the interface for loading/saving settings
// This is implemented by repository.SettingsRepository
type SettingsRepository interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

// Default user settings
const (
	DefaultTimerCompletionSound = "freesound_community-bell-98033.mp3"
	DefaultChimeVolume          = 0.5 // 50% volume
	DefaultAmbienceVolume       = 1.0 // 100% volume
)

// Config returns the full application configuration
func Config() *AppConfig {
	homeDir, _ := os.UserHomeDir()

	return &AppConfig{
		AppName:     AppName,
		AppVersion:  AppVersion,
		DataDir:     filepath.Join(homeDir, DefaultDataDir),
		LogsDir:     filepath.Join(homeDir, DefaultLogsDir),
		SoundsDir:   filepath.Join(homeDir, DefaultSoundsDir),
		ChimesDir:   filepath.Join(homeDir, DefaultChimesDir),
		AmbienceDir: filepath.Join(homeDir, DefaultAmbienceDir),
		DatabasePath: filepath.Join(
			filepath.Join(homeDir, DefaultDataDir),
			DatabaseFileName,
		),
		DaemonPort: DaemonHTTPPort,
	}
}

// AppConfig holds all application configuration
type AppConfig struct {
	AppName      string
	AppVersion   string
	DataDir      string
	LogsDir      string
	SoundsDir    string
	ChimesDir    string
	AmbienceDir  string
	DatabasePath string
	DaemonPort   int
}

// AmbienceSoundConfig holds configuration for a single ambience sound
type AmbienceSoundConfig struct {
	Filename string  `json:"filename"`
	Volume   float64 `json:"volume"`
	Enabled  bool    `json:"enabled"`
}

// UserConfig holds user-editable configuration settings
type UserConfig struct {
	TimerCompletionSound string                `json:"timer_completion_sound"`
	ChimeVolume          float64               `json:"chime_volume"`
	AmbienceEnabled      bool                  `json:"ambience_enabled"` // Global ambience on/off switch
	AmbienceSounds       []AmbienceSoundConfig `json:"ambience_sounds"`
	SkyfieldMode         string                `json:"skyfield_mode"`
	SkyfieldShowNames    bool                  `json:"skyfield_show_names"`
}

// DefaultUserConfig returns the default user configuration
func DefaultUserConfig() *UserConfig {
	return &UserConfig{
		TimerCompletionSound: DefaultTimerCompletionSound,
		ChimeVolume:          DefaultChimeVolume,
		AmbienceEnabled:      true,                    // Ambience sounds enabled by default
		AmbienceSounds:       []AmbienceSoundConfig{}, // Empty by default, populated on first run with available files
		SkyfieldMode:         "random",
		SkyfieldShowNames:    false,
	}
}

// LoadUserConfig loads the user configuration from the database
// Returns default config if settings don't exist
func LoadUserConfig(repo SettingsRepository) *UserConfig {
	cfg := DefaultUserConfig()

	// Load each setting from DB, using defaults if not found
	if val, err := repo.Get("timer_completion_sound"); err == nil && val != "" {
		cfg.TimerCompletionSound = val
	}
	if val, err := repo.Get("chime_volume"); err == nil && val != "" {
		if vol, err := strconv.ParseFloat(val, 64); err == nil {
			cfg.ChimeVolume = vol
		}
	}
	if val, err := repo.Get("ambience_enabled"); err == nil && val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			cfg.AmbienceEnabled = enabled
		}
	}
	if val, err := repo.Get("ambience_sounds"); err == nil && val != "" {
		if err := json.Unmarshal([]byte(val), &cfg.AmbienceSounds); err != nil {
			// Keep default empty slice if unmarshal fails
			cfg.AmbienceSounds = []AmbienceSoundConfig{}
		}
	}
	if val, err := repo.Get("skyfield_mode"); err == nil && val != "" {
		cfg.SkyfieldMode = val
	}
	if val, err := repo.Get("skyfield_show_names"); err == nil && val != "" {
		if show, err := strconv.ParseBool(val); err == nil {
			cfg.SkyfieldShowNames = show
		}
	}

	return cfg
}

// Save persists the user configuration to the database
func (c *UserConfig) Save(repo SettingsRepository) error {
	// Save each setting
	if err := repo.Set("timer_completion_sound", c.TimerCompletionSound); err != nil {
		return fmt.Errorf("failed to save timer_completion_sound: %w", err)
	}
	if err := repo.Set("chime_volume", strconv.FormatFloat(c.ChimeVolume, 'f', 2, 64)); err != nil {
		return fmt.Errorf("failed to save chime_volume: %w", err)
	}
	if err := repo.Set("ambience_enabled", strconv.FormatBool(c.AmbienceEnabled)); err != nil {
		return fmt.Errorf("failed to save ambience_enabled: %w", err)
	}
	soundsJSON, err := json.Marshal(c.AmbienceSounds)
	if err != nil {
		return fmt.Errorf("failed to marshal ambience_sounds: %w", err)
	}
	if err := repo.Set("ambience_sounds", string(soundsJSON)); err != nil {
		return fmt.Errorf("failed to save ambience_sounds: %w", err)
	}
	if err := repo.Set("skyfield_mode", c.SkyfieldMode); err != nil {
		return fmt.Errorf("failed to save skyfield_mode: %w", err)
	}
	if err := repo.Set("skyfield_show_names", strconv.FormatBool(c.SkyfieldShowNames)); err != nil {
		return fmt.Errorf("failed to save skyfield_show_names: %w", err)
	}

	return nil
}
