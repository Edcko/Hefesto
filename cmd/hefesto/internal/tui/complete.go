package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// CompleteModel is the completion screen shown after successful installation
type CompleteModel struct {
	width  int
	height int

	configPath string

	// InstallDuration is the total time taken for installation
	InstallDuration time.Duration

	// InstalledComponents tracks what was installed
	InstalledComponents []InstalledComponent
}

// InstalledComponent represents an installed item
type InstalledComponent struct {
	Name        string
	Description string
}

// NewCompleteModel creates a new complete screen
func NewCompleteModel(configPath string, width, height int) *CompleteModel {
	return &CompleteModel{
		configPath: configPath,
		InstalledComponents: []InstalledComponent{
			{Name: "Config files", Description: "(AGENTS.md, opencode.json)"},
			{Name: "25 AI skills", Description: "(Angular, React, SDD...)"},
			{Name: "6 SDD phase agents", Description: "(init→plan→spec→...)"},
			{Name: "Fuego/Forge theme", Description: "(amber/copper)"},
			{Name: "Engram", Description: "(persistent memory)"},
			{Name: "Background agents", Description: "(async delegation)"},
			{Name: "Plugins", Description: "(engram + background)"},
		},
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

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *CompleteModel) View() string {
	width := ResolveContentWidth(m.width)

	// Wizard progress: all done
	wizardSteps := []WizardStep{
		{Label: "Welcome", Done: true},
		{Label: "Detect", Done: true},
		{Label: "Select", Done: true},
		{Label: "Backup", Done: true},
		{Label: "Install", Done: true},
		{Label: "Complete", Done: true},
	}

	var b strings.Builder

	// Wizard progress — all green
	b.WriteString(RenderWizardProgress(wizardSteps, width))
	b.WriteString(strings.Repeat("\n", SpaceLG))

	// Hero section — borderless celebration
	b.WriteString(RenderCenteredHero(BannerAnvil, "Installation Complete!", "", width))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	// Installed components
	for _, comp := range m.InstalledComponents {
		check := GreenText(IconCheck)
		line := fmt.Sprintf("  %s  %-22s %s", check, WhiteText(comp.Name), GrayText(comp.Description))
		b.WriteString(CenterText(line, width))
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("\n", SpaceSM))

	// Installation time
	durationStr := "Installed in "
	if m.InstallDuration > 0 {
		durationStr += m.InstallDuration.Round(time.Millisecond * 100).String()
	} else {
		durationStr += "< 1s"
	}
	b.WriteString(CenterText(AmberText(durationStr), width))
	b.WriteString(strings.Repeat("\n", SpaceMD))

	// Next steps section
	b.WriteString(CenterText(CopperText("Start using:"), width))
	b.WriteString("\n")
	b.WriteString(CenterText(AmberText("$ opencode"), width))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	b.WriteString(CenterText(CopperText("Check status:"), width))
	b.WriteString("\n")
	b.WriteString(CenterText(AmberText("$ hefesto status"), width))
	b.WriteString(strings.Repeat("\n", SpaceMD))

	// Forge on message
	b.WriteString(CenterText(AmberText("Forge on!"), width))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	// Help bar
	hints := []KeyHint{
		{Key: "q", Action: "Exit"},
	}
	b.WriteString(RenderHelpBar(hints, width))

	return RenderScreenFrame(b.String(), FrameOptions{
		Width:  m.width,
		Height: m.height,
		Border: BorderNone,
	})
}
