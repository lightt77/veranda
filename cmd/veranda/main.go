package main

import (
	"fmt"
	"os"

	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/db"
	"github.com/lightt77/veranda/internal/repository"
	"github.com/lightt77/veranda/internal/service"
)

func main() {
	cfg := config.Config()

	// Ensure directories exist
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating data directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(cfg.LogsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logs directory: %v\n", err)
		os.Exit(1)
	}

	// Open database
	database, err := db.Open(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize services
	timerService := service.NewTimerService(repository.NewTimerRepository(database))
	stopwatchService := service.NewStopwatchService(
		repository.NewStopwatchRepository(database),
		repository.NewStopwatchLapRepository(database),
	)
	journalService := service.NewJournalService(repository.NewJournalRepository(database))
	settingsService := service.NewSettingsService(repository.NewSettingsRepository(database))

	fmt.Printf("Veranda %s\n", cfg.AppVersion)
	fmt.Printf("Database: %s\n", cfg.DatabasePath)
	fmt.Printf("Daemon port: %d\n", cfg.DaemonPort)
	fmt.Println()

	// Demo: Create a test timer
	timer, err := timerService.Create("Test Timer (Pomodoro)", 25*60*1000)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating timer: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created timer #%d: %s (%s)\n", timer.ID, timer.Label, service.FormatDurationMs(timer.DurationMs))

	// Demo: Create a test stopwatch
	sw, err := stopwatchService.Create("Test Stopwatch")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stopwatch: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created stopwatch #%d: %s\n", sw.ID, sw.Label)

	// Demo: Create a journal entry
	entry, err := journalService.Create("Veranda CLI initialized successfully!")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating journal entry: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created journal entry #%d\n", entry.ID)

	// Demo: Set a setting
	if err := settingsService.Set("initialized_at", fmt.Sprintf("%d", service.GenerateTimestampMs())); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting value: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Saved initialization timestamp to settings")

	fmt.Println()
	fmt.Println("All services working correctly!")
	fmt.Println("TODO: Implement CLI commands and TUI")
}
