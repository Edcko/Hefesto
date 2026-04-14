package tui

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Edcko/Hefesto/cmd/hefesto/internal/install"
)

// DetectModel is the environment detection screen.
type DetectModel struct {
	width  int
	height int

	// Detection state
	detecting bool
	spinner   int

	// Detection results
	results []DetectResult
}

// DetectResult represents a single detection result.
type DetectResult struct {
	Name     string
	Found    bool
	Details  string
	Checking bool
}

// DetectCompleteMsg signals detection is complete.
type DetectCompleteMsg struct {
	Results []DetectResult
}

// NewDetectModel creates a new detection screen.
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

// Init implements tea.Model.
func (m *DetectModel) Init() tea.Cmd {
	return tea.Batch(
		Tick(100*time.Millisecond),
		m.runDetection(),
	)
}

// runDetection performs the actual environment detection
// using the shared install.Detect() function from the install package.
func (m *DetectModel) runDetection() tea.Cmd {
	return func() tea.Msg {
		results := make([]DetectResult, 4)

		env, err := install.Detect()

		// Detect OpenCode CLI
		results[0] = DetectResult{Name: "OpenCode CLI"}
		if err != nil {
			results[0].Details = "Detection failed"
		} else if env.OpenCodeInstalled {
			results[0].Found = true
			if env.OpenCodeVersion != "" {
				results[0].Details = fmt.Sprintf("Installed (%s)", env.OpenCodeVersion)
			} else {
				results[0].Details = "Installed (version unknown)"
			}
		} else {
			results[0].Details = "Not found (will be configured)"
		}

		// Detect config directory
		results[1] = DetectResult{Name: "Config directory"}
		if err == nil && env.ConfigExists {
			results[1].Found = true
			results[1].Details = "~/.config/opencode/ exists"
		} else {
			results[1].Details = "~/.config/opencode/ (will be created)"
		}

		// Detect existing config
		results[2] = DetectResult{Name: "Existing config"}
		if err == nil && env.ExistingConfig != "none" {
			results[2].Found = true
			switch env.ExistingConfig {
			case "gentleman-dots":
				results[2].Details = "Gentleman.Dots detected"
			case "hefesto":
				results[2].Details = "Hefesto configuration detected"
			case "custom":
				results[2].Details = "Custom configuration found"
			default:
				results[2].Details = "Configuration detected"
			}
		} else {
			results[2].Details = "No existing configuration"
		}

		// System info
		results[3] = DetectResult{
			Name:    "System info",
			Found:   true,
			Details: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		}

		return DetectCompleteMsg{Results: results}
	}
}

// Update implements tea.Model.
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
				return m, TransitionTo(ScreenComponentSelect)
			}
		}
	}

	return m, nil
}

// wizardSteps returns the wizard progress steps for the detect screen.
func (m *DetectModel) wizardSteps() []WizardStep {
	return []WizardStep{
		{Label: "Welcome", Done: true},
		{Label: "Detect", Active: true},
		{Label: "Select"},
		{Label: "Install"},
	}
}

// View implements tea.Model.
func (m *DetectModel) View() string {
	width := ResolveContentWidth(m.width)

	// ===== Wizard progress indicator =====
	progress := RenderWizardProgress(m.wizardSteps(), width)

	// ===== Section title =====
	title := RenderSectionTitle("Environment Detection", width)

	// ===== Status line =====
	var statusLine string
	if m.detecting {
		spinnerChar := string(IconSpinner[m.spinner])
		statusLine = CenterText(AmberText(spinnerChar+" Detecting environment..."), width)
	} else {
		statusLine = CenterText(GreenText(IconCheck+" Detection complete"), width)
	}

	// ===== Results =====
	var resultLines strings.Builder
	for _, result := range m.results {
		resultLines.WriteString(m.renderResult(result, width))
		resultLines.WriteString("\n")
	}

	// ===== Help bar =====
	var helpBar string
	if !m.detecting {
		helpBar = RenderHelpBar([]KeyHint{
			{Key: "Enter", Action: "Continue"},
			{Key: "Esc", Action: "Back"},
		}, width)
	}

	// ===== Assemble with spacing rhythm =====
	var b strings.Builder
	b.WriteString(progress)
	b.WriteString("\n")
	b.WriteString(title)
	b.WriteString(strings.Repeat("\n", SpaceSM))
	b.WriteString(statusLine)
	b.WriteString(strings.Repeat("\n", SpaceSM))
	b.WriteString(resultLines.String())

	if helpBar != "" {
		b.WriteString(strings.Repeat("\n", SpaceSM))
		b.WriteString(helpBar)
	}

	content := b.String()

	// Wrap in rounded border frame, centered in terminal.
	return RenderScreenFrame(content, FrameOptions{
		Width:  width,
		Height: m.height,
		Border: BorderRounded,
	})
}

// renderResult renders a single detection result line.
func (m *DetectModel) renderResult(result DetectResult, width int) string {
	var icon string
	var styledName string

	if result.Checking {
		spinnerChar := string(IconSpinner[m.spinner])
		icon = AmberText(spinnerChar)
		styledName = WhiteText(result.Name)
	} else if result.Found {
		icon = GreenText(IconCheck)
		styledName = WhiteText(result.Name)
	} else {
		icon = GrayText(IconBullet)
		styledName = GrayText(result.Name)
	}

	details := MutedStyle.Render(result.Details)
	line := fmt.Sprintf("  %s %-20s %s", icon, styledName, details)
	return CenterText(line, width)
}
