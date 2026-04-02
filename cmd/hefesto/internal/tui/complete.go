package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CompleteModel is the completion screen
type CompleteModel struct {
	width  int
	height int

	configPath string
}

// NewCompleteModel creates a new complete screen
func NewCompleteModel(configPath string, width, height int) *CompleteModel {
	return &CompleteModel{
		configPath: configPath,
	}
}

// Init implements tea.Model
func (m *CompleteModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *CompleteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

// View implements tea.Model
func (m *CompleteModel) View() string {
	var b strings.Builder

	b.WriteString("\n")

	// Success icon and message
	success := SuccessStyle.Render("✓ Hefesto installed successfully!")
	b.WriteString(CenterText(success, 60))
	b.WriteString("\n\n")

	// Summary box
	summary := []string{
		"Configuration installed to:",
		BoldStyle.Render(m.configPath),
		"",
		"What's included:",
		"  • AI agent configuration (AGENTS.md)",
		"  • 21 coding skills for various frameworks",
		"  • 5 SDD slash commands",
		"  • Engram persistent memory plugin",
		"  • Fuego/Forge theme",
	}

	boxContent := strings.Join(summary, "\n")
	box := BoxStyle.Render(boxContent)
	b.WriteString(CenterText(box, 60))
	b.WriteString("\n\n")

	// Next steps
	nextSteps := SubtitleStyle.Render("Next Steps")
	b.WriteString(CenterText(nextSteps, 60))
	b.WriteString("\n")

	steps := []string{
		fmt.Sprintf("  1. %s to start OpenCode", BoldStyle.Render("Run `opencode`")),
		fmt.Sprintf("  2. %s", BoldStyle.Render("Configure your API key with `opencode providers`")),
		fmt.Sprintf("  3. %s", BoldStyle.Render("Start coding with AI assistance!")),
	}

	for _, step := range steps {
		b.WriteString(CenterText(BodyStyle.Render(step), 60))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Exit instruction
	exitMsg := MutedStyle.Render("Press q to exit")
	b.WriteString(CenterText(exitMsg, 60))

	// Fire emoji
	b.WriteString("\n")
	fire := lipgloss.NewStyle().Foreground(Primary).Render("🔥")
	b.WriteString(CenterText(fire+" Happy coding! "+fire, 60))

	return b.String()
}
