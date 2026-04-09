// Package tui provides the Terminal User Interface using Bubble Tea.
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightt77/veranda/internal/audio"
	"github.com/lightt77/veranda/internal/config"
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
	activeTab     int  // 0 = timers, 1 = stopwatches, 2 = ambient sounds, 3 = preferences
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

	// Preferences
	userConfig          *config.UserConfig
	configDir           string
	chimesDir           string
	ambienceDir         string
	settingsMessage     string
	settingsMessageTime time.Time
	selectingSoundFile  bool // dropdown mode for selecting chime sound file
	availableSoundFiles []string
	selectedSoundIdx    int

	// Chime volume editing
	editingChimeVolume bool

	// Ambient Sounds tab
	availableAmbientFiles []string
	ambientSelectedIdx    int
	ambientTestPlaying    map[string]bool   // tracks which sounds are being tested
	ambientStopFuncs      map[string]func() // stop functions for test playback
	editingAmbientVolume  bool

	// Audio player for test playback (initialized on first use)
	testAudioPlayer *audio.Player
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

	// Load user config
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, config.DefaultDataDir)
	chimesDir := filepath.Join(homeDir, config.DefaultChimesDir)
	ambienceDir := filepath.Join(homeDir, config.DefaultAmbienceDir)
	userConfig := config.LoadUserConfig(configDir)

	// Load available sound files
	availableSounds := loadAvailableSoundFiles(chimesDir)
	availableAmbience := loadAvailableSoundFiles(ambienceDir)

	// Initialize ambient sounds if empty (first run)
	if len(userConfig.AmbientSounds) == 0 && len(availableAmbience) > 0 {
		// Create config entries for all available files
		// First one enabled by default, rest disabled
		for i, filename := range availableAmbience {
			enabled := i == 0 // Only first one enabled
			userConfig.AmbientSounds = append(userConfig.AmbientSounds, config.AmbientSoundConfig{
				Filename: filename,
				Volume:   config.DefaultAmbientVolume,
				Enabled:  enabled,
			})
		}
		// Save the initialized config
		_ = userConfig.Save(configDir)
	}

	return Model{
		client:                daemon.NewClient(port),
		activeTab:             0,
		lastUpdate:            time.Now(),
		starfieldMode:         ModeRandom,
		randomStarfield:       starfield.NewRandomStarfield(),
		realisticStarfield:    starfield.NewRealisticStarfield(observer),
		citiesRepo:            citiesRepo,
		currentCity:           defaultCity,
		userConfig:            userConfig,
		configDir:             configDir,
		chimesDir:             chimesDir,
		ambienceDir:           ambienceDir,
		availableSoundFiles:   availableSounds,
		availableAmbientFiles: availableAmbience,
		ambientTestPlaying:    make(map[string]bool),
		ambientStopFuncs:      make(map[string]func()),
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

		// Handle ambient volume editing mode
		if m.activeTab == 2 && m.editingAmbientVolume {
			switch msg.String() {
			case "esc", "enter":
				m.editingAmbientVolume = false
			case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
				// Set volume to key value * 10%
				vol := float64(msg.String()[0]-'0') / 10.0
				if m.ambientSelectedIdx < len(m.userConfig.AmbientSounds) {
					m.userConfig.AmbientSounds[m.ambientSelectedIdx].Volume = vol
				}
			case "up", "k":
				if m.ambientSelectedIdx < len(m.userConfig.AmbientSounds) {
					newVol := m.userConfig.AmbientSounds[m.ambientSelectedIdx].Volume + 0.1
					if newVol > 1.0 {
						newVol = 1.0
					}
					m.userConfig.AmbientSounds[m.ambientSelectedIdx].Volume = newVol
				}
			case "down", "j":
				if m.ambientSelectedIdx < len(m.userConfig.AmbientSounds) {
					newVol := m.userConfig.AmbientSounds[m.ambientSelectedIdx].Volume - 0.1
					if newVol < 0.0 {
						newVol = 0.0
					}
					m.userConfig.AmbientSounds[m.ambientSelectedIdx].Volume = newVol
				}
			}
			return m, nil
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
			m.stopAllTestPlayback()
			return m, tea.Quit
		case "tab", "right":
			m.activeTab = (m.activeTab + 1) % 4
			m.selectedIdx = 0
			m.selectingSoundFile = false
			m.editingAmbientVolume = false
			m.editingChimeVolume = false
		case "shift+tab", "left":
			m.activeTab = (m.activeTab - 1 + 4) % 4
			m.selectedIdx = 0
			m.selectingSoundFile = false
			m.editingAmbientVolume = false
			m.editingChimeVolume = false
		case "up", "k":
			if m.activeTab == 2 {
				if m.ambientSelectedIdx > 0 {
					m.ambientSelectedIdx--
				}
			} else {
				m.selectedIdx = max(0, m.selectedIdx-1)
			}
		case "down", "j":
			if m.activeTab == 2 {
				maxIdx := len(m.userConfig.AmbientSounds) - 1
				m.ambientSelectedIdx = min(maxIdx, m.ambientSelectedIdx+1)
			} else {
				maxIdx := len(m.timers) - 1
				if m.activeTab == 1 {
					maxIdx = len(m.stopwatches) - 1
				}
				m.selectedIdx = min(maxIdx, m.selectedIdx+1)
			}
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
			if m.activeTab == 2 {
				// Reset ambient sounds to defaults
				m.resetAmbientSounds()
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
			} else if m.activeTab == 2 && m.ambientSelectedIdx < len(m.userConfig.AmbientSounds) {
				// Test toggle for ambient sound
				m.toggleAmbientTest()
			}
		case "s":
			if m.activeTab == 1 {
				// Quick stopwatch creation
				return m, createStopwatchCmd(m.client)
			} else if m.activeTab == 2 {
				// Save ambient sound settings
				m.saveAmbientSettings()
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
		case "e":
			// Edit timer completion sound file (only in preferences tab)
			if m.activeTab == 3 && !m.selectingSoundFile {
				m.selectingSoundFile = true
				// Find current sound file index
				m.selectedSoundIdx = 0
				for i, f := range m.availableSoundFiles {
					if f == m.userConfig.TimerCompletionSound {
						m.selectedSoundIdx = i
						break
					}
				}
			}
		case "space":
			// Toggle enable/disable ambient sound
			if m.activeTab == 2 && m.ambientSelectedIdx < len(m.userConfig.AmbientSounds) {
				m.userConfig.AmbientSounds[m.ambientSelectedIdx].Enabled =
					!m.userConfig.AmbientSounds[m.ambientSelectedIdx].Enabled
			}
		case "v":
			// Edit volume
			if m.activeTab == 2 && m.ambientSelectedIdx < len(m.userConfig.AmbientSounds) {
				m.editingAmbientVolume = true
			} else if m.activeTab == 3 {
				m.editingChimeVolume = true
			}
		}

		// Handle dropdown navigation when selecting sound file
		if m.selectingSoundFile {
			switch msg.String() {
			case "esc":
				m.selectingSoundFile = false
			case "enter":
				// Save the selected sound file
				if m.selectedSoundIdx < len(m.availableSoundFiles) {
					m.userConfig.TimerCompletionSound = m.availableSoundFiles[m.selectedSoundIdx]
					if err := m.userConfig.Save(m.configDir); err != nil {
						m.setSettingsMessage(fmt.Sprintf("Error saving: %v", err))
					} else {
						m.setSettingsMessage("Settings saved!")
					}
				}
				m.selectingSoundFile = false
			case "up", "k":
				if m.selectedSoundIdx > 0 {
					m.selectedSoundIdx--
				}
			case "down", "j":
				if m.selectedSoundIdx < len(m.availableSoundFiles)-1 {
					m.selectedSoundIdx++
				}
			}
		}

	case tickMsg:
		m.lastUpdate = time.Time(msg)
		// Clear settings message after 3 seconds
		if m.settingsMessage != "" && time.Since(m.settingsMessageTime) > 3*time.Second {
			m.settingsMessage = ""
		}
		// Sync test playback state with actual audio player state
		// This handles the case where a test finished naturally
		if m.testAudioPlayer != nil {
			for filename := range m.ambientTestPlaying {
				if !m.testAudioPlayer.IsTestPlaying(filename) {
					// Test finished naturally, clean up
					delete(m.ambientTestPlaying, filename)
					delete(m.ambientStopFuncs, filename)
				}
			}
		}
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

// setSettingsMessage sets a temporary settings message
func (m *Model) setSettingsMessage(msg string) {
	m.settingsMessage = msg
	m.settingsMessageTime = time.Now()
}

// toggleAmbientTest toggles test playback for the selected ambient sound
func (m *Model) toggleAmbientTest() {
	if m.testAudioPlayer == nil {
		// Initialize test audio player
		cfg := config.Config()
		player, err := audio.NewPlayer(cfg, m.configDir)
		if err != nil {
			m.setSettingsMessage(fmt.Sprintf("Audio error: %v", err))
			return
		}
		m.testAudioPlayer = player
	}

	sound := m.userConfig.AmbientSounds[m.ambientSelectedIdx]

	// Check actual playback state (in case it finished naturally)
	isActuallyPlaying := m.testAudioPlayer.IsTestPlaying(sound.Filename)
	isMarkedPlaying := m.ambientTestPlaying[sound.Filename]

	// If marked as playing but actually stopped (finished naturally), clean up
	if isMarkedPlaying && !isActuallyPlaying {
		m.ambientTestPlaying[sound.Filename] = false
		delete(m.ambientStopFuncs, sound.Filename)
		isMarkedPlaying = false
	}

	// Toggle based on current state
	if isMarkedPlaying {
		// Stop it
		if stopFunc, ok := m.ambientStopFuncs[sound.Filename]; ok && stopFunc != nil {
			stopFunc()
		}
		m.testAudioPlayer.StopTestPlay(sound.Filename) // Ensure it's stopped
		m.ambientTestPlaying[sound.Filename] = false
		delete(m.ambientStopFuncs, sound.Filename)
	} else {
		// Start it
		stopFunc, err := m.testAudioPlayer.TestPlay(sound.Filename, sound.Volume)
		if err != nil {
			m.setSettingsMessage(fmt.Sprintf("Play error: %v", err))
			return
		}
		m.ambientTestPlaying[sound.Filename] = true
		m.ambientStopFuncs[sound.Filename] = stopFunc
	}
}

// stopAllTestPlayback stops all ambient test playback
func (m *Model) stopAllTestPlayback() {
	if m.testAudioPlayer != nil {
		m.testAudioPlayer.StopAllTestPlays()
	}
	m.ambientTestPlaying = make(map[string]bool)
	m.ambientStopFuncs = make(map[string]func())
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

			symbol := starfield.GetObjectSymbol(obj.Object, obj.Brightness, obj.TwinklePhase)
			twinkleBrightness := starfield.GetTwinkleBrightness(obj.Brightness, obj.TwinklePhase, obj.Object.Type)
			color := starfield.GetObjectColor(obj.Object, twinkleBrightness)

			if len(symbol) > 0 {
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
					{obj.ScreenX + 3, obj.ScreenY},
					{obj.ScreenX - nameLen - 2, obj.ScreenY},
					{obj.ScreenX - nameLen/2, obj.ScreenY + 2},
					{obj.ScreenX - nameLen/2, obj.ScreenY - 2},
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
								grid[nameY][nameX+colOffset] = Cell{
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

	// Tabs - now 4 tabs
	timersTab := inactiveTabStyle.Render("Timers [t]")
	stopwatchesTab := inactiveTabStyle.Render("Stopwatches [s]")
	ambientTab := inactiveTabStyle.Render("Ambient [a]")
	prefsTab := inactiveTabStyle.Render("Prefs [e]")

	switch m.activeTab {
	case 0:
		timersTab = activeTabStyle.Render("Timers [t]")
	case 1:
		stopwatchesTab = activeTabStyle.Render("Stopwatches [s]")
	case 2:
		ambientTab = activeTabStyle.Render("Ambient [a]")
	case 3:
		prefsTab = activeTabStyle.Render("Prefs [e]")
	}

	s += leftPad + lipgloss.JoinHorizontal(lipgloss.Left, timersTab, stopwatchesTab, ambientTab, prefsTab) + "\n"

	// Content
	if m.err != nil {
		s += leftPad + fmt.Sprintf("Error: %v\n", m.err)
	} else {
		switch m.activeTab {
		case 0:
			s += leftPad + strings.ReplaceAll(m.renderTimers(), "\n", "\n"+leftPad)
		case 1:
			s += leftPad + strings.ReplaceAll(m.renderStopwatches(), "\n", "\n"+leftPad)
		case 2:
			s += leftPad + strings.ReplaceAll(m.renderAmbientSounds(), "\n", "\n"+leftPad)
		case 3:
			s += leftPad + strings.ReplaceAll(m.renderPreferences(), "\n", "\n"+leftPad)
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
	case 2:
		// Ambient Sounds tab
		if m.editingAmbientVolume {
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
				lipgloss.NewStyle().Foreground(mauve).Render("s") +
				lipgloss.NewStyle().Foreground(overlay0).Render(":save ") +
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

	// Sort ambient sounds by filename for consistent display
	sortedSounds := make([]config.AmbientSoundConfig, len(m.userConfig.AmbientSounds))
	copy(sortedSounds, m.userConfig.AmbientSounds)
	sort.Slice(sortedSounds, func(i, j int) bool {
		return sortedSounds[i].Filename < sortedSounds[j].Filename
	})

	// Find the index of the currently selected sound in sorted list
	selectedFilename := ""
	if m.ambientSelectedIdx < len(m.userConfig.AmbientSounds) {
		selectedFilename = m.userConfig.AmbientSounds[m.ambientSelectedIdx].Filename
	}
	selectedSortedIdx := 0
	for i, s := range sortedSounds {
		if s.Filename == selectedFilename {
			selectedSortedIdx = i
			break
		}
	}

	// Display list
	for i, sound := range sortedSounds {
		// Find the original index for this sorted sound
		originalIdx := -1
		for j, orig := range m.userConfig.AmbientSounds {
			if orig.Filename == sound.Filename {
				originalIdx = j
				break
			}
		}

		// Selection indicator
		prefix := "  "
		if i == selectedSortedIdx {
			prefix = selectedStyle.Render("> ")
		}

		// Checkbox
		checkbox := "[ ]"
		if sound.Enabled {
			checkbox = "[" + lipgloss.NewStyle().Foreground(green).Render("✓") + "]"
		}

		// Filename
		filename := sound.Filename
		if i == selectedSortedIdx {
			filename = lipgloss.NewStyle().Foreground(lavender).Bold(true).Render(filename)
		} else {
			filename = lipgloss.NewStyle().Foreground(text).Render(filename)
		}

		// Volume bar
		volBar := renderVolumeBar(sound.Volume, 10)
		volPercent := int(sound.Volume * 100)
		volStr := fmt.Sprintf("%3d%%", volPercent)
		if originalIdx == m.ambientSelectedIdx && m.editingAmbientVolume {
			volStr = lipgloss.NewStyle().Foreground(yellow).Render(volStr)
		}

		// Test indicator - check both our state and actual player state
		testIndicator := "  "
		isActuallyPlaying := m.testAudioPlayer != nil && m.testAudioPlayer.IsTestPlaying(sound.Filename)
		isMarkedPlaying := m.ambientTestPlaying[sound.Filename]
		if isActuallyPlaying || isMarkedPlaying {
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

// renderVolumeBar creates a visual volume bar
func renderVolumeBar(volume float64, width int) string {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	filled := int(volume * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	filledBar := strings.Repeat("█", filled)
	emptyBar := strings.Repeat("░", empty)

	return progressBarStyle.Render(filledBar) + progressBarEmptyStyle.Render(emptyBar)
}

// renderPreferences renders the preferences/settings page
func (m Model) renderPreferences() string {
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

// loadAvailableSoundFiles scans the sounds directory and returns a list of MP3 files
func loadAvailableSoundFiles(soundsDir string) []string {
	entries, err := os.ReadDir(soundsDir)
	if err != nil {
		return []string{}
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".mp3") {
			files = append(files, name)
		}
	}

	// Sort files for consistent ordering
	sort.Strings(files)

	return files
}
