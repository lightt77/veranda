// Package tui provides the Terminal User Interface using Bubble Tea.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightt77/veranda/internal/daemon"
	"github.com/lightt77/veranda/internal/repository"
	"github.com/lightt77/veranda/internal/tui/starfield"
)

// StarfieldMode represents the starfield display mode
type StarfieldMode int

const (
	ModeRandom StarfieldMode = iota
	ModeRealistic
)

// twinkleMsg is sent to update star twinkling
type twinkleMsg time.Time

// starfieldUpdateMsg is sent to update realistic starfield positions
type starfieldUpdateMsg time.Time

// Model represents the TUI state
type Model struct {
	client        *daemon.Client
	timers        []map[string]interface{}
	stopwatches   []map[string]interface{}
	activeTab     int  // 0 = timers, 1 = stopwatches
	showCompleted bool // toggle to show/hide completed timers
	selectedIdx   int  // currently selected item index
	width         int
	height        int
	err           error
	lastUpdate    time.Time

	// Starfield mode
	starfieldMode      StarfieldMode
	randomStarfield    *starfield.RandomStarfield
	realisticStarfield *starfield.RealisticStarfield

	// Location
	citiesRepo  *repository.CitiesRepository
	currentCity *repository.City

	// City picker
	showCityPicker   bool
	cityPickerSearch string
	cityPickerCities []repository.City
	cityPickerIndex  int

	// Object name display toggle
	showObjectNames bool // Toggle to show/hide celestial object names
}

// Catppuccin Mocha Color Palette
var (
	// Base colors
	base     = lipgloss.Color("#1e1e2e") // Background
	text     = lipgloss.Color("#cdd6f4") // Foreground
	surface0 = lipgloss.Color("#313244") // Surface
	surface1 = lipgloss.Color("#45475a") // Surface border
	overlay0 = lipgloss.Color("#6c7086") // Subtle text

	// Accent colors (Purples)
	mauve    = lipgloss.Color("#cba6f7") // Primary accent
	lavender = lipgloss.Color("#b4befe") // Secondary accent
	pink     = lipgloss.Color("#f5c2e7") // Tertiary accent

	// Status colors
	green  = lipgloss.Color("#a6e3a1") // Running
	yellow = lipgloss.Color("#f9e2af") // Paused
	red    = lipgloss.Color("#f38ba8") // Error/Stopped
	blue   = lipgloss.Color("#89b4fa") // Info
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(mauve).
			MarginLeft(2)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(base).
			Background(mauve).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(overlay0).
				Padding(0, 2)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(surface1).
			Padding(1, 2).
			Width(60)

	helpStyle = lipgloss.NewStyle().
			Foreground(overlay0)

	runningStyle = lipgloss.NewStyle().
			Foreground(green)

	pausedStyle = lipgloss.NewStyle().
			Foreground(yellow)

	completedStyle = lipgloss.NewStyle().
			Foreground(overlay0)

	remainingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lavender)

	selectedStyle = lipgloss.NewStyle().
			Foreground(mauve).
			Bold(true)

	progressBarStyle = lipgloss.NewStyle().
				Foreground(mauve)

	progressBarEmptyStyle = lipgloss.NewStyle().
				Foreground(surface0)
)

// tickMsg is sent on every tick
type tickMsg time.Time

// refreshMsg is sent to refresh data
type refreshMsg struct{}

// errMsg wraps an error
type errMsg struct {
	err error
}

// dataMsg contains fetched data
type dataMsg struct {
	timers      []map[string]interface{}
	stopwatches []map[string]interface{}
}

// New creates a new TUI model
func New(port int, citiesRepo *repository.CitiesRepository) Model {
	// Get default city (Mumbai)
	var defaultCity *repository.City
	if citiesRepo != nil {
		defaultCity, _ = citiesRepo.DefaultCity()
	}

	// Create observer from city or use Mumbai defaults
	var observer starfield.Observer
	if defaultCity != nil {
		observer = starfield.Observer{
			Latitude:  defaultCity.Latitude,
			Longitude: defaultCity.Longitude,
			Timezone:  defaultCity.Timezone,
		}
	} else {
		observer = starfield.MumbaiObserver()
	}

	return Model{
		client:             daemon.NewClient(port),
		activeTab:          0,
		lastUpdate:         time.Now(),
		starfieldMode:      ModeRandom, // Default to random mode
		randomStarfield:    starfield.NewRandomStarfield(),
		realisticStarfield: starfield.NewRealisticStarfield(observer),
		citiesRepo:         citiesRepo,
		currentCity:        defaultCity,
	}
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		refreshCmd(m.client),
		twinkleCmd(),
		starfieldUpdateCmd(),
	)
}

// twinkleCmd creates a command that triggers star twinkling
func twinkleCmd() tea.Cmd {
	return tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
		return twinkleMsg(t)
	})
}

// starfieldUpdateCmd creates a command for realistic starfield updates (120s)
func starfieldUpdateCmd() tea.Cmd {
	return tea.Tick(120*time.Second, func(t time.Time) tea.Msg {
		return starfieldUpdateMsg(t)
	})
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate star area height
		contentHeight := 14
		paddingBottom := 2
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
		// Only twinkle in random mode
		if m.starfieldMode == ModeRandom {
			m.randomStarfield.Twinkle()
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
			switch msg.String() {
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
				if len(msg.String()) == 1 {
					m.cityPickerSearch += msg.String()
					m.searchCities()
				} else if msg.String() == "backspace" && len(m.cityPickerSearch) > 0 {
					m.cityPickerSearch = m.cityPickerSearch[:len(m.cityPickerSearch)-1]
					m.searchCities()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "right":
			m.activeTab = (m.activeTab + 1) % 2
			m.selectedIdx = 0
		case "shift+tab", "left":
			m.activeTab = (m.activeTab - 1 + 2) % 2
			m.selectedIdx = 0
		case "up", "k":
			m.selectedIdx = max(0, m.selectedIdx-1)
		case "down", "j":
			maxIdx := len(m.timers) - 1
			if m.activeTab == 1 {
				maxIdx = len(m.stopwatches) - 1
			}
			m.selectedIdx = min(maxIdx, m.selectedIdx+1)
		case " ", "p":
			// Pause/resume selected timer or stopwatch
			if m.activeTab == 0 && m.selectedIdx < len(m.timers) {
				return m, toggleTimerCmd(m.client, m.timers, m.selectedIdx, m.showCompleted)
			} else if m.activeTab == 1 && m.selectedIdx < len(m.stopwatches) {
				return m, toggleStopwatchCmd(m.client, m.stopwatches, m.selectedIdx)
			}
		case "d":
			// Delete selected timer or stopwatch
			if m.activeTab == 0 && m.selectedIdx < len(m.timers) {
				return m, deleteTimerCmd(m.client, m.timers, m.selectedIdx, m.showCompleted)
			} else if m.activeTab == 1 && m.selectedIdx < len(m.stopwatches) {
				return m, deleteStopwatchCmd(m.client, m.stopwatches, m.selectedIdx)
			}
		case "r":
			return m, refreshCmd(m.client)
		case "h":
			// Toggle showing completed timers
			if m.activeTab == 0 {
				m.showCompleted = !m.showCompleted
				m.selectedIdx = 0
			}
		case "t":
			// Quick timer creation placeholder
			if m.activeTab == 0 {
				return m, createTimerCmd(m.client)
			}
		case "s":
			// Quick stopwatch creation placeholder
			if m.activeTab == 1 {
				return m, createStopwatchCmd(m.client)
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
				m.cityPickerIndex = 0
				m.cityPickerSearch = ""
				m.loadAllCities()
			}
		case "n":
			// Toggle showing celestial object names
			m.showObjectNames = !m.showObjectNames
		}

	case tickMsg:
		m.lastUpdate = time.Time(msg)
		return m, tickCmd()

	case refreshMsg:
		return m, refreshCmd(m.client)

	case dataMsg:
		m.timers = msg.timers
		m.stopwatches = msg.stopwatches
		m.err = nil
		// Auto-refresh every 500ms
		return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return refreshMsg{}
		})

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
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

// loadAllCities loads all cities for the picker
func (m *Model) loadAllCities() {
	if m.citiesRepo == nil {
		return
	}

	cities, err := m.citiesRepo.GetAll()
	if err != nil {
		m.cityPickerCities = []repository.City{}
		return
	}

	m.cityPickerCities = cities
	m.cityPickerIndex = 0
}

// View renders the UI
func (m Model) View() string {
	if !m.client.IsRunning() {
		return "\n  ⚠️  Daemon is not running.\n\n  Start it with: veranda daemon\n\n  Press 'q' to quit.\n"
	}

	// Handle case where dimensions aren't set yet
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Generate stars if we don't have any
	if m.starfieldMode == ModeRandom && len(m.randomStarfield.GetStars()) == 0 {
		contentHeight := 14
		paddingBottom := 2
		starAreaHeight := m.height - contentHeight - paddingBottom
		if starAreaHeight < 8 {
			starAreaHeight = 8
		}
		m.randomStarfield.Resize(m.width, m.height, starAreaHeight)
	}

	// Calculate star area dimensions
	contentHeight := 14
	paddingBottom := 2
	starAreaHeight := m.height - contentHeight - paddingBottom
	if starAreaHeight < 8 {
		starAreaHeight = 8
	}

	// Build starfield
	var starfieldStr string
	if m.showCityPicker {
		starfieldStr = m.renderCityPicker(starAreaHeight)
	} else if m.starfieldMode == ModeRealistic {
		starfieldStr = m.renderRealisticStarfield(starAreaHeight)
	} else {
		starfieldStr = m.renderRandomStarfield(starAreaHeight)
	}

	// Get content
	content := m.renderContent()

	// Use lipgloss to place content at bottom
	contentArea := lipgloss.NewStyle().
		Height(contentHeight).
		Width(m.width).
		Align(lipgloss.Left, lipgloss.Bottom).
		Render(content)

	// Join starfield and content
	return starfieldStr + "\n" + contentArea
}

// renderRandomStarfield renders the random starfield
func (m Model) renderRandomStarfield(starAreaHeight int) string {
	rows := make([]string, starAreaHeight)

	// Map to track star positions
	starMap := make(map[[2]int]starfield.Star)
	for _, star := range m.randomStarfield.GetStars() {
		if star.Y < starAreaHeight && star.X < m.width {
			starMap[[2]int{star.X, star.Y}] = star
		}
	}

	// Render starfield rows
	for y := 0; y < starAreaHeight; y++ {
		var row strings.Builder
		for x := 0; x < m.width; x++ {
			if star, ok := starMap[[2]int{x, y}]; ok {
				row.WriteString(lipgloss.NewStyle().Foreground(star.Color).Render(star.Char))
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
	// grid[y][x] = {char, style}
	type Cell struct {
		char  string
		color lipgloss.Color
		isSet bool
	}

	grid := make([][]Cell, starAreaHeight)
	for y := range grid {
		grid[y] = make([]Cell, m.width)
	}

	// Get visible objects
	objects := m.realisticStarfield.GetVisibleObjects()

	// Track which positions have objects to avoid overlapping names
	occupiedPositions := make(map[[2]int]bool)

	// Render each object
	for _, obj := range objects {
		if obj.ScreenY >= 0 && obj.ScreenY < starAreaHeight &&
			obj.ScreenX >= 0 && obj.ScreenX < m.width {

			// Get symbol based on brightness
			symbol := starfield.GetObjectSymbol(obj.Object, obj.Brightness)

			// Get color
			color := starfield.GetObjectColor(obj.Object, obj.Brightness)

			// Place symbol in grid (single character/rune)
			if len(symbol) > 0 {
				// Convert to runes to handle multi-byte UTF-8 characters properly
				runes := []rune(symbol)
				if len(runes) > 0 {
					grid[obj.ScreenY][obj.ScreenX] = Cell{
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
				obj.ScreenX >= 0 && obj.ScreenX < m.width &&
				obj.Brightness > 0.3 { // Only show names for brighter objects

				// Place name to the right of the object
				nameX := obj.ScreenX + 2
				nameY := obj.ScreenY

				// Get shortened name
				name := obj.Object.Name
				if len(name) > 8 {
					name = name[:7] + "."
				}

				// Check if there's enough space and position isn't occupied
				if nameX < m.width-len(name) && nameY < starAreaHeight {
					canPlace := true
					for i := 0; i < len(name) && nameX+i < m.width; i++ {
						if occupiedPositions[[2]int{nameX + i, nameY}] || grid[nameY][nameX+i].isSet {
							canPlace = false
							break
						}
					}

					if canPlace {
						// Place name characters in grid
						colOffset := 0
						for _, ch := range name {
							if nameX+colOffset < m.width {
								grid[nameY][nameX+colOffset] = Cell{
									char:  string(ch),
									color: overlay0,
									isSet: true,
								}
								colOffset++
							}
						}
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
			cell := grid[y][x]
			if cell.isSet {
				style := lipgloss.NewStyle().Foreground(cell.color)
				row.WriteString(style.Render(cell.char))
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

// renderContent renders the content at bottom-left
func (m Model) renderContent() string {
	var s string
	paddingLeft := 2
	paddingBottom := 1

	// Add bottom padding (empty lines)
	for i := 0; i < paddingBottom; i++ {
		s += "\n"
	}

	// Add left padding
	leftPad := strings.Repeat(" ", paddingLeft)

	// Title
	s += leftPad + titleStyle.Render("⏱️  Veranda") + "\n"

	// Tabs
	timersTab := inactiveTabStyle.Render("Timers [t]")
	stopwatchesTab := inactiveTabStyle.Render("Stopwatches [s]")

	if m.activeTab == 0 {
		timersTab = activeTabStyle.Render("Timers [t]")
	} else {
		stopwatchesTab = activeTabStyle.Render("Stopwatches [s]")
	}

	s += leftPad + lipgloss.JoinHorizontal(lipgloss.Left, timersTab, stopwatchesTab) + "\n"

	// Content
	if m.err != nil {
		s += leftPad + fmt.Sprintf("Error: %v\n", m.err)
	} else if m.activeTab == 0 {
		s += leftPad + strings.ReplaceAll(m.renderTimers(), "\n", "\n"+leftPad)
	} else {
		s += leftPad + strings.ReplaceAll(m.renderStopwatches(), "\n", "\n"+leftPad)
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
		fmt.Sprintf("Mode: %s [%s] | Loc: %s [%s] | Names: %s [%s] | ", modeText, lipgloss.NewStyle().Foreground(mauve).Render("m"), locationText, lipgloss.NewStyle().Foreground(mauve).Render("l"), namesText, lipgloss.NewStyle().Foreground(mauve).Render("n")),
	)

	// Help
	helpText := lipgloss.NewStyle().Foreground(mauve).Render("↑↓") +
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

		// Get current remaining time (server calculates this for running timers)
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

		// Get elapsed time - for running stopwatches, calculate locally for smooth UI
		var elapsedMs int64
		if status == "running" {
			// For running stopwatches, calculate elapsed time based on local clock
			// for smooth millisecond updates between server refreshes
			if startedAt, ok := sw["started_at_ms"]; ok && startedAt != nil {
				startedAtMs := int64(startedAt.(float64))
				// elapsed_ms is the time accumulated before this run started
				baseElapsed := int64(sw["elapsed_ms"].(float64))
				// Add time since start
				elapsedMs = baseElapsed + now.UnixMilli() - startedAtMs
			} else {
				// Fallback to server-calculated value
				if curr, ok := sw["current_elapsed_ms"]; ok {
					elapsedMs = int64(curr.(float64))
				} else {
					elapsedMs = int64(sw["elapsed_ms"].(float64))
				}
			}
		} else {
			// For stopped/paused stopwatches, use server value
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

// Commands

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

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

// formatDuration formats milliseconds as MM:SS
func formatDuration(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	seconds := ms / 1000
	minutes := seconds / 60
	seconds = seconds % 60
	hours := minutes / 60
	minutes = minutes % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// formatStopwatch formats milliseconds as stopwatch time with milliseconds
// Format: 0.000 -> 1:00.000 -> 1:00:00.000
func formatStopwatch(ms int64) string {
	if ms < 0 {
		ms = 0
	}

	hours := ms / (60 * 60 * 1000)
	remaining := ms % (60 * 60 * 1000)
	minutes := remaining / (60 * 1000)
	remaining = remaining % (60 * 1000)
	seconds := remaining / 1000
	milliseconds := remaining % 1000

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d:%02d.%03d", minutes, seconds, milliseconds)
	}
	return fmt.Sprintf("%d.%03d", seconds, milliseconds)
}

// Run starts the TUI
func Run(port int, citiesRepo *repository.CitiesRepository) error {
	p := tea.NewProgram(New(port, citiesRepo), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// renderProgressBar creates a liquid filling style progress bar
func renderProgressBar(progress float64, width int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}

	empty := width - filled

	// Liquid filling style: ▓ = dark filled (liquid), ░ = light empty (container)
	filledBar := ""
	emptyBar := ""
	for i := 0; i < filled; i++ {
		filledBar += "▓"
	}
	for i := 0; i < empty; i++ {
		emptyBar += "░"
	}

	return progressBarStyle.Render(filledBar) + progressBarEmptyStyle.Render(emptyBar)
}
