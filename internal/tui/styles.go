// Package tui provides the Terminal User Interface using Bubble Tea.
package tui

import "github.com/charmbracelet/lipgloss"

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

// Styles used throughout the TUI
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
