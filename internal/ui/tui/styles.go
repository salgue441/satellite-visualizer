package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all lipgloss styles for the dashboard.
type Styles struct {
	// Panel borders
	BorderActive   lipgloss.Style
	BorderInactive lipgloss.Style

	// Title styles
	Title    lipgloss.Style
	Subtitle lipgloss.Style

	// List styles
	Selected   lipgloss.Style
	Unselected lipgloss.Style

	// Status bar
	StatusBar   lipgloss.Style
	StatusLabel lipgloss.Style
	StatusValue lipgloss.Style
	StatusWarn  lipgloss.Style

	// Help
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style
}

// DefaultStyles returns the default dashboard styles.
func DefaultStyles() Styles {
	return Styles{
		BorderActive: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")), // cyan
		BorderInactive: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")), // gray
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginLeft(1),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).          // bright yellow text
			Background(lipgloss.Color("236")),          // dark gray background
		Unselected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			MarginLeft(1),
		StatusLabel: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")),
		StatusValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		StatusWarn: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")), // red
		HelpKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")),
		HelpDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
	}
}
