package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Edcko/Hefesto/cmd/hefesto/internal/embed"
	"github.com/Edcko/Hefesto/cmd/hefesto/internal/install"
)

// InstallStep represents a single installation step
type InstallStep struct {
	Name      string
	Status    StepStatus
	Message   string
	StartTime time.Time
	EndTime   time.Time
	Progress  float64 // 0.0 to 1.0 for progress tracking
	Detail    string  // Current file/operation being processed
}

// StepStatus represents the status of an install step
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepComplete
	StepError
)

// String returns the string representation
func (s StepStatus) String() string {
	return [...]string{"pending", "running", "complete", "error"}[s]
}

// InstallModel is the installation progress screen
type InstallModel struct {
	width  int
	height int

	configPath     string
	backupPath     string
	existingConfig bool

	// Component selection — controls what gets installed
	selection *ComponentSelection

	// Installation steps
	steps    []InstallStep
	current  int
	complete bool

	// Spinner animation
	spinnerIndex int
	spinnerFrame string

	// Progress tracking
	startTime time.Time

	// Real installer components
	env *install.Environment

	// Installation results for final report
	verifyResult *install.VerifyResult
}

// StepCompleteMsg signals a step has completed
type StepCompleteMsg struct {
	Index   int
	Success bool
	Message string
	Error   error
}

// StepProgressMsg signals progress update during a step
type StepProgressMsg struct {
	StepIndex int
	Progress  float64 // 0.0 to 1.0
	Detail    string  // current file/operation being processed
}

// InstallCompleteMsg signals all installation is complete
type InstallCompleteMsg struct {
	Success bool
}

// NewInstallModel creates a new install screen
func NewInstallModel(configPath, backupPath string, existingConfig bool, selection *ComponentSelection, width, height int) *InstallModel {
	if selection == nil {
		selection = DefaultComponentSelection()
	}

	// Build steps based on component selection
	steps := []InstallStep{
		{Name: "Detect environment", Status: StepPending},
	}

	if existingConfig {
		steps = append(steps, InstallStep{Name: "Backup existing config", Status: StepPending})
	}

	if selection.IsSelected(ComponentSkills) || selection.IsSelected(ComponentTheme) ||
		selection.IsSelected(ComponentPersonality) || selection.IsSelected(ComponentCommands) ||
		selection.IsSelected(ComponentPlugins) || selection.IsSelected(ComponentAgents) ||
		selection.IsSelected(ComponentOpenCode) {
		steps = append(steps, InstallStep{Name: "Copying configuration", Status: StepPending})
	}

	if selection.IsSelected(ComponentEngram) {
		steps = append(steps, InstallStep{Name: "Install Engram", Status: StepPending})
	}

	if selection.IsSelected(ComponentPlugins) {
		steps = append(steps, InstallStep{Name: "Install dependencies", Status: StepPending})
	}

	steps = append(steps, InstallStep{Name: "Verify installation", Status: StepPending})

	return &InstallModel{
		configPath:     configPath,
		backupPath:     backupPath,
		existingConfig: existingConfig,
		selection:      selection,
		steps:          steps,
		startTime:      time.Now(),
		spinnerFrame:   "⠋",
		spinnerIndex:   0,
	}
}

// Init implements tea.Model
func (m *InstallModel) Init() tea.Cmd {
	return tea.Batch(
		Tick(100*time.Millisecond),
		m.runStep(0),
	)
}

// runStep executes a single installation step
func (m *InstallModel) runStep(index int) tea.Cmd {
	return func() tea.Msg {
		configPath := expandHomePath(m.configPath)

		var message string

		stepName := m.steps[index].Name

		switch stepName {
		case "Detect environment":
			m.steps[index].StartTime = time.Now()
			env, detectErr := install.Detect()
			if detectErr != nil {
				return StepCompleteMsg{
					Index:   index,
					Success: false,
					Message: fmt.Sprintf("Failed to detect environment: %v", detectErr),
					Error:   detectErr,
				}
			}
			m.env = env
			message = fmt.Sprintf("Detected %s/%s, OpenCode %s", env.Platform, env.Arch, env.OpenCodeVersion)

		case "Backup existing config":
			m.steps[index].StartTime = time.Now()
			if m.env != nil && m.env.ConfigExists && m.env.ExistingConfig != "none" {
				backupPath, backupErr := install.Backup(configPath)
				if backupErr != nil {
					// Non-fatal - continue even if backup fails
					message = fmt.Sprintf("Backup skipped: %v", backupErr)
				} else {
					m.backupPath = backupPath
					message = fmt.Sprintf("Backup created: %s", backupPath)
				}
			} else {
				message = "No existing config to backup"
			}

		case "Copying configuration":
			m.steps[index].StartTime = time.Now()
			if err := install.CopyConfig(embed.ConfigFiles, configPath); err != nil {
				return StepCompleteMsg{
					Index:   index,
					Success: false,
					Message: fmt.Sprintf("Failed to copy configuration: %v", err),
					Error:   err,
				}
			}
			message = "Configuration copied successfully"

		case "Install Engram":
			m.steps[index].StartTime = time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// Check if engram is already installed
			env, detectErr := install.Detect()
			if detectErr != nil {
				return StepCompleteMsg{
					Index:   index,
					Success: false,
					Message: fmt.Sprintf("Failed to detect environment: %v", detectErr),
					Error:   detectErr,
				}
			}

			if env.EngramInstalled {
				message = fmt.Sprintf("Engram already installed (%s)", env.EngramVersion)
			} else {
				// Install engram
				if installErr := install.InstallEngram(ctx); installErr != nil {
					// Non-fatal - continue even if engram install fails
					message = fmt.Sprintf("Engram install skipped: %v", installErr)
				} else {
					message = "Engram installed successfully"
				}
			}

		case "Install dependencies":
			m.steps[index].StartTime = time.Now()
			// npm install is non-fatal - continue even if it fails
			if err := install.NpmInstall(configPath); err != nil {
				// Log the warning but don't fail the install
				message = fmt.Sprintf("npm install skipped: %v", err)
			} else {
				message = "Dependencies installed successfully"
			}

		case "Verify installation":
			m.steps[index].StartTime = time.Now()
			result, err := install.Verify(configPath)
			if err != nil {
				return StepCompleteMsg{
					Index:   index,
					Success: false,
					Message: fmt.Sprintf("Verification failed: %v", err),
					Error:   err,
				}
			}

			// Check critical verifications
			if !result.ConfigCopied {
				err := fmt.Errorf("config files not properly installed")
				return StepCompleteMsg{
					Index:   index,
					Success: false,
					Message: "Configuration verification failed",
					Error:   err,
				}
			}

			m.verifyResult = result
			message = fmt.Sprintf("Verified (Config: %v, NPM: %v, OpenCode: %v)",
				result.ConfigCopied, result.NpmInstalled, result.OpenCodeWorks)

		default:
			return StepCompleteMsg{
				Index:   index,
				Success: false,
				Message: fmt.Sprintf("Unknown step: %s", stepName),
				Error:   fmt.Errorf("unknown step: %s", stepName),
			}
		}

		return StepCompleteMsg{
			Index:   index,
			Success: true,
			Message: message,
		}
	}
}

// Update implements tea.Model
func (m *InstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		// Update spinner animation
		spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		m.spinnerIndex = (m.spinnerIndex + 1) % len(spinnerFrames)
		m.spinnerFrame = spinnerFrames[m.spinnerIndex]
		if !m.complete {
			return m, Tick(100 * time.Millisecond)
		}
		return m, nil

	case StepProgressMsg:
		// Update progress for the step
		if msg.StepIndex >= 0 && msg.StepIndex < len(m.steps) {
			m.steps[msg.StepIndex].Progress = msg.Progress
			m.steps[msg.StepIndex].Detail = msg.Detail
		}
		return m, nil

	case StepCompleteMsg:
		if msg.Success {
			m.steps[msg.Index].Status = StepComplete
			m.steps[msg.Index].Message = msg.Message
			m.steps[msg.Index].EndTime = time.Now()
			m.steps[msg.Index].Progress = 1.0

			// Move to next step
			m.current++

			if m.current >= len(m.steps) {
				// All steps complete
				m.complete = true
				return m, func() tea.Msg {
					time.Sleep(500 * time.Millisecond)
					return InstallCompleteMsg{Success: true}
				}
			}

			// Start next step
			m.steps[m.current].Status = StepRunning
			m.steps[m.current].StartTime = time.Now()
			return m, m.runStep(m.current)
		}

		// Step failed
		m.steps[msg.Index].Status = StepError
		m.steps[msg.Index].Message = msg.Message
		m.steps[msg.Index].EndTime = time.Now()
		return m, NewErrorMsg(m.steps[msg.Index].Name, msg.Message, msg.Error)

	case InstallCompleteMsg:
		if msg.Success {
			return m, TransitionTo(ScreenComplete)
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *InstallModel) View() string {
	width := ResolveContentWidth(m.width)

	// Calculate inner content width accounting for frame border + padding.
	// RoundedBorder adds 2 chars (1 per side), PadBox adds 4 chars (2 per side) = 6 total.
	innerWidth := width - (PadBox*2 + 2)
	if innerWidth < 20 {
		innerWidth = 20
	}

	// Wizard progress: Welcome → Detect → Select → Backup → [Install] → Complete
	wizardSteps := []WizardStep{
		{Label: "Welcome", Done: true},
		{Label: "Detect", Done: true},
		{Label: "Select", Done: true},
		{Label: "Backup", Done: true},
		{Label: "Install", Active: true},
		{Label: "Complete"},
	}

	var b strings.Builder

	// Wizard progress indicator
	b.WriteString(RenderWizardProgress(wizardSteps, innerWidth))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	// Section title
	b.WriteString(RenderSectionTitle("Installing", innerWidth))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	// Step list
	for i, step := range m.steps {
		b.WriteString(m.renderStep(i, step, innerWidth))
	}

	// Current step detail line
	detail := m.getCurrentDetail()
	b.WriteString(strings.Repeat("\n", SpaceXS))
	b.WriteString(CenterText(GrayText(detail), innerWidth))

	return RenderScreenFrame(b.String(), FrameOptions{
		Width:  m.width,
		Height: m.height,
		Border: BorderRounded,
	})
}

// renderStep renders a single step with icon and status
func (m *InstallModel) renderStep(index int, step InstallStep, width int) string {
	var icon string
	var nameStyle lipgloss.Style

	switch step.Status {
	case StepPending:
		icon = DimTextStyle.Render("○")
		nameStyle = DimTextStyle
	case StepRunning:
		icon = AmberText(m.spinnerFrame)
		nameStyle = lipgloss.NewStyle().Foreground(ColorAmber)
	case StepComplete:
		icon = GreenText(IconCheck)
		nameStyle = lipgloss.NewStyle().Foreground(ColorGreen)
	case StepError:
		icon = RedText(IconCross)
		nameStyle = lipgloss.NewStyle().Foreground(ColorRed)
	}

	stepLine := fmt.Sprintf("  %s %s", icon, nameStyle.Render(step.Name))

	// Add timing for completed steps
	if step.Status == StepComplete && !step.EndTime.IsZero() && !step.StartTime.IsZero() {
		duration := step.EndTime.Sub(step.StartTime)
		timing := fmt.Sprintf("(%.1fs)", duration.Seconds())
		stepLine += " " + GrayText(timing)
	}

	var lines []string
	lines = append(lines, CenterText(stepLine, width))

	// Progress bar for running copy step
	if step.Status == StepRunning && step.Name == "Copying configuration" {
		// Scale progress bar to fit within the available inner width.
		// Reserve space for: " " + "100%" + " " + detail(~15) = ~21 chars
		progressBarWidth := width - 21
		if progressBarWidth > 20 {
			progressBarWidth = 20
		}
		if progressBarWidth < 5 {
			progressBarWidth = 5
		}

		progressBar := renderProgressBar(progressBarWidth, step.Progress)
		percent := int(step.Progress * 100)

		// Style bar and percent separately in amber, detail in gray.
		// Avoids double-styling: the old code wrapped everything in
		// Foreground(ColorAmber) including the already-gray detail text.
		progressText := fmt.Sprintf("%s %s", AmberText(progressBar), AmberText(fmt.Sprintf("%d%%", percent)))
		if step.Detail != "" {
			detail := step.Detail
			runes := []rune(detail)
			if len(runes) > 15 {
				detail = "..." + string(runes[len(runes)-12:])
			}
			progressText += " " + GrayText(detail)
		}

		lines = append(lines, CenterText(progressText, width))
	}

	return strings.Join(lines, "\n") + "\n"
}

// getCurrentDetail returns the detail text for the current operation
func (m *InstallModel) getCurrentDetail() string {
	if m.current < 0 || m.current >= len(m.steps) {
		return "Preparing..."
	}

	step := m.steps[m.current]

	if step.Status != StepRunning {
		return "Preparing next step..."
	}

	// Return step-specific detail based on name
	switch step.Name {
	case "Detect environment":
		return "Detecting system environment..."
	case "Backup existing config":
		if m.env != nil && m.env.ConfigExists {
			return "Creating backup of existing configuration..."
		}
		return "No backup needed"
	case "Copying configuration":
		if step.Detail != "" {
			return fmt.Sprintf("Copying: %s", step.Detail)
		}
		return "Copying configuration files..."
	case "Install Engram":
		return "Installing Engram for persistent memory..."
	case "Install dependencies":
		return "Running npm install..."
	case "Verify installation":
		return "Verifying installation integrity..."
	default:
		return "Processing..."
	}
}

// renderProgressBar renders a progress bar with the given width and progress
func renderProgressBar(width int, progress float64) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filledWidth := int(float64(width) * progress)

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", width-filledWidth)

	return filled + empty
}

// expandHomePath expands ~ to the user's home directory
func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, strings.TrimPrefix(path, "~/"))
	}
	return path
}
