package cmd

import (
	"fmt"

	"github.com/lightt77/veranda/internal/db"
	"github.com/lightt77/veranda/internal/repository"
	"github.com/lightt77/veranda/internal/tui"
	"github.com/spf13/cobra"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the Terminal User Interface",
	Long:  `Launch an interactive TUI for managing timers, stopwatches, and viewing activity.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure daemon is running before starting TUI
		if _, err := ensureDaemon(); err != nil {
			fmt.Printf("Error starting daemon: %v\n", err)
			return
		}

		// Open database for cities repository
		cfg := getConfig()
		database, err := db.Open(cfg)
		if err != nil {
			fmt.Printf("Error opening database: %v\n", err)
			return
		}
		defer database.Close()

		citiesRepo := repository.NewCitiesRepository(database)

		if err := tui.Run(cfg.DaemonPort, citiesRepo); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(tuiCmd)
}
