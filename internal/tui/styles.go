package tui

import "github.com/charmbracelet/lipgloss"

// Style definitions for the TUI
var (
	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D7FF")).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(0, 1)

	// Search prompt style
	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00"))

	// Input field style
	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	// Matched text style
	matchStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF005F")).
			Underline(true)

	// File path style
	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF")).
			Bold(true)

	// Line number style
	lineNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF87"))

	// Line content style
	lineContentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	// Selected item style
	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D7FF")).
			Background(lipgloss.Color("#1A1A1A"))

	// Status bar style
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(0, 1)

	// Help style
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999"))

	// Quit message style
	quitStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5F00"))
)
