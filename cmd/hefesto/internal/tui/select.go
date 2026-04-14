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

// wizardSteps returns the wizard progress steps for the select screen.
func (m *SelectModel) wizardSteps() []WizardStep {
	return []WizardStep{
		{Label: "Welcome", Done: true},
		{Label: "Detect", Done: true},
		{Label: "Select", Active: true},
		{Label: "Install"},
	}
}

// View implements tea.Model.
func (m *SelectModel) View() string {
	width := ResolveContentWidth(m.width)

	// Calculate inner content width accounting for border + padding.
	// RoundedBorder adds 2 chars per side, PadBox adds 2 per side = 8 total.
	innerWidth := width - (PadBox*2 + 2*2)
	if innerWidth < 20 {
		innerWidth = 20
	}

	// ===== Wizard progress =====
	progress := RenderWizardProgress(m.wizardSteps(), width)

	// ===== Section title =====
	title := RenderSectionTitle("Select Components to Install", width)

	// ===== Component checklist =====
	var itemLines strings.Builder
	for i, item := range m.items.Items {
		itemLines.WriteString(m.renderItemLine(i, item, innerWidth))
		itemLines.WriteString("\n")
	}

	// ===== Continue action =====
	continueText := "[Enter] Continue with selection"
	if m.continueRow {
		continueText = AmberText(continueText)
	} else {
		continueText = MutedStyle.Render(continueText)
	}

	// ===== Help bar =====
	helpBar := RenderHelpBar([]KeyHint{
		{Key: "↑↓", Action: "Navigate"},
		{Key: "Space", Action: "Toggle"},
		{Key: "a", Action: "Toggle all"},
		{Key: "Esc", Action: "Back"},
		{Key: "q", Action: "Quit"},
	}, width)

	// ===== Assemble with spacing rhythm =====
	var b strings.Builder
	b.WriteString(progress)
	b.WriteString("\n")
	b.WriteString(title)
	b.WriteString(strings.Repeat("\n", SpaceSM))
	b.WriteString(itemLines.String())
	b.WriteString(CenterText(continueText, innerWidth))
	b.WriteString(strings.Repeat("\n", SpaceSM))
	b.WriteString(helpBar)

	content := b.String()

	// Wrap in rounded border frame, centered in terminal.
	return RenderScreenFrame(content, FrameOptions{
		Width:  m.width,
		Height: m.height,
		Border: BorderRounded,
	})
}

// renderItemLine renders a single component item line with checkbox.
// innerWidth is the available content width inside the border.
func (m *SelectModel) renderItemLine(index int, item ComponentItem, innerWidth int) string {
	isActive := !m.continueRow && m.cursor == index

	// Checkbox: ✓ for selected, ○ for unselected
	var checkbox string
	if item.Selected {
		checkbox = GreenText(IconCheck)
	} else {
		checkbox = DimTextStyle.Render("○")
	}

	// Cursor indicator
	var cursor string
	if isActive {
		cursor = AmberText("❯")
	} else {
		cursor = " "
	}

	// Build display: cursor + checkbox + name + description
	var nameText string
	if isActive {
		nameText = AmberText(item.Name)
	} else if item.Required {
		nameText = GrayText(item.Name)
	} else {
		nameText = WhiteText(item.Name)
	}

	descText := MutedStyle.Render(item.Description)
	displayText := fmt.Sprintf("%s %s", nameText, descText)

	// Truncate if wider than available space
	maxDisplayWidth := innerWidth - lipgloss.Width(cursor) - lipgloss.Width(checkbox) - 4 // spaces
	if lipgloss.Width(displayText) > maxDisplayWidth {
		displayText = truncateVisual(displayText, maxDisplayWidth)
	}

	line := fmt.Sprintf("%s %s %s", cursor, checkbox, displayText)

	// Apply full-line styling for active item highlight
	if isActive {
		return lipgloss.NewStyle().Bold(true).Render(line)
	}
	return line
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
