// Package tui provides the Bubbletea TUI screens for Hefesto installer.
package tui

import (
	"runtime/debug"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ===== Fuego/Forge Theme Colors =====
// The aesthetic of the forge - fire, metal, and craftsmanship

var (
	// ColorAmber - Primary for titles, success, highlights
	ColorAmber = lipgloss.Color("#FF8C00")
	// Primary - Alias for ColorAmber (backward compatibility)
	Primary = ColorAmber
	// PrimaryDark - Alias for ColorAmber (backward compatibility)
	PrimaryDark = ColorAmber
	// ColorCopper - Borders, subtitles, accents
	ColorCopper = lipgloss.Color("#B87333")
	// Secondary - Alias for ColorCopper (backward compatibility)
	Secondary = ColorCopper
	// ColorDeepBlack - Main background
	ColorDeepBlack = lipgloss.Color("#1A1A1A")
	// Background - Alias for ColorDeepBlack (backward compatibility)
	Background = ColorDeepBlack
	// ColorSurface - Cards, boxes (slightly lighter)
	ColorSurface = lipgloss.Color("#2A2A2A")
	// Surface - Alias for ColorSurface (backward compatibility)
	Surface = ColorSurface
	// ColorRed - Errors
	ColorRed = lipgloss.Color("#FF4444")
	// Error - Alias for ColorRed (backward compatibility)
	Error = ColorRed
	// ColorGray - Muted text (adjusted from #666666 for readability)
	ColorGray = lipgloss.Color("#7C7C7C")
	// ColorMutedBorder - Subtle borders that don't compete with content
	ColorMutedBorder = lipgloss.Color("#3A3A3A")
	// ColorDimText - Very dim text for disabled/deemphasized content
	ColorDimText = lipgloss.Color("#4A4A4A")
	// ColorWhite - Normal text
	ColorWhite = lipgloss.Color("#FFFFFF")
	// ColorGreen - Success indicators
	ColorGreen = lipgloss.Color("#22C55E")
	// Success - Alias for ColorGreen (backward compatibility)
	Success = ColorGreen
	// ColorYellow - Warnings/pending
	ColorYellow = lipgloss.Color("#EAB308")
	// Warning - Alias for ColorYellow (backward compatibility)
	Warning = ColorYellow

	// Text - Alias for ColorWhite (backward compatibility)
	Text = ColorWhite
	// TextMuted - Alias for ColorGray (backward compatibility)
	TextMuted = ColorGray
	// TextBold - Alias for ColorWhite (backward compatibility)
	TextBold = ColorWhite
)

// ===== Base Styles =====

var (
	// TitleStyle - Main titles in amber, bold
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorAmber).
			Bold(true).
			Padding(0, 1)

	// SubtitleStyle - Subtitles with copper accent
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorCopper).
			Padding(0, 1)

	// BodyStyle - Regular body text in white
	BodyStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Padding(0, 1)

	// BoldStyle - Emphasized text
	BoldStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Bold(true)

	// MutedStyle - Dimmed/disabled text in gray
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	// HighlightStyle - Amber background, black text for emphasis
	HighlightStyle = lipgloss.NewStyle().
			Background(ColorAmber).
			Foreground(ColorDeepBlack).
			Bold(true).
			Padding(0, 1)
)

// ===== Semantic Styles =====

var (
	// SuccessStyle - Green with checkmark for success states
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true).
			SetString("✅ ")

	// ErrorStyle - Red with X for errors
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true).
			SetString("❌ ")

	// WarningStyle - Yellow for warnings
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorYellow).
			Bold(true)

	// InfoStyle - Amber for info
	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorAmber)
)

// ===== Box Styles =====

var (
	// BorderStyle - Copper border for boxes
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorCopper).
			Background(ColorDeepBlack).
			Padding(1, 2)

	// BoxStyle - Card/box with rounded border
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCopper).
			Background(ColorSurface).
			Padding(1, 2).
			Margin(1, 2)

	// BoxFocusedStyle - Focused box with amber border
	BoxFocusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAmber).
			Background(ColorSurface).
			Padding(1, 2).
			Margin(1, 2)
)

// ===== Typography Hierarchy =====
//
// These styles define the visual voice of the TUI. Use HeroTitleStyle for
// splash moments, SectionTitleStyle for screen headings, and DimTextStyle
// for content that should recede visually.

var (
	// HeroTitleStyle - Main hero title for splash/celebration screens (no padding)
	HeroTitleStyle = lipgloss.NewStyle().
			Foreground(ColorAmber).
			Bold(true)

	// SectionTitleStyle - Section headings within screens
	SectionTitleStyle = lipgloss.NewStyle().
				Foreground(ColorAmber).
				Bold(true)

	// DimTextStyle - Very dim text for disabled/deemphasized content
	DimTextStyle = lipgloss.NewStyle().
			Foreground(ColorDimText)

	// MutedBorderTextStyle - For subtle border decoration
	MutedBorderTextStyle = lipgloss.NewStyle().
				Foreground(ColorMutedBorder)
)

// ===== Progress Bar Styles =====

var (
	// ProgressContainerStyle - Container for progress bar
	ProgressContainerStyle = lipgloss.NewStyle().
				Foreground(ColorGray)

	// ProgressFilledStyle - Filled portion of progress bar in amber
	ProgressFilledStyle = lipgloss.NewStyle().
				Foreground(ColorAmber).
				Bold(true)

	// ProgressEmptyStyle - Empty portion of progress bar
	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorGray)
)

// ===== Status Icons =====

const (
	IconCheck   = "✓"
	IconCross   = "✗"
	IconSpinner = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
	IconArrow   = "→"
	IconBullet  = "•"
	IconFire    = "*"
)

// ===== ASCII Art Banner - The Anvil with Flames =====

// BannerAnvil is the ASCII art anvil with flame spark.
// Uses single-width characters only to avoid alignment issues in terminals.
const BannerAnvil = `
     *
    ╱│╲
   ╱ │ ╲
  ╱  │  ╲
 ╱___▼___╲
  ║███████║
  ║███████║
  ╰═══════╯`

// Logo - Alias for BannerAnvil (backward compatibility)
var Logo = BannerAnvil

// BannerStyle returns the banner styled with amber
func BannerStyle() string {
	return lipgloss.NewStyle().
		Foreground(ColorAmber).
		SetString(BannerAnvil).
		String()
}

// ===== Version Handling =====

// Version holds the current Hefesto version. It is initialized from build
// info at package load time, but should be overridden by the main package's
// ldflags-driven version via SetVersion for consistency.
var Version = initVersion()

// initVersion extracts version from build info or returns a default.
func initVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				if len(setting.Value) >= 7 {
					return "dev-" + setting.Value[:7]
				}
			}
		}
	}
	return "dev"
}

// SetVersion overrides the package-level Version. Call this from main after
// flag parsing so the TUI uses the same ldflags-driven version as the CLI.
func SetVersion(v string) {
	if v != "" {
		Version = v
	}
}

// VersionStyle returns styled version string
func VersionStyle() string {
	return lipgloss.NewStyle().
		Foreground(ColorAmber).
		Bold(true).
		Render("HEFESTO " + Version)
}

// TaglineStyle returns styled tagline
func TaglineStyle() string {
	return lipgloss.NewStyle().
		Foreground(ColorCopper).
		Render("AI Dev Environment Forge")
}

// ===== Utility Functions =====

// CenterText centers text within a given width
func CenterText(text string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(text)
}

// CenterBlock centers a block of text within a given width
func CenterBlock(text string, width int) string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, CenterText(line, width))
	}
	return strings.Join(result, "\n")
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

// BulletItem creates a bullet point item with amber bullet
func BulletItem(text string) string {
	bullet := lipgloss.NewStyle().Foreground(ColorAmber).Render(IconBullet)
	return "  " + bullet + "  " + text
}

// AmberText renders text in amber color
func AmberText(text string) string {
	return lipgloss.NewStyle().Foreground(ColorAmber).Render(text)
}

// CopperText renders text in copper color
func CopperText(text string) string {
	return lipgloss.NewStyle().Foreground(ColorCopper).Render(text)
}

// WhiteText renders text in white color
func WhiteText(text string) string {
	return lipgloss.NewStyle().Foreground(ColorWhite).Render(text)
}

// GrayText renders text in gray color
func GrayText(text string) string {
	return lipgloss.NewStyle().Foreground(ColorGray).Render(text)
}

// GreenText renders text in green color
func GreenText(text string) string {
	return lipgloss.NewStyle().Foreground(ColorGreen).Render(text)
}

// RedText renders text in red color
func RedText(text string) string {
	return lipgloss.NewStyle().Foreground(ColorRed).Render(text)
}

// stripAnsi removes ANSI escape codes from a string for length calculation.
// Used internally for visual-width calculations.
func stripAnsi(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
