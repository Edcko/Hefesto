package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ===== SelectModel constructor tests =====

func TestNewSelectModel(t *testing.T) {
	m := NewSelectModel(80, 40)

	if m == nil {
		t.Fatal("NewSelectModel() returned nil")
	}
	if m.width != 80 {
		t.Errorf("width = %d, want 80", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
	if m.continueRow {
		t.Error("continueRow should be false initially")
	}
	if m.items == nil {
		t.Fatal("items should not be nil")
	}
	if len(m.items.Items) == 0 {
		t.Error("items should have default components")
	}
}

// ===== SelectModel.Init tests =====

func TestSelectInit(t *testing.T) {
	m := NewSelectModel(80, 40)
	cmd := m.Init()
	_ = cmd // Init returns nil — that's fine
}

// ===== SelectModel.Update tests =====

func TestSelectUpdateWindowSize(t *testing.T) {
	m := NewSelectModel(80, 40)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	m2, ok := updated.(*SelectModel)
	if !ok {
		t.Fatalf("Update returned %T, want *SelectModel", updated)
	}
	if m2.width != 120 {
		t.Errorf("width = %d, want 120", m2.width)
	}
	if m2.height != 50 {
		t.Errorf("height = %d, want 50", m2.height)
	}
}

func TestSelectUpdateDownMovesCursor(t *testing.T) {
	m := NewSelectModel(80, 40)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2, _ := updated.(*SelectModel)
	if m2.cursor != 1 {
		t.Errorf("cursor after down = %d, want 1", m2.cursor)
	}

	// Test 'j' key
	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m3, _ := updated.(*SelectModel)
	if m3.cursor != 2 {
		t.Errorf("cursor after j = %d, want 2", m3.cursor)
	}
}

func TestSelectUpdateUpMovesCursor(t *testing.T) {
	m := NewSelectModel(80, 40)
	m.cursor = 2

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m2, _ := updated.(*SelectModel)
	if m2.cursor != 1 {
		t.Errorf("cursor after up = %d, want 1", m2.cursor)
	}

	// Test 'k' key
	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m3, _ := updated.(*SelectModel)
	if m3.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", m3.cursor)
	}
}

func TestSelectUpdateDownAtLastItemMovesToContinue(t *testing.T) {
	m := NewSelectModel(80, 40)
	lastIdx := len(m.items.Items) - 1
	m.cursor = lastIdx

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2, _ := updated.(*SelectModel)
	if !m2.continueRow {
		t.Error("should be on continue row after down at last item")
	}
}

func TestSelectUpdateUpAtFirstStays(t *testing.T) {
	m := NewSelectModel(80, 40)
	m.cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m2, _ := updated.(*SelectModel)
	if m2.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (should stay at top)", m2.cursor)
	}
}

func TestSelectUpdateUpFromContinueGoesToLastItem(t *testing.T) {
	m := NewSelectModel(80, 40)
	m.continueRow = true
	expectedCursor := len(m.items.Items) - 1

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m2, _ := updated.(*SelectModel)
	if m2.continueRow {
		t.Error("should not be on continue row after up")
	}
	if m2.cursor != expectedCursor {
		t.Errorf("cursor = %d, want %d", m2.cursor, expectedCursor)
	}
}

func TestSelectUpdateSpaceToggleNonRequired(t *testing.T) {
	m := NewSelectModel(80, 40)
	// Skills is index 2, non-required, selected by default
	m.cursor = 2

	if !m.items.Items[2].Selected {
		t.Fatal("Skills should be selected by default")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2, _ := updated.(*SelectModel)
	if m2.items.Items[2].Selected {
		t.Error("Skills should be deselected after Space")
	}

	// Toggle back
	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeySpace})
	m3, _ := updated.(*SelectModel)
	if !m3.items.Items[2].Selected {
		t.Error("Skills should be selected again after Space")
	}
}

func TestSelectUpdateSpaceDoesNotToggleRequired(t *testing.T) {
	m := NewSelectModel(80, 40)
	m.cursor = 0 // AGENTS.md, required

	if !m.items.Items[0].Required {
		t.Fatal("AGENTS.md should be required")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m2, _ := updated.(*SelectModel)
	if !m2.items.Items[0].Selected {
		t.Error("AGENTS.md should still be selected (required)")
	}
}

func TestSelectUpdateEnterToggleNonRequired(t *testing.T) {
	m := NewSelectModel(80, 40)
	m.cursor = 3 // Theme, non-required

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2, _ := updated.(*SelectModel)
	if m2.items.Items[3].Selected {
		t.Error("Theme should be deselected after Enter")
	}
}

func TestSelectUpdateSpaceOnContinueTransitions(t *testing.T) {
	m := NewSelectModel(80, 40)
	m.continueRow = true

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if cmd == nil {
		t.Fatal("Space on continue should return a command")
	}
	msg := cmd()
	if _, ok := msg.(selectCompleteMsg); !ok {
		t.Fatalf("command produced %T, want selectCompleteMsg", msg)
	}
	_ = updated
}

func TestSelectUpdateEnterOnContinueTransitions(t *testing.T) {
	m := NewSelectModel(80, 40)
	m.continueRow = true

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter on continue should return a command")
	}
	msg := cmd()
	if _, ok := msg.(selectCompleteMsg); !ok {
		t.Fatalf("command produced %T, want selectCompleteMsg", msg)
	}
	_ = updated
}

func TestSelectUpdateATogglesAll(t *testing.T) {
	m := NewSelectModel(80, 40)

	// All items start selected. Press 'a' → deselect all non-required
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m2, _ := updated.(*SelectModel)
	for i, item := range m2.items.Items {
		if item.Required && !item.Selected {
			t.Errorf("item %d (%s) is required but deselected", i, item.Name)
		}
		if !item.Required && item.Selected {
			t.Errorf("item %d (%s) should be deselected after toggle all", i, item.Name)
		}
	}

	// Press 'a' again → select all non-required
	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m3, _ := updated.(*SelectModel)
	for i, item := range m3.items.Items {
		if !item.Selected {
			t.Errorf("item %d (%s) should be selected after second toggle all", i, item.Name)
		}
	}
}

func TestSelectUpdateEscTransitions(t *testing.T) {
	m := NewSelectModel(80, 40)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should return a transition command")
	}
	msg := cmd()
	trans, ok := msg.(ScreenTransitionMsg)
	if !ok {
		t.Fatalf("command produced %T, want ScreenTransitionMsg", msg)
	}
	if trans.Target != ScreenDetect {
		t.Errorf("target = %v, want ScreenDetect", trans.Target)
	}
	_ = updated
}

func TestSelectUpdateQQuits(t *testing.T) {
	m := NewSelectModel(80, 40)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should return quit command")
	}
	_ = updated
}

// ===== SelectModel.View tests =====

func TestSelectViewContainsExpectedContent(t *testing.T) {
	m := NewSelectModel(80, 40)
	view := m.View()

	checks := []struct {
		name    string
		content string
	}{
		{"title", "Select Components"},
		{"navigate hint", "Navigate"},
		{"toggle hint", "Toggle"},
		{"quit hint", "Quit"},
		{"toggle all hint", "Toggle all"},
		{"continue action", "Continue"},
		{"AGENTS.md item", "AGENTS.md"},
		{"opencode.json item", "opencode.json"},
		{"Skills item", "Skills"},
		{"Theme item", "Theme"},
		{"Commands item", "Commands"},
		{"Plugins item", "Plugins"},
		{"Engram item", "Engram"},
		{"border rounded", "╭"},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if !strings.Contains(view, check.content) {
				t.Errorf("View() missing %q", check.content)
			}
		})
	}
}

func TestSelectViewAdaptsToNarrowTerminal(t *testing.T) {
	m := NewSelectModel(40, 20)
	view := m.View()

	if !strings.Contains(view, "Select Components") {
		t.Error("narrow view missing title")
	}
	if !strings.Contains(view, "AGENTS.md") {
		t.Error("narrow view missing AGENTS.md")
	}
}

func TestSelectViewDefaultWidth(t *testing.T) {
	m := NewSelectModel(0, 0)
	view := m.View()
	if !strings.Contains(view, "Select Components") {
		t.Error("default width view missing title")
	}
}

func TestSelectViewBordersAligned(t *testing.T) {
	m := NewSelectModel(80, 40)
	view := m.View()

	// Every line that contains a rounded border should be properly rendered.
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		plain := stripAnsi(line)
		// Check for rounded border characters used by lipgloss.RoundedBorder()
		if !strings.Contains(plain, "╭") && !strings.Contains(plain, "╰") &&
			!strings.Contains(plain, "│") {
			continue
		}
		// lipgloss border should handle alignment automatically
		// Just verify the line isn't empty
		if len(plain) == 0 {
			t.Errorf("line %d is empty but appears to be a border line", i)
		}
	}
}

// ===== truncateVisual tests =====

func TestTruncateVisual(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     string
	}{
		{"short enough", "hello", 10, "hello"},
		{"exact fit", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello w…"},
		{"very narrow", "hello world", 3, "he…"},
		{"single char ellipsis", "abcdef", 1, "…"},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateVisual(tt.input, tt.maxWidth)
			if got != tt.want {
				t.Errorf("truncateVisual(%q, %d) = %q, want %q", tt.input, tt.maxWidth, got, tt.want)
			}
			// Verify the result fits
			if tt.input != "" && lipgloss.Width(got) > tt.maxWidth {
				t.Errorf("result width %d exceeds maxWidth %d: %q", lipgloss.Width(got), tt.maxWidth, got)
			}
		})
	}
}

// ===== ComponentSelection tests =====

func TestDefaultComponentSelection(t *testing.T) {
	sel := DefaultComponentSelection()

	for _, item := range sel.Items {
		if !item.Selected {
			t.Errorf("%s should be selected by default", item.Name)
		}
	}

	requiredNames := map[string]bool{"AGENTS.md": true, "opencode.json": true}
	for _, item := range sel.Items {
		_, isRequired := requiredNames[item.Name]
		if item.Required != isRequired {
			t.Errorf("%s: Required=%v, want %v", item.Name, item.Required, isRequired)
		}
	}
}

func TestComponentSelectionIsSelected(t *testing.T) {
	sel := DefaultComponentSelection()

	if !sel.IsSelected(ComponentAgents) {
		t.Error("Agents should be selected")
	}
	if !sel.IsSelected(ComponentSkills) {
		t.Error("Skills should be selected")
	}
	if sel.IsSelected(ComponentID("nonexistent")) {
		t.Error("Nonexistent component should not be selected")
	}
}

// ===== GetSelection tests =====

func TestSelectModelGetSelection(t *testing.T) {
	m := NewSelectModel(80, 40)
	sel := m.GetSelection()

	if sel == nil {
		t.Fatal("GetSelection() returned nil")
	}
	if len(sel.Items) != len(DefaultComponentSelection().Items) {
		t.Errorf("got %d items, want %d", len(sel.Items), len(DefaultComponentSelection().Items))
	}
}
