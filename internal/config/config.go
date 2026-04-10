// Package config contains application-wide configuration constants and settings.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// User config file
const (
	ConfigFileName = "config.json"
)

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

// LoadUserConfig loads the user configuration from disk
// Returns default config if file doesn't exist or is invalid
func LoadUserConfig(configDir string) *UserConfig {
	configPath := filepath.Join(configDir, ConfigFileName)

	// If file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultUserConfig()
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultUserConfig()
	}

	// Parse JSON
	var cfg UserConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultUserConfig()
	}

	// Set defaults for empty values
	if cfg.TimerCompletionSound == "" {
		cfg.TimerCompletionSound = DefaultTimerCompletionSound
	}
	if cfg.ChimeVolume == 0 {
		cfg.ChimeVolume = DefaultChimeVolume
	}
	if cfg.SkyfieldMode == "" {
		cfg.SkyfieldMode = "random"
	}
	// Note: AmbienceSounds is intentionally not auto-populated here.
	// The TUI will initialize it with available files on first access.

	return &cfg
}

// Save persists the user configuration to disk
func (c *UserConfig) Save(configDir string) error {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, ConfigFileName)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
