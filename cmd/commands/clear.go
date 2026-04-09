package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/lightt77/veranda/internal/db"
	"github.com/spf13/cobra"
)

var (
	clearForce bool

	clearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear all data (timers, stopwatches, journal entries)",
		Long: `Clear all data from the Veranda database.

This will soft-delete all timers, stopwatches, journal entries, and settings.
This action cannot be undone!

Use --force or -f to skip the confirmation prompt.`,
		Run: runClear,
	}
)

func init() {
	RootCmd.AddCommand(clearCmd)
	clearCmd.Flags().BoolVarP(&clearForce, "force", "f", false, "Skip confirmation prompt")
}

func runClear(cmd *cobra.Command, args []string) {
	cfg := getConfig()

	// Confirmation prompt
	if !clearForce {
		fmt.Print("⚠️  This will delete ALL data (timers, stopwatches, journal entries, settings).\n")
		fmt.Print("Are you sure? Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			fmt.Println("Cancelled.")
			return
		}
	}

	// Open database directly (no need for daemon)
	database, err := db.Open(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Clear all data
	if err := database.ClearAllData(); err != nil {
		fmt.Fprintf(os.Stderr, "Error clearing data: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ All data cleared successfully.")
}
