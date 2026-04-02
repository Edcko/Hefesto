// Package tui provides the Bubbletea TUI screens for Hefesto installer.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Fuego/Forge theme colors - the aesthetic of the forge
var (
	// Primary - Amber/Orange tones for the forge fire
	Primary     = lipgloss.Color("#FFA500") // Amber - titles, highlights
	PrimaryDark = lipgloss.Color("#FF8C00") // Dark Orange - borders, accents
	Secondary   = lipgloss.Color("#B87333") // Copper - borders, muted text

	// Backgrounds - Dark forge atmosphere
	Background = lipgloss.Color("#1A1A1A") // Main background
	Surface    = lipgloss.Color("#2A2A2A") // Cards, boxes

	// Semantic colors
	Success = lipgloss.Color("#4CAF50") // Checkmarks, done
	Error   = lipgloss.Color("#FF5252") // Errors, warnings
	Warning = lipgloss.Color("#FFC107") // Warnings

	// Text
	Text      = lipgloss.Color("#E0E0E0") // Body text
	TextMuted = lipgloss.Color("#888888") // Muted/disabled text
	TextBold  = lipgloss.Color("#FFFFFF") // Bold text
)

// Base styles
var (
	// TitleStyle - Main titles in amber
	TitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Padding(0, 1)

	// SubtitleStyle - Subtitles with copper accent
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Padding(0, 1)

	// BodyStyle - Regular body text
	BodyStyle = lipgloss.NewStyle().
			Foreground(Text).
			Padding(0, 1)

	// BoldStyle - Emphasized text
	BoldStyle = lipgloss.NewStyle().
			Foreground(TextBold).
			Bold(true)

	// MutedStyle - Dimmed/disabled text
	MutedStyle = lipgloss.NewStyle().
			Foreground(TextMuted)
)

// Box styles
var (
	// BoxStyle - Card/box with border
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Secondary).
			Background(Surface).
			Padding(1, 2).
			Margin(1, 2)

	// BoxFocusedStyle - Focused box with amber border
	BoxFocusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Background(Surface).
			Padding(1, 2).
			Margin(1, 2)
)

// Semantic styles
var (
	// SuccessStyle - Green for success states
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	// ErrorStyle - Red for errors
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	// WarningStyle - Yellow for warnings
	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	// InfoStyle - Amber for info
	InfoStyle = lipgloss.NewStyle().
			Foreground(Primary)
)

// Progress bar styles
var (
	// ProgressContainerStyle - Container for progress bar
	ProgressContainerStyle = lipgloss.NewStyle().
				Foreground(TextMuted)

	// ProgressFilledStyle - Filled portion of progress bar
	ProgressFilledStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Bold(true)

	// ProgressEmptyStyle - Empty portion of progress bar
	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(TextMuted)
)

// Checkmark and status icons
const (
	IconCheck   = "✓"
	IconCross   = "✗"
	IconSpinner = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
	IconArrow   = "→"
	IconBullet  = "•"
)

// Logo - ASCII art anvil/forge symbol for Hefesto
var Logo = `
    ╔═══════════════════════════════════╗
    ║                                   ║
    ║     ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄      ║
    ║    █░░░░░░░░░░░░░░░░░░░░░░░█     ║
    ║    █░░░░░░░░░░░░░░░░░░░░░░░█     ║
    ║    █░░░░░░░░░░░░░░░░░░░░░░░█     ║
    ║    ████████▀▀▀▀▀▀▀▀████████      ║
    ║    ░░░░░░░░        ░░░░░░░░      ║
    ║   ░░░░░░░░░░      ░░░░░░░░░░     ║
    ║  ░░░░░░░░░░░░    ░░░░░░░░░░░░    ║
    ║ ░░░░░░░░░░░░░░  ░░░░░░░░░░░░░░   ║
    ║                                   ║
    ╚═══════════════════════════════════╝
`

// LogoStyle returns the logo styled with amber gradient
func LogoStyle() string {
	return lipgloss.NewStyle().
		Foreground(Primary).
		SetString(Logo).
		String()
}

// Version display
const Version = "v1.0.0"

// VersionStyle returns styled version string
func VersionStyle() string {
	return MutedStyle.Render("Hefesto " + Version)
}

// CenterText centers text within a given width
func CenterText(text string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(text)
}

// ProgressBar renders a progress bar
func ProgressBar(width int, filled, total int) string {
	if total == 0 {
		return ""
	}

	percent := float64(filled) / float64(total)
	filledWidth := int(float64(width) * percent)

	filledBar := ProgressFilledStyle.Render(repeatStr("█", filledWidth))
	emptyBar := ProgressEmptyStyle.Render(repeatStr("░", width-filledWidth))

	return filledBar + emptyBar
}

// repeatStr repeats a string n times
func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
