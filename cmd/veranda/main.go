package main

import (
	"fmt"
	"os"

	"github.com/lightt77/veranda/internal/config"
)

func main() {
	cfg := config.Config()

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating data directory: %v\n", err)
		os.Exit(1)
	}

	// Ensure logs directory exists
	if err := os.MkdirAll(cfg.LogsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logs directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Veranda %s\n", cfg.AppVersion)
	fmt.Printf("Data directory: %s\n", cfg.DataDir)
	fmt.Printf("Logs directory: %s\n", cfg.LogsDir)
	fmt.Printf("Database: %s\n", cfg.DatabasePath)
	fmt.Printf("Sounds directory: %s\n", cfg.SoundsDir)
	fmt.Printf("Daemon port: %d\n", cfg.DaemonPort)
	fmt.Println("\nTODO: Implement CLI commands (timer, stopwatch, tui, etc.)")
}
