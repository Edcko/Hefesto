package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ===== App constructor tests =====

func TestNewApp(t *testing.T) {
	app := NewApp()

	if app == nil {
		t.Fatal("NewApp() returned nil")
	}
	if app.screen != ScreenWelcome {
		t.Errorf("initial screen = %v, want ScreenWelcome", app.screen)
	}
	if app.configPath != AppConfigPath {
		t.Errorf("configPath = %q, want %q", app.configPath, AppConfigPath)
	}
	if app.quitting {
		t.Error("app should not be quitting initially")
	}
}

// ===== App.Init tests =====

func TestAppInit(t *testing.T) {
	app := NewApp()
	cmd := app.Init()

	// welcome.Init() returns nil, so tea.Batch(nil) returns nil
	// What matters is the welcome model gets created
	if app.welcome == nil {
		t.Error("welcome model not initialized after Init()")
	}
	// cmd may be nil if welcome.Init() returns nil — that's fine
	t.Logf("App.Init() cmd = %v", cmd)
}

// ===== App.Update - global key tests (table-driven) =====

func TestAppUpdateCtrlCQuits(t *testing.T) {
	app := NewApp()
	_ = app.Init()

	updated, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Fatal("Update(CtrlC) returned nil command")
	}

	appModel, ok := updated.(*App)
	if !ok {
		t.Fatalf("Update returned %T, want *App", updated)
	}
	if !appModel.quitting {
		t.Error("app.quitting = false, want true after Ctrl+C")
	}
}

func TestAppUpdateQKeyOnCompleteScreen(t *testing.T) {
	app := NewApp()
	_ = app.Init()
	app.screen = ScreenComplete
	app.complete = NewCompleteModel("~/.config/opencode/", 80, 40)

	updated, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Fatal("Update(q) on complete screen returned nil command")
	}

	appModel, ok := updated.(*App)
	if !ok {
		t.Fatalf("Update returned %T, want *App", updated)
	}
	if !appModel.quitting {
		t.Error("app.quitting = false, want true after q on complete screen")
	}
}

func TestAppUpdateQKeyOnErrorScreen(t *testing.T) {
	app := NewApp()
	_ = app.Init()
	app.screen = ScreenError
	app.err = NewErrorModel(&InstallError{Step: "test", Message: "fail"})
	app.err.OnAction = func(action ErrorAction) tea.Cmd {
		return tea.Quit
	}

	updated, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Fatal("Update(q) on error screen returned nil command")
	}

	appModel, ok := updated.(*App)
	if !ok {
		t.Fatalf("Update returned %T, want *App", updated)
	}
	if !appModel.quitting {
		t.Error("app.quitting = false, want true after q on error screen")
	}
}

func TestAppUpdateQKeyOnWelcomeScreenDoesNotQuit(t *testing.T) {
	app := NewApp()
	_ = app.Init()

	// q on welcome screen delegates to WelcomeModel, which doesn't handle q
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	appModel, ok := updated.(*App)
	if !ok {
		t.Fatalf("Update returned %T, want *App", updated)
	}
	if appModel.quitting {
		t.Error("app should NOT quit on q from welcome screen")
	}
}

// ===== App.Update - window size propagation tests =====

func TestAppUpdateWindowSizePropagatesToWelcome(t *testing.T) {
	app := NewApp()
	_ = app.Init()

	_, _ = app.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if app.width != 100 {
		t.Errorf("app.width = %d, want 100", app.width)
	}
	if app.height != 50 {
		t.Errorf("app.height = %d, want 50", app.height)
	}
	if app.welcome != nil {
		if app.welcome.width != 100 {
			t.Errorf("welcome.width = %d, want 100", app.welcome.width)
		}
		if app.welcome.height != 50 {
			t.Errorf("welcome.height = %d, want 50", app.welcome.height)
		}
	}
}

// ===== App.Update - ScreenTransitionMsg tests (table-driven) =====

func TestAppHandleScreenTransition(t *testing.T) {
	tests := []struct {
		name         string
		target       Screen
		expectModel  string // "detect", "select", "backup", "install", "complete"
		expectScreen Screen
	}{
		{
			name:         "transition to detect",
			target:       ScreenDetect,
			expectModel:  "detect",
			expectScreen: ScreenDetect,
		},
		{
			name:         "transition to component select",
			target:       ScreenComponentSelect,
			expectModel:  "select",
			expectScreen: ScreenComponentSelect,
		},
		{
			name:         "transition to backup",
			target:       ScreenBackup,
			expectModel:  "backup",
			expectScreen: ScreenBackup,
		},
		{
			name:         "transition to install",
			target:       ScreenInstall,
			expectModel:  "install",
			expectScreen: ScreenInstall,
		},
		{
			name:         "transition to complete",
			target:       ScreenComplete,
			expectModel:  "complete",
			expectScreen: ScreenComplete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			_ = app.Init()

			updated, cmd := app.Update(ScreenTransitionMsg{Target: tt.target})

			appModel, ok := updated.(*App)
			if !ok {
				t.Fatalf("Update returned %T, want *App", updated)
			}

			if appModel.screen != tt.expectScreen {
				t.Errorf("screen = %v, want %v", appModel.screen, tt.expectScreen)
			}

			switch tt.expectModel {
			case "detect":
				if appModel.detect == nil {
					t.Error("detect model is nil after transition")
				}
			case "select":
				if appModel.select_ == nil {
					t.Error("select model is nil after transition")
				}
			case "backup":
				if appModel.backup == nil {
					t.Error("backup model is nil after transition")
				}
			case "install":
				if appModel.install == nil {
					t.Error("install model is nil after transition")
				}
			case "complete":
				if appModel.complete == nil {
					t.Error("complete model is nil after transition")
				}
			}

			// Some screens' Init() returns nil (select, backup, complete) — that's valid
			// Only detect and install return actual commands
			_ = cmd // cmd may be nil depending on screen's Init()
		})
	}
}

// ===== App.Update - ErrorMsg tests =====

func TestAppUpdateErrorMsg(t *testing.T) {
	app := NewApp()
	_ = app.Init()

	errMsg := ErrorMsg{
		Step:    "install",
		Message: "something went wrong",
		Err:     nil,
	}

	updated, cmd := app.Update(errMsg)

	// err.Init() returns nil, so cmd from ErrorMsg handler is nil
	t.Logf("Update(ErrorMsg) cmd = %v", cmd)

	appModel, ok := updated.(*App)
	if !ok {
		t.Fatalf("Update returned %T, want *App", updated)
	}

	if appModel.screen != ScreenError {
		t.Errorf("screen = %v, want ScreenError", appModel.screen)
	}
	if appModel.err == nil {
		t.Error("error model is nil after ErrorMsg")
	}
	if appModel.lastError == nil {
		t.Error("lastError is nil after ErrorMsg")
	}
}

// ===== App.View tests =====

func TestAppViewQuitting(t *testing.T) {
	app := NewApp()
	_ = app.Init()
	app.quitting = true

	view := app.View()
	if !strings.Contains(view, "Thanks for using Hefesto") {
		t.Errorf("View() when quitting = %q, want to contain 'Thanks for using Hefesto'", view)
	}
}

func TestAppViewWelcomeScreen(t *testing.T) {
	app := NewApp()
	_ = app.Init()
	app.width = 80
	app.height = 40

	view := app.View()
	if !strings.Contains(view, "HEFESTO") {
		t.Error("View() on welcome screen missing 'HEFESTO'")
	}
}

func TestAppViewCompleteScreen(t *testing.T) {
	app := NewApp()
	_ = app.Init()
	app.screen = ScreenComplete
	app.complete = NewCompleteModel("~/.config/opencode/", 80, 40)
	app.width = 80
	app.height = 40

	view := app.View()
	if !strings.Contains(view, "Installation Complete") {
		t.Error("View() on complete screen missing 'Installation Complete'")
	}
}

func TestAppViewErrorScreen(t *testing.T) {
	app := NewApp()
	_ = app.Init()
	app.screen = ScreenError
	app.err = NewErrorModel(&InstallError{Step: "test", Message: "fail"})
	app.width = 80
	app.height = 40

	view := app.View()
	if !strings.Contains(view, "Installation Failed") {
		t.Error("View() on error screen missing 'Installation Failed'")
	}
}

func TestAppViewComponentSelectScreen(t *testing.T) {
	app := NewApp()
	_ = app.Init()
	app.screen = ScreenComponentSelect
	app.select_ = NewSelectModel(80, 40)
	app.width = 80
	app.height = 40

	view := app.View()
	if !strings.Contains(view, "Select Components") {
		t.Error("View() on select screen missing 'Select Components'")
	}
}

// ===== App.SetDetectionResults tests =====

func TestAppSetDetectionResults(t *testing.T) {
	app := NewApp()

	app.SetDetectionResults(true, "v1.0.0", true, true)

	if !app.openCodeInstalled {
		t.Error("openCodeInstalled = false, want true")
	}
	if app.openCodeVersion != "v1.0.0" {
		t.Errorf("openCodeVersion = %q, want 'v1.0.0'", app.openCodeVersion)
	}
	if !app.existingConfig {
		t.Error("existingConfig = false, want true")
	}
	if !app.isGentlemanDots {
		t.Error("isGentlemanDots = false, want true")
	}
}

// ===== App.SetBackupPath tests =====

func TestAppSetBackupPath(t *testing.T) {
	app := NewApp()

	app.SetBackupPath("/tmp/backup-123")

	if app.backupPath != "/tmp/backup-123" {
		t.Errorf("backupPath = %q, want '/tmp/backup-123'", app.backupPath)
	}
}

// ===== TransitionTo command tests =====

func TestTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		screen Screen
	}{
		{"welcome", ScreenWelcome},
		{"detect", ScreenDetect},
		{"component-select", ScreenComponentSelect},
		{"backup", ScreenBackup},
		{"install", ScreenInstall},
		{"complete", ScreenComplete},
		{"error", ScreenError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := TransitionTo(tt.screen)
			if cmd == nil {
				t.Fatal("TransitionTo() returned nil command")
			}

			msg := cmd()
			transition, ok := msg.(ScreenTransitionMsg)
			if !ok {
				t.Fatalf("command produced %T, want ScreenTransitionMsg", msg)
			}
			if transition.Target != tt.screen {
				t.Errorf("target = %v, want %v", transition.Target, tt.screen)
			}
		})
	}
}

// ===== NewErrorMsg tests =====

func TestNewErrorMsg(t *testing.T) {
	cmd := NewErrorMsg("test-step", "test message", nil)
	if cmd == nil {
		t.Fatal("NewErrorMsg() returned nil command")
	}

	msg := cmd()
	errMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("command produced %T, want ErrorMsg", msg)
	}
	if errMsg.Step != "test-step" {
		t.Errorf("step = %q, want 'test-step'", errMsg.Step)
	}
	if errMsg.Message != "test message" {
		t.Errorf("message = %q, want 'test message'", errMsg.Message)
	}
}

// ===== StepStatus.String() tests =====

func TestStepStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status StepStatus
		want   string
	}{
		{"pending", StepPending, "pending"},
		{"running", StepRunning, "running"},
		{"complete", StepComplete, "complete"},
		{"error", StepError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("StepStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

// ===== UndoCompleteMsg flow test =====

func TestAppUndoCompleteQuits(t *testing.T) {
	app := NewApp()
	_ = app.Init()

	updated, cmd := app.Update(UndoCompleteMsg{})

	appModel, ok := updated.(*App)
	if !ok {
		t.Fatalf("Update returned %T, want *App", updated)
	}
	if !appModel.quitting {
		t.Error("app should be quitting after UndoCompleteMsg")
	}
	if cmd == nil {
		t.Error("Update(UndoCompleteMsg) returned nil command")
	}
}

// ===== UpdateInstallStepTracking tests =====

func TestUpdateInstallStepTrackingWithNilInstall(t *testing.T) {
	app := NewApp()
	// install is nil — should not panic
	app.UpdateInstallStepTracking()

	if len(app.completedSteps) != 0 {
		t.Error("completedSteps should remain empty with nil install")
	}
}

// ===== ComponentSelection integration with App =====

func TestAppInstallTransitionWithDefaultSelection(t *testing.T) {
	app := NewApp()
	_ = app.Init()

	// Transition to install with no component selection set — should use defaults
	updated, cmd := app.Update(ScreenTransitionMsg{Target: ScreenInstall})

	appModel, ok := updated.(*App)
	if !ok {
		t.Fatalf("Update returned %T, want *App", updated)
	}

	if appModel.install == nil {
		t.Fatal("install model is nil after transition")
	}

	if len(appModel.install.steps) == 0 {
		t.Error("install steps are empty, expected default steps")
	}

	// First step should be "Install OpenCode CLI" (auto-selected when not installed)
	// because DefaultComponentSelection includes it as required when OpenCode is not installed
	if appModel.install.steps[0].Name != "Install OpenCode CLI" {
		t.Errorf("first step = %q, want 'Install OpenCode CLI'", appModel.install.steps[0].Name)
	}

	if cmd == nil {
		t.Error("transition to install returned nil command")
	}
}

func TestAppInstallTransitionWithCustomSelection(t *testing.T) {
	app := NewApp()
	_ = app.Init()

	// Set a minimal component selection
	sel := &ComponentSelection{
		Items: []ComponentItem{
			{ID: ComponentAgents, Name: "AGENTS.md", Selected: true, Required: true},
			{ID: ComponentOpenCode, Name: "opencode.json", Selected: true, Required: true},
		},
	}
	app.componentSelection = sel

	updated, _ := app.Update(ScreenTransitionMsg{Target: ScreenInstall})

	appModel, ok := updated.(*App)
	if !ok {
		t.Fatalf("Update returned %T, want *App", updated)
	}

	if appModel.install == nil {
		t.Fatal("install model is nil")
	}

	// Should have Detect + Copying (agents/opencode selected) + Verify = 3 steps
	// No backup step (existingConfig=false), no engram, no npm
	if len(appModel.install.steps) < 2 {
		t.Errorf("got %d steps, expected at least 2 (detect + verify)", len(appModel.install.steps))
	}
}
