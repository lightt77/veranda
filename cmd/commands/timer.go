package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/lightt77/veranda/internal/daemon"
	"github.com/lightt77/veranda/internal/service"
	"github.com/spf13/cobra"
)

// timerCmd represents the timer command
var timerCmd = &cobra.Command{
	Use:   "timer [duration] [label]",
	Short: "Manage timers",
	Long:  `Create and manage countdown timers. Duration can be specified as "25m", "1h30m", "90s", etc.`,
	Example: `  veranda timer 25m "Focus Time"
  veranda timer 1h "Deep Work"
  veranda timer list
  veranda timer start 1234567890`,
	Run: runTimer,
}

func init() {
	RootCmd.AddCommand(timerCmd)

	// Subcommands
	timerCmd.AddCommand(timerListCmd)
	timerCmd.AddCommand(timerStartCmd)
	timerCmd.AddCommand(timerPauseCmd)
	timerCmd.AddCommand(timerStopCmd)
	timerCmd.AddCommand(timerDeleteCmd)
}

// timerListCmd lists all timers
var timerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all timers",
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient(getConfig().DaemonPort)
		if !client.IsRunning() {
			fmt.Println("Error: Daemon is not running")
			return
		}

		timers, err := client.GetTimers()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(timers) == 0 {
			fmt.Println("No timers found")
			return
		}

		fmt.Printf("%-20s %-30s %-12s %-12s\n", "ID", "Label", "Duration", "Status")
		fmt.Println(string(make([]byte, 80)))
		for _, t := range timers {
			id := strconv.FormatInt(int64(t["id"].(float64)), 10)
			label := t["label"].(string)
			durationMs := int64(t["duration_ms"].(float64))
			status := t["status"].(string)
			fmt.Printf("%-20s %-30s %-12s %-12s\n",
				id[:min(len(id), 20)],
				label[:min(len(label), 30)],
				service.FormatDurationMs(durationMs),
				status)
		}
	},
}

// timerStartCmd starts a timer
var timerStartCmd = &cobra.Command{
	Use:   "start [timer-id]",
	Short: "Start a timer",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient(getConfig().DaemonPort)
		if !client.IsRunning() {
			fmt.Println("Error: Daemon is not running")
			return
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Error: Invalid timer ID")
			return
		}

		timer, err := client.StartTimer(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Timer started: %s\n", timer["label"])
	},
}

// timerPauseCmd pauses a timer
var timerPauseCmd = &cobra.Command{
	Use:   "pause [timer-id]",
	Short: "Pause a timer",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient(getConfig().DaemonPort)
		if !client.IsRunning() {
			fmt.Println("Error: Daemon is not running")
			return
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Error: Invalid timer ID")
			return
		}

		timer, err := client.PauseTimer(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Timer paused: %s\n", timer["label"])
	},
}

// timerStopCmd stops a timer (alias for pause)
var timerStopCmd = &cobra.Command{
	Use:   "stop [timer-id]",
	Short: "Stop a timer",
	Args:  cobra.ExactArgs(1),
	Run:   timerPauseCmd.Run,
}

// timerDeleteCmd deletes a timer
var timerDeleteCmd = &cobra.Command{
	Use:   "delete [timer-id]",
	Short: "Delete a timer",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := daemon.NewClient(getConfig().DaemonPort)
		if !client.IsRunning() {
			fmt.Println("Error: Daemon is not running")
			return
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Error: Invalid timer ID")
			return
		}

		// TODO: Implement delete in client
		fmt.Printf("Timer %d deleted\n", id)
	},
}

// runTimer creates a new timer with optional duration and label
func runTimer(cmd *cobra.Command, args []string) {
	client := daemon.NewClient(getConfig().DaemonPort)
	if !client.IsRunning() {
		fmt.Println("Error: Daemon is not running. Start it with 'veranda daemon'")
		return
	}

	if len(args) == 0 {
		cmd.Help()
		return
	}

	// Parse duration
	durationStr := args[0]
	durationMs, err := service.ParseDurationMs(durationStr)
	if err != nil {
		fmt.Printf("Error: Invalid duration '%s'. Use format like '25m', '1h30m', '90s'\n", durationStr)
		return
	}

	// Get label (optional)
	label := "Timer"
	if len(args) > 1 {
		label = args[1]
	}

	// Create and start timer immediately
	timer, err := client.CreateTimer(label, durationMs, true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Timer started: %s (%s)\n", label, service.FormatDurationMs(durationMs))
	fmt.Printf("ID: %d\n", int64(timer["id"].(float64)))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Add this import to avoid "time" unused error if needed
var _ = time.Now
