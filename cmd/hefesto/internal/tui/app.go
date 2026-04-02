package tui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen represents the current screen in the TUI
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenDetect
	ScreenBackup
	ScreenInstall
	ScreenComplete
	ScreenError
)

// String returns the screen name
func (s Screen) String() string {
	return [...]string{"welcome", "detect", "backup", "install", "complete", "error"}[s]
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
	configPath        string
	backupPath        string
	openCodeVersion   string
	openCodeInstalled bool
	existingConfig    bool
	isGentlemanDots   bool

	// Current screen models
	welcome  *WelcomeModel
	detect   *DetectModel
	backup   *BackupModel
	install  *InstallModel
	complete *CompleteModel
	err      *ErrorModel

	// Error state
	lastError *InstallError

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
		case "ctrl+c", "q":
			if a.screen == ScreenComplete || a.screen == ScreenError {
				a.quitting = true
				return a, tea.Quit
			}
			if a.screen == ScreenWelcome {
				a.quitting = true
				return a, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Propagate to all screens - ignore commands for resize
		if a.welcome != nil {
			if m, _ := a.welcome.Update(msg); m != nil {
				a.welcome = m.(*WelcomeModel)
			}
		}
		if a.detect != nil {
			if m, _ := a.detect.Update(msg); m != nil {
				a.detect = m.(*DetectModel)
			}
		}
		if a.backup != nil {
			if m, _ := a.backup.Update(msg); m != nil {
				a.backup = m.(*BackupModel)
			}
		}
		if a.install != nil {
			if m, _ := a.install.Update(msg); m != nil {
				a.install = m.(*InstallModel)
			}
		}
		if a.complete != nil {
			if m, _ := a.complete.Update(msg); m != nil {
				a.complete = m.(*CompleteModel)
			}
		}
		if a.err != nil {
			if m, _ := a.err.Update(msg); m != nil {
				a.err = m.(*ErrorModel)
			}
		}

	// Screen transition messages
	case ScreenTransitionMsg:
		return a.handleScreenTransition(msg)

	case ErrorMsg:
		a.lastError = &InstallError{
			Step:    msg.Step,
			Message: msg.Message,
			Err:     msg.Err,
		}
		a.err = NewErrorModel(a.lastError)
		a.screen = ScreenError
		return a, a.err.Init()
	}

	// Update current screen
	switch a.screen {
	case ScreenWelcome:
		if a.welcome != nil {
			if m, cmd := a.welcome.Update(msg); m != nil {
				a.welcome = m.(*WelcomeModel)
				cmds = append(cmds, cmd)
			}
		}

	case ScreenDetect:
		if a.detect != nil {
			if m, cmd := a.detect.Update(msg); m != nil {
				a.detect = m.(*DetectModel)
				cmds = append(cmds, cmd)
			}
		}

	case ScreenBackup:
		if a.backup != nil {
			if m, cmd := a.backup.Update(msg); m != nil {
				a.backup = m.(*BackupModel)
				cmds = append(cmds, cmd)
			}
		}

	case ScreenInstall:
		if a.install != nil {
			if m, cmd := a.install.Update(msg); m != nil {
				a.install = m.(*InstallModel)
				cmds = append(cmds, cmd)
			}
		}

	case ScreenComplete:
		if a.complete != nil {
			if m, cmd := a.complete.Update(msg); m != nil {
				a.complete = m.(*CompleteModel)
				cmds = append(cmds, cmd)
			}
		}

	case ScreenError:
		if a.err != nil {
			if m, cmd := a.err.Update(msg); m != nil {
				a.err = m.(*ErrorModel)
				cmds = append(cmds, cmd)
			}
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

	case ScreenBackup:
		a.backup = NewBackupModel(
			a.configPath,
			a.existingConfig,
			a.width,
			a.height,
		)
		a.screen = ScreenBackup
		return a, a.backup.Init()

	case ScreenInstall:
		a.install = NewInstallModel(
			a.configPath,
			a.backupPath,
			a.existingConfig,
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
