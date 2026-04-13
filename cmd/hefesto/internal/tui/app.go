package tui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	installPkg "github.com/Edcko/Hefesto/cmd/hefesto/internal/install"
)

// Screen represents the current screen in the TUI
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenDetect
	ScreenComponentSelect
	ScreenBackup
	ScreenInstall
	ScreenComplete
	ScreenError
)

// String returns the screen name
func (s Screen) String() string {
	return [...]string{"welcome", "detect", "component-select", "backup", "install", "complete", "error"}[s]
}

// InstallError wraps an error with context for the error screen
type InstallError struct {
	Step    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *InstallError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Step, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Step, e.Message)
}

// AppConfigPath is the default OpenCode config directory
const AppConfigPath = "~/.config/opencode/"

// App is the main application model that manages screen transitions
type App struct {
	screen Screen
	width  int
	height int

	// Shared state between screens
	configPath         string
	backupPath         string
	openCodeVersion    string
	openCodeInstalled  bool
	existingConfig     bool
	isGentlemanDots    bool
	componentSelection *ComponentSelection

	// Current screen models
	welcome  *WelcomeModel
	detect   *DetectModel
	select_  *SelectModel
	backup   *BackupModel
	install  *InstallModel
	complete *CompleteModel
	err      *ErrorModel

	// Error state
	lastError *InstallError

	// Partial install tracking for error screen
	completedSteps []string
	pendingSteps   []string
	failedStep     string

	// Quit channel
	quitting bool
}

// NewApp creates a new application instance
func NewApp() *App {
	return &App{
		screen:     ScreenWelcome,
		configPath: AppConfigPath,
	}
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	a.welcome = NewWelcomeModel()
	return tea.Batch(
		a.welcome.Init(),
	)
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle global keys
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			a.quitting = true
			return a, tea.Quit
		case "q":
			if a.screen == ScreenComplete || a.screen == ScreenError {
				a.quitting = true
				return a, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Propagate to all screens - ignore commands for resize
		if a.welcome != nil {
			m, _ := a.welcome.Update(msg)
			a.welcome = m.(*WelcomeModel)
		}
		if a.detect != nil {
			m, _ := a.detect.Update(msg)
			a.detect = m.(*DetectModel)
		}
		if a.select_ != nil {
			m, _ := a.select_.Update(msg)
			a.select_ = m.(*SelectModel)
		}
		if a.backup != nil {
			m, _ := a.backup.Update(msg)
			a.backup = m.(*BackupModel)
		}
		if a.install != nil {
			m, _ := a.install.Update(msg)
			a.install = m.(*InstallModel)
		}
		if a.complete != nil {
			m, _ := a.complete.Update(msg)
			a.complete = m.(*CompleteModel)
		}
		if a.err != nil {
			m, _ := a.err.Update(msg)
			a.err = m.(*ErrorModel)
		}

	// Screen transition messages
	case ScreenTransitionMsg:
		return a.handleScreenTransition(msg)

	case selectCompleteMsg:
		// Save the component selection
		if a.select_ != nil {
			a.componentSelection = a.select_.GetSelection()
		}
		// Decide next screen: backup if existing config, else install
		if a.existingConfig {
			return a.handleScreenTransition(ScreenTransitionMsg{Target: ScreenBackup})
		}
		return a.handleScreenTransition(ScreenTransitionMsg{Target: ScreenInstall})

	case ErrorMsg:
		// Snapshot install step tracking before transitioning to error screen
		a.UpdateInstallStepTracking()

		a.lastError = &InstallError{
			Step:    msg.Step,
			Message: msg.Message,
			Err:     msg.Err,
		}
		a.err = NewErrorModel(a.lastError)

		// Pass partial install step info from the install model
		a.err.SetSteps(a.completedSteps, a.pendingSteps, a.failedStep)

		// Wire up action callbacks
		a.err.OnAction = func(action ErrorAction) tea.Cmd {
			switch action {
			case ErrorActionRetry:
				// Reset install state and go back to install screen
				a.completedSteps = nil
				a.pendingSteps = nil
				a.failedStep = ""
				return TransitionTo(ScreenInstall)

			case ErrorActionUndo:
				// Perform rollback/uninstall of partial install, then quit
				return a.performUndo()

			case ErrorActionQuit:
				a.quitting = true
				return tea.Quit
			}
			return tea.Quit
		}

		a.screen = ScreenError
		return a, a.err.Init()

	case UndoCompleteMsg:
		return a.handleUndoComplete()
	}

	// Update current screen
	switch a.screen {
	case ScreenWelcome:
		if a.welcome != nil {
			m, cmd := a.welcome.Update(msg)
			a.welcome = m.(*WelcomeModel)
			cmds = append(cmds, cmd)
		}

	case ScreenDetect:
		if a.detect != nil {
			m, cmd := a.detect.Update(msg)
			a.detect = m.(*DetectModel)
			cmds = append(cmds, cmd)
		}

	case ScreenComponentSelect:
		if a.select_ != nil {
			m, cmd := a.select_.Update(msg)
			a.select_ = m.(*SelectModel)
			cmds = append(cmds, cmd)
		}

	case ScreenBackup:
		if a.backup != nil {
			m, cmd := a.backup.Update(msg)
			a.backup = m.(*BackupModel)
			cmds = append(cmds, cmd)
		}

	case ScreenInstall:
		if a.install != nil {
			m, cmd := a.install.Update(msg)
			a.install = m.(*InstallModel)
			cmds = append(cmds, cmd)
		}

	case ScreenComplete:
		if a.complete != nil {
			m, cmd := a.complete.Update(msg)
			a.complete = m.(*CompleteModel)
			cmds = append(cmds, cmd)
		}

	case ScreenError:
		if a.err != nil {
			m, cmd := a.err.Update(msg)
			a.err = m.(*ErrorModel)
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// View implements tea.Model
func (a *App) View() string {
	if a.quitting {
		return "\n" + MutedStyle.Render("Thanks for using Hefesto! 🔥") + "\n"
	}

	var content string

	switch a.screen {
	case ScreenWelcome:
		if a.welcome != nil {
			content = a.welcome.View()
		}
	case ScreenDetect:
		if a.detect != nil {
			content = a.detect.View()
		}
	case ScreenComponentSelect:
		if a.select_ != nil {
			content = a.select_.View()
		}
	case ScreenBackup:
		if a.backup != nil {
			content = a.backup.View()
		}
	case ScreenInstall:
		if a.install != nil {
			content = a.install.View()
		}
	case ScreenComplete:
		if a.complete != nil {
			content = a.complete.View()
		}
	case ScreenError:
		if a.err != nil {
			content = a.err.View()
		}
	}

	// Center the content
	return lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// handleScreenTransition processes screen transition messages
func (a *App) handleScreenTransition(msg ScreenTransitionMsg) (tea.Model, tea.Cmd) {
	switch msg.Target {
	case ScreenDetect:
		a.detect = NewDetectModel(a.width, a.height)
		a.screen = ScreenDetect
		return a, a.detect.Init()

	case ScreenComponentSelect:
		a.select_ = NewSelectModel(a.width, a.height)
		a.screen = ScreenComponentSelect
		return a, a.select_.Init()

	case ScreenBackup:
		// Save component selection before transitioning
		if a.select_ != nil {
			a.componentSelection = a.select_.GetSelection()
		}
		a.backup = NewBackupModel(
			a.configPath,
			a.existingConfig,
			a.width,
			a.height,
		)
		a.screen = ScreenBackup
		return a, a.backup.Init()

	case ScreenInstall:
		// Pass component selection to install model
		selection := a.componentSelection
		if selection == nil {
			selection = DefaultComponentSelection()
		}
		a.install = NewInstallModel(
			a.configPath,
			a.backupPath,
			a.existingConfig,
			selection,
			a.width,
			a.height,
		)
		a.screen = ScreenInstall
		return a, a.install.Init()

	case ScreenComplete:
		a.complete = NewCompleteModel(a.configPath, a.width, a.height)
		a.screen = ScreenComplete
		return a, a.complete.Init()
	}

	return a, nil
}

// SetDetectionResults updates the app state with detection results
func (a *App) SetDetectionResults(openCodeInstalled bool, version string, existingConfig, isGentlemanDots bool) {
	a.openCodeInstalled = openCodeInstalled
	a.openCodeVersion = version
	a.existingConfig = existingConfig
	a.isGentlemanDots = isGentlemanDots
}

// SetBackupPath sets the backup path
func (a *App) SetBackupPath(path string) {
	a.backupPath = path
}

// ===== Message Types =====

// ScreenTransitionMsg signals a screen transition
type ScreenTransitionMsg struct {
	Target Screen
}

// TransitionTo creates a screen transition command
func TransitionTo(screen Screen) tea.Cmd {
	return func() tea.Msg {
		return ScreenTransitionMsg{Target: screen}
	}
}

// ErrorMsg wraps an error for display
type ErrorMsg struct {
	Step    string
	Message string
	Err     error
}

// UndoCompleteMsg signals that the undo/rollback operation has completed
type UndoCompleteMsg struct{}

// NewErrorMsg creates an error message
func NewErrorMsg(step, message string, err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{Step: step, Message: message, Err: err}
	}
}

// ===== Tick messages for animations =====

// TickMsg is sent periodically for animations
type TickMsg time.Time

// Tick creates a tick command
func Tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// performUndo rolls back a partial installation and returns a quit command
func (a *App) performUndo() tea.Cmd {
	return func() tea.Msg {
		configPath := expandHomePath(a.configPath)

		// Try to restore from backup if one was made
		if a.backupPath != "" {
			backupPath := expandHomePath(a.backupPath)
			if _, err := os.Stat(backupPath); err == nil {
				// Remove partial install
				_ = os.RemoveAll(configPath)
				// Restore backup using the install package's CopyDirectory
				_ = installPkg.CopyDirectory(backupPath, configPath)
			}
		} else {
			// No backup — just remove the partial config
			_ = os.RemoveAll(configPath)
		}

		return UndoCompleteMsg{}
	}
}

// handleUndoComplete transitions to quit after undo finishes
func (a *App) handleUndoComplete() (tea.Model, tea.Cmd) {
	a.quitting = true
	return a, tea.Quit
}

// UpdateInstallStepTracking records completed/pending/failed steps from the install model
func (a *App) UpdateInstallStepTracking() {
	if a.install == nil {
		return
	}

	var completed []string
	var pending []string
	var failed string

	for _, step := range a.install.steps {
		switch step.Status {
		case StepComplete:
			completed = append(completed, step.Name)
		case StepPending:
			pending = append(pending, step.Name)
		case StepError:
			failed = step.Name
		}
	}

	a.completedSteps = completed
	a.pendingSteps = pending
	a.failedStep = failed
}

// Run starts the TUI application
func Run() error {
	app := NewApp()
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}

// Exit exits the program gracefully
func Exit(code int) {
	os.Exit(code)
}
