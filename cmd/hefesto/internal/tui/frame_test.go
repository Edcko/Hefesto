package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// ===== ResolveContentWidth tests =====

func TestResolveContentWidth(t *testing.T) {
	tests := []struct {
		name      string
		termWidth int
		wantWidth int
	}{
		{"zero defaults to ContentWidth", 0, ContentWidth},
		{"negative defaults to ContentWidth", -1, ContentWidth},
		{"narrow clamps to MinContentWidth", 20, MinContentWidth},
		{"exact min passes through", MinContentWidth, MinContentWidth},
		{"wide clamps to MaxContentWidth", 200, MaxContentWidth},
		{"exact max passes through", MaxContentWidth, MaxContentWidth},
		{"normal value passes through", 65, 65},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveContentWidth(tt.termWidth)
			if got != tt.wantWidth {
				t.Errorf("ResolveContentWidth(%d) = %d, want %d", tt.termWidth, got, tt.wantWidth)
			}
		})
	}
}

// ===== Layout constant sanity tests =====

func TestLayoutConstantsRelationship(t *testing.T) {
	if MinContentWidth >= ContentWidth {
		t.Errorf("MinContentWidth (%d) should be less than ContentWidth (%d)", MinContentWidth, ContentWidth)
	}
	if ContentWidth >= MaxContentWidth {
		t.Errorf("ContentWidth (%d) should be less than MaxContentWidth (%d)", ContentWidth, MaxContentWidth)
	}
	if SpaceXS >= SpaceSM {
		t.Errorf("SpaceXS (%d) should be less than SpaceSM (%d)", SpaceXS, SpaceSM)
	}
	if SpaceSM >= SpaceMD {
		t.Errorf("SpaceSM (%d) should be less than SpaceMD (%d)", SpaceSM, SpaceMD)
	}
}

// ===== RenderCenteredHero tests =====

func TestRenderCenteredHeroAllFields(t *testing.T) {
	result := RenderCenteredHero("BANNER", "TITLE", "subtitle", 40)

	if !strings.Contains(result, "BANNER") {
		t.Error("RenderCenteredHero missing banner text")
	}
	if !strings.Contains(result, "TITLE") {
		t.Error("RenderCenteredHero missing title text")
	}
	if !strings.Contains(result, "subtitle") {
		t.Error("RenderCenteredHero missing subtitle text")
	}
}

func TestRenderCenteredHeroEmptyBanner(t *testing.T) {
	result := RenderCenteredHero("", "TITLE", "sub", 40)

	if strings.Contains(result, "BANNER") {
		t.Error("RenderCenteredHero should not contain banner when empty")
	}
	if !strings.Contains(result, "TITLE") {
		t.Error("RenderCenteredHero missing title text")
	}
}

func TestRenderCenteredHeroAllEmpty(t *testing.T) {
	result := RenderCenteredHero("", "", "", 40)

	// Should produce minimal output (just newlines)
	if len(result) > 2 {
		t.Errorf("RenderCenteredHero with all empty should be minimal, got %d bytes: %q", len(result), result)
	}
}

// ===== RenderScreenFrame tests =====

func TestRenderScreenFrameNoBorder(t *testing.T) {
	opts := FrameOptions{Width: 60, Height: 20, Border: BorderNone}
	result := RenderScreenFrame("hello", opts)

	if !strings.Contains(result, "hello") {
		t.Error("RenderScreenFrame missing content")
	}
}

func TestRenderScreenFrameRoundedBorder(t *testing.T) {
	opts := FrameOptions{Width: 80, Height: 24, Border: BorderRounded}
	result := RenderScreenFrame("test content", opts)

	if !strings.Contains(result, "test content") {
		t.Error("RenderScreenFrame missing content with rounded border")
	}
	// Rounded border should produce curve chars
	if !strings.Contains(result, "╭") && !strings.Contains(result, "╮") {
		t.Error("RenderScreenFrame with RoundedBorder missing border chars")
	}
}

func TestRenderScreenFrameZeroDimensions(t *testing.T) {
	opts := FrameOptions{Width: 0, Height: 0, Border: BorderNone}
	result := RenderScreenFrame("content", opts)

	if !strings.Contains(result, "content") {
		t.Error("RenderScreenFrame with zero dims should still contain content")
	}
}

// ===== RenderHelpBar tests =====

func TestRenderHelpBar(t *testing.T) {
	hints := []KeyHint{
		{Key: "Enter", Action: "Continue"},
		{Key: "q", Action: "Quit"},
	}
	result := RenderHelpBar(hints, 60)

	if !strings.Contains(result, "Enter") {
		t.Error("RenderHelpBar missing Enter key")
	}
	if !strings.Contains(result, "Continue") {
		t.Error("RenderHelpBar missing Continue action")
	}
	if !strings.Contains(result, "Quit") {
		t.Error("RenderHelpBar missing Quit action")
	}
}

func TestRenderHelpBarEmpty(t *testing.T) {
	result := RenderHelpBar(nil, 60)
	if result != "" {
		t.Errorf("RenderHelpBar with nil hints should return empty, got %q", result)
	}
}

// ===== RenderSectionTitle tests =====

func TestRenderSectionTitle(t *testing.T) {
	result := RenderSectionTitle("Test Section", 60)

	if !strings.Contains(result, "Test Section") {
		t.Error("RenderSectionTitle missing title text")
	}
}

// ===== RenderWizardProgress tests =====

func TestRenderWizardProgressEmpty(t *testing.T) {
	result := RenderWizardProgress(nil, 60)
	if result != "" {
		t.Errorf("RenderWizardProgress with nil steps should return empty, got %q", result)
	}
}

func TestRenderWizardProgressSingleActiveStep(t *testing.T) {
	steps := []WizardStep{
		{Label: "Welcome", Active: true},
	}
	result := RenderWizardProgress(steps, 60)

	if !strings.Contains(result, "●") {
		t.Error("RenderWizardProgress missing active dot for active step")
	}
}

func TestRenderWizardProgressDoneAndPending(t *testing.T) {
	steps := []WizardStep{
		{Label: "Welcome", Done: true},
		{Label: "Detect", Active: true},
		{Label: "Install"},
	}
	result := RenderWizardProgress(steps, 60)

	if !strings.Contains(result, "✓") {
		t.Error("RenderWizardProgress missing check for done step")
	}
	if !strings.Contains(result, "●") {
		t.Error("RenderWizardProgress missing active dot")
	}
	if !strings.Contains(result, "○") {
		t.Error("RenderWizardProgress missing empty dot for pending step")
	}
}

// ===== padLineToBoxWidth tests =====

func TestPadLineToBoxWidth(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		boxWidth int
		wantLen  int // visual width of padded line
	}{
		{
			name:     "short line gets padded",
			line:     "│  hello",
			boxWidth: 20,
			wantLen:  21, // boxWidth + 1 (accounting for right border)
		},
		{
			name:     "exact fit no extra padding",
			line:     "│" + strings.Repeat("x", 20),
			boxWidth: 20,
			wantLen:  21,
		},
		{
			name:     "line with ANSI codes measured visually",
			line:     "│  \x1b[33mstyled\x1b[0m text",
			boxWidth: 30,
			wantLen:  31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := padLineToBoxWidth(tt.line, tt.boxWidth)
			visualW := lipglossWidth(padded)
			if visualW != tt.wantLen {
				t.Errorf("padLineToBoxWidth visual width = %d, want %d (line=%q)",
					visualW, tt.wantLen, padded)
			}
		})
	}
}

// lipglossWidth is a test helper that wraps lipgloss.Width.
func lipglossWidth(s string) int {
	return lipgloss.Width(s)
}

// ===== BorderType tests =====

func TestBorderTypeValues(t *testing.T) {
	if BorderNone != 0 {
		t.Errorf("BorderNone = %d, want 0", BorderNone)
	}
	if BorderRounded != 1 {
		t.Errorf("BorderRounded = %d, want 1", BorderRounded)
	}
}

// ===== KeyHint tests =====

func TestKeyHintUsage(t *testing.T) {
	hints := []KeyHint{
		{Key: "↑↓", Action: "Navigate"},
		{Key: "Space", Action: "Toggle"},
	}

	if hints[0].Key != "↑↓" {
		t.Errorf("KeyHint.Key = %q, want '↑↓'", hints[0].Key)
	}
	if hints[1].Action != "Toggle" {
		t.Errorf("KeyHint.Action = %q, want 'Toggle'", hints[1].Action)
	}
}

// ===== WizardStep tests =====

func TestWizardStepStates(t *testing.T) {
	done := WizardStep{Label: "Step 1", Done: true}
	active := WizardStep{Label: "Step 2", Active: true}
	pending := WizardStep{Label: "Step 3"}

	if !done.Done {
		t.Error("done step should have Done=true")
	}
	if !active.Active {
		t.Error("active step should have Active=true")
	}
	if pending.Done || pending.Active {
		t.Error("pending step should have neither Done nor Active")
	}
}
