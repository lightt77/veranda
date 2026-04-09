package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the Terminal User Interface",
	Long:  `Launch an interactive TUI for managing timers, stopwatches, and viewing activity.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TUI not yet implemented")
		fmt.Println("Coming soon: Interactive terminal interface with Bubble Tea")
	},
}

func init() {
	RootCmd.AddCommand(tuiCmd)
}
