// Package tui provides the Terminal User Interface using Bubble Tea.
package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightt77/veranda/internal/daemon"
)

// Model represents the TUI state
type Model struct {
	client      *daemon.Client
	timers      []map[string]interface{}
	stopwatches []map[string]interface{}
	activeTab   int // 0 = timers, 1 = stopwatches
	width       int
	height      int
	err         error
	lastUpdate  time.Time
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5F87")).
			MarginLeft(2)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888")).
				Padding(0, 2)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF5F87")).
			Padding(1, 2).
			Width(60)

	listStyle = lipgloss.NewStyle().
			MarginLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666")).
			MarginTop(1)

	runningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0f0"))

	pausedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0"))

	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888"))
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
func New(port int) Model {
	return Model{
		client:     daemon.NewClient(port),
		activeTab:  0,
		lastUpdate: time.Now(),
	}
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		refreshCmd(m.client),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "right":
			m.activeTab = (m.activeTab + 1) % 2
		case "shift+tab", "left":
			m.activeTab = (m.activeTab - 1 + 2) % 2
		case "r":
			return m, refreshCmd(m.client)
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

// View renders the UI
func (m Model) View() string {
	if !m.client.IsRunning() {
		return "\n  ⚠️  Daemon is not running.\n\n  Start it with: veranda daemon\n\n  Press 'q' to quit.\n"
	}

	var s string

	// Title
	s += titleStyle.Render("⏱️  Veranda") + "\n\n"

	// Tabs
	timersTab := inactiveTabStyle.Render("Timers [t]")
	stopwatchesTab := inactiveTabStyle.Render("Stopwatches [s]")

	if m.activeTab == 0 {
		timersTab = activeTabStyle.Render("Timers [t]")
	} else {
		stopwatchesTab = activeTabStyle.Render("Stopwatches [s]")
	}

	s += lipgloss.JoinHorizontal(lipgloss.Left, timersTab, stopwatchesTab) + "\n\n"

	// Content
	if m.err != nil {
		s += fmt.Sprintf("  Error: %v\n", m.err)
	} else if m.activeTab == 0 {
		s += m.renderTimers()
	} else {
		s += m.renderStopwatches()
	}

	// Help
	s += "\n" + helpStyle.Render("tab: switch • t: new timer • s: new stopwatch • r: refresh • q: quit")

	return s
}

// renderTimers renders the timer list
func (m Model) renderTimers() string {
	if len(m.timers) == 0 {
		return boxStyle.Render("  No timers yet.\n\n  Press 't' to create one.")
	}

	var content string
	for _, t := range m.timers {
		id := int64(t["id"].(float64))
		label := t["label"].(string)
		status := t["status"].(string)

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

		content += fmt.Sprintf("  %s %s %s %s\n",
			statusStyle.Render(statusIcon),
			fmt.Sprintf("%d", id),
			label,
			statusStyle.Render(status),
		)
	}

	return boxStyle.Render(content)
}

// renderStopwatches renders the stopwatch list
func (m Model) renderStopwatches() string {
	if len(m.stopwatches) == 0 {
		return boxStyle.Render("  No stopwatches yet.\n\n  Press 's' to create one.")
	}

	var content string
	for _, sw := range m.stopwatches {
		id := int64(sw["id"].(float64))
		label := sw["label"].(string)
		status := sw["status"].(string)
		elapsedMs := int64(sw["elapsed_ms"].(float64))

		statusIcon := "○"
		statusStyle := inactiveTabStyle

		if status == "running" {
			statusIcon = "▶"
			statusStyle = runningStyle
		}

		elapsedStr := formatDuration(elapsedMs)

		content += fmt.Sprintf("  %s %s %s %s %s\n",
			statusStyle.Render(statusIcon),
			fmt.Sprintf("%d", id),
			label,
			elapsedStr,
			statusStyle.Render(status),
		)
	}

	return boxStyle.Render(content)
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

// Run starts the TUI
func Run(port int) error {
	p := tea.NewProgram(New(port), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
