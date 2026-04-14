package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// WelcomeModel is the welcome screen model.
type WelcomeModel struct {
	width  int
	height int
}

// NewWelcomeModel creates a new welcome screen.
func NewWelcomeModel() *WelcomeModel {
	return &WelcomeModel{}
}

// Init implements tea.Model.
func (m *WelcomeModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m *WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			return m, TransitionTo(ScreenDetect)
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m *WelcomeModel) View() string {
	width := ResolveContentWidth(m.width)

	// ===== Hero section: banner + title + subtitle =====
	hero := RenderCenteredHero(BannerAnvil, "HEFESTO "+Version, "AI Dev Environment Forge", width)

	// ===== Description =====
	desc := CenterText(WhiteText("Hefesto will configure your OpenCode environment"), width) + "\n" +
		CenterText(WhiteText("with agents, skills, themes and persistence."), width)

	// ===== What will be installed =====
	installHeader := CenterText(CopperText("What will be installed:"), width)

	items := []string{
		"25 AI skills (Angular, React, SDD...)",
		"SDD orchestrator + 6 phase agents",
		"Fuego/Forge theme (amber/copper)",
		"Engram persistent memory",
		"Background agents plugin",
	}

	var itemLines strings.Builder
	for _, item := range items {
		itemLines.WriteString(CenterText(BulletItem(WhiteText(item)), width))
		itemLines.WriteString("\n")
	}

	// ===== Help bar =====
	helpBar := RenderHelpBar([]KeyHint{
		{Key: "Enter", Action: "Continue"},
	}, width)

	// ===== Assemble with consistent spacing =====
	var b strings.Builder
	b.WriteString(strings.Repeat("\n", SpaceLG))
	b.WriteString(hero)
	b.WriteString(strings.Repeat("\n", SpaceSM))
	b.WriteString(desc)
	b.WriteString(strings.Repeat("\n", SpaceMD))
	b.WriteString(installHeader)
	b.WriteString("\n")
	b.WriteString(itemLines.String())
	b.WriteString(strings.Repeat("\n", SpaceSM))
	b.WriteString(helpBar)
	b.WriteString(strings.Repeat("\n", SpaceLG))

	content := b.String()

	// Center in terminal — splash screen, no border.
	return RenderScreenFrame(content, FrameOptions{
		Width:  m.width,
		Height: m.height,
		Border: BorderNone,
	})
}
