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
		{"next steps header", "Next Steps:"},
		{"hefesto status", "hefesto status"},
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

func TestCompleteViewFreshInstallShowsSourceStep(t *testing.T) {
	m := NewCompleteModel("~/.config/opencode/", 80, 40)
	m.OpenCodeInstallAttempted = true
	m.OpenCodeInstallSuccess = true
	m.OpenCodeInstallVersion = "1.0.0"
	m.OpenCodeWasAlreadyInstalled = false

	view := m.View()

	// Fresh install should show source step and numbered steps
	if !strings.Contains(view, "1.") {
		t.Error("View() missing numbered step '1.' for fresh install")
	}
	if !strings.Contains(view, "2.") {
		t.Error("View() missing numbered step '2.' for fresh install")
	}
	if !strings.Contains(view, "source") {
		t.Error("View() missing 'source' command for fresh install")
	}
	if !strings.Contains(view, "required!") {
		t.Error("View() missing '(required!)' hint for source step")
	}
	if !strings.Contains(view, "OpenCode CLI installed") {
		t.Error("View() missing 'OpenCode CLI installed' for successful install")
	}
	if !strings.Contains(view, "opencode") {
		t.Error("View() missing 'opencode' command")
	}
}

func TestCompleteViewAlreadyInstalledSkipsSource(t *testing.T) {
	m := NewCompleteModel("~/.config/opencode/", 80, 40)
	m.OpenCodeInstallAttempted = true
	m.OpenCodeInstallSuccess = true
	m.OpenCodeInstallVersion = "1.0.0"
	m.OpenCodeWasAlreadyInstalled = true

	view := m.View()

	// Already installed should NOT show source step
	if strings.Contains(view, "source") {
		t.Error("View() should NOT show 'source' when OpenCode was already installed")
	}
	if strings.Contains(view, "required!") {
		t.Error("View() should NOT show '(required!)' hint when OpenCode was already installed")
	}
	if !strings.Contains(view, "$ opencode") {
		t.Error("View() missing '$ opencode' run prompt for already installed")
	}
}

func TestCompleteViewFailedInstallShowsFallback(t *testing.T) {
	m := NewCompleteModel("~/.config/opencode/", 80, 40)
	m.OpenCodeInstallAttempted = true
	m.OpenCodeInstallSuccess = false
	m.OpenCodeInstallError = "OpenCode CLI install failed (non-fatal): connection refused"

	view := m.View()

	// Failed install should show curl fallback and numbered steps
	if !strings.Contains(view, "Install OpenCode CLI manually:") {
		t.Error("View() missing manual install prompt for failed install")
	}
	if !strings.Contains(view, "curl -fsSL https://opencode.ai/install | bash") {
		t.Error("View() missing curl command for failed install")
	}
	if !strings.Contains(view, "1.") {
		t.Error("View() missing numbered step '1.' for failed install fallback")
	}
	if !strings.Contains(view, "source") {
		t.Error("View() missing 'source' step in failed install fallback")
	}
}

func TestCompleteViewNoInstallAttemptedShowsSimplePrompt(t *testing.T) {
	m := NewCompleteModel("~/.config/opencode/", 80, 40)
	// OpenCodeInstallAttempted is false by default

	view := m.View()

	// No install attempted — should show simple "Run $ opencode" prompt
	if !strings.Contains(view, "$ opencode") {
		t.Error("View() missing '$ opencode' for no-install-attempt case")
	}
	if strings.Contains(view, "source") {
		t.Error("View() should NOT show 'source' when no install was attempted")
	}
	if strings.Contains(view, "required!") {
		t.Error("View() should NOT show '(required!)' when no install was attempted")
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
