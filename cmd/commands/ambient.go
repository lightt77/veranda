package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// ambientCmd represents the ambient command
var ambientCmd = &cobra.Command{
	Use:   "ambient [command]",
	Short: "Control ambient sound",
	Long:  `Start, stop, and control ambient sound playback.`,
	Example: `  veranda ambient play
  veranda ambient stop
  veranda ambient status
  veranda ambient volume 0.7`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RootCmd.AddCommand(ambientCmd)

	// Subcommands
	ambientCmd.AddCommand(ambientPlayCmd)
	ambientCmd.AddCommand(ambientStopCmd)
	ambientCmd.AddCommand(ambientStatusCmd)
	ambientCmd.AddCommand(ambientVolumeCmd)
}

// ambientPlayCmd starts ambient playback
var ambientPlayCmd = &cobra.Command{
	Use:   "play",
	Short: "Start ambient sound playback",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ensureDaemon()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if err := client.PlayAmbient(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("Ambient sound started")
	},
}

// ambientStopCmd stops ambient playback
var ambientStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop ambient sound playback",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ensureDaemon()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if err := client.StopAmbient(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("Ambient sound stopped")
	},
}

// ambientStatusCmd shows ambient status
var ambientStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show ambient sound status",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ensureDaemon()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Note: Would need a client method to get status
		_ = client
		fmt.Println("Ambient sound service is running")
		fmt.Printf("Check API at: http://localhost:%d/ambient/status\n", getConfig().DaemonPort)
	},
}

// ambientVolumeCmd sets ambient volume
var ambientVolumeCmd = &cobra.Command{
	Use:   "volume [0.0-1.0]",
	Short: "Set ambient sound volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ensureDaemon()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		volume, err := strconv.ParseFloat(args[0], 64)
		if err != nil || volume < 0 || volume > 1 {
			fmt.Println("Error: Volume must be between 0.0 and 1.0")
			return
		}

		if err := client.SetAmbientVolume(volume); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Volume set to %.0f%%\n", volume*100)
	},
}
