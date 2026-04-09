// Package config contains application-wide configuration constants and settings.
package config

import (
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
	AudioFadeInDuration  = 3 * 1000       // 3 seconds fade in
	AudioFadeOutDuration = 5 * 1000       // 5 seconds fade out
)

// Default paths (relative to home directory)
const (
	DefaultDataDir   = ".veranda/data"
	DefaultLogsDir   = ".veranda/logs"
	DefaultSoundsDir = ".veranda/sounds"
)

// Database
const (
	DatabaseFileName = "veranda.db"
)

// Config returns the full application configuration
func Config() *AppConfig {
	homeDir, _ := os.UserHomeDir()

	return &AppConfig{
		AppName:    AppName,
		AppVersion: AppVersion,
		DataDir:    filepath.Join(homeDir, DefaultDataDir),
		LogsDir:    filepath.Join(homeDir, DefaultLogsDir),
		SoundsDir:  filepath.Join(homeDir, DefaultSoundsDir),
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
	DatabasePath string
	DaemonPort   int
}
