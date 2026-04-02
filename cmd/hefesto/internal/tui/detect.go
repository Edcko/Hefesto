package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetectModel is the environment detection screen
type DetectModel struct {
	width  int
	height int

	// Detection state
	detecting bool
	spinner   int

	// Detection results
	results []DetectResult

	// App state to update
	app *App
}

// DetectResult represents a single detection result
type DetectResult struct {
	Name     string
	Found    bool
	Details  string
	Checking bool
}

// DetectCompleteMsg signals detection is complete
type DetectCompleteMsg struct {
	Results []DetectResult
}

// NewDetectModel creates a new detection screen
func NewDetectModel(width, height int) *DetectModel {
	return &DetectModel{
		width:     width,
		height:    height,
		detecting: true,
		results: []DetectResult{
			{Name: "OpenCode CLI", Checking: true},
			{Name: "Config directory", Checking: true},
			{Name: "Existing config", Checking: true},
			{Name: "System info", Checking: true},
		},
	}
}

// Init implements tea.Model
func (m *DetectModel) Init() tea.Cmd {
	return tea.Batch(
		Tick(100*time.Millisecond),
		m.runDetection(),
	)
}

// runDetection performs the actual environment detection
func (m *DetectModel) runDetection() tea.Cmd {
	return func() tea.Msg {
		results := make([]DetectResult, 4)

		// Detect OpenCode CLI
		results[0] = m.detectOpenCode()

		// Detect config directory
		results[1] = m.detectConfigDir()

		// Detect existing config
		results[2] = m.detectExistingConfig()

		// System info
		results[3] = m.detectSystemInfo()

		return DetectCompleteMsg{Results: results}
	}
}

// detectOpenCode checks if OpenCode CLI is installed
func (m *DetectModel) detectOpenCode() DetectResult {
	result := DetectResult{Name: "OpenCode CLI"}

	// Try to find opencode in PATH
	path, err := exec.LookPath("opencode")
	if err != nil {
		result.Found = false
		result.Details = "Not found (will be configured)"
		return result
	}

	// Try to get version
	cmd := exec.Command(path, "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Found = true
		result.Details = "Installed (version unknown)"
		return result
	}

	version := strings.TrimSpace(string(output))
	result.Found = true
	result.Details = fmt.Sprintf("Installed (%s)", version)
	return result
}

// detectConfigDir checks if the config directory exists
func (m *DetectModel) detectConfigDir() DetectResult {
	result := DetectResult{Name: "Config directory"}

	configPath := os.ExpandEnv(AppConfigPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		result.Found = false
		result.Details = "~/.config/opencode/ (will be created)"
		return result
	}

	result.Found = true
	result.Details = "~/.config/opencode/ exists"
	return result
}

// detectExistingConfig checks for existing configuration
func (m *DetectModel) detectExistingConfig() DetectResult {
	result := DetectResult{Name: "Existing config"}

	configPath := os.ExpandEnv(AppConfigPath)

	// Check for AGENTS.md (indicates Gentleman.Dots or similar)
	agentsPath := configPath + "AGENTS.md"
	if _, err := os.Stat(agentsPath); err == nil {
		result.Found = true
		result.Details = "Gentleman.Dots or custom config detected"
		return result
	}

	// Check for opencode.json
	jsonPath := configPath + "opencode.json"
	if _, err := os.Stat(jsonPath); err == nil {
		result.Found = true
		result.Details = "Custom configuration found"
		return result
	}

	result.Found = false
	result.Details = "No existing configuration"
	return result
}

// detectSystemInfo gathers system information
func (m *DetectModel) detectSystemInfo() DetectResult {
	result := DetectResult{
		Name:  "System info",
		Found: true,
	}

	details := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	result.Details = details
	return result
}

// Update implements tea.Model
func (m *DetectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		if m.detecting {
			m.spinner = (m.spinner + 1) % len(IconSpinner)
			return m, Tick(100 * time.Millisecond)
		}
		return m, nil

	case DetectCompleteMsg:
		m.results = msg.Results
		m.detecting = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			if !m.detecting {
				// Check if we need backup screen
				existingConfig := m.results[2].Found
				if existingConfig {
					return m, TransitionTo(ScreenBackup)
				}
				return m, TransitionTo(ScreenInstall)
			}
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *DetectModel) View() string {
	var b strings.Builder

	// Title
	title := TitleStyle.Render("Environment Detection")
	b.WriteString(CenterText(title, 60))
	b.WriteString("\n\n")

	// Status
	if m.detecting {
		spinnerChar := string(IconSpinner[m.spinner])
		status := InfoStyle.Render(spinnerChar + " Detecting environment...")
		b.WriteString(CenterText(status, 60))
	} else {
		status := SuccessStyle.Render(IconCheck + " Detection complete")
		b.WriteString(CenterText(status, 60))
	}
	b.WriteString("\n\n")

	// Results
	for _, result := range m.results {
		b.WriteString(m.renderResult(result))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Instructions
	if !m.detecting {
		instruction := MutedStyle.Render("Press Enter to continue")
		b.WriteString(CenterText(instruction, 60))
	}

	return b.String()
}

// renderResult renders a single detection result
func (m *DetectModel) renderResult(result DetectResult) string {
	var icon string
	var style lipgloss.Style

	if result.Checking {
		spinnerChar := string(IconSpinner[m.spinner])
		icon = spinnerChar
		style = InfoStyle
	} else if result.Found {
		icon = IconCheck
		style = SuccessStyle
	} else {
		icon = IconBullet
		style = MutedStyle
	}

	name := BodyStyle.Render(result.Name)
	details := MutedStyle.Render(result.Details)

	line := fmt.Sprintf("  %s %-20s %s", style.Render(icon), name, details)
	return CenterText(line, 60)
}
