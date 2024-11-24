package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	// Base colors
	Text     = lipgloss.Color("#CDD6F4") // Text
	Subtext1 = lipgloss.Color("#BAC2DE") // Subtext1
	Overlay0 = lipgloss.Color("#6C7086") // Overlay0
	Mauve    = lipgloss.Color("#CBA6F7") // Mauve
	Green    = lipgloss.Color("#A6E3A1") // Green
	Red      = lipgloss.Color("#F38BA8") // Red
	Peach    = lipgloss.Color("#FAB387") // Peach
	Sky      = lipgloss.Color("#89B4FA") // Sky
	Lavender = lipgloss.Color("#B4BEFE") // Lavender
)

// Common styles used across the application
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Overlay0)

	// Box styles
	BoxStyle = BaseStyle.Copy().
		Padding(0, 1).
		Border(lipgloss.NormalBorder()).
		BorderForeground(Overlay0).
		Width(44)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Mauve).
		Padding(0, 1)

	// Label styles
	LabelStyle = lipgloss.NewStyle().
		Foreground(Subtext1).
		Width(10).
		Align(lipgloss.Left)

	// Value styles
	ValueStyle = lipgloss.NewStyle().
		Foreground(Text).
		Bold(true)

	// Empty state styles
	EmptyStyle = lipgloss.NewStyle().
		Foreground(Overlay0).
		Italic(true).
		Align(lipgloss.Center)

	// Prompt styles
	PromptStyle = lipgloss.NewStyle().
		Foreground(Lavender)

	// Info styles
	InfoStyle = lipgloss.NewStyle().
		Foreground(Subtext1)

	// Error styles
	ErrorStyle = lipgloss.NewStyle().
		Foreground(Red).
		Bold(true)

	// Container styles
	ContainerStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Overlay0).
		Padding(1).
		Align(lipgloss.Center)

	// Pair styles
	PairStyle = lipgloss.NewStyle().
		Foreground(Sky).
		Bold(true)

	// Confirm styles
	ConfirmStyle = lipgloss.NewStyle().
		Foreground(Sky).
		Bold(true)

	// Balance styles
	BalanceStyle = lipgloss.NewStyle().
		Foreground(Green).
		Bold(true)

	// PnL styles
	PnLPositiveStyle = lipgloss.NewStyle().
		Foreground(Green).
		Bold(true)

	PnLNegativeStyle = lipgloss.NewStyle().
		Foreground(Red).
		Bold(true)

	// Risk styles
	RiskStyle = lipgloss.NewStyle().
		Foreground(Peach).
		Bold(true)

	// Input styles
	InputStyle = lipgloss.NewStyle().
		Foreground(Text)
)
