package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lightt77/veranda/internal/audio"
	"github.com/lightt77/veranda/internal/daemon"
	"github.com/lightt77/veranda/internal/db"
	"github.com/lightt77/veranda/internal/notify"
	"github.com/lightt77/veranda/internal/repository"
	"github.com/lightt77/veranda/internal/service"
	"github.com/spf13/cobra"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the Veranda daemon",
	Long: `Start the Veranda daemon server which handles:
- Timer and stopwatch management
- Ambient sound playback
- Desktop notifications
- HTTP API for CLI commands`,
	Run: runDaemon,
}

func init() {
	RootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStopCmd)
}

// daemonStopCmd represents the daemon stop command
var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Veranda daemon",
	Long:  `Send a shutdown signal to the running Veranda daemon.`,
	Run:   runDaemonStop,
}

func runDaemonStop(cmd *cobra.Command, args []string) {
	cfg := getConfig()
	client := daemon.NewClient(cfg.DaemonPort)

	if !client.IsRunning() {
		fmt.Println("Daemon is not running")
		return
	}

	fmt.Println("Stopping daemon...")
	if err := client.StopDaemon(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping daemon: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Daemon stopped")
}

func runDaemon(cmd *cobra.Command, args []string) {
	cfg := getConfig()

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
	if err := os.MkdirAll(cfg.ChimesDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating chimes directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(cfg.AmbienceDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating ambience directory: %v\n", err)
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

	// Initialize ambient sound service
	var ambientService *audio.AmbientService
	var notifyAudioPlayer *audio.Player
	ambientService, err = audio.NewAmbientService(cfg, cfg.DataDir, timerService, stopwatchService)
	if err != nil {
		fmt.Printf("Note: Ambient sound service not available: %v\n", err)
	} else {
		if err := ambientService.Start(); err != nil {
			fmt.Printf("Note: Failed to start ambient monitoring: %v\n", err)
		} else {
			fmt.Println("Ambient sound monitoring: started")
		}
	}

	// Get the audio player for notifications (either from ambient service or create standalone)
	if ambientService != nil {
		notifyAudioPlayer = ambientService.GetPlayer()
	}
	if notifyAudioPlayer == nil {
		// Create standalone player for notifications if ambient service failed
		notifyAudioPlayer, err = audio.NewPlayer(cfg, cfg.DataDir)
		if err != nil {
			fmt.Printf("Note: Notification sounds not available: %v\n", err)
		} else {
			fmt.Println("Notification sounds: initialized")
		}
	} else {
		fmt.Println("Notification sounds: shared with ambient service")
	}

	// Initialize notification service (config will be reloaded from disk each time)
	var notifier *notify.Notifier
	notifyService := notify.NewService(cfg.AppName, notifyAudioPlayer, cfg.DataDir)
	if notify.IsSupported() {
		notifier = notify.NewNotifier(notifyService, timerService)
		notifier.Start()
		stopwatchService.SetLapNotifier(notifier)
		fmt.Println("Desktop notifications: enabled")
	} else {
		fmt.Printf("Desktop notifications: not supported on %s\n", notify.GetPlatform())
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

	fmt.Printf("Veranda %s daemon started\n", cfg.AppVersion)
	fmt.Printf("Database: %s\n", cfg.DatabasePath)
	fmt.Printf("API: http://localhost:%d\n", cfg.DaemonPort)
	fmt.Println("\nPress Ctrl+C to stop")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	daemonServer.Stop()
}
