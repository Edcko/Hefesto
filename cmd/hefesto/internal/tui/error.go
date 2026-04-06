package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Colors for the error screen
var (
	colorErrorRed      = lipgloss.Color("#FF4444")
	colorWarningYellow = lipgloss.Color("#EAB308")
	colorGray          = lipgloss.Color("#666666")
)

// ErrorAction represents user actions on the error screen
type ErrorAction int

const (
	ErrorActionNone ErrorAction = iota
	ErrorActionRetry
	ErrorActionUndo
	ErrorActionQuit
)

// ErrorModel is the error screen shown when installation fails
type ErrorModel struct {
	width  int
	height int

	err *InstallError

	// StepTracking for showing partial installation progress
	CompletedSteps []string
	PendingSteps   []string
	FailedStep     string

	// User action callback
	OnAction func(ErrorAction) tea.Cmd
}

// NewErrorModel creates a new error screen
func NewErrorModel(err *InstallError) *ErrorModel {
	return &ErrorModel{
		err:            err,
		CompletedSteps: []string{},
		PendingSteps: []string{
			"Install Engram",
			"Install dependencies",
			"Verify installation",
		},
	}
}

// SetSteps sets the completed and pending steps for partial install display
func (m *ErrorModel) SetSteps(completed, pending []string, failed string) {
	m.CompletedSteps = completed
	m.PendingSteps = pending
	m.FailedStep = failed
}

// Init implements tea.Model
func (m *ErrorModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *ErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			if m.OnAction != nil {
				return m, m.OnAction(ErrorActionRetry)
			}
			return m, tea.Quit
		case "u":
			if m.OnAction != nil {
				return m, m.OnAction(ErrorActionUndo)
			}
			return m, tea.Quit
		case "q", "ctrl+c":
			if m.OnAction != nil {
				return m, m.OnAction(ErrorActionQuit)
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *ErrorModel) View() string {
	// Box dimensions
	boxWidth := 54

	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(colorErrorRed).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().
		Foreground(colorErrorRed).
		Bold(true)

	errorStyle := lipgloss.NewStyle().
		Foreground(colorErrorRed)

	warningStyle := lipgloss.NewStyle().
		Foreground(colorWarningYellow)

	checkStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#22C55E"))

	grayStyle := lipgloss.NewStyle().
		Foreground(colorGray)

	amberStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF8C00")).
		Bold(true)

	// Build content lines
	var lines []string

	// Empty line at top
	lines = append(lines, "")

	// Title with X emojis
	title := "    ❌  Installation Failed  ❌    "
	lines = append(lines, titleStyle.Render(title))

	// Empty line
	lines = append(lines, "")

	// Error details
	if m.err != nil {
		// Step
		stepLine := fmt.Sprintf("  Step: %s", m.err.Step)
		lines = append(lines, stepLine)

		// Error message
		errMsg := m.err.Message
		if m.err.Err != nil {
			errMsg = fmt.Sprintf("%s: %v", m.err.Message, m.err.Err)
		}
		// Wrap long error messages
		wrapped := wrapText(errMsg, 46)
		for _, line := range wrapped {
			lines = append(lines, "  Error: "+errorStyle.Render(line))
		}
	}

	// Empty line
	lines = append(lines, "")

	// Suggested fixes
	lines = append(lines, "  "+warningStyle.Render("Suggested fixes:"))
	fixes := m.getSuggestedFixes()
	for _, fix := range fixes {
		lines = append(lines, "    "+fix)
	}

	// Empty line
	lines = append(lines, "")

	// Partial installation status (if any steps completed)
	if len(m.CompletedSteps) > 0 || len(m.PendingSteps) > 0 {
		lines = append(lines, "  Partial installation detected:")

		// Completed steps
		for _, step := range m.CompletedSteps {
			lines = append(lines, "    "+checkStyle.Render("✅")+" "+step)
		}

		// Failed step
		if m.FailedStep != "" {
			lines = append(lines, "    "+errorStyle.Render("❌")+" "+m.FailedStep)
		} else if m.err != nil {
			lines = append(lines, "    "+errorStyle.Render("❌")+" "+m.err.Step)
		}

		// Pending steps
		for _, step := range m.PendingSteps {
			lines = append(lines, "    "+grayStyle.Render("⏳")+" "+grayStyle.Render(step))
		}

		// Empty line
		lines = append(lines, "")
	}

	// Options
	optionsLine := fmt.Sprintf("  Options: [%s] %s  [%s] %s  [%s] %s",
		amberStyle.Render("r"), "Retry",
		amberStyle.Render("u"), "Undo partial install",
		amberStyle.Render("q"), "Quit")
	lines = append(lines, optionsLine)

	// Empty line at bottom
	lines = append(lines, "")

	// Build the box content
	content := strings.Join(lines, "\n")

	// Apply border
	box := borderStyle.
		Width(boxWidth).
		Render(content)

	// Center the box
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(box)
}

// getSuggestedFixes returns context-aware fix suggestions based on the error
func (m *ErrorModel) getSuggestedFixes() []string {
	if m.err == nil {
		return []string{
			"1. Check directory permissions",
			"   chmod 755 ~/.config/opencode/",
			"2. Run with appropriate permissions",
			"3. Check disk space",
		}
	}

	var fixes []string

	// Context-aware suggestions based on error
	errMsg := m.err.Message
	if m.err.Err != nil {
		errMsg = m.err.Err.Error()
	}

	switch {
	case strings.Contains(errMsg, "permission denied"):
		fixes = []string{
			"1. Check directory permissions",
			"   chmod 755 ~/.config/opencode/",
			"2. Run with appropriate permissions",
			"3. Check disk space",
		}
	case strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection"):
		fixes = []string{
			"1. Check your internet connection",
			"2. Try again in a few moments",
			"3. Check if a firewall is blocking",
		}
	case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "no such file"):
		fixes = []string{
			"1. Ensure OpenCode is installed",
			"2. Run 'opencode --version' to verify",
			"3. Reinstall OpenCode if needed",
		}
	case strings.Contains(errMsg, "npm") || strings.Contains(errMsg, "node"):
		fixes = []string{
			"1. Install Node.js and npm",
			"2. Run 'node --version' to verify",
			"3. Update npm: npm install -g npm",
		}
	default:
		fixes = []string{
			"1. Check directory permissions",
			"   chmod 755 ~/.config/opencode/",
			"2. Run with appropriate permissions",
			"3. Check disk space",
		}
	}

	return fixes
}

// wrapText wraps text to a maximum width
func wrapText(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 > maxWidth {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		} else {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	if len(lines) == 0 {
		return []string{text}
	}

	return lines
}
