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

		// Handle command mode first
		if m.commandMode {
			switch msg.String() {
			case "esc":
				m.commandMode = false
				m.commandInput = ""
				m.commandError = ""
				return m, nil
			case "enter":
				return m.executeCommand()
			case "backspace":
				if len(m.commandInput) > 0 {
					m.commandInput = m.commandInput[:len(m.commandInput)-1]
				}
				return m, nil
			default:
				// Append character to command input
				if len(msg.String()) == 1 {
					m.commandInput += msg.String()
				}
				return m, nil
			}
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
			} else if m.activeTab == 0 {
				m.navigateTimersStopwatches(-1)
			} else {
				m.selectedIdx = max(0, m.selectedIdx-1)
			}
		case "down", "j":
			if m.activeTab == 1 {
				newM, _, _ := m.handleAmbienceKey(msg.String())
				return newM, nil
			} else if m.activeTab == 0 {
				m.navigateTimersStopwatches(1)
			} else {
				maxIdx := len(m.timers) - 1
				if m.activeTab == 3 {
					maxIdx = len(m.stopwatches) - 1
				}
				m.selectedIdx = min(maxIdx, m.selectedIdx+1)
			}
		case "left":
			if m.activeTab == 0 && m.showTimers && m.showStopwatches {
				m.activeColumn = 0
			}
		case "right":
			if m.activeTab == 0 && m.showTimers && m.showStopwatches {
				m.activeColumn = 1
			}
		case " ", "p":
			// Pause/resume selected timer or stopwatch
			if m.activeTab == 0 {
				return m.handleTimersStopwatchesToggle()
			} else if m.activeTab == 1 {
				// Toggle ambience enabled
				newM, _, _ := m.handleAmbienceKey(" ")
				return newM, nil
			}
		case "d":
			// Delete selected timer or stopwatch
			if m.activeTab == 0 {
				return m.handleTimersStopwatchesDelete()
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
			// Toggle showing completed/archived items
			if m.activeTab == 0 {
				m.showCompleted = !m.showCompleted
				// Reset selections
				m.timerSelectedIdx = 0
				m.stopwatchSelectedIdx = 0
				m.timerScrollOffset = 0
				m.stopwatchScrollOffset = 0
			}
		case "t":
			if m.activeTab == 0 {
				// Toggle show timers
				m.showTimers = !m.showTimers
				if !m.showTimers && !m.showStopwatches {
					m.showStopwatches = true // Ensure at least one is visible
				}
				m.activeColumn = 0
			} else if m.activeTab == 1 {
				// Test toggle for ambience sound - use async command
				if err := m.initTestAudioPlayer(); err != nil {
					m.setSettingsMessage(fmt.Sprintf("Audio error: %v", err))
					return m, nil
				}
				newM, cmd, _ := m.handleAmbienceKey("t")
				return newM, cmd
			}
		case "s":
			if m.activeTab == 0 {
				// Toggle show stopwatches
				m.showStopwatches = !m.showStopwatches
				if !m.showTimers && !m.showStopwatches {
					m.showTimers = true // Ensure at least one is visible
				}
				m.activeColumn = 1
			}
		case "/":
			// Enter command mode
			if m.activeTab == 0 {
				m.commandMode = true
				m.commandInput = ""
				m.commandError = ""
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
		if m.commandMode {
			// Command mode
			helpText = lipgloss.NewStyle().Foreground(mauve).Render("enter") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":execute ") +
				lipgloss.NewStyle().Foreground(mauve).Render("esc") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":cancel")
		} else {
			// Timers/Stopwatches tab
			helpText = lipgloss.NewStyle().Foreground(mauve).Render("↑↓") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":navigate ") +
				lipgloss.NewStyle().Foreground(mauve).Render("←/→") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":column ") +
				lipgloss.NewStyle().Foreground(mauve).Render("space/p") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":pause ") +
				lipgloss.NewStyle().Foreground(mauve).Render("d") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":delete ") +
				lipgloss.NewStyle().Foreground(mauve).Render("tab") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":switch ") +
				lipgloss.NewStyle().Foreground(mauve).Render("t/s/h") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":toggle ") +
				lipgloss.NewStyle().Foreground(mauve).Render("/") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":command ") +
				lipgloss.NewStyle().Foreground(mauve).Render("q") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":quit")
		}
	}

	// Show command mode input line if active
	if m.commandMode {
		cmdLine := lipgloss.NewStyle().Foreground(mauve).Render(":"+m.commandInput) + "_"
		if m.commandError != "" {
			cmdLine += " " + lipgloss.NewStyle().Foreground(red).Render(m.commandError)
		}
		s += leftPad + cmdLine + "\n"
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

// renderTimersStopwatches renders the combined timer and stopwatch list
// Shows max 3 rows per section with scrolling, supports 1 or 2 column layout
func (m Model) renderTimersStopwatches() string {
	var content string

	// Header with toggle status
	content += "Timers/Stopwatches\n\n"

	// Show toggle status line
	timersStatus := "off"
	if m.showTimers {
		timersStatus = "on"
	}
	stopwatchesStatus := "off"
	if m.showStopwatches {
		stopwatchesStatus = "on"
	}
	completedStatus := "hidden"
	if m.showCompleted {
		completedStatus = "shown"
	}

	content += lipgloss.NewStyle().Foreground(overlay0).Render(
		fmt.Sprintf("[t]timers:%s [s]stopwatches:%s [h]completed:%s",
			lipgloss.NewStyle().Foreground(mauve).Render(timersStatus),
			lipgloss.NewStyle().Foreground(mauve).Render(stopwatchesStatus),
			lipgloss.NewStyle().Foreground(mauve).Render(completedStatus))) + "\n\n"

	// Handle different view modes
	if !m.showTimers && !m.showStopwatches {
		return content + lipgloss.NewStyle().Foreground(overlay0).Render("Press 't' to show timers or 's' to show stopwatches") + "\n"
	}

	// Get visible items
	visibleTimers := m.getVisibleTimers()
	visibleStopwatches := m.getVisibleStopwatches()

	// Single column layout (only one section visible)
	if (m.showTimers && !m.showStopwatches) || (!m.showTimers && m.showStopwatches) {
		if m.showTimers {
			content += m.renderTimersSection(visibleTimers, true)
		} else {
			content += m.renderStopwatchesSection(visibleStopwatches, true)
		}
		return content
	}

	// Two column layout (both visible)
	// Render side by side
	timersContent := m.renderTimersSection(visibleTimers, false)
	stopwatchesContent := m.renderStopwatchesSection(visibleStopwatches, false)

	// Split into lines and combine side by side
	timersLines := strings.Split(timersContent, "\n")
	stopwatchLines := strings.Split(stopwatchesContent, "\n")

	maxLines := max(len(timersLines), len(stopwatchLines))
	for i := 0; i < maxLines; i++ {
		timerLine := ""
		if i < len(timersLines) {
			timerLine = timersLines[i]
		}
		stopwatchLine := ""
		if i < len(stopwatchLines) {
			stopwatchLine = stopwatchLines[i]
		}

		// Pad timer line to consistent width (40 chars)
		padding := 40 - len(timerLine)
		if padding < 0 {
			padding = 0
		}

		content += timerLine + strings.Repeat(" ", padding) + stopwatchLine + "\n"
	}

	return content
}

// getVisibleTimers returns filtered timers based on showCompleted setting
func (m Model) getVisibleTimers() []map[string]interface{} {
	var visible []map[string]interface{}
	for _, t := range m.timers {
		status := t["status"].(string)
		if status == "completed" && !m.showCompleted {
			continue
		}
		visible = append(visible, t)
	}
	return visible
}

// getVisibleStopwatches returns all non-deleted stopwatches
func (m Model) getVisibleStopwatches() []map[string]interface{} {
	return m.stopwatches
}

// renderTimersSection renders the timers section (max 3 visible rows with scrolling)
func (m Model) renderTimersSection(timers []map[string]interface{}, fullWidth bool) string {
	var content string

	// Header
	content += lipgloss.NewStyle().Foreground(mauve).Bold(true).Render("Timers") + "\n"

	if len(timers) == 0 {
		content += lipgloss.NewStyle().Foreground(overlay0).Render("  No timers. Use /t to create") + "\n"
		return content
	}

	// Calculate scroll window (max 3 rows visible)
	const maxVisible = 3
	totalTimers := len(timers)
	startIdx := m.timerScrollOffset
	if startIdx > totalTimers-maxVisible {
		startIdx = max(0, totalTimers-maxVisible)
	}
	endIdx := min(startIdx+maxVisible, totalTimers)

	// Show scroll indicator if needed
	if startIdx > 0 {
		content += lipgloss.NewStyle().Foreground(overlay0).Render("  ↑ more") + "\n"
	}

	// Render visible timers
	for i := startIdx; i < endIdx; i++ {
		t := timers[i]
		label := t["label"].(string)
		status := t["status"].(string)

		// Selection indicator (only if this column is active)
		selected := "  "
		if m.activeColumn == 0 && i == m.timerSelectedIdx {
			selected = selectedStyle.Render("> ")
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

		// Get time info
		var remainingMs int64
		if curr, ok := t["current_remaining_ms"]; ok {
			remainingMs = int64(curr.(float64))
		} else {
			remainingMs = int64(t["remaining_ms"].(float64))
		}

		remainingTime := formatDuration(remainingMs)

		content += fmt.Sprintf("%s%s %s %s\n",
			selected,
			statusStyle.Render(statusIcon),
			label,
			remainingStyle.Render(remainingTime),
		)
	}

	// Show scroll indicator if more below
	if endIdx < totalTimers {
		content += lipgloss.NewStyle().Foreground(overlay0).Render("  ↓ more") + "\n"
	}

	return content
}

// renderStopwatchesSection renders the stopwatches section (max 3 visible rows with scrolling)
func (m Model) renderStopwatchesSection(stopwatches []map[string]interface{}, fullWidth bool) string {
	var content string

	// Header
	content += lipgloss.NewStyle().Foreground(mauve).Bold(true).Render("Stopwatches") + "\n"

	if len(stopwatches) == 0 {
		content += lipgloss.NewStyle().Foreground(overlay0).Render("  No stopwatches. Use /s to create") + "\n"
		return content
	}

	// Calculate scroll window (max 3 rows visible)
	const maxVisible = 3
	totalStopwatches := len(stopwatches)
	startIdx := m.stopwatchScrollOffset
	if startIdx > totalStopwatches-maxVisible {
		startIdx = max(0, totalStopwatches-maxVisible)
	}
	endIdx := min(startIdx+maxVisible, totalStopwatches)

	// Show scroll indicator if needed
	if startIdx > 0 {
		content += lipgloss.NewStyle().Foreground(overlay0).Render("  ↑ more") + "\n"
	}

	now := time.Now()

	// Render visible stopwatches
	for i := startIdx; i < endIdx; i++ {
		sw := stopwatches[i]
		label := sw["label"].(string)
		status := sw["status"].(string)

		// Selection indicator (only if this column is active)
		selected := "  "
		if m.activeColumn == 1 && i == m.stopwatchSelectedIdx {
			selected = selectedStyle.Render("> ")
		}

		statusIcon := "○"
		statusStyle := inactiveTabStyle

		if status == "running" {
			statusIcon = "▶"
			statusStyle = runningStyle
		}

		// Get elapsed time
		var elapsedMs int64
		if status == "running" {
			if startedAt, ok := sw["started_at_ms"]; ok && startedAt != nil {
				startedAtMs := int64(startedAt.(float64))
				baseElapsed := int64(sw["elapsed_ms"].(float64))
				elapsedMs = baseElapsed + now.UnixMilli() - startedAtMs
			} else {
				elapsedMs = int64(sw["elapsed_ms"].(float64))
			}
		} else {
			elapsedMs = int64(sw["elapsed_ms"].(float64))
		}

		elapsedStr := formatStopwatch(elapsedMs)

		content += fmt.Sprintf("%s%s %s %s\n",
			selected,
			statusStyle.Render(statusIcon),
			label,
			elapsedStr,
		)
	}

	// Show scroll indicator if more below
	if endIdx < totalStopwatches {
		content += lipgloss.NewStyle().Foreground(overlay0).Render("  ↓ more") + "\n"
	}

	return content
}

// navigateTimersStopwatches handles navigation within the timers/stopwatches tab
// delta: -1 for up, 1 for down
func (m *Model) navigateTimersStopwatches(delta int) {
	if m.activeColumn == 0 && m.showTimers {
		// Navigate timers
		visibleTimers := m.getVisibleTimers()
		if len(visibleTimers) == 0 {
			return
		}

		newIdx := m.timerSelectedIdx + delta
		if newIdx < 0 {
			newIdx = 0
		}
		if newIdx >= len(visibleTimers) {
			newIdx = len(visibleTimers) - 1
		}

		m.timerSelectedIdx = newIdx

		// Update scroll offset to keep selection visible
		const maxVisible = 3
		if m.timerSelectedIdx < m.timerScrollOffset {
			m.timerScrollOffset = m.timerSelectedIdx
		}
		if m.timerSelectedIdx >= m.timerScrollOffset+maxVisible {
			m.timerScrollOffset = m.timerSelectedIdx - maxVisible + 1
		}
	} else if m.activeColumn == 1 && m.showStopwatches {
		// Navigate stopwatches
		visibleStopwatches := m.getVisibleStopwatches()
		if len(visibleStopwatches) == 0 {
			return
		}

		newIdx := m.stopwatchSelectedIdx + delta
		if newIdx < 0 {
			newIdx = 0
		}
		if newIdx >= len(visibleStopwatches) {
			newIdx = len(visibleStopwatches) - 1
		}

		m.stopwatchSelectedIdx = newIdx

		// Update scroll offset to keep selection visible
		const maxVisible = 3
		if m.stopwatchSelectedIdx < m.stopwatchScrollOffset {
			m.stopwatchScrollOffset = m.stopwatchSelectedIdx
		}
		if m.stopwatchSelectedIdx >= m.stopwatchScrollOffset+maxVisible {
			m.stopwatchScrollOffset = m.stopwatchSelectedIdx - maxVisible + 1
		}
	}
}

// handleTimersStopwatchesToggle handles pause/resume for the selected timer or stopwatch
func (m Model) handleTimersStopwatchesToggle() (tea.Model, tea.Cmd) {
	if m.activeColumn == 0 && m.showTimers {
		visibleTimers := m.getVisibleTimers()
		if m.timerSelectedIdx < len(visibleTimers) {
			// Find the actual index in m.timers
			actualIdx := m.findTimerActualIndex(visibleTimers[m.timerSelectedIdx])
			if actualIdx >= 0 {
				return m, toggleTimerCmd(m.client, m.timers, actualIdx, m.showCompleted)
			}
		}
	} else if m.activeColumn == 1 && m.showStopwatches {
		visibleStopwatches := m.getVisibleStopwatches()
		if m.stopwatchSelectedIdx < len(visibleStopwatches) {
			// Find the actual index in m.stopwatches
			actualIdx := m.findStopwatchActualIndex(visibleStopwatches[m.stopwatchSelectedIdx])
			if actualIdx >= 0 {
				return m, toggleStopwatchCmd(m.client, m.stopwatches, actualIdx)
			}
		}
	}
	return m, nil
}

// handleTimersStopwatchesDelete handles delete for the selected timer or stopwatch
func (m Model) handleTimersStopwatchesDelete() (tea.Model, tea.Cmd) {
	if m.activeColumn == 0 && m.showTimers {
		visibleTimers := m.getVisibleTimers()
		if m.timerSelectedIdx < len(visibleTimers) {
			actualIdx := m.findTimerActualIndex(visibleTimers[m.timerSelectedIdx])
			if actualIdx >= 0 {
				return m, deleteTimerCmd(m.client, m.timers, actualIdx, m.showCompleted)
			}
		}
	} else if m.activeColumn == 1 && m.showStopwatches {
		visibleStopwatches := m.getVisibleStopwatches()
		if m.stopwatchSelectedIdx < len(visibleStopwatches) {
			actualIdx := m.findStopwatchActualIndex(visibleStopwatches[m.stopwatchSelectedIdx])
			if actualIdx >= 0 {
				return m, deleteStopwatchCmd(m.client, m.stopwatches, actualIdx)
			}
		}
	}
	return m, nil
}

// executeCommand parses and executes a command from command mode
// Commands: :t 25m "name" (timer), :s "name" (stopwatch)
func (m Model) executeCommand() (tea.Model, tea.Cmd) {
	cmd := strings.TrimSpace(m.commandInput)
	if cmd == "" {
		m.commandMode = false
		m.commandInput = ""
		return m, nil
	}

	// Remove leading colon if present
	if strings.HasPrefix(cmd, ":") {
		cmd = cmd[1:]
	}

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		m.commandError = "Empty command"
		return m, nil
	}

	cmdType := parts[0]

	switch cmdType {
	case "t", "timer":
		return m.parseAndCreateTimer(parts[1:])
	case "s", "sw", "stopwatch":
		return m.parseAndCreateStopwatch(parts[1:])
	default:
		m.commandError = fmt.Sprintf("Unknown command: %s", cmdType)
		return m, nil
	}
}

// parseAndCreateTimer parses timer creation arguments and returns a command
// Format: 25m "name" or 1h30m "name" or just 25m
func (m Model) parseAndCreateTimer(args []string) (tea.Model, tea.Cmd) {
	if len(args) == 0 {
		m.commandError = "Usage: :t <duration> [\"name\"]"
		return m, nil
	}

	// Parse duration
	durationMs, err := parseDuration(args[0])
	if err != nil {
		m.commandError = fmt.Sprintf("Invalid duration: %v", err)
		return m, nil
	}

	// Parse label (optional, can be quoted)
	label := "Timer"
	if len(args) > 1 {
		label = strings.Join(args[1:], " ")
		label = strings.Trim(label, "\"")
	}

	m.commandMode = false
	m.commandInput = ""
	m.commandError = ""

	return m, createTimerWithLabelCmd(m.client, label, durationMs)
}

// parseAndCreateStopwatch parses stopwatch creation arguments and returns a command
// Format: "name" or (no args for default)
func (m Model) parseAndCreateStopwatch(args []string) (tea.Model, tea.Cmd) {
	label := "Stopwatch"
	if len(args) > 0 {
		label = strings.Join(args, " ")
		label = strings.Trim(label, "\"")
	}

	m.commandMode = false
	m.commandInput = ""
	m.commandError = ""

	return m, createStopwatchWithLabelCmd(m.client, label)
}

// parseDuration parses a duration string like "25m", "1h30m", "30s"
func parseDuration(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}

	var totalMs int64
	var numStr string

	for _, ch := range s {
		switch ch {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			numStr += string(ch)
		case 'h':
			if numStr == "" {
				return 0, fmt.Errorf("invalid duration format")
			}
			hours := parseInt64(numStr)
			totalMs += hours * 60 * 60 * 1000
			numStr = ""
		case 'm':
			if numStr == "" {
				return 0, fmt.Errorf("invalid duration format")
			}
			mins := parseInt64(numStr)
			totalMs += mins * 60 * 1000
			numStr = ""
		case 's':
			if numStr == "" {
				return 0, fmt.Errorf("invalid duration format")
			}
			secs := parseInt64(numStr)
			totalMs += secs * 1000
			numStr = ""
		default:
			return 0, fmt.Errorf("invalid character in duration: %c", ch)
		}
	}

	if numStr != "" {
		return 0, fmt.Errorf("duration must end with h, m, or s")
	}

	if totalMs == 0 {
		return 0, fmt.Errorf("duration cannot be zero")
	}

	return totalMs, nil
}

// parseInt64 parses a string to int64
func parseInt64(s string) int64 {
	var result int64
	for _, ch := range s {
		result = result*10 + int64(ch-'0')
	}
	return result
}

// findTimerActualIndex finds the actual index in m.timers for a visible timer
func (m Model) findTimerActualIndex(visibleTimer map[string]interface{}) int {
	visibleID := int64(visibleTimer["id"].(float64))
	for i, t := range m.timers {
		if int64(t["id"].(float64)) == visibleID {
			return i
		}
	}
	return -1
}

// findStopwatchActualIndex finds the actual index in m.stopwatches for a visible stopwatch
func (m Model) findStopwatchActualIndex(visibleStopwatch map[string]interface{}) int {
	visibleID := int64(visibleStopwatch["id"].(float64))
	for i, sw := range m.stopwatches {
		if int64(sw["id"].(float64)) == visibleID {
			return i
		}
	}
	return -1
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
