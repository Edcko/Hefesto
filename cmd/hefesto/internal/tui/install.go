package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InstallStep represents a single installation step
type InstallStep struct {
	Name      string
	Status    StepStatus
	Message   string
	StartTime time.Time
	EndTime   time.Time
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
	spinner  int

	// Progress tracking
	startTime time.Time
}

// StepCompleteMsg signals a step has completed
type StepCompleteMsg struct {
	Index   int
	Success bool
	Message string
	Error   error
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
			{Name: "Creating config directory", Status: StepPending},
			{Name: "Copying Hefesto configuration", Status: StepPending},
			{Name: "Installing dependencies", Status: StepPending},
			{Name: "Configuring API provider", Status: StepPending},
			{Name: "Verifying installation", Status: StepPending},
		},
		startTime: time.Now(),
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
		// This is a stub - the actual implementation will be in internal/install/
		// For now, we simulate the steps with delays
		time.Sleep(500 * time.Millisecond)

		// Simulate success for now
		// In real implementation, this would call installer functions
		return StepCompleteMsg{
			Index:   index,
			Success: true,
			Message: "Completed",
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
		m.spinner = (m.spinner + 1) % len(IconSpinner)
		if !m.complete {
			return m, Tick(100 * time.Millisecond)
		}
		return m, nil

	case StepCompleteMsg:
		if msg.Success {
			m.steps[msg.Index].Status = StepComplete
			m.steps[msg.Index].Message = msg.Message
			m.steps[msg.Index].EndTime = time.Now()

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

	// Title
	title := TitleStyle.Render("Installing Hefesto")
	b.WriteString(CenterText(title, 60))
	b.WriteString("\n\n")

	// Progress bar
	progress := m.current
	if m.complete {
		progress = len(m.steps)
	}
	progressBar := ProgressBar(40, progress, len(m.steps))
	b.WriteString(CenterText(progressBar, 60))

	percent := float64(progress) / float64(len(m.steps)) * 100
	percentStr := fmt.Sprintf("%.0f%%", percent)
	b.WriteString(CenterText(MutedStyle.Render(percentStr), 60))
	b.WriteString("\n\n")

	// Steps
	for i, step := range m.steps {
		b.WriteString(m.renderStep(i, step))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Elapsed time
	elapsed := time.Since(m.startTime)
	timeStr := fmt.Sprintf("Elapsed: %s", elapsed.Round(time.Second))
	b.WriteString(CenterText(MutedStyle.Render(timeStr), 60))

	return b.String()
}

// renderStep renders a single installation step
func (m *InstallModel) renderStep(index int, step InstallStep) string {
	var icon string
	var nameStyle lipgloss.Style

	switch step.Status {
	case StepPending:
		icon = MutedStyle.Render("○")
		nameStyle = MutedStyle
	case StepRunning:
		spinnerChar := string(IconSpinner[m.spinner])
		icon = InfoStyle.Render(spinnerChar)
		nameStyle = BodyStyle
	case StepComplete:
		icon = SuccessStyle.Render(IconCheck)
		nameStyle = BodyStyle
	case StepError:
		icon = ErrorStyle.Render(IconCross)
		nameStyle = ErrorStyle
	}

	name := nameStyle.Render(step.Name)

	line := fmt.Sprintf("  %s %s", icon, name)

	// Add message if present
	if step.Message != "" && step.Status != StepPending {
		msg := MutedStyle.Render(fmt.Sprintf("(%s)", step.Message))
		line = fmt.Sprintf("%s %s", line, msg)
	}

	return CenterText(line, 60)
}
