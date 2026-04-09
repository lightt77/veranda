package cmd

import (
	"fmt"

	"github.com/lightt77/veranda/internal/tui"
	"github.com/spf13/cobra"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the Terminal User Interface",
	Long:  `Launch an interactive TUI for managing timers, stopwatches, and viewing activity.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := tui.Run(getConfig().DaemonPort); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(tuiCmd)
}
