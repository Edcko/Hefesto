package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ===== WelcomeModel constructor tests =====

func TestNewWelcomeModel(t *testing.T) {
	m := NewWelcomeModel()
	if m == nil {
		t.Fatal("NewWelcomeModel() returned nil")
	}
	if m.width != 0 {
		t.Errorf("initial width = %d, want 0", m.width)
	}
	if m.height != 0 {
		t.Errorf("initial height = %d, want 0", m.height)
	}
}

// ===== WelcomeModel.Init tests =====

func TestWelcomeInit(t *testing.T) {
	m := NewWelcomeModel()
	cmd := m.Init()
	if cmd != nil {
		t.Errorf("WelcomeModel.Init() = %v, want nil", cmd)
	}
}

// ===== WelcomeModel.Update tests (table-driven) =====

func TestWelcomeUpdateEnterKeyTransitionsToDetect(t *testing.T) {
	m := NewWelcomeModel()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should return a command (TransitionTo)
	if cmd == nil {
		t.Fatal("Update(Enter) returned nil command, expected TransitionTo command")
	}

	// Execute the command and check the message type
	msg := cmd()
	transition, ok := msg.(ScreenTransitionMsg)
	if !ok {
		t.Fatalf("command produced %T, want ScreenTransitionMsg", msg)
	}
	if transition.Target != ScreenDetect {
		t.Errorf("transition target = %v, want ScreenDetect", transition.Target)
	}

	// Model itself should not change type
	if _, ok := updated.(*WelcomeModel); !ok {
		t.Fatalf("Update returned %T, want *WelcomeModel", updated)
	}
}

func TestWelcomeUpdateSpaceKeyTransitionsToDetect(t *testing.T) {
	m := NewWelcomeModel()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	if cmd == nil {
		t.Fatal("Update(Space) returned nil command, expected TransitionTo command")
	}

	msg := cmd()
	transition, ok := msg.(ScreenTransitionMsg)
	if !ok {
		t.Fatalf("command produced %T, want ScreenTransitionMsg", msg)
	}
	if transition.Target != ScreenDetect {
		t.Errorf("transition target = %v, want ScreenDetect", transition.Target)
	}

	if _, ok := updated.(*WelcomeModel); !ok {
		t.Fatalf("Update returned %T, want *WelcomeModel", updated)
	}
}

func TestWelcomeUpdateOtherKeysNoOp(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"escape key", tea.KeyMsg{Type: tea.KeyEsc}},
		{"tab key", tea.KeyMsg{Type: tea.KeyTab}},
		{"random letter", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}},
		{"q key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewWelcomeModel()
			updated, cmd := m.Update(tt.key)

			if cmd != nil {
				t.Errorf("Update(%v) returned non-nil command, expected nil", tt.key)
			}
			if _, ok := updated.(*WelcomeModel); !ok {
				t.Fatalf("Update returned %T, want *WelcomeModel", updated)
			}
		})
	}
}

func TestWelcomeUpdateWindowSize(t *testing.T) {
	m := NewWelcomeModel()
	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if cmd != nil {
		t.Errorf("Update(WindowSize) returned command, expected nil")
	}

	model, ok := updated.(*WelcomeModel)
	if !ok {
		t.Fatalf("Update returned %T, want *WelcomeModel", updated)
	}
	if model.width != 100 {
		t.Errorf("width = %d, want 100", model.width)
	}
	if model.height != 50 {
		t.Errorf("height = %d, want 50", model.height)
	}
}

func TestWelcomeUpdateIgnoresNonKeyNonSizeMessages(t *testing.T) {
	m := NewWelcomeModel()
	// Send a random message type
	updated, cmd := m.Update(TickMsg{})

	if cmd != nil {
		t.Errorf("Update(TickMsg) returned command, expected nil")
	}
	if _, ok := updated.(*WelcomeModel); !ok {
		t.Fatalf("Update returned %T, want *WelcomeModel", updated)
	}
}

// ===== WelcomeModel.View tests =====

func TestWelcomeViewContainsExpectedContent(t *testing.T) {
	tests := []struct {
		name     string
		contains string
	}{
		{"banner", "*"},
		{"version", "HEFESTO"},
		{"tagline", "AI Dev Environment Forge"},
		{"description", "Set up your AI-powered dev environment in seconds"},
		{"install header", "What will be installed"},
		{"enter key", "Enter"},
		{"help bar continue", "Continue"},
	}

	m := NewWelcomeModel()
	view := m.View()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(view, tt.contains) {
				t.Errorf("View() missing expected content %q.\nGot:\n%s", tt.contains, view)
			}
		})
	}
}

func TestWelcomeViewWithCustomWidth(t *testing.T) {
	m := NewWelcomeModel()
	m.width = 80
	m.height = 40

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestWelcomeViewDefaultWidth(t *testing.T) {
	m := NewWelcomeModel()
	// width is 0, should default to 60 internally
	view := m.View()
	if view == "" {
		t.Error("View() returned empty string with default width")
	}
}
