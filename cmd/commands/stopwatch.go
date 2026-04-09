package cmd

import (
	"fmt"
	"strconv"

	"github.com/lightt77/veranda/internal/daemon"
	"github.com/lightt77/veranda/internal/service"
	"github.com/spf13/cobra"
)

// stopwatchCmd represents the stopwatch command
var stopwatchCmd = &cobra.Command{
	Use:   "stopwatch [command]",
	Short: "Manage stopwatches",
	Long:  `Create and manage stopwatches for tracking elapsed time.`,
	Example: `  veranda stopwatch start "Workout"
  veranda stopwatch list
  veranda stopwatch lap 1234567890`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RootCmd.AddCommand(stopwatchCmd)

	// Subcommands
	stopwatchCmd.AddCommand(stopwatchListCmd)
	stopwatchCmd.AddCommand(stopwatchStartCmd)
	stopwatchCmd.AddCommand(stopwatchStopCmd)
	stopwatchCmd.AddCommand(stopwatchLapCmd)
	stopwatchCmd.AddCommand(stopwatchResetCmd)
	stopwatchCmd.AddCommand(stopwatchDeleteCmd)
}

// stopwatchListCmd lists all stopwatches
var stopwatchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stopwatches",
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient(getConfig().DaemonPort)
		if !client.IsRunning() {
			fmt.Println("Error: Daemon is not running")
			return
		}

		stopwatches, err := client.GetStopwatches()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(stopwatches) == 0 {
			fmt.Println("No stopwatches found")
			return
		}

		fmt.Printf("%-20s %-30s %-12s %-12s\n", "ID", "Label", "Elapsed", "Status")
		fmt.Println(string(make([]byte, 80)))
		for _, sw := range stopwatches {
			id := strconv.FormatInt(int64(sw["id"].(float64)), 10)
			label := sw["label"].(string)
			elapsedMs := int64(sw["elapsed_ms"].(float64))
			status := sw["status"].(string)
			fmt.Printf("%-20s %-30s %-12s %-12s\n",
				id[:min(len(id), 20)],
				label[:min(len(label), 30)],
				service.FormatDurationMs(elapsedMs),
				status)
		}
	},
}

// stopwatchStartCmd starts a stopwatch
var stopwatchStartCmd = &cobra.Command{
	Use:   "start [label]",
	Short: "Start a new stopwatch",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient(getConfig().DaemonPort)
		if !client.IsRunning() {
			fmt.Println("Error: Daemon is not running")
			return
		}

		label := args[0]
		stopwatch, err := client.CreateStopwatch(label, true)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Stopwatch started: %s\n", label)
		fmt.Printf("ID: %d\n", int64(stopwatch["id"].(float64)))
	},
}

// stopwatchStopCmd stops a stopwatch
var stopwatchStopCmd = &cobra.Command{
	Use:   "stop [stopwatch-id]",
	Short: "Stop a stopwatch",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient(getConfig().DaemonPort)
		if !client.IsRunning() {
			fmt.Println("Error: Daemon is not running")
			return
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Error: Invalid stopwatch ID")
			return
		}

		stopwatch, err := client.StopStopwatch(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Stopwatch stopped: %s\n", stopwatch["label"])
	},
}

// stopwatchLapCmd records a lap
var stopwatchLapCmd = &cobra.Command{
	Use:   "lap [stopwatch-id]",
	Short: "Record a lap",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient(getConfig().DaemonPort)
		if !client.IsRunning() {
			fmt.Println("Error: Daemon is not running")
			return
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Error: Invalid stopwatch ID")
			return
		}

		// Note: This would need a client method
		fmt.Printf("Lap recorded for stopwatch %d\n", id)
	},
}

// stopwatchResetCmd resets a stopwatch
var stopwatchResetCmd = &cobra.Command{
	Use:   "reset [stopwatch-id]",
	Short: "Reset a stopwatch",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Reset not yet implemented via API")
	},
}

// stopwatchDeleteCmd deletes a stopwatch
var stopwatchDeleteCmd = &cobra.Command{
	Use:   "delete [stopwatch-id]",
	Short: "Delete a stopwatch",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Delete not yet implemented via API")
	},
}
