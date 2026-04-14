// Package tui provides shared rendering primitives for the Hefesto TUI screens.
//
// frame.go contains the unified visual system: layout constants, border
// helpers, and reusable rendering functions that replace ad-hoc
// border/centering/padding code across screens.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ===== Layout Constants =====
//
// These constants define the visual grid all screens share. Every screen
// should resolve its effective width via ResolveContentWidth rather than
// hard-coding magic numbers.

const (
	// ContentWidth is the standard content width for all screens.
	ContentWidth = 60

	// MaxContentWidth caps content for wide terminals.
	MaxContentWidth = 80

	// MinContentWidth is the narrowest usable layout.
	MinContentWidth = 40
)

// Spacing rhythm — consistent vertical and horizontal gaps.
const (
	SpaceXS = 1 // Tight: between related items
	SpaceSM = 2 // Standard: between sections
	SpaceMD = 3 // Comfortable: major breaks
	SpaceLG = 4 // Generous: before/after hero
	SpaceXL = 6 // Breathing room: splash screens
)

// Standard padding values used across screens.
const (
	PadScreen = 1 // Padding between screen edge and content
	PadBox    = 2 // Padding inside bordered boxes
	PadItem   = 1 // Padding around individual items
)

// ===== Border Types =====

// BorderType defines the visual border style for a screen frame.
type BorderType int

const (
	// BorderNone renders no border — used for splash/celebration screens.
	BorderNone BorderType = iota
	// BorderRounded renders lightweight rounded corners — for interactive containers.
	BorderRounded
)

// ===== Rendering Option Types =====

// FrameOptions configures how RenderScreenFrame wraps content.
type FrameOptions struct {
	Width  int        // Terminal width (0 → ContentWidth)
	Height int        // Terminal height (0 → no vertical centering)
	Border BorderType // Border style to use
}

// KeyHint represents a keyboard shortcut for the help bar.
type KeyHint struct {
	Key    string // e.g. "Enter", "↑↓"
	Action string // e.g. "Continue"
}

// WizardStep represents one step in a multi-step wizard flow.
type WizardStep struct {
	Label  string
	Done   bool
	Active bool
}

// ===== Core Rendering Primitives =====

// ResolveContentWidth returns the effective content width for a given
// terminal width, clamped between MinContentWidth and MaxContentWidth.
func ResolveContentWidth(termWidth int) int {
	if termWidth <= 0 {
		return ContentWidth
	}
	if termWidth < MinContentWidth {
		return MinContentWidth
	}
	if termWidth > MaxContentWidth {
		return MaxContentWidth
	}
	return termWidth
}

// RenderCenteredHero renders borderless centered content (banner, title,
// subtitle) for splash screens like welcome and complete. Each field is
// optional — pass "" to skip.
func RenderCenteredHero(banner, title, subtitle string, width int) string {
	var b strings.Builder

	if banner != "" {
		b.WriteString(CenterText(banner, width))
		b.WriteString("\n")
	}

	if title != "" {
		b.WriteString(CenterText(HeroTitleStyle.Render(title), width))
		b.WriteString("\n")
	}

	if subtitle != "" {
		b.WriteString(CenterText(CopperText(subtitle), width))
		b.WriteString("\n")
	}

	return b.String()
}

// RenderScreenFrame wraps content in a standard frame with optional border
// and centers it in the terminal.
func RenderScreenFrame(content string, opts FrameOptions) string {
	termWidth := opts.Width
	if termWidth == 0 {
		termWidth = ContentWidth
	}
	termHeight := opts.Height

	// Apply border if specified.
	var framed string
	switch opts.Border {
	case BorderRounded:
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCopper).
			Padding(PadBox, PadBox)
		framed = style.Render(content)
	default:
		framed = content
	}

	// Center in terminal.
	if termWidth > 0 && termHeight > 0 {
		return lipgloss.NewStyle().
			Width(termWidth).
			Height(termHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Render(framed)
	}

	if termWidth > 0 {
		return lipgloss.NewStyle().
			Width(termWidth).
			Align(lipgloss.Center).
			Render(framed)
	}

	return framed
}

// RenderHelpBar renders a horizontal bar of keyboard shortcut hints,
// dimmed and centered within the given width.
func RenderHelpBar(hints []KeyHint, width int) string {
	if len(hints) == 0 {
		return ""
	}

	var parts []string
	for _, h := range hints {
		key := AmberText(h.Key)
		part := fmt.Sprintf("[%s] %s", key, h.Action)
		parts = append(parts, part)
	}

	bar := strings.Join(parts, "  ")
	return CenterText(MutedStyle.Render(bar), width)
}

// RenderSectionTitle renders a section title with consistent styling,
// centered within the given width.
func RenderSectionTitle(title string, width int) string {
	styled := SectionTitleStyle.Render(title)
	return CenterText(styled, width)
}

// RenderWizardProgress renders a step progress indicator for wizard flows.
// Steps are shown as connected dots: ✓ for done, ● for active, ○ for pending.
func RenderWizardProgress(steps []WizardStep, width int) string {
	if len(steps) == 0 {
		return ""
	}

	var dots []string
	for _, s := range steps {
		var dot string
		switch {
		case s.Done:
			dot = GreenText(IconCheck)
		case s.Active:
			dot = AmberText("●")
		default:
			dot = DimTextStyle.Render("○")
		}
		dots = append(dots, dot)
	}

	connector := DimTextStyle.Render("─")
	progress := strings.Join(dots, connector)

	return CenterText(progress, width)
}

// ===== Box Rendering Helpers =====

// padLineToBoxWidth right-pads a line so that, when a trailing border char
// is appended, the total visual width equals boxWidth+2 (left+right borders).
//
// line must already include the left border and any leading spacing.
// The caller should append the right border (e.g. "│\n") after calling this.
func padLineToBoxWidth(line string, boxWidth int) string {
	target := boxWidth + 1 // +1 accounts for the right border that will be appended
	visual := lipgloss.Width(line)
	padding := target - visual
	if padding > 0 {
		return line + strings.Repeat(" ", padding)
	}
	return line
}
