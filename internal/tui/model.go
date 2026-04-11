// Package tui provides the Terminal User Interface using Bubble Tea.
package tui

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

// ContentAreaHeight is the fixed height of the content viewport at the bottom
const ContentAreaHeight = 10

// ScrollableContentHeight is the height of the scrollable content area (ContentAreaHeight - tabs - status - help - cmd)
const ScrollableContentHeight = 6

// StatusMessageTimeout is how long status messages remain visible
const StatusMessageTimeout = 3 * time.Second

// statusMessage represents a queued status message
type statusMessage struct {
	text      string
	timestamp time.Time
}

// Model represents the TUI state
type Model struct {
	client      *daemon.Client
	timers      []map[string]interface{}
	stopwatches []map[string]interface{}
	activeTab   int // 0 = timers/stopwatches, 1 = ambience, 2 = skyfield, 3 = prefs
	selectedIdx int // currently selected item index (for prefs tab)

	// Content viewport scroll position (resets on tab switch)
	contentScrollOffset int

	// Status message queue for L8
	statusMessageQueue []statusMessage
	currentStatusMsg   string
	statusMsgExpiry    time.Time

	// Debug mode for diagnostics
	debugMode bool

	// Timers/Stopwatches tab state
	showTimers            bool // 't' key - show timers section
	showStopwatches       bool // 's' key - show stopwatches section
	showCompleted         bool // 'h' key - show completed/archived items
	timerSelectedIdx      int  // selection within visible timers
	stopwatchSelectedIdx  int  // selection within visible stopwatches
	activeColumn          int  // 0 = timers, 1 = stopwatches (when both visible)
	timerScrollOffset     int  // scroll position for timers
	stopwatchScrollOffset int  // scroll position for stopwatches

	// Command mode for creating timers/stopwatches
	commandMode  bool   // '/' enters command mode
	commandInput string // current command being typed
	commandError string // error message to display
	width        int
	height       int
	err          error
	lastUpdate   time.Time

	// Skyfield tab state
	skyfieldSelectedIdx int // 0 = mode, 1 = names, 2 = location

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
	settingsRepo        config.SettingsRepository
	configDir           string
	chimesDir           string
	ambienceDir         string
	selectingSoundFile  bool // dropdown mode for selecting chime sound file
	availableSoundFiles []string
	selectedSoundIdx    int

	// Chime volume editing
	editingChimeVolume bool

	// Ambience tab
	availableAmbienceFiles []string
	ambienceSelectedIdx    int   // index in ambienceDisplayOrder
	ambienceDisplayOrder   []int // maps display index -> config index (sorted by filename)
	editingAmbienceVolume  bool

	// Audio player for test playback (initialized on first use)
	testAudioPlayer *audio.Player
}

// twinkleMsg is sent to update star twinkling
type twinkleMsg time.Time

// starfieldUpdateMsg is sent to update realistic starfield positions
type starfieldUpdateMsg time.Time

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

// testStartedMsg is sent when ambience test playback starts
type testStartedMsg struct {
	filename string
}

// testStoppedMsg is sent when ambience test playback stops
type testStoppedMsg struct {
	filename string
}

// New creates a new TUI model
func New(port int, citiesRepo *repository.CitiesRepository, settingsRepo config.SettingsRepository) Model {
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

	// Load user config from database
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, config.DefaultDataDir)
	chimesDir := filepath.Join(homeDir, config.DefaultChimesDir)
	ambienceDir := filepath.Join(homeDir, config.DefaultAmbienceDir)
	userConfig := config.LoadUserConfig(settingsRepo)

	// Load available sound files
	availableSounds := loadAvailableSoundFiles(chimesDir)
	availableAmbience := loadAvailableSoundFiles(ambienceDir)

	// Initialize ambience sounds if empty (first run)
	if len(userConfig.AmbienceSounds) == 0 && len(availableAmbience) > 0 {
		// Create config entries for all available files
		// First one enabled by default, rest disabled
		for i, filename := range availableAmbience {
			enabled := i == 0 // Only first one enabled
			userConfig.AmbienceSounds = append(userConfig.AmbienceSounds, config.AmbienceSoundConfig{
				Filename: filename,
				Volume:   config.DefaultAmbienceVolume,
				Enabled:  enabled,
			})
		}
		// Save the initialized config
		_ = userConfig.Save(settingsRepo)
	}

	// Initialize display order (sorted by filename)
	displayOrder := initAmbienceDisplayOrder(userConfig.AmbienceSounds)

	return Model{
		client:                 daemon.NewClient(port),
		activeTab:              0,
		lastUpdate:             time.Now(),
		starfieldMode:          ModeRandom,
		randomStarfield:        starfield.NewRandomStarfield(),
		realisticStarfield:     starfield.NewRealisticStarfield(observer),
		citiesRepo:             citiesRepo,
		currentCity:            defaultCity,
		userConfig:             userConfig,
		settingsRepo:           settingsRepo,
		configDir:              configDir,
		chimesDir:              chimesDir,
		ambienceDir:            ambienceDir,
		availableSoundFiles:    availableSounds,
		availableAmbienceFiles: availableAmbience,
		ambienceDisplayOrder:   displayOrder,
		ambienceSelectedIdx:    0,
		// Timers/Stopwatches tab defaults
		showTimers:            true,
		showStopwatches:       true,
		showCompleted:         false,
		timerSelectedIdx:      0,
		stopwatchSelectedIdx:  0,
		activeColumn:          0,
		timerScrollOffset:     0,
		stopwatchScrollOffset: 0,
		// Debug mode - disabled for production
		debugMode: false,
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

// initAmbienceDisplayOrder creates a sorted mapping from display index to config index
func initAmbienceDisplayOrder(sounds []config.AmbienceSoundConfig) []int {
	order := make([]int, len(sounds))
	for i := range sounds {
		order[i] = i
	}
	// Sort by filename
	sort.Slice(order, func(i, j int) bool {
		return sounds[order[i]].Filename < sounds[order[j]].Filename
	})
	return order
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

// loadAvailableSoundFiles loads MP3 files from a directory
func loadAvailableSoundFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) > 4 && name[len(name)-4:] == ".mp3" {
			files = append(files, name)
		}
	}
	return files
}
