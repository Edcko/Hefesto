package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Colors for the completion screen
var (
	colorAmber  = lipgloss.Color("#FF8C00")
	colorCopper = lipgloss.Color("#B87333")
	colorGreen  = lipgloss.Color("#22C55E")
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
			{Name: "26 AI skills", Description: "(Angular, React, SDD...)"},
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
	// Box dimensions
	boxWidth := 54

	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(colorCopper).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().
		Foreground(colorAmber).
		Bold(true)

	checkStyle := lipgloss.NewStyle().
		Foreground(colorGreen)

	amberStyle := lipgloss.NewStyle().
		Foreground(colorAmber)

	amberBoldStyle := lipgloss.NewStyle().
		Foreground(colorAmber).
		Bold(true)

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	// Build content lines
	var lines []string

	// Empty line at top
	lines = append(lines, "")

	// Title with fire emojis
	fireEmoji := "🔥"
	title := fmt.Sprintf("    %s  Installation Complete!  %s    ", fireEmoji, fireEmoji)
	lines = append(lines, titleStyle.Render(title))

	// Empty line
	lines = append(lines, "")

	// Installed components
	for _, comp := range m.InstalledComponents {
		check := checkStyle.Render("✅")
		line := fmt.Sprintf("  %s %-22s %s", check, comp.Name, mutedStyle.Render(comp.Description))
		lines = append(lines, line)
	}

	// Empty line
	lines = append(lines, "")

	// Installation time
	durationStr := "⚡ Installed in "
	if m.InstallDuration > 0 {
		durationStr += m.InstallDuration.Round(time.Millisecond * 100).String()
	} else {
		durationStr += "< 1s"
	}
	lines = append(lines, "  "+amberStyle.Render(durationStr))

	// Empty line
	lines = append(lines, "")

	// Start using section
	lines = append(lines, "  Start using:")
	lines = append(lines, "    "+amberStyle.Render("$ opencode"))

	// Empty line
	lines = append(lines, "")

	// Check status section
	lines = append(lines, "  Check status:")
	lines = append(lines, "    "+amberStyle.Render("$ hefesto status"))

	// Empty line
	lines = append(lines, "")

	// Forge on message
	forgeOn := amberBoldStyle.Render("  Forge on! 🛠️")
	lines = append(lines, forgeOn)

	// Empty line
	lines = append(lines, "")

	// Exit instruction
	lines = append(lines, mutedStyle.Render("  Press q to exit"))

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
