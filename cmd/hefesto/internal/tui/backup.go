package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Edcko/Hefesto/cmd/hefesto/internal/install"
)

// BackupModel is the backup confirmation and execution screen
type BackupModel struct {
	width  int
	height int

	configPath     string
	existingConfig bool
	backupPath     string

	// UI state
	confirming bool
	selected   int // 0 = Yes, 1 = No
	backingUp  bool
	spinner    int
	complete   bool
	error      string

	// Items to backup
	backupItems []string
}

// BackupCompleteMsg signals backup is complete
type BackupCompleteMsg struct {
	Path  string
	Error string
}

// NewBackupModel creates a new backup screen
func NewBackupModel(configPath string, existingConfig bool, width, height int) *BackupModel {
	// Generate backup path with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("~/.config/opencode-backup-%s/", timestamp)

	return &BackupModel{
		configPath:     configPath,
		existingConfig: existingConfig,
		backupPath:     backupPath,
		confirming:     true,
		selected:       0,
	}
}

// Init implements tea.Model
func (m *BackupModel) Init() tea.Cmd {
	// Scan for items to backup
	m.scanBackupItems()
	return nil
}

// scanBackupItems finds files to backup
func (m *BackupModel) scanBackupItems() {
	configPath := os.ExpandEnv(m.configPath)

	items := []string{}
	entries, err := os.ReadDir(configPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files and .git
		if strings.HasPrefix(name, ".") {
			continue
		}
		items = append(items, name)
	}

	m.backupItems = items
}

// runBackup executes the backup
func (m *BackupModel) runBackup() tea.Cmd {
	return func() tea.Msg {
		configPath := os.ExpandEnv(m.configPath)
		backupPath := os.ExpandEnv(m.backupPath)

		// Create backup directory
		if err := os.MkdirAll(backupPath, 0750); err != nil {
			return BackupCompleteMsg{Error: err.Error()}
		}

		// Copy each item
		for _, item := range m.backupItems {
			src := filepath.Join(configPath, item)
			dst := filepath.Join(backupPath, item)

			if err := install.CopyPath(src, dst); err != nil {
				return BackupCompleteMsg{Error: err.Error()}
			}
		}

		return BackupCompleteMsg{Path: m.backupPath}
	}
}

// Update implements tea.Model
func (m *BackupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		if m.backingUp {
			m.spinner = (m.spinner + 1) % len(IconSpinner)
			return m, Tick(100 * time.Millisecond)
		}
		return m, nil

	case BackupCompleteMsg:
		m.backingUp = false
		if msg.Error != "" {
			m.error = msg.Error
		} else {
			m.complete = true
		}
		return m, nil

	case tea.KeyMsg:
		if m.confirming {
			switch msg.String() {
			case "left", "h":
				if m.selected > 0 {
					m.selected--
				}
			case "right", "l":
				if m.selected < 1 {
					m.selected++
				}
			case "enter", " ":
				if m.selected == 0 {
					// Yes - start backup
					m.confirming = false
					m.backingUp = true
					return m, tea.Batch(
						Tick(100*time.Millisecond),
						m.runBackup(),
					)
				} else {
					// No - skip backup, go to install
					return m, TransitionTo(ScreenInstall)
				}
			}
		} else if m.complete {
			switch msg.String() {
			case "enter", " ":
				return m, TransitionTo(ScreenInstall)
			}
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *BackupModel) View() string {
	width := ResolveContentWidth(m.width)

	// Wizard progress: Welcome → Detect → Select → [Backup] → Install → Complete
	wizardSteps := []WizardStep{
		{Label: "Welcome", Done: true},
		{Label: "Detect", Done: true},
		{Label: "Select", Done: true},
		{Label: "Backup", Active: true},
		{Label: "Install"},
		{Label: "Complete"},
	}

	var b strings.Builder

	// Wizard progress indicator
	b.WriteString(RenderWizardProgress(wizardSteps, width))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	// Section title
	b.WriteString(RenderSectionTitle("Backup Configuration", width))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	// Content area
	if m.error != "" {
		m.renderError(&b, width)
	} else if m.complete {
		m.renderComplete(&b, width)
	} else if m.backingUp {
		m.renderProgress(&b, width)
	} else {
		m.renderConfirmation(&b, width)
	}

	// Help bar
	hints := m.getHelpHints()
	if len(hints) > 0 {
		b.WriteString(strings.Repeat("\n", SpaceSM))
		b.WriteString(RenderHelpBar(hints, width))
	}

	return RenderScreenFrame(b.String(), FrameOptions{
		Width:  m.width,
		Height: m.height,
		Border: BorderRounded,
	})
}

// renderConfirmation renders the Yes/No backup confirmation
func (m *BackupModel) renderConfirmation(b *strings.Builder, width int) {
	desc := WhiteText("An existing configuration was found.")
	b.WriteString(CenterText(desc, width))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	backupPath := GrayText("Backup location: " + m.backupPath)
	b.WriteString(CenterText(backupPath, width))
	b.WriteString(strings.Repeat("\n", SpaceMD))

	// Items section header
	itemsHeader := CopperText("Items to backup:")
	b.WriteString(CenterText(itemsHeader, width))
	b.WriteString("\n")

	for _, item := range m.backupItems {
		b.WriteString(CenterText(BulletItem(WhiteText(item)), width))
		b.WriteString("\n")
	}
	b.WriteString(strings.Repeat("\n", SpaceMD))

	// Yes/No selection
	yesStyle := MutedStyle
	noStyle := MutedStyle

	if m.selected == 0 {
		yesStyle = BoldStyle.Foreground(ColorAmber)
	} else {
		noStyle = BoldStyle.Foreground(ColorAmber)
	}

	yes := yesStyle.Render("[ Yes ]")
	no := noStyle.Render("[ No ]")

	selection := fmt.Sprintf("      %s        %s", yes, no)
	b.WriteString(CenterText(selection, width))
}

// renderProgress renders the backup-in-progress state
func (m *BackupModel) renderProgress(b *strings.Builder, width int) {
	spinnerChar := string(IconSpinner[m.spinner])
	status := AmberText(spinnerChar + " Creating backup...")
	b.WriteString(CenterText(status, width))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	for _, item := range m.backupItems {
		b.WriteString(CenterText(BulletItem(WhiteText(item)), width))
		b.WriteString("\n")
	}
}

// renderComplete renders the backup-success state
func (m *BackupModel) renderComplete(b *strings.Builder, width int) {
	success := GreenText(IconCheck + " Backup complete!")
	b.WriteString(CenterText(success, width))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	path := GrayText("Saved to: " + m.backupPath)
	b.WriteString(CenterText(path, width))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	instruction := MutedStyle.Render("Press Enter to continue")
	b.WriteString(CenterText(instruction, width))
}

// renderError renders the backup-error state
func (m *BackupModel) renderError(b *strings.Builder, width int) {
	errMsg := RedText(IconCross + " Backup failed: " + m.error)
	b.WriteString(CenterText(errMsg, width))
	b.WriteString(strings.Repeat("\n", SpaceSM))

	instruction := MutedStyle.Render("Press Enter to continue")
	b.WriteString(CenterText(instruction, width))
}

// getHelpHints returns context-sensitive key hints
func (m *BackupModel) getHelpHints() []KeyHint {
	if m.confirming {
		return []KeyHint{
			{Key: "← →", Action: "Select"},
			{Key: "Enter", Action: "Confirm"},
		}
	}
	if m.complete || m.error != "" {
		return []KeyHint{
			{Key: "Enter", Action: "Continue"},
		}
	}
	return nil
}
