package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// ambienceCmd represents the ambience command
var ambienceCmd = &cobra.Command{
	Use:   "ambience [command]",
	Short: "Control ambience sound",
	Long:  `Start, stop, and control ambience sound playback.`,
	Example: `  veranda ambience play
  veranda ambience stop
  veranda ambience status
  veranda ambience volume 0.7`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RootCmd.AddCommand(ambienceCmd)

	// Subcommands
	ambienceCmd.AddCommand(ambiencePlayCmd)
	ambienceCmd.AddCommand(ambienceStopCmd)
	ambienceCmd.AddCommand(ambienceStatusCmd)
	ambienceCmd.AddCommand(ambienceVolumeCmd)
}

// ambiencePlayCmd starts ambience playback
var ambiencePlayCmd = &cobra.Command{
	Use:   "play",
	Short: "Start ambience sound playback",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ensureDaemon()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if err := client.PlayAmbience(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("Ambience sound started")
	},
}

// ambienceStopCmd stops ambience playback
var ambienceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop ambience sound playback",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ensureDaemon()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if err := client.StopAmbience(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("Ambience sound stopped")
	},
}

// ambienceStatusCmd shows ambience status
var ambienceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show ambience sound status",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := ensureDaemon()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Note: Would need a client method to get status
		_ = client
		fmt.Println("Ambience sound service is running")
		fmt.Printf("Check API at: http://localhost:%d/ambience/status\n", getConfig().DaemonPort)
	},
}

// ambienceVolumeCmd sets ambience volume
var ambienceVolumeCmd = &cobra.Command{
	Use:   "volume [0.0-1.0]",
	Short: "Set ambience sound volume",
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

		if err := client.SetAmbienceVolume(volume); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Volume set to %.0f%%\n", volume*100)
	},
}
