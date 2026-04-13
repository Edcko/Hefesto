package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ComponentID identifies a selectable install component.
type ComponentID string

const (
	ComponentAgents      ComponentID = "agents"
	ComponentOpenCode    ComponentID = "opencode"
	ComponentSkills      ComponentID = "skills"
	ComponentTheme       ComponentID = "theme"
	ComponentPersonality ComponentID = "personality"
	ComponentCommands    ComponentID = "commands"
	ComponentPlugins     ComponentID = "plugins"
	ComponentEngram      ComponentID = "engram"
)

// ComponentItem represents a single selectable component in the checklist.
type ComponentItem struct {
	ID          ComponentID
	Name        string
	Description string
	Selected    bool
	Required    bool // If true, item cannot be deselected
}

// ComponentSelection holds the user's component choices.
type ComponentSelection struct {
	Items []ComponentItem
}

// IsSelected returns whether a given component is selected.
func (cs *ComponentSelection) IsSelected(id ComponentID) bool {
	for _, item := range cs.Items {
		if item.ID == id {
			return item.Selected
		}
	}
	return false
}

// DefaultComponentSelection returns all components selected (opinionated defaults).
func DefaultComponentSelection() *ComponentSelection {
	return &ComponentSelection{
		Items: []ComponentItem{
			{
				ID:          ComponentAgents,
				Name:        "AGENTS.md",
				Description: "Configuration rules",
				Selected:    true,
				Required:    true,
			},
			{
				ID:          ComponentOpenCode,
				Name:        "opencode.json",
				Description: "Agent definitions",
				Selected:    true,
				Required:    true,
			},
			{
				ID:          ComponentSkills,
				Name:        "Skills",
				Description: "30 available",
				Selected:    true,
				Required:    false,
			},
			{
				ID:          ComponentTheme,
				Name:        "Theme",
				Description: "hefesto.json",
				Selected:    true,
				Required:    false,
			},
			{
				ID:          ComponentPersonality,
				Name:        "Personality",
				Description: "hefesto.md",
				Selected:    true,
				Required:    false,
			},
			{
				ID:          ComponentCommands,
				Name:        "Commands",
				Description: "5 slash commands",
				Selected:    true,
				Required:    false,
			},
			{
				ID:          ComponentPlugins,
				Name:        "Plugins",
				Description: "engram.ts + background-agents.ts",
				Selected:    true,
				Required:    false,
			},
			{
				ID:          ComponentEngram,
				Name:        "Engram binary",
				Description: "latest version",
				Selected:    true,
				Required:    false,
			},
		},
	}
}

// SelectModel is the component selection screen where users choose what to install.
type SelectModel struct {
	width  int
	height int

	cursor int
	items  *ComponentSelection

	// Track the "continue" action row (rendered after items)
	continueRow bool
}

// NewSelectModel creates a new component selection screen.
func NewSelectModel(width, height int) *SelectModel {
	return &SelectModel{
		width:  width,
		height: height,
		items:  DefaultComponentSelection(),
	}
}

// GetSelection returns the current component selection.
func (m *SelectModel) GetSelection() *ComponentSelection {
	return m.items
}

// Init implements tea.Model.
func (m *SelectModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m *SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// The cursor can be 0..len(items)-1 for items, or len(items) for "Continue"
		totalRows := len(m.items.Items)

		switch msg.String() {
		case "up", "k":
			if m.continueRow {
				m.continueRow = false
				m.cursor = totalRows - 1
			} else if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.continueRow {
				// Already at continue, stay
				return m, nil
			}
			if m.cursor < totalRows-1 {
				m.cursor++
			} else {
				// Move to continue row
				m.continueRow = true
			}
			return m, nil

		case " ":
			if m.continueRow {
				return m, func() tea.Msg { return selectCompleteMsg{} }
			}
			// Toggle selection (unless required)
			if !m.items.Items[m.cursor].Required {
				m.items.Items[m.cursor].Selected = !m.items.Items[m.cursor].Selected
			}
			return m, nil

		case "enter":
			if m.continueRow {
				return m, func() tea.Msg { return selectCompleteMsg{} }
			}
			// Toggle selection (unless required)
			if !m.items.Items[m.cursor].Required {
				m.items.Items[m.cursor].Selected = !m.items.Items[m.cursor].Selected
			}
			return m, nil

		case "a":
			// Toggle all non-required items
			// Determine if all non-required are currently selected
			allSelected := true
			for _, item := range m.items.Items {
				if !item.Required && !item.Selected {
					allSelected = false
					break
				}
			}
			// Toggle: if all selected, deselect all non-required; otherwise select all
			newState := !allSelected
			for i := range m.items.Items {
				if !m.items.Items[i].Required {
					m.items.Items[i].Selected = newState
				}
			}
			return m, nil

		case "esc":
			// Go back to detect screen
			return m, TransitionTo(ScreenDetect)

		case "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m *SelectModel) View() string {
	boxWidth := 50

	var b strings.Builder

	// Top border
	b.WriteString("╭")
	b.WriteString(strings.Repeat("─", boxWidth))
	b.WriteString("╮\n")

	// Title
	title := "  \U0001F6E0  Select Components to Install"
	titleLine := fmt.Sprintf("│%s", title)
	padding := boxWidth - len(title) + 1
	if padding > 0 {
		titleLine += strings.Repeat(" ", padding)
	}
	titleLine += "│\n"
	b.WriteString(lipgloss.NewStyle().Foreground(Primary).Bold(true).Render(titleLine))

	// Empty line
	b.WriteString("│")
	b.WriteString(strings.Repeat(" ", boxWidth))
	b.WriteString("│\n")

	// Key hints
	hints := "  \u2191/k  Navigate    Space/Enter  Toggle"
	hintLine := fmt.Sprintf("│%s", hints)
	padding = boxWidth - len(hints) + 1
	if padding > 0 {
		hintLine += strings.Repeat(" ", padding)
	}
	hintLine += "│\n"
	b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(hintLine))

	hints2 := "  a    Toggle all   q  Quit"
	hintLine2 := fmt.Sprintf("│%s", hints2)
	padding = boxWidth - len(hints2) + 1
	if padding > 0 {
		hintLine2 += strings.Repeat(" ", padding)
	}
	hintLine2 += "│\n"
	b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(hintLine2))

	// Empty line
	b.WriteString("│")
	b.WriteString(strings.Repeat(" ", boxWidth))
	b.WriteString("│\n")

	// Component items
	for i, item := range m.items.Items {
		b.WriteString(m.renderItemLine(i, item, boxWidth))
	}

	// Empty line
	b.WriteString("│")
	b.WriteString(strings.Repeat(" ", boxWidth))
	b.WriteString("│\n")

	// Continue action
	continueText := "[Enter] Continue with selection"
	continueLine := fmt.Sprintf("│  %s", continueText)
	padding = boxWidth - len(continueText) - 1
	if padding > 0 {
		continueLine += strings.Repeat(" ", padding)
	}
	continueLine += "│\n"

	if m.continueRow {
		b.WriteString(lipgloss.NewStyle().Foreground(Primary).Bold(true).Render(continueLine))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(TextMuted).Render(continueLine))
	}

	// Bottom border
	b.WriteString("╰")
	b.WriteString(strings.Repeat("─", boxWidth))
	b.WriteString("╯\n")

	return CenterText(b.String(), 60)
}

// renderItemLine renders a single component item line with checkbox.
func (m *SelectModel) renderItemLine(index int, item ComponentItem, boxWidth int) string {
	isActive := !m.continueRow && m.cursor == index

	var checkbox string
	if item.Selected {
		checkbox = "✅"
	} else {
		checkbox = "⬜"
	}

	// Build the display text
	displayName := item.Name
	desc := item.Description

	var lineText string
	if item.Required {
		// Required items: show with a lock indicator and description
		lineText = fmt.Sprintf("  %s %s (%s)", checkbox, displayName, desc)
	} else {
		lineText = fmt.Sprintf("  %s %s (%s)", checkbox, displayName, desc)
	}

	// Cursor indicator
	var cursor string
	if isActive {
		cursor = "❯"
	} else {
		cursor = " "
	}

	fullLine := fmt.Sprintf("│%s%s", cursor, lineText)

	// Calculate padding
	plainLen := len(cursor) + len(lineText)
	pad := boxWidth - plainLen + 1
	if pad > 0 {
		fullLine += strings.Repeat(" ", pad)
	}
	fullLine += "│\n"

	// Apply styling based on active state and required status
	if isActive {
		fullLine = lipgloss.NewStyle().Foreground(Primary).Render(fullLine)
	} else if item.Required {
		// Required items shown slightly differently - still styled but dimmer
		fullLine = lipgloss.NewStyle().Foreground(TextMuted).Render(fullLine)
	}

	return fullLine
}

// selectCompleteMsg signals that the user confirmed their component selection.
// The App's Update handler decides whether to go to Backup or Install next.
type selectCompleteMsg struct{}
