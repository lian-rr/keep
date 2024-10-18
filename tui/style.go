package tui

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor = "#F7FAF7"
	borderColor  = "#5f87ff"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Bold(true).
			Foreground(lipgloss.Color(primaryColor)).
			MarginTop(1).
			PaddingBottom(1)

	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(borderColor))

	helpStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			Padding(0, 2, 0)

	containerStyle = lipgloss.NewStyle().
			Padding(0, 2, 0)
)
