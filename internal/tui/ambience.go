// Package tui provides the Terminal User Interface using Bubble Tea.
package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightt77/veranda/internal/audio"
	"github.com/lightt77/veranda/internal/config"
)

// renderAmbience renders the ambience sounds configuration tab
func (m Model) renderAmbience() string {
	var content string

	// Title
	content += "Ambience\n\n"

	// Global ambience toggle
	globalStatus := "OFF"
	globalColor := red
	if m.userConfig.AmbienceEnabled {
		globalStatus = "ON"
		globalColor = green
	}
	content += "Ambience sounds are " + lipgloss.NewStyle().Foreground(globalColor).Bold(true).Render(globalStatus)
	content += lipgloss.NewStyle().Foreground(overlay0).Render(" (press 'a' to toggle)\n")

	if !m.userConfig.AmbienceEnabled {
		content += "\n" + lipgloss.NewStyle().Foreground(overlay0).Render("Ambience sounds are disabled. Press 'a' to enable.\n")
		return content
	}

	if len(m.userConfig.AmbienceSounds) == 0 {
		content += "\n" + lipgloss.NewStyle().Foreground(red).Render("No ambience sound files found.\n")
		content += lipgloss.NewStyle().Foreground(overlay0).Render("Place .mp3 files in ~/.veranda/sounds/ambience/\n")
		return content
	}

	content += "\n"

	// Display list using ambienceDisplayOrder for sorted display
	for displayIdx, configIdx := range m.ambienceDisplayOrder {
		if configIdx < 0 || configIdx >= len(m.userConfig.AmbienceSounds) {
			continue
		}
		sound := m.userConfig.AmbienceSounds[configIdx]

		// Selection indicator
		prefix := "  "
		if displayIdx == m.ambienceSelectedIdx {
			prefix = selectedStyle.Render("> ")
		}

		// Checkbox
		checkbox := "[ ]"
		if sound.Enabled {
			checkbox = "[" + lipgloss.NewStyle().Foreground(green).Render("✓") + "]"
		}

		// Filename
		filename := sound.Filename
		if displayIdx == m.ambienceSelectedIdx {
			filename = lipgloss.NewStyle().Foreground(lavender).Bold(true).Render(filename)
		} else {
			filename = lipgloss.NewStyle().Foreground(text).Render(filename)
		}

		// Volume bar
		volBar := renderVolumeBar(sound.Volume, 10)
		volPercent := int(sound.Volume * 100)
		volStr := fmt.Sprintf("%3d%%", volPercent)
		if displayIdx == m.ambienceSelectedIdx && m.editingAmbienceVolume {
			volStr = lipgloss.NewStyle().Foreground(yellow).Render(volStr)
		}

		// Test indicator - check actual player state
		testIndicator := "  "
		if m.testAudioPlayer != nil && m.testAudioPlayer.IsTestPlaying(sound.Filename) {
			testIndicator = lipgloss.NewStyle().Foreground(green).Render("▶ ")
		}

		content += fmt.Sprintf("%s%s %s %s %s %s\n",
			prefix,
			checkbox,
			filename,
			volBar,
			volStr,
			testIndicator,
		)
	}

	// Instructions
	content += "\n" + lipgloss.NewStyle().Foreground(overlay0).Render(
		"a:ambience on/off • space:toggle sound • t:test • v:volume • r:reset") + "\n"

	// Show settings message if any
	if m.settingsMessage != "" {
		content += "\n" + lipgloss.NewStyle().Foreground(green).Render(m.settingsMessage) + "\n"
	}

	return content
}

// handleAmbienceKey handles key presses in the ambience tab
// Returns (newModel, cmd, handled) where handled indicates if the key was processed
func (m Model) handleAmbienceKey(key string) (Model, tea.Cmd, bool) {
	// Handle volume editing mode first
	if m.editingAmbienceVolume {
		switch key {
		case "esc", "enter":
			m.editingAmbienceVolume = false
			return m, nil, true
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Set volume to key value * 10%
			vol := float64(key[0]-'0') / 10.0
			m.setAmbienceVolume(vol)
			return m, nil, true
		case "up", "k":
			m.adjustAmbienceVolume(0.1)
			return m, nil, true
		case "down", "j":
			m.adjustAmbienceVolume(-0.1)
			return m, nil, true
		}
		return m, nil, true // Consume all keys in volume editing mode
	}

	switch key {
	case "up", "k":
		if m.ambienceSelectedIdx > 0 {
			m.ambienceSelectedIdx--
		}
		return m, nil, true
	case "down", "j":
		maxIdx := len(m.ambienceDisplayOrder) - 1
		if m.ambienceSelectedIdx < maxIdx {
			m.ambienceSelectedIdx++
		}
		return m, nil, true
	case "a":
		// Toggle global ambience on/off
		m.toggleAmbienceGlobal()
		return m, nil, true
	case " ":
		// Toggle enable/disable (only if ambience is enabled)
		if m.userConfig.AmbienceEnabled {
			m.toggleAmbienceEnabled()
		}
		return m, nil, true
	case "t":
		// Test toggle - use async command (only if ambience is enabled)
		if m.userConfig.AmbienceEnabled && m.ambienceSelectedIdx < len(m.ambienceDisplayOrder) {
			return m, toggleAmbienceTestCmd(m.testAudioPlayer, m.userConfig.AmbienceSounds, m.ambienceDisplayOrder, m.ambienceSelectedIdx), true
		}
		return m, nil, true
	case "v":
		// Enter volume editing mode (only if ambience is enabled)
		if m.userConfig.AmbienceEnabled && m.ambienceSelectedIdx < len(m.ambienceDisplayOrder) {
			m.editingAmbienceVolume = true
		}
		return m, nil, true
	case "r":
		// Reset to defaults
		m.resetAmbience()
		return m, nil, true
	}

	return m, nil, false
}

// handleTestMessages processes test started/stopped messages
func (m Model) handleTestMessages(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case testStartedMsg:
		m.setSettingsMessage(fmt.Sprintf("Playing: %s", msg.filename))
	case testStoppedMsg:
		m.setSettingsMessage(fmt.Sprintf("Stopped: %s", msg.filename))
	case errMsg:
		m.setSettingsMessage(fmt.Sprintf("Error: %v", msg.err))
	}
	return m, nil
}

// setAmbienceVolume sets the volume for the currently selected ambience sound
func (m *Model) setAmbienceVolume(vol float64) {
	if vol < 0 {
		vol = 0
	}
	if vol > 1 {
		vol = 1
	}
	if m.ambienceSelectedIdx < len(m.ambienceDisplayOrder) {
		configIdx := m.ambienceDisplayOrder[m.ambienceSelectedIdx]
		if configIdx >= 0 && configIdx < len(m.userConfig.AmbienceSounds) {
			m.userConfig.AmbienceSounds[configIdx].Volume = vol
			// Update playing stream volume in real-time
			if m.testAudioPlayer != nil {
				m.testAudioPlayer.SetTestVolume(m.userConfig.AmbienceSounds[configIdx].Filename, vol)
			}
			// Auto-save the volume change
			m.saveAmbienceSettings()
		}
	}
}

// adjustAmbienceVolume adjusts the volume by the given delta
func (m *Model) adjustAmbienceVolume(delta float64) {
	if m.ambienceSelectedIdx < len(m.ambienceDisplayOrder) {
		configIdx := m.ambienceDisplayOrder[m.ambienceSelectedIdx]
		if configIdx >= 0 && configIdx < len(m.userConfig.AmbienceSounds) {
			newVol := m.userConfig.AmbienceSounds[configIdx].Volume + delta
			if newVol < 0 {
				newVol = 0
			}
			if newVol > 1 {
				newVol = 1
			}
			m.userConfig.AmbienceSounds[configIdx].Volume = newVol
			// Update playing stream volume in real-time
			if m.testAudioPlayer != nil {
				m.testAudioPlayer.SetTestVolume(m.userConfig.AmbienceSounds[configIdx].Filename, newVol)
			}
			// Auto-save the volume change
			m.saveAmbienceSettings()
		}
	}
}

// toggleAmbienceGlobal toggles the global ambience sounds on/off
func (m *Model) toggleAmbienceGlobal() {
	m.userConfig.AmbienceEnabled = !m.userConfig.AmbienceEnabled
	// Auto-save the change
	m.saveAmbienceSettings()
	// Stop test playback if disabling
	if !m.userConfig.AmbienceEnabled {
		m.stopAllTestPlaybackSync()
	}
}

// toggleAmbienceEnabled toggles the enabled state of the currently selected ambience sound
func (m *Model) toggleAmbienceEnabled() {
	if m.ambienceSelectedIdx < len(m.ambienceDisplayOrder) {
		configIdx := m.ambienceDisplayOrder[m.ambienceSelectedIdx]
		if configIdx >= 0 && configIdx < len(m.userConfig.AmbienceSounds) {
			m.userConfig.AmbienceSounds[configIdx].Enabled = !m.userConfig.AmbienceSounds[configIdx].Enabled
			// Auto-save the change immediately
			m.saveAmbienceSettings()
		}
	}
}

// saveAmbienceSettings saves the ambience sound configuration
func (m *Model) saveAmbienceSettings() {
	if err := m.userConfig.Save(m.settingsRepo); err != nil {
		m.setSettingsMessage(fmt.Sprintf("Error saving: %v", err))
	} else {
		m.setSettingsMessage("Ambience settings saved!")
	}
}

// resetAmbience resets ambience sounds to default (first enabled)
func (m *Model) resetAmbience() {
	// Reset to default: only first one enabled at 100%
	for i := range m.userConfig.AmbienceSounds {
		m.userConfig.AmbienceSounds[i].Enabled = i == 0
		m.userConfig.AmbienceSounds[i].Volume = config.DefaultAmbienceVolume
	}
	m.saveAmbienceSettings()
	m.setSettingsMessage("Reset to defaults")
}

// setSettingsMessage sets a message to display in the settings area
func (m *Model) setSettingsMessage(msg string) {
	m.settingsMessage = msg
	m.settingsMessageTime = time.Now()
}

// stopAllTestPlayback stops all ambient test playback
func (m *Model) stopAllTestPlayback() tea.Cmd {
	return stopAllTestPlaybackCmd(m.testAudioPlayer)
}

// stopAllTestPlaybackSync stops all ambient test playback synchronously with fade out
func (m *Model) stopAllTestPlaybackSync() {
	if m.testAudioPlayer != nil {
		m.testAudioPlayer.StopAllTestPlaysWithFade()
	}
}

// initTestAudioPlayer initializes the test audio player if needed
func (m *Model) initTestAudioPlayer() error {
	if m.testAudioPlayer == nil {
		cfg := config.Config()
		player, err := audio.NewPlayer(cfg, m.configDir)
		if err != nil {
			return err
		}
		m.testAudioPlayer = player
	}
	return nil
}
