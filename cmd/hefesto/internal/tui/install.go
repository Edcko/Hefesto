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

// Step icons for visual state indicators
const (
	IconCompleted  = "✅"
	IconInProgress = "🔄"
	IconPending    = "⏳"
	IconError      = "❌"
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
func NewInstallModel(configPath, backupPath string, existingConfig bool, width, height int) *InstallModel {
	return &InstallModel{
		configPath:     configPath,
		backupPath:     backupPath,
		existingConfig: existingConfig,
		steps: []InstallStep{
			{Name: "Detect environment", Status: StepPending},
			{Name: "Backup existing config", Status: StepPending},
			{Name: "Copying configuration", Status: StepPending},
			{Name: "Install Engram", Status: StepPending},
			{Name: "Install dependencies", Status: StepPending},
			{Name: "Verify installation", Status: StepPending},
		},
		startTime:    time.Now(),
		spinnerFrame: "⠋",
		spinnerIndex: 0,
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

		switch index {
		case 0:
			// Step 1: Detect environment
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

		case 1:
			// Step 2: Backup existing config if present
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

		case 2:
			// Step 3: Copy Hefesto configuration from embedded files
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

		case 3:
			// Step 4: Configure API provider (Engram)
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

		case 4:
			// Step 5: Install dependencies (npm)
			m.steps[index].StartTime = time.Now()
			// npm install is non-fatal - continue even if it fails
			if err := install.NpmInstall(configPath); err != nil {
				// Log the warning but don't fail the install
				message = fmt.Sprintf("npm install skipped: %v", err)
			} else {
				message = "Dependencies installed successfully"
			}

		case 5:
			// Step 6: Verify installation
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
				Message: fmt.Sprintf("Unknown step index: %d", index),
				Error:   fmt.Errorf("unknown step index: %d", index),
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
	var b strings.Builder

	// Box width for layout
	boxWidth := 46

	// Top border
	b.WriteString("╭")
	b.WriteString(strings.Repeat("─", boxWidth))
	b.WriteString("╮\n")

	// Title
	title := "🔥 HEFESTO — Installing..."
	titleLine := fmt.Sprintf("│  %s", title)
	titleLine = titleLine + strings.Repeat(" ", boxWidth-len(titleLine)+2) + "│\n"
	b.WriteString(lipgloss.NewStyle().Foreground(Primary).Render(titleLine))

	// Empty line
	b.WriteString("│")
	b.WriteString(strings.Repeat(" ", boxWidth))
	b.WriteString("│\n")

	// Steps with visual indicators
	for i, step := range m.steps {
		b.WriteString(m.renderStepLine(i, step, boxWidth))
	}

	// Empty line
	b.WriteString("│")
	b.WriteString(strings.Repeat(" ", boxWidth))
	b.WriteString("│\n")

	// Separator line for current detail
	separator := strings.Repeat("━", boxWidth)
	b.WriteString("│")
	b.WriteString(lipgloss.NewStyle().Foreground(Secondary).Render(separator))
	b.WriteString("│\n")

	// Current step detail
	currentDetail := m.getCurrentDetail()
	detailLine := fmt.Sprintf("│  %s", currentDetail)
	detailLine = detailLine + strings.Repeat(" ", boxWidth-len(currentDetail)-2) + "│\n"
	b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(detailLine))

	// Bottom border
	b.WriteString("╰")
	b.WriteString(strings.Repeat("─", boxWidth))
	b.WriteString("╯\n")

	return CenterText(b.String(), 60)
}

// renderStepLine renders a single step line with icon, name, and status
func (m *InstallModel) renderStepLine(index int, step InstallStep, boxWidth int) string {
	var icon string
	var nameStyle lipgloss.Style

	switch step.Status {
	case StepPending:
		icon = IconPending
		nameStyle = lipgloss.NewStyle().Foreground(TextMuted)
	case StepRunning:
		icon = m.spinnerFrame
		nameStyle = lipgloss.NewStyle().Foreground(Primary)
	case StepComplete:
		icon = IconCompleted
		nameStyle = lipgloss.NewStyle().Foreground(Success)
	case StepError:
		icon = IconError
		nameStyle = lipgloss.NewStyle().Foreground(Error)
	}

	// Build the step line
	stepText := fmt.Sprintf("%s %s", icon, step.Name)
	styledStep := nameStyle.Render(stepText)

	// Add timing for completed steps
	if step.Status == StepComplete && !step.EndTime.IsZero() && !step.StartTime.IsZero() {
		duration := step.EndTime.Sub(step.StartTime)
		timing := fmt.Sprintf("(%.1fs)", duration.Seconds())
		styledStep = styledStep + " " + lipgloss.NewStyle().Foreground(TextMuted).Render(timing)
	}

	// Add progress bar for running steps (especially the copy step)
	if step.Status == StepRunning && index == 2 {
		return m.renderStepWithProgress(index, step, boxWidth, icon, nameStyle)
	}

	// Format the line with padding
	line := fmt.Sprintf("│  %s", styledStep)
	padding := boxWidth - len(line) + 2
	if padding > 0 {
		line = line + strings.Repeat(" ", padding)
	}
	line = line + "│\n"

	return line
}

// renderStepWithProgress renders a step with a progress bar (for copy step)
func (m *InstallModel) renderStepWithProgress(index int, step InstallStep, boxWidth int, icon string, nameStyle lipgloss.Style) string {
	var b strings.Builder

	// Step name line
	stepText := fmt.Sprintf("%s %s", icon, step.Name)
	styledStep := nameStyle.Render(stepText)
	line := fmt.Sprintf("│  %s", styledStep)
	padding := boxWidth - len(line) + 2
	if padding > 0 {
		line = line + strings.Repeat(" ", padding)
	}
	line = line + "│\n"
	b.WriteString(line)

	// Progress bar line
	progressBarWidth := 20
	progressBar := renderProgressBar(progressBarWidth, step.Progress)

	// Add percentage and detail
	percent := int(step.Progress * 100)
	progressText := fmt.Sprintf("     %s %d%%", progressBar, percent)

	// Add detail if available
	if step.Detail != "" {
		// Truncate detail if too long
		detail := step.Detail
		if len(detail) > 15 {
			detail = "..." + detail[len(detail)-12:]
		}
		progressText = fmt.Sprintf("     %s %d%% — %s", progressBar, percent, detail)
	}

	progressLine := lipgloss.NewStyle().Foreground(PrimaryDark).Render(progressText)
	fullLine := fmt.Sprintf("│%s", progressLine)

	// Pad to box width
	lineLen := len(progressText) + 1 // +1 for the │
	padding = boxWidth - lineLen + 1
	if padding > 0 {
		fullLine = fullLine + strings.Repeat(" ", padding)
	}
	fullLine = fullLine + "│\n"
	b.WriteString(fullLine)

	return b.String()
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

// getCurrentDetail returns the detail text for the current operation
func (m *InstallModel) getCurrentDetail() string {
	if m.current < 0 || m.current >= len(m.steps) {
		return "Preparing..."
	}

	step := m.steps[m.current]

	if step.Status != StepRunning {
		return "Preparing next step..."
	}

	// Return step-specific detail
	switch m.current {
	case 0:
		return "Detecting system environment..."
	case 1:
		if m.env != nil && m.env.ConfigExists {
			return "Creating backup of existing configuration..."
		}
		return "No backup needed"
	case 2:
		if step.Detail != "" {
			return fmt.Sprintf("Copying: %s", step.Detail)
		}
		return "Copying configuration files..."
	case 3:
		return "Installing Engram for persistent memory..."
	case 4:
		return "Running npm install..."
	case 5:
		return "Verifying installation integrity..."
	default:
		return "Processing..."
	}
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
