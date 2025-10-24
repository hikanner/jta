package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color definitions
var (
	ColorPrimary   = lipgloss.Color("#00D9FF") // Cyan
	ColorSuccess   = lipgloss.Color("#00FF87") // Green
	ColorWarning   = lipgloss.Color("#FFD700") // Yellow
	ColorError     = lipgloss.Color("#FF5F87") // Red
	ColorInfo      = lipgloss.Color("#8787FF") // Blue
	ColorSubtle    = lipgloss.Color("#626262") // Gray
	ColorHighlight = lipgloss.Color("#FF87D7") // Pink
)

// Style definitions
var (
	// HeaderStyle is for section headers
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginTop(1).
			MarginBottom(1)

	// SuccessStyle is for success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// ErrorStyle is for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// WarningStyle is for warning messages
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	// InfoStyle is for informational messages
	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo)

	// SubtleStyle is for subtle/dimmed text
	SubtleStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle)

	// HighlightStyle is for highlighted text
	HighlightStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true)

	// BoxStyle is for bordered boxes
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	// StatsStyle is for statistics display
	StatsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorInfo).
			Padding(0, 2).
			MarginTop(1)

	// LabelStyle is for labels
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle)

	// ValueStyle is for values
	ValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)
)

// Icons for various states
const (
	IconSuccess = "✓"
	IconError   = "✗"
	IconWarning = "⚠"
	IconInfo    = "ℹ"
	IconLoading = "⏳"
	IconRocket  = "🚀"
	IconBook    = "📚"
	IconRobot   = "🤖"
	IconMagnify = "🔍"
	IconSparkle = "✨"
	IconFile    = "📄"
	IconSave    = "💾"
	IconChart   = "📊"
	IconTarget  = "🎯"
	IconCheck   = "☑"
	IconCross   = "☒"
)
