package cmd

import (
	"fmt"

	"github.com/lightt77/veranda/internal/daemon"
	"github.com/spf13/cobra"
)

// journalCmd represents the journal command
var journalCmd = &cobra.Command{
	Use:   "journal [command]",
	Short: "Manage journal entries",
	Long:  `Create and view journal entries for logging your thoughts and progress.`,
	Example: `  veranda journal write
  veranda journal list
  veranda journal recent 5`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RootCmd.AddCommand(journalCmd)

	// Subcommands
	journalCmd.AddCommand(journalWriteCmd)
	journalCmd.AddCommand(journalListCmd)
	journalCmd.AddCommand(journalRecentCmd)
}

// journalWriteCmd opens editor to write a journal entry
var journalWriteCmd = &cobra.Command{
	Use:   "write [text]",
	Short: "Write a journal entry",
	Long:  `Write a journal entry. If no text is provided, opens your $EDITOR.`,
	Example: `  veranda journal write "Completed the project today!"
  veranda journal write`,
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient(getConfig().DaemonPort)
		if !client.IsRunning() {
			fmt.Println("Error: Daemon is not running")
			return
		}

		var log string
		if len(args) > 0 {
			log = args[0]
		} else {
			// TODO: Open editor
			fmt.Println("Editor mode not yet implemented. Provide text as argument.")
			return
		}

		entry, err := client.CreateJournalEntry(log)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Journal entry created: #%d\n", int64(entry["id"].(float64)))
	},
}

// journalListCmd lists all journal entries
var journalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all journal entries",
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient(getConfig().DaemonPort)
		if !client.IsRunning() {
			fmt.Println("Error: Daemon is not running")
			return
		}

		// TODO: Implement client method
		fmt.Println("List not yet implemented via API")
	},
}

// journalRecentCmd shows recent entries
var journalRecentCmd = &cobra.Command{
	Use:   "recent [count]",
	Short: "Show recent journal entries",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient(getConfig().DaemonPort)
		if !client.IsRunning() {
			fmt.Println("Error: Daemon is not running")
			return
		}

		// TODO: Implement client method
		fmt.Println("Recent entries not yet implemented via API")
	},
}
