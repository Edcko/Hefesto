package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WelcomeModel is the welcome screen model
type WelcomeModel struct {
	width  int
	height int
}

// NewWelcomeModel creates a new welcome screen
func NewWelcomeModel() *WelcomeModel {
	return &WelcomeModel{}
}

// Init implements tea.Model
func (m *WelcomeModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			// Proceed to detection screen
			return m, TransitionTo(ScreenDetect)
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *WelcomeModel) View() string {
	var b strings.Builder

	// Logo
	logoStyle := lipgloss.NewStyle().
		Foreground(Primary).
		SetString(Logo)

	b.WriteString(logoStyle.String())
	b.WriteString("\n")

	// Title
	title := TitleStyle.Render("Welcome to Hefesto")
	b.WriteString(CenterText(title, 50))
	b.WriteString("\n\n")

	// Description
	desc := BodyStyle.Render("AI-powered development environment configuration")
	b.WriteString(CenterText(desc, 50))
	b.WriteString("\n\n")

	// Tagline
	tagline := MutedStyle.Render("Forge your perfect coding setup")
	b.WriteString(CenterText(tagline, 50))
	b.WriteString("\n\n")

	// Instructions box
	instructions := []string{
		"Press " + BoldStyle.Render("Enter") + " to start installation",
		"Press " + BoldStyle.Render("q") + " to quit",
	}

	boxContent := strings.Join(instructions, "\n")
	box := BoxStyle.Render(boxContent)
	b.WriteString(CenterText(box, 50))
	b.WriteString("\n\n")

	// Version
	b.WriteString(CenterText(VersionStyle(), 50))

	return b.String()
}
