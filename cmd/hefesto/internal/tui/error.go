package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
	width := ResolveContentWidth(m.width)

	var b strings.Builder

	// Error hero — no border, clean impact
	b.WriteString(RenderCenteredHero("", "Installation Failed", "Something went wrong during setup", width))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	// Error details
	if m.err != nil {
		// Step
		stepLabel := CopperText("Step:")
		stepValue := WhiteText(m.err.Step)
		b.WriteString(CenterText(fmt.Sprintf("  %s %s", stepLabel, stepValue), width))
		b.WriteString("\n")

		// Error message
		errMsg := m.err.Message
		if m.err.Err != nil {
			errMsg = fmt.Sprintf("%s: %v", m.err.Message, m.err.Err)
		}
		// Wrap long error messages
		wrapped := wrapText(errMsg, width-8)
		for _, line := range wrapped {
			b.WriteString(CenterText(RedText("  "+line), width))
			b.WriteString("\n")
		}
	}

	b.WriteString(strings.Repeat("\n", SpaceSM))

	// Suggested fixes
	b.WriteString(CenterText(CopperText("Suggested fixes:"), width))
	b.WriteString("\n")

	fixes := m.getSuggestedFixes()
	for _, fix := range fixes {
		b.WriteString(CenterText("  "+GrayText(fix), width))
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("\n", SpaceSM))

	// Partial installation status
	if len(m.CompletedSteps) > 0 || len(m.PendingSteps) > 0 {
		b.WriteString(CenterText(CopperText("Partial installation status:"), width))
		b.WriteString("\n")

		for _, step := range m.CompletedSteps {
			b.WriteString(CenterText(fmt.Sprintf("  %s %s", GreenText(IconCheck), WhiteText(step)), width))
			b.WriteString("\n")
		}

		if m.FailedStep != "" {
			b.WriteString(CenterText(fmt.Sprintf("  %s %s", RedText(IconCross), WhiteText(m.FailedStep)), width))
			b.WriteString("\n")
		} else if m.err != nil {
			b.WriteString(CenterText(fmt.Sprintf("  %s %s", RedText(IconCross), WhiteText(m.err.Step)), width))
			b.WriteString("\n")
		}

		for _, step := range m.PendingSteps {
			b.WriteString(CenterText(fmt.Sprintf("  %s %s", DimTextStyle.Render("○"), GrayText(step)), width))
			b.WriteString("\n")
		}

		b.WriteString(strings.Repeat("\n", SpaceSM))
	}

	// Help bar with options
	hints := []KeyHint{
		{Key: "r", Action: "Retry"},
		{Key: "u", Action: "Undo"},
		{Key: "q", Action: "Quit"},
	}
	b.WriteString(RenderHelpBar(hints, width))

	return RenderScreenFrame(b.String(), FrameOptions{
		Width:  m.width,
		Height: m.height,
		Border: BorderRounded,
	})
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
