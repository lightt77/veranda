package main

import (
	"fmt"
	"os"

	"github.com/lightt77/veranda/internal/audio"
	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/daemon"
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
	if err := os.MkdirAll(cfg.SoundsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating sounds directory: %v\n", err)
		os.Exit(1)
	}

	// Open database
	database, err := db.Open(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize repositories
	timerRepo := repository.NewTimerRepository(database)
	stopwatchRepo := repository.NewStopwatchRepository(database)
	lapRepo := repository.NewStopwatchLapRepository(database)
	journalRepo := repository.NewJournalRepository(database)
	settingsRepo := repository.NewSettingsRepository(database)

	// Initialize services
	timerService := service.NewTimerService(timerRepo)
	stopwatchService := service.NewStopwatchService(stopwatchRepo, lapRepo)
	journalService := service.NewJournalService(journalRepo)
	settingsService := service.NewSettingsService(settingsRepo)

	// Initialize ambient sound service (optional - won't fail if no audio files)
	var ambientService *audio.AmbientService
	ambientService, err = audio.NewAmbientService(cfg, timerService, stopwatchService)
	if err != nil {
		fmt.Printf("Note: Ambient sound not available: %v\n", err)
	} else {
		if err := ambientService.Start(); err != nil {
			fmt.Printf("Note: Failed to start ambient monitoring: %v\n", err)
		}
	}

	// Initialize and start daemon server
	daemonServer := daemon.NewServer(
		cfg,
		timerService,
		stopwatchService,
		journalService,
		settingsService,
		ambientService,
	)

	if err := daemonServer.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting daemon: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Veranda %s\n", cfg.AppVersion)
	fmt.Printf("Database: %s\n", cfg.DatabasePath)
	fmt.Printf("Daemon: http://localhost:%d\n", cfg.DaemonPort)
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

	// Check for ambient sounds
	if ambientService != nil {
		fmt.Printf("\nAmbient sound service: initialized\n")
		fmt.Printf("Sounds directory: %s\n", cfg.SoundsDir)
		fmt.Println("Tip: Add MP3 files to the sounds directory for ambient playback")
	}

	fmt.Println("\nDaemon is running. API available at http://localhost:17342")

	// Keep the daemon running
	select {}
}
