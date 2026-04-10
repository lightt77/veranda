// Package tui provides the Terminal User Interface using Bubble Tea.
package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightt77/veranda/internal/config"
	"github.com/lightt77/veranda/internal/repository"
	"github.com/lightt77/veranda/internal/tui/starfield"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle test messages first (from async audio commands)
	m, cmd := m.handleTestMessages(msg)
	if cmd != nil {
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate star area height
		contentHeight := 6
		paddingBottom := 0
		starAreaHeight := m.height - contentHeight - paddingBottom
		if starAreaHeight < 8 {
			starAreaHeight = 8
		}

		// Resize both starfields
		m.randomStarfield.Resize(m.width, m.height, starAreaHeight)
		m.realisticStarfield.Resize(m.width, m.height, starAreaHeight)

		// Trigger initial realistic starfield update
		m.realisticStarfield.Update(time.Now())

	case twinkleMsg:
		// Twinkle in both modes
		if m.starfieldMode == ModeRandom {
			m.randomStarfield.Twinkle()
		} else {
			m.realisticStarfield.Twinkle()
		}
		return m, twinkleCmd()

	case starfieldUpdateMsg:
		// Update realistic starfield positions
		if m.starfieldMode == ModeRealistic {
			m.realisticStarfield.Update(time.Now())
		}
		return m, starfieldUpdateCmd()

	case tea.KeyMsg:
		// Handle city picker mode first
		if m.showCityPicker {
			return m.handleCityPickerKey(msg.String())
		}

		// Handle ambience volume editing mode
		if m.activeTab == 1 && m.editingAmbienceVolume {
			newM, cmd, handled := m.handleAmbienceKey(msg.String())
			if handled {
				return newM, cmd
			}
		}

		// Handle chime volume editing mode
		if m.activeTab == 3 && m.editingChimeVolume {
			switch msg.String() {
			case "esc", "enter":
				m.editingChimeVolume = false
			case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
				// Set volume to key value * 10%
				m.userConfig.ChimeVolume = float64(msg.String()[0]-'0') / 10.0
			case "up", "k":
				newVol := m.userConfig.ChimeVolume + 0.1
				if newVol > 1.0 {
					newVol = 1.0
				}
				m.userConfig.ChimeVolume = newVol
			case "down", "j":
				newVol := m.userConfig.ChimeVolume - 0.1
				if newVol < 0.0 {
					newVol = 0.0
				}
				m.userConfig.ChimeVolume = newVol
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			// Stop any test playback before quitting
			return m, tea.Batch(m.stopAllTestPlayback(), tea.Quit)
		case "tab":
			// Stop test playback when leaving ambience tab
			if m.activeTab == 1 {
				m.stopAllTestPlaybackSync()
			}
			m.activeTab = (m.activeTab + 1) % 4
			m.selectedIdx = 0
			m.selectingSoundFile = false
			m.editingAmbienceVolume = false
			m.editingChimeVolume = false
			m.editingChimeVolume = false
		case "up", "k":
			if m.activeTab == 1 {
				newM, _, _ := m.handleAmbienceKey(msg.String())
				return newM, nil
			} else {
				m.selectedIdx = max(0, m.selectedIdx-1)
			}
		case "down", "j":
			if m.activeTab == 1 {
				newM, _, _ := m.handleAmbienceKey(msg.String())
				return newM, nil
			} else {
				maxIdx := len(m.timers) - 1
				if m.activeTab == 0 {
					maxIdx = len(m.stopwatches) - 1
				}
				m.selectedIdx = min(maxIdx, m.selectedIdx+1)
			}
		case " ", "p":
			// Pause/resume selected timer or stopwatch
			if m.activeTab == 0 && m.selectedIdx < len(m.timers) {
				return m, toggleTimerCmd(m.client, m.timers, m.selectedIdx, m.showCompleted)
			} else if m.activeTab == 0 && m.selectedIdx < len(m.stopwatches) {
				return m, toggleStopwatchCmd(m.client, m.stopwatches, m.selectedIdx)
			} else if m.activeTab == 1 {
				// Toggle ambience enabled
				newM, _, _ := m.handleAmbienceKey(" ")
				return newM, nil
			}
		case "d":
			// Delete selected timer or stopwatch
			if m.activeTab == 0 && m.selectedIdx < len(m.timers) {
				return m, deleteTimerCmd(m.client, m.timers, m.selectedIdx, m.showCompleted)
			} else if m.activeTab == 0 && m.selectedIdx < len(m.stopwatches) {
				return m, deleteStopwatchCmd(m.client, m.stopwatches, m.selectedIdx)
			}
		case "r":
			if m.activeTab == 1 {
				// Reset ambience sounds to defaults
				newM, _, _ := m.handleAmbienceKey("r")
				return newM, nil
			} else {
				return m, refreshCmd(m.client)
			}
		case "h":
			// Toggle showing completed timers
			if m.activeTab == 0 {
				m.showCompleted = !m.showCompleted
				m.selectedIdx = 0
			}
		case "t":
			if m.activeTab == 0 {
				// Quick timer creation
				return m, createTimerCmd(m.client)
			} else if m.activeTab == 1 {
				// Test toggle for ambience sound - use async command
				if err := m.initTestAudioPlayer(); err != nil {
					m.setSettingsMessage(fmt.Sprintf("Audio error: %v", err))
					return m, nil
				}
				newM, cmd, _ := m.handleAmbienceKey("t")
				return newM, cmd
			}
		case "a":
			if m.activeTab == 1 {
				// Toggle global ambience on/off
				newM, _, _ := m.handleAmbienceKey("a")
				return newM, nil
			}
		case "m":
			// Toggle starfield mode
			if m.starfieldMode == ModeRandom {
				m.starfieldMode = ModeRealistic
				m.realisticStarfield.Update(time.Now())
			} else {
				m.starfieldMode = ModeRandom
			}
		case "l":
			// Open city picker
			if m.citiesRepo != nil {
				m.showCityPicker = true
				m.cityPickerSearch = ""
				m.loadAllCities()
			}
		case "n":
			// Toggle object names display
			m.showObjectNames = !m.showObjectNames
		case "e":
			// Edit sound file selection
			if m.activeTab == 3 {
				m.selectingSoundFile = true
				m.selectedSoundIdx = 0
				// Find index of current sound
				for i, f := range m.availableSoundFiles {
					if f == m.userConfig.TimerCompletionSound {
						m.selectedSoundIdx = i
						break
					}
				}
			}
		case "v":
			// Volume editing
			if m.activeTab == 1 {
				newM, _, _ := m.handleAmbienceKey("v")
				return newM, nil
			} else if m.activeTab == 3 {
				m.editingChimeVolume = true
			}
		case "enter":
			// Confirm sound file selection
			if m.activeTab == 3 && m.selectingSoundFile {
				if m.selectedSoundIdx < len(m.availableSoundFiles) {
					m.userConfig.TimerCompletionSound = m.availableSoundFiles[m.selectedSoundIdx]
					if err := m.userConfig.Save(m.settingsRepo); err != nil {
						m.setSettingsMessage(fmt.Sprintf("Error saving: %v", err))
					} else {
						m.setSettingsMessage("Sound updated!")
					}
				}
				m.selectingSoundFile = false
			}
		}

	case tickMsg:
		m.lastUpdate = time.Now()
		// Refresh timer/stopwatch data from daemon on every tick
		return m, tea.Batch(tickCmd(), refreshCmd(m.client))

	case refreshMsg:
		return m, refreshCmd(m.client)

	case dataMsg:
		m.timers = msg.timers
		m.stopwatches = msg.stopwatches
		m.err = nil

	case errMsg:
		m.err = msg.err
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	var s strings.Builder

	// Calculate star area height
	contentHeight := 6
	paddingBottom := 0
	starAreaHeight := m.height - contentHeight - paddingBottom
	if starAreaHeight < 8 {
		starAreaHeight = 8
	}

	// Render starfield or city picker
	if m.showCityPicker {
		s.WriteString(m.renderCityPicker(starAreaHeight))
	} else {
		s.WriteString(m.renderStarfield(starAreaHeight))
	}

	// Render content at bottom
	s.WriteString(m.renderContent())

	return s.String()
}

// renderStarfield renders the appropriate starfield
func (m Model) renderStarfield(starAreaHeight int) string {
	if m.starfieldMode == ModeRandom {
		return m.renderRandomStarfield(starAreaHeight)
	}
	return m.renderRealisticStarfield(starAreaHeight)
}

// renderRandomStarfield renders the twinkling random starfield
func (m Model) renderRandomStarfield(starAreaHeight int) string {
	stars := m.randomStarfield.GetStars()

	// Create a 2D grid
	type cell struct {
		char  string
		color lipgloss.Color
		isSet bool
	}

	grid := make([][]cell, starAreaHeight)
	for y := range grid {
		grid[y] = make([]cell, m.width)
	}

	// Place stars
	for _, star := range stars {
		if star.Y < starAreaHeight && star.X < m.width {
			grid[star.Y][star.X] = cell{
				char:  star.Char,
				color: star.Color,
				isSet: true,
			}
		}
	}

	// Build output rows
	rows := make([]string, starAreaHeight)
	for y := 0; y < starAreaHeight; y++ {
		var row strings.Builder
		for x := 0; x < m.width; x++ {
			c := grid[y][x]
			if c.isSet {
				style := lipgloss.NewStyle().Foreground(c.color)
				row.WriteString(style.Render(c.char))
			} else {
				row.WriteString(" ")
			}
		}
		rows[y] = row.String()
	}

	return strings.Join(rows, "\n")
}

// renderRealisticStarfield renders the realistic starfield
func (m Model) renderRealisticStarfield(starAreaHeight int) string {
	// Use a grid to track characters and their styles separately
	type cell struct {
		char  string
		color lipgloss.Color
		isSet bool
	}

	grid := make([][]cell, starAreaHeight)
	for y := range grid {
		grid[y] = make([]cell, m.width)
	}

	// Get visible objects
	objects := m.realisticStarfield.GetVisibleObjects()

	// Track which positions have objects to avoid overlapping names
	occupiedPositions := make(map[[2]int]bool)

	// Render each object
	for _, obj := range objects {
		if obj.ScreenY >= 0 && obj.ScreenY < starAreaHeight &&
			obj.ScreenX >= 0 && obj.ScreenX < m.width {

			symbol := starfield.GetObjectSymbol(obj.Object, obj.Brightness, obj.TwinklePhase)
			twinkleBrightness := starfield.GetTwinkleBrightness(obj.Brightness, obj.TwinklePhase, obj.Object.Type)
			color := starfield.GetObjectColor(obj.Object, twinkleBrightness)

			if len(symbol) > 0 {
				runes := []rune(symbol)
				if len(runes) > 0 {
					grid[obj.ScreenY][obj.ScreenX] = cell{
						char:  string(runes[0]),
						color: color,
						isSet: true,
					}
					occupiedPositions[[2]int{obj.ScreenX, obj.ScreenY}] = true
				}
			}
		}
	}

	// Render object names if enabled
	if m.showObjectNames {
		for _, obj := range objects {
			if obj.ScreenY >= 0 && obj.ScreenY < starAreaHeight &&
				obj.ScreenX >= 0 && obj.ScreenX < m.width {

				name := obj.Object.Name
				nameRunes := []rune(name)
				nameLen := len(nameRunes)

				positions := [][2]int{
					{obj.ScreenX + 2, obj.ScreenY},
					{obj.ScreenX - nameLen - 1, obj.ScreenY},
					{obj.ScreenX - nameLen/2, obj.ScreenY + 1},
					{obj.ScreenX - nameLen/2, obj.ScreenY - 1},
					{obj.ScreenX + 2, obj.ScreenY + 1},
					{obj.ScreenX - nameLen - 1, obj.ScreenY + 1},
					{obj.ScreenX + 2, obj.ScreenY - 1},
					{obj.ScreenX - nameLen - 1, obj.ScreenY - 1},
				}

				for _, pos := range positions {
					nameX := pos[0]
					nameY := pos[1]

					if nameX < 0 || nameX+nameLen > m.width || nameY < 0 || nameY >= starAreaHeight {
						continue
					}

					canPlace := true
					for i := 0; i < nameLen && nameX+i < m.width; i++ {
						if occupiedPositions[[2]int{nameX + i, nameY}] || grid[nameY][nameX+i].isSet {
							canPlace = false
							break
						}
					}

					if canPlace {
						colOffset := 0
						for _, ch := range nameRunes {
							if nameX+colOffset < m.width {
								grid[nameY][nameX+colOffset] = cell{
									char:  string(ch),
									color: overlay0,
									isSet: true,
								}
								colOffset++
							}
						}
						break
					}
				}
			}
		}
	}

	// Build output rows
	rows := make([]string, starAreaHeight)
	for y := 0; y < starAreaHeight; y++ {
		var row strings.Builder
		for x := 0; x < m.width; x++ {
			c := grid[y][x]
			if c.isSet {
				style := lipgloss.NewStyle().Foreground(c.color)
				row.WriteString(style.Render(c.char))
			} else {
				row.WriteString(" ")
			}
		}
		rows[y] = row.String()
	}

	return strings.Join(rows, "\n")
}

// renderCityPicker renders the city selection interface
func (m Model) renderCityPicker(starAreaHeight int) string {
	var s strings.Builder

	// Title
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(mauve).Render("📍 Select Location\n\n"))

	// Search box
	searchStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	searchText := m.cityPickerSearch
	if searchText == "" {
		searchText = "Type to search cities..."
	}
	s.WriteString(searchStyle.Render(searchText) + "\n\n")

	// City list
	maxDisplay := starAreaHeight - 5
	if maxDisplay < 5 {
		maxDisplay = 5
	}

	startIdx := 0
	if m.cityPickerIndex >= maxDisplay {
		startIdx = m.cityPickerIndex - maxDisplay + 1
	}

	endIdx := startIdx + maxDisplay
	if endIdx > len(m.cityPickerCities) {
		endIdx = len(m.cityPickerCities)
	}

	for i := startIdx; i < endIdx && i < len(m.cityPickerCities); i++ {
		city := m.cityPickerCities[i]
		prefix := "  "
		if i == m.cityPickerIndex {
			prefix = selectedStyle.Render("> ")
		}

		line := fmt.Sprintf("%s%s, %s\n", prefix, city.City, city.Country)
		s.WriteString(line)
	}

	// Instructions
	s.WriteString("\n" + lipgloss.NewStyle().Foreground(overlay0).Render("↑↓:select  enter:confirm  esc:cancel"))

	// Pad to fill the star area
	lines := strings.Split(s.String(), "\n")
	for len(lines) < starAreaHeight {
		lines = append(lines, "")
	}

	return strings.Join(lines[:starAreaHeight], "\n")
}

// handleCityPickerKey handles keys in city picker mode
func (m Model) handleCityPickerKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "q":
		m.showCityPicker = false
		m.cityPickerSearch = ""
		m.cityPickerCities = nil
	case "up", "k":
		if m.cityPickerIndex > 0 {
			m.cityPickerIndex--
		}
	case "down", "j":
		if m.cityPickerIndex < len(m.cityPickerCities)-1 {
			m.cityPickerIndex++
		}
	case "enter":
		if m.cityPickerIndex < len(m.cityPickerCities) {
			selected := m.cityPickerCities[m.cityPickerIndex]
			m.currentCity = &selected

			// Update observer
			observer := starfield.Observer{
				Latitude:  selected.Latitude,
				Longitude: selected.Longitude,
				Timezone:  selected.Timezone,
			}
			m.realisticStarfield.SetObserver(observer)
			m.realisticStarfield.Update(time.Now())

			// Close picker
			m.showCityPicker = false
			m.cityPickerSearch = ""
			m.cityPickerCities = nil
		}
	default:
		// Handle search input
		if len(key) == 1 {
			m.cityPickerSearch += key
			m.searchCities()
		} else if key == "backspace" && len(m.cityPickerSearch) > 0 {
			m.cityPickerSearch = m.cityPickerSearch[:len(m.cityPickerSearch)-1]
			m.searchCities()
		}
	}
	return m, nil
}

// loadAllCities loads all cities for the picker
func (m *Model) loadAllCities() {
	if m.citiesRepo == nil {
		return
	}
	cities, err := m.citiesRepo.Search("")
	if err != nil {
		m.cityPickerCities = []repository.City{}
		return
	}
	m.cityPickerCities = cities
	m.cityPickerIndex = 0
}

// searchCities searches for cities based on current search string
func (m *Model) searchCities() {
	if m.citiesRepo == nil {
		return
	}

	if m.cityPickerSearch == "" {
		m.loadAllCities()
		return
	}

	cities, err := m.citiesRepo.Search(m.cityPickerSearch)
	if err != nil {
		m.cityPickerCities = []repository.City{}
		return
	}

	m.cityPickerCities = cities
	m.cityPickerIndex = 0
}

// renderContent renders the content at bottom-left
func (m Model) renderContent() string {
	var s string
	paddingLeft := 2
	paddingBottom := 0

	// Add bottom padding (empty lines)
	for i := 0; i < paddingBottom; i++ {
		s += "\n"
	}

	// Add left padding
	leftPad := strings.Repeat(" ", paddingLeft)

	// Title
	s += leftPad + titleStyle.Render("⏱️  Veranda") + "\n"

	// Tabs - now 4 tabs: timers/stopwatches, ambience, skyfield, prefs
	timersStopwatchesTab := inactiveTabStyle.Render("Timers/Stopwatches")
	ambienceTab := inactiveTabStyle.Render("Ambience")
	skyfieldTab := inactiveTabStyle.Render("Skyfield")
	prefsTab := inactiveTabStyle.Render("Prefs")

	switch m.activeTab {
	case 0:
		timersStopwatchesTab = activeTabStyle.Render("Timers/Stopwatches")
	case 1:
		ambienceTab = activeTabStyle.Render("Ambience")
	case 2:
		skyfieldTab = activeTabStyle.Render("Skyfield")
	case 3:
		prefsTab = activeTabStyle.Render("Prefs")
	}

	s += leftPad + lipgloss.JoinHorizontal(lipgloss.Left, timersStopwatchesTab, ambienceTab, skyfieldTab, prefsTab) + "\n"

	// Content
	if m.err != nil {
		s += leftPad + fmt.Sprintf("Error: %v\n", m.err)
	} else {
		switch m.activeTab {
		case 0:
			s += leftPad + strings.ReplaceAll(m.renderTimersStopwatches(), "\n", "\n"+leftPad)
		case 1:
			s += leftPad + strings.ReplaceAll(m.renderAmbience(), "\n", "\n"+leftPad)
		case 2:
			s += leftPad + strings.ReplaceAll(m.renderSkyfield(), "\n", "\n"+leftPad)
		case 3:
			s += leftPad + strings.ReplaceAll(m.renderPrefs(), "\n", "\n"+leftPad)
		}
	}

	// Status line with mode, location, and names toggle
	modeText := "random"
	if m.starfieldMode == ModeRealistic {
		modeText = "realistic"
	}

	locationText := "Mumbai"
	if m.currentCity != nil {
		locationText = m.currentCity.City
	}

	namesText := "off"
	if m.showObjectNames {
		namesText = "on"
	}

	statusLine := lipgloss.NewStyle().Foreground(overlay0).Render(
		fmt.Sprintf("Mode: %s [%s] | Loc: %s [%s] | Names: %s [%s] | ",
			modeText, lipgloss.NewStyle().Foreground(mauve).Render("m"),
			locationText, lipgloss.NewStyle().Foreground(mauve).Render("l"),
			namesText, lipgloss.NewStyle().Foreground(mauve).Render("n")),
	)

	// Help - different based on active tab
	var helpText string
	switch m.activeTab {
	case 1:
		// Ambience tab
		if m.editingAmbienceVolume {
			helpText = lipgloss.NewStyle().Foreground(mauve).Render("0-9") +
				lipgloss.NewStyle().Foreground(overlay0).Render("/↑↓:volume ") +
				lipgloss.NewStyle().Foreground(mauve).Render("enter/esc") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":done")
		} else {
			helpText = lipgloss.NewStyle().Foreground(mauve).Render("↑↓") +
				lipgloss.NewStyle().Foreground(overlay0).Render("/jk:select ") +
				lipgloss.NewStyle().Foreground(mauve).Render("space") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":toggle ") +
				lipgloss.NewStyle().Foreground(mauve).Render("t") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":test ") +
				lipgloss.NewStyle().Foreground(mauve).Render("v") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":vol ") +
				lipgloss.NewStyle().Foreground(mauve).Render("r") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":reset")
		}
	case 3:
		// Preferences tab
		if m.selectingSoundFile {
			helpText = lipgloss.NewStyle().Foreground(mauve).Render("↑↓") +
				lipgloss.NewStyle().Foreground(overlay0).Render("/jk:select ") +
				lipgloss.NewStyle().Foreground(mauve).Render("enter") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":confirm ") +
				lipgloss.NewStyle().Foreground(mauve).Render("esc") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":cancel")
		} else if m.editingChimeVolume {
			helpText = lipgloss.NewStyle().Foreground(mauve).Render("0-9") +
				lipgloss.NewStyle().Foreground(overlay0).Render("/↑↓:volume ") +
				lipgloss.NewStyle().Foreground(mauve).Render("enter/esc") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":done")
		} else {
			helpText = lipgloss.NewStyle().Foreground(mauve).Render("e") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":chime ") +
				lipgloss.NewStyle().Foreground(mauve).Render("v") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":chime-vol ") +
				lipgloss.NewStyle().Foreground(mauve).Render("tab") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":switch ") +
				lipgloss.NewStyle().Foreground(mauve).Render("q") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":quit")
		}
	default:
		// Timer/stopwatch tabs
		helpText = lipgloss.NewStyle().Foreground(mauve).Render("↑↓") +
			lipgloss.NewStyle().Foreground(overlay0).Render("/jk:select ") +
			lipgloss.NewStyle().Foreground(mauve).Render("space/p") +
			lipgloss.NewStyle().Foreground(overlay0).Render(":pause ") +
			lipgloss.NewStyle().Foreground(mauve).Render("d") +
			lipgloss.NewStyle().Foreground(overlay0).Render(":delete ") +
			lipgloss.NewStyle().Foreground(mauve).Render("tab") +
			lipgloss.NewStyle().Foreground(overlay0).Render(":switch ") +
			lipgloss.NewStyle().Foreground(mauve).Render("t/s") +
			lipgloss.NewStyle().Foreground(overlay0).Render(":new ") +
			lipgloss.NewStyle().Foreground(mauve).Render("r") +
			lipgloss.NewStyle().Foreground(overlay0).Render(":refresh ") +
			lipgloss.NewStyle().Foreground(mauve).Render("q") +
			lipgloss.NewStyle().Foreground(overlay0).Render(":quit")
	}

	s += leftPad + statusLine + helpText + "\n"

	return s
}

// renderTimers renders the timer list
func (m Model) renderTimers() string {
	if len(m.timers) == 0 {
		return "No timers yet. Press 't' to create one.\n"
	}

	var content string
	completedCount := 0
	displayedCount := 0
	visibleIdx := 0

	for _, t := range m.timers {
		label := t["label"].(string)
		status := t["status"].(string)

		// Filter out completed timers unless showCompleted is true
		if status == "completed" && !m.showCompleted {
			completedCount++
			continue
		}

		// Selection indicator
		selected := "  "
		if visibleIdx == m.selectedIdx {
			selected = selectedStyle.Render("> ")
		}
		visibleIdx++
		displayedCount++

		durationMs := int64(t["duration_ms"].(float64))

		// Get current remaining time
		var remainingMs int64
		if curr, ok := t["current_remaining_ms"]; ok {
			remainingMs = int64(curr.(float64))
		} else {
			remainingMs = int64(t["remaining_ms"].(float64))
		}

		// Get progress
		var progress float64
		if p, ok := t["progress"]; ok {
			progress = p.(float64)
		}

		statusIcon := "○"
		statusStyle := inactiveTabStyle

		switch status {
		case "running":
			statusIcon = "▶"
			statusStyle = runningStyle
		case "paused":
			statusIcon = "⏸"
			statusStyle = pausedStyle
		case "completed":
			statusIcon = "✓"
			statusStyle = completedStyle
		}

		// Format times
		totalTime := formatDuration(durationMs)
		remainingTime := formatDuration(remainingMs)

		// Build progress bar
		progressBar := renderProgressBar(progress, 25)

		// Calculate percentage
		percent := int(progress * 100)

		content += fmt.Sprintf("%s%s %s\n",
			selected,
			statusStyle.Render(statusIcon),
			label,
		)
		content += fmt.Sprintf("    %s / %s %s %d%%\n",
			remainingStyle.Render(remainingTime),
			totalTime,
			progressBar,
			percent,
		)
		content += "\n"
	}

	// Show message if all timers are completed and hidden
	if displayedCount == 0 && completedCount > 0 {
		content = lipgloss.NewStyle().Foreground(mauve).Render("✓ All timers completed\n\n")
	}

	// Show count of hidden completed timers
	if completedCount > 0 {
		content += fmt.Sprintf("%s %s %s\n",
			lipgloss.NewStyle().Foreground(overlay0).Render(fmt.Sprintf("%d completed timer(s) hidden", completedCount)),
			lipgloss.NewStyle().Foreground(surface0).Render("·"),
			lipgloss.NewStyle().Foreground(mauve).Render("press 'h' to show"),
		)
	}

	return content
}

// renderStopwatches renders the stopwatch list
func (m Model) renderStopwatches() string {
	if len(m.stopwatches) == 0 {
		return "No stopwatches yet. Press 's' to create one.\n"
	}

	var content string
	now := time.Now()

	for i, sw := range m.stopwatches {
		label := sw["label"].(string)
		status := sw["status"].(string)

		// Get elapsed time
		var elapsedMs int64
		if status == "running" {
			if startedAt, ok := sw["started_at_ms"]; ok && startedAt != nil {
				startedAtMs := int64(startedAt.(float64))
				baseElapsed := int64(sw["elapsed_ms"].(float64))
				elapsedMs = baseElapsed + now.UnixMilli() - startedAtMs
			} else {
				if curr, ok := sw["current_elapsed_ms"]; ok {
					elapsedMs = int64(curr.(float64))
				} else {
					elapsedMs = int64(sw["elapsed_ms"].(float64))
				}
			}
		} else {
			if curr, ok := sw["current_elapsed_ms"]; ok {
				elapsedMs = int64(curr.(float64))
			} else {
				elapsedMs = int64(sw["elapsed_ms"].(float64))
			}
		}

		// Selection indicator
		selected := "  "
		if i == m.selectedIdx {
			selected = selectedStyle.Render("> ")
		}

		statusIcon := "○"
		statusStyle := inactiveTabStyle

		if status == "running" {
			statusIcon = "▶"
			statusStyle = runningStyle
		}

		elapsedStr := formatStopwatch(elapsedMs)

		content += fmt.Sprintf("%s%s %s %s %s\n",
			selected,
			statusStyle.Render(statusIcon),
			label,
			elapsedStr,
			statusStyle.Render(status),
		)
	}

	return content
}

// renderPrefs renders the preferences/settings page
func (m Model) renderPrefs() string {
	var content string

	// Title
	content += "Settings\n\n"

	// Timer completion sound setting
	content += "Timer Completion Sound\n"

	if m.selectingSoundFile {
		// Show dropdown with available sound files
		content += lipgloss.NewStyle().Foreground(mauve).Render("Select a sound file:") + "\n\n"

		if len(m.availableSoundFiles) == 0 {
			content += lipgloss.NewStyle().Foreground(red).Render("  No MP3 files found in chimes directory") + "\n"
			content += lipgloss.NewStyle().Foreground(overlay0).Render("  Place .mp3 files in ~/.veranda/sounds/chimes/") + "\n"
		} else {
			// Show up to 5 files at a time with scrolling
			maxDisplay := 5
			startIdx := 0
			if m.selectedSoundIdx >= maxDisplay {
				startIdx = m.selectedSoundIdx - maxDisplay + 1
			}
			endIdx := startIdx + maxDisplay
			if endIdx > len(m.availableSoundFiles) {
				endIdx = len(m.availableSoundFiles)
			}

			for i := startIdx; i < endIdx; i++ {
				prefix := "  "
				if i == m.selectedSoundIdx {
					prefix = selectedStyle.Render("> ")
				}
				fileName := m.availableSoundFiles[i]
				// Highlight current selection
				if i == m.selectedSoundIdx {
					fileName = lipgloss.NewStyle().Foreground(lavender).Bold(true).Render(fileName)
				} else {
					fileName = lipgloss.NewStyle().Foreground(text).Render(fileName)
				}
				content += prefix + fileName + "\n"
			}

			// Show count if there are more files
			if len(m.availableSoundFiles) > maxDisplay {
				content += lipgloss.NewStyle().Foreground(overlay0).Render(
					fmt.Sprintf("  (%d more files)", len(m.availableSoundFiles)-maxDisplay)) + "\n"
			}
		}

		content += "\n" + lipgloss.NewStyle().Foreground(overlay0).Render("↑↓:select  enter:confirm  esc:cancel") + "\n"
	} else {
		// Show current value with edit hint
		soundFile := m.userConfig.TimerCompletionSound
		content += lipgloss.NewStyle().Foreground(lavender).Render(soundFile) + "\n"
		content += lipgloss.NewStyle().Foreground(overlay0).Render("       [e]dit sound") + "\n"
	}

	content += "\n"

	// Chime Volume setting
	content += "Chime Volume\n"
	volBar := renderVolumeBar(m.userConfig.ChimeVolume, 10)
	volPercent := int(m.userConfig.ChimeVolume * 100)
	volStr := fmt.Sprintf("%d%%", volPercent)
	if m.editingChimeVolume {
		volStr = lipgloss.NewStyle().Foreground(yellow).Render(volStr)
	}
	content += fmt.Sprintf("%s %s\n", volBar, volStr)
	content += lipgloss.NewStyle().Foreground(overlay0).Render("       [v]olume (0-9)") + "\n"

	// Show settings message if any
	if m.settingsMessage != "" {
		content += "\n" + lipgloss.NewStyle().Foreground(green).Render(m.settingsMessage) + "\n"
	}

	return content
}

// renderTimersStopwatches renders the combined timer and stopwatch list
func (m Model) renderTimersStopwatches() string {
	// For now, just render timers (will be implemented in Phase 3)
	return m.renderTimers()
}

// renderSkyfield renders the skyfield settings tab
func (m Model) renderSkyfield() string {
	var content string

	// Title
	content += "Skyfield\n\n"

	// Mode setting
	modeText := m.userConfig.SkyfieldMode
	if modeText == "" {
		modeText = "random"
	}
	content += fmt.Sprintf("Mode:      %s\n", modeText)

	// Names setting
	namesText := "off"
	if m.userConfig.SkyfieldShowNames {
		namesText = "on"
	}
	content += fmt.Sprintf("Names:     %s\n", namesText)

	// Location
	locationText := "Mumbai"
	if m.currentCity != nil {
		locationText = m.currentCity.City
	}
	content += fmt.Sprintf("Location:  %s\n", locationText)

	content += "\n" + lipgloss.NewStyle().Foreground(overlay0).Render(
		"[←/→]:toggle mode [space]:toggle names [l]:change location") + "\n"

	return content
}

// Run starts the TUI
func Run(port int, citiesRepo *repository.CitiesRepository, settingsRepo config.SettingsRepository) error {
	p := tea.NewProgram(New(port, citiesRepo, settingsRepo), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
