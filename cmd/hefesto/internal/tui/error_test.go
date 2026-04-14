package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ===== ErrorModel constructor tests =====

func TestNewErrorModel(t *testing.T) {
	err := &InstallError{Step: "test-step", Message: "something failed"}
	m := NewErrorModel(err)

	if m == nil {
		t.Fatal("NewErrorModel() returned nil")
	}
	if m.err != err {
		t.Error("error not set correctly")
	}
	if len(m.CompletedSteps) != 0 {
		t.Errorf("CompletedSteps = %v, want empty", m.CompletedSteps)
	}
	if len(m.PendingSteps) == 0 {
		t.Error("PendingSteps is empty, expected default pending steps")
	}
}

// ===== ErrorModel.Init tests =====

func TestErrorInit(t *testing.T) {
	err := &InstallError{Step: "test", Message: "fail"}
	m := NewErrorModel(err)
	cmd := m.Init()
	if cmd != nil {
		t.Errorf("ErrorModel.Init() = %v, want nil", cmd)
	}
}

// ===== ErrorModel.SetSteps tests =====

func TestErrorSetSteps(t *testing.T) {
	err := &InstallError{Step: "test", Message: "fail"}
	m := NewErrorModel(err)

	completed := []string{"Step 1", "Step 2"}
	pending := []string{"Step 4", "Step 5"}
	failed := "Step 3"

	m.SetSteps(completed, pending, failed)

	if len(m.CompletedSteps) != 2 {
		t.Errorf("CompletedSteps len = %d, want 2", len(m.CompletedSteps))
	}
	if len(m.PendingSteps) != 2 {
		t.Errorf("PendingSteps len = %d, want 2", len(m.PendingSteps))
	}
	if m.FailedStep != "Step 3" {
		t.Errorf("FailedStep = %q, want 'Step 3'", m.FailedStep)
	}
}

// ===== ErrorModel.Update tests (table-driven) =====

func TestErrorUpdateRetryKey(t *testing.T) {
	var capturedAction ErrorAction
	err := &InstallError{Step: "install", Message: "failed"}
	m := NewErrorModel(err)
	m.OnAction = func(action ErrorAction) tea.Cmd {
		capturedAction = action
		return tea.Quit
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if cmd == nil {
		t.Fatal("Update('r') returned nil command")
	}
	if capturedAction != ErrorActionRetry {
		t.Errorf("OnAction called with %v, want ErrorActionRetry", capturedAction)
	}
}

func TestErrorUpdateUndoKey(t *testing.T) {
	var capturedAction ErrorAction
	err := &InstallError{Step: "install", Message: "failed"}
	m := NewErrorModel(err)
	m.OnAction = func(action ErrorAction) tea.Cmd {
		capturedAction = action
		return tea.Quit
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})

	if cmd == nil {
		t.Fatal("Update('u') returned nil command")
	}
	if capturedAction != ErrorActionUndo {
		t.Errorf("OnAction called with %v, want ErrorActionUndo", capturedAction)
	}
}

func TestErrorUpdateQuitKey(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"q key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedAction ErrorAction
			err := &InstallError{Step: "install", Message: "failed"}
			m := NewErrorModel(err)
			m.OnAction = func(action ErrorAction) tea.Cmd {
				capturedAction = action
				return tea.Quit
			}

			_, cmd := m.Update(tt.key)

			if cmd == nil {
				t.Fatal("Update() returned nil command")
			}
			if capturedAction != ErrorActionQuit {
				t.Errorf("OnAction called with %v, want ErrorActionQuit", capturedAction)
			}
		})
	}
}

func TestErrorUpdateWithoutCallbackFallsBackToQuit(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"r without callback", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}},
		{"u without callback", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}},
		{"q without callback", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"ctrl+c without callback", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &InstallError{Step: "install", Message: "failed"}
			m := NewErrorModel(err)
			// OnAction is nil

			_, cmd := m.Update(tt.key)

			if cmd == nil {
				t.Errorf("Update(%v) with nil OnAction returned nil, expected tea.Quit", tt.key)
			}
		})
	}
}

func TestErrorUpdateOtherKeysNoOp(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"enter", tea.KeyMsg{Type: tea.KeyEnter}},
		{"escape", tea.KeyMsg{Type: tea.KeyEsc}},
		{"random letter", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &InstallError{Step: "test", Message: "fail"}
			m := NewErrorModel(err)
			_, cmd := m.Update(tt.key)

			if cmd != nil {
				t.Errorf("Update(%v) returned command, expected nil", tt.key)
			}
		})
	}
}

func TestErrorUpdateWindowSize(t *testing.T) {
	err := &InstallError{Step: "test", Message: "fail"}
	m := NewErrorModel(err)
	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if cmd != nil {
		t.Errorf("Update(WindowSize) returned command, expected nil")
	}

	model, ok := updated.(*ErrorModel)
	if !ok {
		t.Fatalf("Update returned %T, want *ErrorModel", updated)
	}
	if model.width != 100 {
		t.Errorf("width = %d, want 100", model.width)
	}
}

// ===== ErrorModel.View tests =====

func TestErrorViewContainsExpectedContent(t *testing.T) {
	err := &InstallError{Step: "Copying config", Message: "permission denied"}
	m := NewErrorModel(err)
	view := m.View()

	tests := []struct {
		name     string
		contains string
	}{
		{"title", "Installation Failed"},
		{"x icon", IconCross},
		{"step name", "Copying config"},
		{"error message", "permission denied"},
		{"options retry", "Retry"},
		{"options undo", "Undo"},
		{"options quit", "Quit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(view, tt.contains) {
				t.Errorf("View() missing %q", tt.contains)
			}
		})
	}
}

func TestErrorViewWithPartialInstall(t *testing.T) {
	err := &InstallError{Step: "install", Message: "failed"}
	m := NewErrorModel(err)
	m.SetSteps(
		[]string{"Detect environment", "Backup config"},
		[]string{"Install dependencies", "Verify installation"},
		"Copying config",
	)

	view := m.View()

	if !strings.Contains(view, "Partial installation status") {
		t.Error("View() missing partial install header")
	}
	if !strings.Contains(view, "Detect environment") {
		t.Error("View() missing completed step")
	}
	if !strings.Contains(view, "Copying config") {
		t.Error("View() missing failed step")
	}
	if !strings.Contains(view, "Install dependencies") {
		t.Error("View() missing pending step")
	}
}

func TestErrorViewWithNilError(t *testing.T) {
	m := NewErrorModel(nil)
	view := m.View()

	// Should still render the error screen
	if !strings.Contains(view, "Installation Failed") {
		t.Error("View() missing title for nil error")
	}
	// Should show default suggested fixes
	if !strings.Contains(view, "Suggested fixes") {
		t.Error("View() missing suggested fixes for nil error")
	}
}

// ===== ErrorAction constants test =====

func TestErrorActionValues(t *testing.T) {
	tests := []struct {
		name   string
		action ErrorAction
		value  int
	}{
		{"none", ErrorActionNone, 0},
		{"retry", ErrorActionRetry, 1},
		{"undo", ErrorActionUndo, 2},
		{"quit", ErrorActionQuit, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.action) != tt.value {
				t.Errorf("ErrorAction %s = %d, want %d", tt.name, tt.action, tt.value)
			}
		})
	}
}

// ===== getSuggestedFixes tests (table-driven) =====

func TestGetSuggestedFixes(t *testing.T) {
	tests := []struct {
		name     string
		err      *InstallError
		contains string
	}{
		{
			name:     "nil error returns default fixes",
			err:      nil,
			contains: "Check directory permissions",
		},
		{
			name:     "permission denied error",
			err:      &InstallError{Step: "copy", Message: "permission denied", Err: fmt.Errorf("permission denied")},
			contains: "chmod 755",
		},
		{
			name:     "network error",
			err:      &InstallError{Step: "download", Message: "network timeout", Err: fmt.Errorf("network connection refused")},
			contains: "internet connection",
		},
		{
			name:     "not found error",
			err:      &InstallError{Step: "detect", Message: "not found", Err: fmt.Errorf("no such file or directory")},
			contains: "OpenCode is installed",
		},
		{
			name:     "npm error",
			err:      &InstallError{Step: "deps", Message: "npm failed", Err: fmt.Errorf("npm command not found")},
			contains: "OpenCode is installed", // "not found" case matches before "npm" case
		},
		{
			name:     "generic error returns default",
			err:      &InstallError{Step: "unknown", Message: "something weird happened", Err: fmt.Errorf("unknown error")},
			contains: "Check directory permissions",
		},
		{
			name:     "message only without wrapped error",
			err:      &InstallError{Step: "test", Message: "permission denied access"},
			contains: "chmod 755",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewErrorModel(tt.err)
			fixes := m.getSuggestedFixes()

			if len(fixes) == 0 {
				t.Fatal("getSuggestedFixes() returned empty slice")
			}

			found := false
			for _, fix := range fixes {
				if strings.Contains(fix, tt.contains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("getSuggestedFixes() = %v, want to contain %q", fixes, tt.contains)
			}
		})
	}
}

// ===== InstallError.Error() tests =====

func TestInstallErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *InstallError
		contains string
	}{
		{
			name:     "with wrapped error",
			err:      &InstallError{Step: "copy", Message: "failed to copy", Err: fmt.Errorf("disk full")},
			contains: "disk full",
		},
		{
			name:     "without wrapped error",
			err:      &InstallError{Step: "detect", Message: "not found"},
			contains: "detect: not found",
		},
		{
			name:     "step in output",
			err:      &InstallError{Step: "install", Message: "broken", Err: fmt.Errorf("oops")},
			contains: "install",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !strings.Contains(got, tt.contains) {
				t.Errorf("InstallError.Error() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}
