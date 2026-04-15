package tui

import (
	"fmt"
	"os"
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

	// OpenCode CLI install status from the install screen
	OpenCodeInstallAttempted    bool
	OpenCodeInstallSuccess      bool
	OpenCodeInstallVersion      string
	OpenCodeInstallError        string
	OpenCodeWasAlreadyInstalled bool // true when OpenCode was present before this run
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
			{Name: "30 AI skills", Description: "(Angular, React, SDD...)"},
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

	var b strings.Builder

	// Hero section — borderless celebration
	b.WriteString(RenderCenteredHero(BannerAnvil, "Installation Complete!", "", width))
	b.WriteString(strings.Repeat("\n", SpaceXS))

	// Installed components
	for _, comp := range m.InstalledComponents {
		check := GreenText(IconCheck)
		line := fmt.Sprintf("  %s  %-22s %s", check, WhiteText(comp.Name), GrayText(comp.Description))
		b.WriteString(CenterText(line, width))
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("\n", SpaceXS))

	// Installation time
	durationStr := "Installed in "
	if m.InstallDuration > 0 {
		durationStr += m.InstallDuration.Round(time.Millisecond * 100).String()
	} else {
		durationStr += "< 1s"
	}
	b.WriteString(CenterText(AmberText(durationStr), width))
	b.WriteString(strings.Repeat("\n", SpaceXS))

	// Next Steps — prominent, numbered actions
	b.WriteString(CenterText(AmberText("Next Steps:"), width))
	b.WriteString("\n")

	// Show OpenCode CLI install status and version info
	if m.OpenCodeInstallAttempted && m.OpenCodeInstallSuccess {
		versionInfo := ""
		if m.OpenCodeInstallVersion != "" {
			versionInfo = fmt.Sprintf(" v%s", m.OpenCodeInstallVersion)
		}
		b.WriteString(CenterText(
			fmt.Sprintf("%s %s%s", GreenText(IconCheck), WhiteText("OpenCode CLI installed"), GrayText(versionInfo)),
			width,
		))
		b.WriteString("\n")
	}

	// When OpenCode was freshly installed, PATH needs to be refreshed
	if m.OpenCodeInstallAttempted && m.OpenCodeInstallSuccess && !m.OpenCodeWasAlreadyInstalled {
		rcHint := getRCHint()
		b.WriteString(CenterText(
			fmt.Sprintf("  %s %s  %s",
				AmberText("1."),
				AmberText(fmt.Sprintf("source %s", rcHint)),
				CopperText("← refreshes PATH (required!)"),
			),
			width,
		))
		b.WriteString("\n")
		b.WriteString(CenterText(
			fmt.Sprintf("  %s %s  %s",
				AmberText("2."),
				AmberText("opencode"),
				CopperText("← start coding!"),
			),
			width,
		))
		b.WriteString("\n")
	} else if m.OpenCodeInstallAttempted && !m.OpenCodeInstallSuccess && m.OpenCodeInstallError != "" {
		// Install failed — show manual fallback
		b.WriteString(CenterText(
			AmberText("  Install OpenCode CLI manually:"),
			width,
		))
		b.WriteString("\n")
		b.WriteString(CenterText(
			AmberText("    curl -fsSL https://opencode.ai/install | bash"),
			width,
		))
		b.WriteString("\n")
		b.WriteString(CenterText(
			fmt.Sprintf("  %s %s  %s",
				AmberText("1."),
				AmberText(fmt.Sprintf("source %s", getRCHint())),
				CopperText("← refreshes PATH"),
			),
			width,
		))
		b.WriteString("\n")
		b.WriteString(CenterText(
			fmt.Sprintf("  %s %s  %s",
				AmberText("2."),
				AmberText("opencode"),
				CopperText("← start coding!"),
			),
			width,
		))
		b.WriteString("\n")
	} else {
		// Already installed or not attempted — simple prompt
		b.WriteString(CenterText(
			CopperText("Run")+AmberText(" $ opencode")+CopperText(" to start"),
			width,
		))
		b.WriteString("\n")
	}

	b.WriteString(CenterText(CopperText("Run")+AmberText(" $ hefesto status")+CopperText(" to check"), width))
	b.WriteString(strings.Repeat("\n", SpaceXS))

	// Forge on message
	b.WriteString(CenterText(AmberText("Forge on!"), width))
	b.WriteString(strings.Repeat("\n", SpaceXS))

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

// getRCHint returns the user's shell RC file name (e.g. "~/.zshrc" or "~/.bashrc")
// so the complete screen can suggest sourcing it to refresh PATH.
func getRCHint() string {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return "~/.zshrc"
	}
	return "~/.bashrc"
}
