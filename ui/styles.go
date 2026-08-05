package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles defines the color palette and formatting for the TUI
type Styles struct {
	NormalText     lipgloss.Style
	CorrectText    lipgloss.Style
	WrongText      lipgloss.Style
	Unset          lipgloss.Style
	CursorText     lipgloss.Style
	Stats          lipgloss.Style
	ActiveOption   lipgloss.Style
	InactiveOption lipgloss.Style
	Separator      lipgloss.Style
	TimerText      lipgloss.Style
	ResultLabel    lipgloss.Style
	ResultValue    lipgloss.Style
	BigWPM         lipgloss.Style
	ChartBar       lipgloss.Style
}

// DefaultStyles provides a monkeytype-inspired theme
func DefaultStyles() Styles {
	return Styles{
		NormalText:     lipgloss.NewStyle().Foreground(lipgloss.Color("#737373")),                                       // dim grey
		CorrectText:    lipgloss.NewStyle().Foreground(lipgloss.Color("#d1d0c5")),                                       // bright text
		WrongText:      lipgloss.NewStyle().Foreground(lipgloss.Color("#ca4754")),                                       // subtle red
		Unset:          lipgloss.NewStyle().Foreground(lipgloss.Color("#737373")),                                       // dim grey
		CursorText:     lipgloss.NewStyle().Background(lipgloss.Color("#d1d0c5")).Foreground(lipgloss.Color("#323437")), // cursor
		Stats:          lipgloss.NewStyle().Foreground(lipgloss.Color("#e2b714")).Bold(true),                             // monkeytype yellow
		ActiveOption:   lipgloss.NewStyle().Foreground(lipgloss.Color("#e2b714")),                                       // yellow for selected
		InactiveOption: lipgloss.NewStyle().Foreground(lipgloss.Color("#737373")),                                       // dim for unselected
		Separator:      lipgloss.NewStyle().Foreground(lipgloss.Color("#737373")),                                       // dim separator
		TimerText:      lipgloss.NewStyle().Foreground(lipgloss.Color("#e2b714")).Bold(true),                             // yellow countdown
		ResultLabel:    lipgloss.NewStyle().Foreground(lipgloss.Color("#737373")),                                       // dim labels
		ResultValue:    lipgloss.NewStyle().Foreground(lipgloss.Color("#d1d0c5")),                                       // bright values
		BigWPM:         lipgloss.NewStyle().Foreground(lipgloss.Color("#e2b714")).Bold(true),                             // big WPM number
		ChartBar:       lipgloss.NewStyle().Foreground(lipgloss.Color("#e2b714")),                                       // sparkline
	}
}