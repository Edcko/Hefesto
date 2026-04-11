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
	var b strings.Builder

	// Title
	title := TitleStyle.Render("Backup Existing Configuration")
	b.WriteString(CenterText(title, 60))
	b.WriteString("\n\n")

	// Status message
	if m.error != "" {
		errMsg := ErrorStyle.Render("Backup failed: " + m.error)
		b.WriteString(CenterText(errMsg, 60))
		b.WriteString("\n\n")
	} else if m.complete {
		success := SuccessStyle.Render(IconCheck + " Backup complete!")
		b.WriteString(CenterText(success, 60))
		b.WriteString("\n\n")

		path := MutedStyle.Render("Saved to: " + m.backupPath)
		b.WriteString(CenterText(path, 60))
		b.WriteString("\n\n")

		instruction := MutedStyle.Render("Press Enter to continue")
		b.WriteString(CenterText(instruction, 60))
		return b.String()
	} else if m.backingUp {
		spinnerChar := string(IconSpinner[m.spinner])
		status := InfoStyle.Render(spinnerChar + " Creating backup...")
		b.WriteString(CenterText(status, 60))
		b.WriteString("\n\n")

		// Show items being backed up
		for _, item := range m.backupItems {
			line := fmt.Sprintf("  %s %s", MutedStyle.Render(IconBullet), BodyStyle.Render(item))
			b.WriteString(CenterText(line, 60))
			b.WriteString("\n")
		}
		return b.String()
	}

	// Confirmation UI
	desc := BodyStyle.Render("An existing configuration was found.")
	b.WriteString(CenterText(desc, 60))
	b.WriteString("\n\n")

	backupPath := MutedStyle.Render("Backup location: " + m.backupPath)
	b.WriteString(CenterText(backupPath, 60))
	b.WriteString("\n\n")

	// Items to backup
	itemsTitle := SubtitleStyle.Render("Items to backup:")
	b.WriteString(CenterText(itemsTitle, 60))
	b.WriteString("\n")

	for _, item := range m.backupItems {
		line := fmt.Sprintf("  %s %s", MutedStyle.Render(IconBullet), BodyStyle.Render(item))
		b.WriteString(CenterText(line, 60))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Yes/No selection
	yesStyle := MutedStyle
	noStyle := MutedStyle

	if m.selected == 0 {
		yesStyle = BoldStyle.Foreground(Primary)
	} else {
		noStyle = BoldStyle.Foreground(Primary)
	}

	yes := yesStyle.Render("[ Yes ]")
	no := noStyle.Render("[ No ]")

	selection := fmt.Sprintf("      %s        %s", yes, no)
	b.WriteString(CenterText(selection, 60))
	b.WriteString("\n\n")

	hint := MutedStyle.Render("← → to select, Enter to confirm")
	b.WriteString(CenterText(hint, 60))

	return b.String()
}
