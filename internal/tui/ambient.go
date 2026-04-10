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

// renderAmbientSounds renders the ambient sounds configuration tab
func (m Model) renderAmbientSounds() string {
	var content string

	// Title
	content += "Ambient Sounds\n\n"

	if len(m.userConfig.AmbientSounds) == 0 {
		content += lipgloss.NewStyle().Foreground(red).Render("No ambient sound files found.\n")
		content += lipgloss.NewStyle().Foreground(overlay0).Render("Place .mp3 files in ~/.veranda/sounds/ambience/\n")
		return content
	}

	// Display list using ambientDisplayOrder for sorted display
	for displayIdx, configIdx := range m.ambientDisplayOrder {
		if configIdx < 0 || configIdx >= len(m.userConfig.AmbientSounds) {
			continue
		}
		sound := m.userConfig.AmbientSounds[configIdx]

		// Selection indicator
		prefix := "  "
		if displayIdx == m.ambientSelectedIdx {
			prefix = selectedStyle.Render("> ")
		}

		// Checkbox
		checkbox := "[ ]"
		if sound.Enabled {
			checkbox = "[" + lipgloss.NewStyle().Foreground(green).Render("✓") + "]"
		}

		// Filename
		filename := sound.Filename
		if displayIdx == m.ambientSelectedIdx {
			filename = lipgloss.NewStyle().Foreground(lavender).Bold(true).Render(filename)
		} else {
			filename = lipgloss.NewStyle().Foreground(text).Render(filename)
		}

		// Volume bar
		volBar := renderVolumeBar(sound.Volume, 10)
		volPercent := int(sound.Volume * 100)
		volStr := fmt.Sprintf("%3d%%", volPercent)
		if displayIdx == m.ambientSelectedIdx && m.editingAmbientVolume {
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
		"space:toggle • t:test • v:volume • s:save • r:reset") + "\n"

	// Show settings message if any
	if m.settingsMessage != "" {
		content += "\n" + lipgloss.NewStyle().Foreground(green).Render(m.settingsMessage) + "\n"
	}

	return content
}

// handleAmbientKey handles key presses in the ambient sounds tab
// Returns (newModel, cmd, handled) where handled indicates if the key was processed
func (m Model) handleAmbientKey(key string) (Model, tea.Cmd, bool) {
	// Handle volume editing mode first
	if m.editingAmbientVolume {
		switch key {
		case "esc", "enter":
			m.editingAmbientVolume = false
			return m, nil, true
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Set volume to key value * 10%
			vol := float64(key[0]-'0') / 10.0
			m.setAmbientVolume(vol)
			return m, nil, true
		case "up", "k":
			m.adjustAmbientVolume(0.1)
			return m, nil, true
		case "down", "j":
			m.adjustAmbientVolume(-0.1)
			return m, nil, true
		}
		return m, nil, true // Consume all keys in volume editing mode
	}

	switch key {
	case "up", "k":
		if m.ambientSelectedIdx > 0 {
			m.ambientSelectedIdx--
		}
		return m, nil, true
	case "down", "j":
		maxIdx := len(m.ambientDisplayOrder) - 1
		if m.ambientSelectedIdx < maxIdx {
			m.ambientSelectedIdx++
		}
		return m, nil, true
	case " ":
		// Toggle enable/disable
		m.toggleAmbientEnabled()
		return m, nil, true
	case "t":
		// Test toggle - use async command
		if m.ambientSelectedIdx < len(m.ambientDisplayOrder) {
			return m, toggleAmbientTestCmd(m.testAudioPlayer, m.userConfig.AmbientSounds, m.ambientDisplayOrder, m.ambientSelectedIdx), true
		}
		return m, nil, true
	case "v":
		// Enter volume editing mode
		if m.ambientSelectedIdx < len(m.ambientDisplayOrder) {
			m.editingAmbientVolume = true
		}
		return m, nil, true
	case "s":
		// Save settings
		m.saveAmbientSettings()
		return m, nil, true
	case "r":
		// Reset to defaults
		m.resetAmbientSounds()
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
	}
	return m, nil
}

// setAmbientVolume sets the volume for the currently selected ambient sound
func (m *Model) setAmbientVolume(vol float64) {
	if vol < 0 {
		vol = 0
	}
	if vol > 1 {
		vol = 1
	}
	if m.ambientSelectedIdx < len(m.ambientDisplayOrder) {
		configIdx := m.ambientDisplayOrder[m.ambientSelectedIdx]
		if configIdx >= 0 && configIdx < len(m.userConfig.AmbientSounds) {
			m.userConfig.AmbientSounds[configIdx].Volume = vol
		}
	}
}

// adjustAmbientVolume adjusts the volume by the given delta
func (m *Model) adjustAmbientVolume(delta float64) {
	if m.ambientSelectedIdx < len(m.ambientDisplayOrder) {
		configIdx := m.ambientDisplayOrder[m.ambientSelectedIdx]
		if configIdx >= 0 && configIdx < len(m.userConfig.AmbientSounds) {
			newVol := m.userConfig.AmbientSounds[configIdx].Volume + delta
			if newVol < 0 {
				newVol = 0
			}
			if newVol > 1 {
				newVol = 1
			}
			m.userConfig.AmbientSounds[configIdx].Volume = newVol
		}
	}
}

// toggleAmbientEnabled toggles the enabled state of the currently selected ambient sound
func (m *Model) toggleAmbientEnabled() {
	if m.ambientSelectedIdx < len(m.ambientDisplayOrder) {
		configIdx := m.ambientDisplayOrder[m.ambientSelectedIdx]
		if configIdx >= 0 && configIdx < len(m.userConfig.AmbientSounds) {
			m.userConfig.AmbientSounds[configIdx].Enabled = !m.userConfig.AmbientSounds[configIdx].Enabled
		}
	}
}

// saveAmbientSettings saves the ambient sound configuration
func (m *Model) saveAmbientSettings() {
	if err := m.userConfig.Save(m.configDir); err != nil {
		m.setSettingsMessage(fmt.Sprintf("Error saving: %v", err))
	} else {
		m.setSettingsMessage("Ambient settings saved!")
	}
}

// resetAmbientSounds resets ambient sounds to default (first enabled)
func (m *Model) resetAmbientSounds() {
	// Reset to default: only first one enabled at 50%
	for i := range m.userConfig.AmbientSounds {
		m.userConfig.AmbientSounds[i].Enabled = i == 0
		m.userConfig.AmbientSounds[i].Volume = config.DefaultAmbientVolume
	}
	m.saveAmbientSettings()
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
