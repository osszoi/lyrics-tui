package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Catppuccin Mocha colors
	mauve    = lipgloss.Color("#cba6f7")
	blue     = lipgloss.Color("#89b4fa")
	lavender = lipgloss.Color("#b4befe")
	green    = lipgloss.Color("#a6e3a1")
	red      = lipgloss.Color("#f38ba8")
	overlay0 = lipgloss.Color("#6c7086")
	yellow   = lipgloss.Color("#f9e2af")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(mauve)

	infoStyle = lipgloss.NewStyle().
			Foreground(blue)

	errorStyle = lipgloss.NewStyle().
			Foreground(red)

	helpStyle = lipgloss.NewStyle().
			Foreground(overlay0)

	activeStyle = lipgloss.NewStyle().
			Foreground(green)

	warningStyle = lipgloss.NewStyle().
			Foreground(yellow)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lavender).
			Padding(1, 2)
)
