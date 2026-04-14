package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ===== CompleteModel constructor tests =====

func TestNewCompleteModel(t *testing.T) {
	m := NewCompleteModel("~/.config/opencode/", 80, 40)

	if m == nil {
		t.Fatal("NewCompleteModel() returned nil")
	}
	if m.configPath != "~/.config/opencode/" {
		t.Errorf("configPath = %q, want ~/.config/opencode/", m.configPath)
	}
	if len(m.InstalledComponents) == 0 {
		t.Error("InstalledComponents is empty, expected default components")
	}
}

// ===== CompleteModel.Init tests =====

func TestCompleteInit(t *testing.T) {
	m := NewCompleteModel("~/.config/opencode/", 80, 40)
	cmd := m.Init()
	if cmd != nil {
		t.Errorf("CompleteModel.Init() = %v, want nil", cmd)
	}
}

// ===== CompleteModel.Update tests (table-driven) =====

func TestCompleteUpdateQuitKeys(t *testing.T) {
	tests := []struct {
		name    string
		key     tea.KeyMsg
		wantNil bool // cmd should be tea.Quit (not nil)
	}{
		{
			name:    "q key quits",
			key:     tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantNil: false,
		},
		{
			name:    "ctrl+c quits",
			key:     tea.KeyMsg{Type: tea.KeyCtrlC},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCompleteModel("~/.config/opencode/", 80, 40)
			_, cmd := m.Update(tt.key)

			if tt.wantNil && cmd != nil {
				t.Errorf("Update(%v) returned non-nil command, expected nil", tt.key)
			}
			if !tt.wantNil && cmd == nil {
				t.Errorf("Update(%v) returned nil command, expected quit command", tt.key)
			}
		})
	}
}

func TestCompleteUpdateOtherKeysNoOp(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"enter key", tea.KeyMsg{Type: tea.KeyEnter}},
		{"escape key", tea.KeyMsg{Type: tea.KeyEsc}},
		{"random letter a", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}},
		{"tab key", tea.KeyMsg{Type: tea.KeyTab}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCompleteModel("~/.config/opencode/", 80, 40)
			_, cmd := m.Update(tt.key)

			if cmd != nil {
				t.Errorf("Update(%v) returned command %v, expected nil", tt.key, cmd)
			}
		})
	}
}

func TestCompleteUpdateWindowSize(t *testing.T) {
	m := NewCompleteModel("~/.config/opencode/", 80, 40)
	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 60})

	if cmd != nil {
		t.Errorf("Update(WindowSize) returned command, expected nil")
	}

	model, ok := updated.(*CompleteModel)
	if !ok {
		t.Fatalf("Update returned %T, want *CompleteModel", updated)
	}
	if model.width != 120 {
		t.Errorf("width = %d, want 120", model.width)
	}
	if model.height != 60 {
		t.Errorf("height = %d, want 60", model.height)
	}
}

// ===== CompleteModel.View tests =====

func TestCompleteViewContainsExpectedContent(t *testing.T) {
	tests := []struct {
		name     string
		contains string
	}{
		{"title", "Installation Complete"},
		{"banner anvil", "║"},
		{"component config files", "Config files"},
		{"component skills", "AI skills"},
		{"component SDD agents", "SDD phase agents"},
		{"component theme", "Fuego/Forge theme"},
		{"component engram", "Engram"},
		{"component background agents", "Background agents"},
		{"component plugins", "Plugins"},
		{"start using opencode", "opencode"},
		{"check status hefesto", "hefesto status"},
		{"forge on message", "Forge on"},
		{"exit key in help bar", "q"},
	}

	m := NewCompleteModel("~/.config/opencode/", 80, 40)
	view := m.View()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(view, tt.contains) {
				t.Errorf("View() missing expected content %q", tt.contains)
			}
		})
	}
}

func TestCompleteViewWithInstallDuration(t *testing.T) {
	m := NewCompleteModel("~/.config/opencode/", 80, 40)
	m.InstallDuration = 5 * time.Second

	view := m.View()
	if !strings.Contains(view, "Installed in") {
		t.Errorf("View() missing 'Installed in' with duration set")
	}
}

func TestCompleteViewWithoutInstallDuration(t *testing.T) {
	m := NewCompleteModel("~/.config/opencode/", 80, 40)
	m.InstallDuration = 0

	view := m.View()
	if !strings.Contains(view, "< 1s") {
		t.Errorf("View() missing '< 1s' fallback when no duration set")
	}
}

func TestCompleteViewCustomComponents(t *testing.T) {
	m := NewCompleteModel("~/.config/opencode/", 80, 40)
	m.InstalledComponents = []InstalledComponent{
		{Name: "Test Component", Description: "(test desc)"},
	}

	view := m.View()
	if !strings.Contains(view, "Test Component") {
		t.Error("View() missing custom component name")
	}
	if !strings.Contains(view, "(test desc)") {
		t.Error("View() missing custom component description")
	}
}
