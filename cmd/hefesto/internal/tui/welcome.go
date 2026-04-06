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
	width := m.width
	if width == 0 {
		width = 60
	}

	// Ensure minimum width for the box
	boxWidth := 54
	if width < boxWidth+4 {
		boxWidth = width - 4
	}

	// ===== Build the welcome content =====

	// Banner - The Anvil with Flames (centered)
	bannerStyled := lipgloss.NewStyle().
		Foreground(ColorAmber).
		Render(BannerAnvil)

	// Version line
	versionLine := VersionStyle()
	taglineLine := TaglineStyle()

	// Build the inner content
	var innerContent strings.Builder
	innerContent.WriteString("\n")
	innerContent.WriteString(CenterText(bannerStyled, boxWidth-4))
	innerContent.WriteString("\n")
	innerContent.WriteString(CenterText(versionLine, boxWidth-4))
	innerContent.WriteString("\n")
	innerContent.WriteString(CenterText(taglineLine, boxWidth-4))
	innerContent.WriteString("\n\n")

	// Description
	description := WhiteText("Hefesto will configure your OpenCode environment")
	innerContent.WriteString(CenterText(description, boxWidth-4))
	innerContent.WriteString("\n")
	description2 := WhiteText("with agents, skills, themes and persistence.")
	innerContent.WriteString(CenterText(description2, boxWidth-4))
	innerContent.WriteString("\n\n")

	// What will be installed header
	installHeader := CopperText("What will be installed:")
	innerContent.WriteString("  " + installHeader)
	innerContent.WriteString("\n")

	// Bullet points with amber bullets
	items := []string{
		"25 AI skills (Angular, React, SDD...)",
		"SDD orchestrator + 6 phase agents",
		"Fuego/Forge theme (amber/copper)",
		"Engram persistent memory",
		"Background agents plugin",
	}

	for _, item := range items {
		innerContent.WriteString(BulletItem(WhiteText(item)))
		innerContent.WriteString("\n")
	}

	innerContent.WriteString("\n")

	// Press Enter prompt - highlighted to draw attention
	enterPrompt := AmberText("Press Enter to continue ") + GrayText("→")
	innerContent.WriteString(CenterText(enterPrompt, boxWidth-4))
	innerContent.WriteString("\n")

	// ===== Create the bordered box =====

	// Border characters for double border
	topLeft := "╔"
	topRight := "╗"
	bottomLeft := "╚"
	bottomRight := "╝"
	horizontal := "═"
	vertical := "║"

	// Style for border
	borderStyle := lipgloss.NewStyle().Foreground(ColorCopper)
	innerStyle := lipgloss.NewStyle().Foreground(ColorWhite)

	// Build top border
	topBorder := borderStyle.Render(topLeft + strings.Repeat(horizontal, boxWidth-2) + topRight)

	// Build bottom border
	bottomBorder := borderStyle.Render(bottomLeft + strings.Repeat(horizontal, boxWidth-2) + bottomRight)

	// Process inner content lines
	contentLines := strings.Split(innerContent.String(), "\n")
	var boxedLines []string
	boxedLines = append(boxedLines, topBorder)

	for _, line := range contentLines {
		// Calculate padding to align right border
		plainLine := stripAnsi(line)
		lineLen := len(plainLine)
		targetLen := boxWidth - 4 // -2 for borders, -2 for padding

		padding := 0
		if lineLen < targetLen {
			padding = targetLen - lineLen
		}

		paddedLine := line + strings.Repeat(" ", padding)
		boxedLines = append(boxedLines, borderStyle.Render(vertical)+" "+innerStyle.Render(paddedLine)+" "+borderStyle.Render(vertical))
	}

	boxedLines = append(boxedLines, bottomBorder)

	// Join all lines
	result := strings.Join(boxedLines, "\n")

	return result
}

// stripAnsi removes ANSI escape codes from a string for length calculation
func stripAnsi(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
