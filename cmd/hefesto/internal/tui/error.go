package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ErrorModel is the error screen
type ErrorModel struct {
	width  int
	height int

	err *InstallError
}

// NewErrorModel creates a new error screen
func NewErrorModel(err *InstallError) *ErrorModel {
	return &ErrorModel{
		err: err,
	}
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
	}

	return m, nil
}

// View implements tea.Model
func (m *ErrorModel) View() string {
	var b strings.Builder

	b.WriteString("\n")

	// Error icon and message
	title := ErrorStyle.Render("✗ Installation Failed")
	b.WriteString(CenterText(title, 60))
	b.WriteString("\n\n")

	// Error details
	if m.err != nil {
		step := fmt.Sprintf("Step: %s", m.err.Step)
		b.WriteString(CenterText(BoldStyle.Render(step), 60))
		b.WriteString("\n\n")

		// Error message
		msg := m.err.Message
		if m.err.Err != nil {
			msg = fmt.Sprintf("%s: %v", msg, m.err.Err)
		}

		// Wrap long error messages
		lines := wrapText(msg, 50)
		for _, line := range lines {
			b.WriteString(CenterText(ErrorStyle.Render(line), 60))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Help text
	help := []string{
		"Please check the following:",
		"  • You have write permissions to ~/.config/",
		"  • You have an active internet connection",
		"  • Node.js and npm are installed",
	}

	for _, line := range help {
		b.WriteString(CenterText(BodyStyle.Render(line), 60))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Exit instruction
	exitMsg := MutedStyle.Render("Press q to exit")
	b.WriteString(CenterText(exitMsg, 60))

	return b.String()
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

	return lines
}
