// Package cmd provides CLI commands using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/lightt77/veranda/internal/config"
	"github.com/spf13/cobra"
)

var (
	// Version is set during build
	Version = "dev"

	// RootCmd is the root command
	RootCmd = &cobra.Command{
		Use:   "veranda",
		Short: "A productivity CLI for timers, stopwatches, and focus",
		Long: `Veranda is a productivity CLI tool with:
- Timers and stopwatches with millisecond precision
- Ambient sound support
- Journal logging
- TUI interface

Run 'veranda --help' for more information.`,
		Version: Version,
	}
)

// Execute runs the root command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	RootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.veranda/config.yaml)")
}

// getConfig returns the application configuration
func getConfig() *config.AppConfig {
	return config.Config()
}
