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
	// Adapt to terminal width like welcome.go does.
	termWidth := m.width
	if termWidth == 0 {
		termWidth = 60
	}

	// Inner content width that lipgloss border+padding will render inside.
	// lipgloss DoubleBorder adds 2 chars per side, Padding(0,1) adds 1 per side = 6 total.
	// So a Width(w) on the style means the inner text area is w chars wide.
	boxInnerWidth := 46
	if termWidth < boxInnerWidth+8 {
		boxInnerWidth = termWidth - 8
	}
	if boxInnerWidth < 20 {
		boxInnerWidth = 20
	}

	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorCopper).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().
		Foreground(ColorAmber).
		Bold(true)

	mutedStyle := lipgloss.NewStyle().
		Foreground(ColorGray)

	// Build content lines
	var lines []string

	// Title
	lines = append(lines, titleStyle.Render("Select Components to Install"))
	lines = append(lines, "")

	// Key hints — split into two short lines that fit inside boxInnerWidth
	lines = append(lines, mutedStyle.Render("up/k Navigate  Space/Enter Toggle"))
	lines = append(lines, mutedStyle.Render("a   Toggle all  q  Quit"))
	lines = append(lines, "")

	// Component items
	for i, item := range m.items.Items {
		lines = append(lines, m.renderItemLine(i, item, boxInnerWidth))
	}

	lines = append(lines, "")

	// Continue action
	continueText := "[Enter] Continue with selection"
	if m.continueRow {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorAmber).Bold(true).Render(continueText))
	} else {
		lines = append(lines, mutedStyle.Render(continueText))
	}

	// Render inside bordered box
	content := strings.Join(lines, "\n")
	box := borderStyle.
		Width(boxInnerWidth).
		Render(content)

	// Center in terminal
	return lipgloss.NewStyle().
		Width(termWidth).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(box)
}

// renderItemLine renders a single component item line with checkbox.
// innerWidth is the available content width inside the box.
func (m *SelectModel) renderItemLine(index int, item ComponentItem, innerWidth int) string {
	isActive := !m.continueRow && m.cursor == index

	var checkbox string
	if item.Selected {
		checkbox = "✅"
	} else {
		checkbox = "⬜"
	}

	// Cursor indicator
	var cursor string
	if isActive {
		cursor = "❯"
	} else {
		cursor = " "
	}

	// Build display text: "✅ Name (description)"
	displayText := fmt.Sprintf("%s %s (%s)", checkbox, item.Name, item.Description)

	// Truncate if wider than available space (cursor + space + displayText)
	maxDisplayWidth := innerWidth - lipgloss.Width(cursor) - 1 // -1 for space after cursor
	if lipgloss.Width(displayText) > maxDisplayWidth {
		displayText = truncateVisual(displayText, maxDisplayWidth)
	}

	line := cursor + " " + displayText

	// Apply styling
	if isActive {
		return lipgloss.NewStyle().Foreground(ColorAmber).Bold(true).Render(line)
	}
	if item.Required {
		return lipgloss.NewStyle().Foreground(ColorGray).Render(line)
	}
	return lipgloss.NewStyle().Foreground(ColorWhite).Render(line)
}

// truncateVisual truncates a string to fit within maxWidth terminal columns,
// appending "…" if truncation was needed.
func truncateVisual(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// Remove characters from the end until it fits
	runes := []rune(s)
	for len(runes) > 0 {
		test := string(runes) + "…"
		if lipgloss.Width(test) <= maxWidth {
			return test
		}
		runes = runes[:len(runes)-1]
	}
	return "…"
}

// selectCompleteMsg signals that the user confirmed their component selection.
// The App's Update handler decides whether to go to Backup or Install next.
type selectCompleteMsg struct{}
