// Package tui provides the Terminal User Interface using Bubble Tea.
package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lightt77/veranda/internal/audio"
	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/daemon"
)

// tickCmd creates a command that ticks every 100ms
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// refreshCmd creates a command to refresh timer and stopwatch data
func refreshCmd(client *daemon.Client) tea.Cmd {
	return func() tea.Msg {
		timers, err := client.GetTimers()
		if err != nil {
			return errMsg{err}
		}

		stopwatches, err := client.GetStopwatches()
		if err != nil {
			return errMsg{err}
		}

		return dataMsg{timers, stopwatches}
	}
}

// createTimerCmd creates a command to create a new timer
func createTimerCmd(client *daemon.Client) tea.Cmd {
	return func() tea.Msg {
		// Create a 25-minute timer with default label
		_, err := client.CreateTimer("Quick Timer", 25*60*1000, true)
		if err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	}
}

// createStopwatchCmd creates a command to create a new stopwatch
func createStopwatchCmd(client *daemon.Client) tea.Cmd {
	return func() tea.Msg {
		// Create a stopwatch with default label
		_, err := client.CreateStopwatch("Quick Stopwatch", true)
		if err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	}
}

// createTimerWithLabelCmd creates a command to create a new timer with custom label and duration
func createTimerWithLabelCmd(client *daemon.Client, label string, durationMs int64) tea.Cmd {
	return func() tea.Msg {
		_, err := client.CreateTimer(label, durationMs, true)
		if err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	}
}

// createStopwatchWithLabelCmd creates a command to create a new stopwatch with custom label
func createStopwatchWithLabelCmd(client *daemon.Client, label string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.CreateStopwatch(label, true)
		if err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	}
}

// getVisibleTimerIndex converts selected index to actual timer index accounting for filtered completed
func getVisibleTimerIndex(timers []map[string]interface{}, visibleIdx int, showCompleted bool) int {
	if showCompleted {
		return visibleIdx
	}
	current := -1
	for i, t := range timers {
		if t["status"].(string) != "completed" {
			current++
			if current == visibleIdx {
				return i
			}
		}
	}
	return -1
}

// toggleTimerCmd creates a command to toggle a timer's pause state
func toggleTimerCmd(client *daemon.Client, timers []map[string]interface{}, selectedIdx int, showCompleted bool) tea.Cmd {
	return func() tea.Msg {
		actualIdx := getVisibleTimerIndex(timers, selectedIdx, showCompleted)
		if actualIdx < 0 || actualIdx >= len(timers) {
			return errMsg{fmt.Errorf("invalid selection")}
		}

		id := int64(timers[actualIdx]["id"].(float64))
		status := timers[actualIdx]["status"].(string)

		var err error
		if status == "running" {
			_, err = client.PauseTimer(id)
		} else {
			_, err = client.StartTimer(id)
		}

		if err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	}
}

// toggleStopwatchCmd creates a command to toggle a stopwatch's run state
func toggleStopwatchCmd(client *daemon.Client, stopwatches []map[string]interface{}, selectedIdx int) tea.Cmd {
	return func() tea.Msg {
		if selectedIdx < 0 || selectedIdx >= len(stopwatches) {
			return errMsg{fmt.Errorf("invalid selection")}
		}

		id := int64(stopwatches[selectedIdx]["id"].(float64))
		status := stopwatches[selectedIdx]["status"].(string)

		var err error
		if status == "running" {
			_, err = client.StopStopwatch(id)
		} else {
			_, err = client.StartStopwatch(id)
		}

		if err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	}
}

// deleteTimerCmd creates a command to delete a timer
func deleteTimerCmd(client *daemon.Client, timers []map[string]interface{}, selectedIdx int, showCompleted bool) tea.Cmd {
	return func() tea.Msg {
		actualIdx := getVisibleTimerIndex(timers, selectedIdx, showCompleted)
		if actualIdx < 0 || actualIdx >= len(timers) {
			return errMsg{fmt.Errorf("invalid selection")}
		}

		id := int64(timers[actualIdx]["id"].(float64))
		if err := client.DeleteTimer(id); err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	}
}

// deleteStopwatchCmd creates a command to delete a stopwatch
func deleteStopwatchCmd(client *daemon.Client, stopwatches []map[string]interface{}, selectedIdx int) tea.Cmd {
	return func() tea.Msg {
		if selectedIdx < 0 || selectedIdx >= len(stopwatches) {
			return errMsg{fmt.Errorf("invalid selection")}
		}

		id := int64(stopwatches[selectedIdx]["id"].(float64))
		if err := client.DeleteStopwatch(id); err != nil {
			return errMsg{err}
		}
		return refreshMsg{}
	}
}

// toggleAmbienceTestCmd creates a command to toggle test playback for the selected ambience sound
// This runs asynchronously to avoid blocking the UI during fade operations
func toggleAmbienceTestCmd(player *audio.Player, sounds []config.AmbienceSoundConfig, displayOrder []int, selectedIdx int) tea.Cmd {
	return func() tea.Msg {
		if player == nil || selectedIdx < 0 || selectedIdx >= len(displayOrder) {
			return nil
		}

		configIdx := displayOrder[selectedIdx]
		if configIdx < 0 || configIdx >= len(sounds) {
			return nil
		}

		sound := sounds[configIdx]
		isPlaying := player.IsTestPlaying(sound.Filename)

		if isPlaying {
			// Stop playback - this will fade out asynchronously
			player.StopTestPlay(sound.Filename)
			return testStoppedMsg{filename: sound.Filename}
		}

		// Start playback
		_, err := player.TestPlay(sound.Filename, sound.Volume)
		if err != nil {
			return errMsg{fmt.Errorf("play error: %w", err)}
		}
		return testStartedMsg{filename: sound.Filename}
	}
}

// stopAllTestPlaybackCmd creates a command to stop all test playback
func stopAllTestPlaybackCmd(player *audio.Player) tea.Cmd {
	return func() tea.Msg {
		if player != nil {
			player.StopAllTestPlays()
		}
		return nil
	}
}
