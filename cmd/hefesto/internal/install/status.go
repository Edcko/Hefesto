// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StatusInfo holds detailed status information about the Hefesto installation.
type StatusInfo struct {
	ConfigPath string
	Installed  bool
	Version    string
	Components ComponentStatus
	Binaries   BinaryStatus
}

// ComponentStatus holds status of individual configuration components.
type ComponentStatus struct {
	AgentsMD     ComponentDetail
	OpenCodeJSON ComponentDetail
	Skills       ComponentDetail
	Plugins      ComponentDetail
	Personality  ComponentDetail
	Theme        ComponentDetail
	Commands     ComponentDetail
}

// ComponentDetail holds details about a specific component.
type ComponentDetail struct {
	Present bool
	Detail  string // e.g., "8,624 bytes" or "27 directories"
}

// BinaryStatus holds status of required binaries.
type BinaryStatus struct {
	Engram   BinaryDetail
	OpenCode BinaryDetail
}

// BinaryDetail holds details about a binary.
type BinaryDetail struct {
	Installed bool
	Version   string
	Path      string
}

// CheckStatus performs a comprehensive check of the Hefesto installation.
func CheckStatus() (*StatusInfo, error) {
	env, err := Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect environment: %w", err)
	}

	status := &StatusInfo{
		ConfigPath: env.ConfigPath,
		Installed:  false, // Will be set properly after checking components
	}

	// If config exists, check components
	if env.ConfigExists {
		status.Components = checkComponents(env.ConfigPath)
		status.Version = detectVersion(env.ConfigPath)

		// Installed requires BOTH AGENTS.md AND opencode.json to exist
		status.Installed = status.Components.AgentsMD.Present && status.Components.OpenCodeJSON.Present
	}

	// Check binaries
	status.Binaries = BinaryStatus{
		Engram: BinaryDetail{
			Installed: env.EngramInstalled,
			Version:   env.EngramVersion,
			Path:      env.EngramPath,
		},
		OpenCode: BinaryDetail{
			Installed: env.OpenCodeInstalled,
			Version:   env.OpenCodeVersion,
			Path:      env.OpenCodePath,
		},
	}

	return status, nil
}

// checkComponents checks the status of all configuration components.
func checkComponents(configPath string) ComponentStatus {
	return ComponentStatus{
		AgentsMD:     checkFile(configPath, "AGENTS.md"),
		OpenCodeJSON: checkFile(configPath, "opencode.json"),
		Skills:       checkDirectory(configPath, "skills"),
		Plugins:      checkDirectory(configPath, "plugins"),
		Personality:  checkFile(configPath, "personality"),
		Theme:        checkThemeDirectory(configPath),
		Commands:     checkDirectory(configPath, "commands"),
	}
}

// checkFile checks if a file exists and returns its details.
func checkFile(configPath, filename string) ComponentDetail {
	filePath := filepath.Join(configPath, filename)
	info, err := os.Stat(filePath)
	if err != nil {
		return ComponentDetail{Present: false, Detail: "Missing"}
	}

	size := formatBytes(info.Size())
	return ComponentDetail{Present: true, Detail: size}
}

// checkDirectory checks if a directory exists and counts its contents.
func checkDirectory(configPath, dirname string) ComponentDetail {
	dirPath := filepath.Join(configPath, dirname)
	info, err := os.Stat(dirPath)
	if err != nil {
		return ComponentDetail{Present: false, Detail: "Missing"}
	}

	if !info.IsDir() {
		return ComponentDetail{Present: false, Detail: "Not a directory"}
	}

	// Count contents
	contents, err := os.ReadDir(dirPath)
	if err != nil {
		return ComponentDetail{Present: true, Detail: "Error reading"}
	}

	count := len(contents)
	if dirname == "skills" {
		// Count subdirectories for skills
		dirCount := 0
		for _, c := range contents {
			if c.IsDir() {
				dirCount++
			}
		}
		return ComponentDetail{Present: true, Detail: fmt.Sprintf("%d directories", dirCount)}
	}

	return ComponentDetail{Present: true, Detail: fmt.Sprintf("%d files", count)}
}

// checkThemeDirectory checks if the themes directory exists and has at least one .json file.
func checkThemeDirectory(configPath string) ComponentDetail {
	themesPath := filepath.Join(configPath, "themes")
	info, err := os.Stat(themesPath)
	if err != nil {
		return ComponentDetail{Present: false, Detail: "Missing"}
	}

	if !info.IsDir() {
		return ComponentDetail{Present: false, Detail: "Not a directory"}
	}

	// Look for .json files in themes directory
	contents, err := os.ReadDir(themesPath)
	if err != nil {
		return ComponentDetail{Present: true, Detail: "Error reading"}
	}

	// Find the first .json file and report its details
	for _, entry := range contents {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			size := formatBytes(info.Size())
			return ComponentDetail{
				Present: true,
				Detail:  fmt.Sprintf("%s (%s)", entry.Name(), size),
			}
		}
	}

	// Themes directory exists but no .json files
	return ComponentDetail{Present: false, Detail: "No .json files"}
}

// detectVersion tries to detect the Hefesto version from configuration files.
func detectVersion(configPath string) string {
	// Try to read version from AGENTS.md
	agentsPath := filepath.Join(configPath, "AGENTS.md")
	content, err := os.ReadFile(agentsPath) //nolint:gosec // G304: agentsPath built from known config directory
	if err == nil {
		// Look for version markers in AGENTS.md
		if strings.Contains(string(content), "Hefesto") {
			return "Hefesto Config"
		}
		if strings.Contains(string(content), "Gentleman") {
			return "Gentleman.Dots"
		}
	}

	// Try opencode.json for version info
	configFile := filepath.Join(configPath, "opencode.json")
	content, err = os.ReadFile(configFile) //nolint:gosec // G304: configFile built from known config directory
	if err == nil {
		// Simple check for version field
		if strings.Contains(string(content), "version") {
			// Could parse JSON properly, but for now just indicate it has version
			return "Custom Config"
		}
	}

	return "Unknown"
}

// formatBytes formats bytes into a human-readable string.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d bytes", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// PrintStatus prints the status information in a formatted way.
func PrintStatus(status *StatusInfo) {
	fmt.Println()
	fmt.Println("🔥 Hefesto Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Show config directory
	fmt.Printf("  Config Dir:   %s\n", formatPath(status.ConfigPath))

	if !status.Installed {
		fmt.Println()
		fmt.Println("  Status: ❌ Not installed")
		fmt.Println()
		fmt.Println("  Run `hefesto install` to get started.")
		fmt.Println()
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
		return
	}

	// Show installed status
	fmt.Printf("  Installed:    ✅ Yes\n")
	fmt.Printf("  Version:      %s\n", status.Version)
	fmt.Println()

	// Show components
	fmt.Println("  Components:")
	printComponent("AGENTS.md", status.Components.AgentsMD)
	printComponent("opencode.json", status.Components.OpenCodeJSON)
	printComponent("Skills", status.Components.Skills)
	printComponent("Plugins", status.Components.Plugins)
	printComponent("Personality", status.Components.Personality)
	printComponent("Theme", status.Components.Theme)
	printComponent("Commands", status.Components.Commands)
	fmt.Println()

	// Show binaries
	fmt.Println("  Binaries:")
	printBinary("engram", status.Binaries.Engram)
	printBinary("opencode", status.Binaries.OpenCode)

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// PrintStatusVerbose prints detailed status information.
func PrintStatusVerbose(status *StatusInfo) {
	fmt.Println()
	fmt.Println("🔥 Hefesto Status (Verbose)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Show config directory
	fmt.Printf("  Config Directory: %s\n", formatPath(status.ConfigPath))
	fmt.Println()

	if !status.Installed {
		fmt.Println("  Installation Status: ❌ Not installed")
		fmt.Println()
		fmt.Println("  Run `hefesto install` to get started.")
		fmt.Println()
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
		return
	}

	// Show installed status
	fmt.Println("  Installation Status: ✅ Installed")
	fmt.Printf("  Version:             %s\n", status.Version)
	fmt.Println()

	// Show components with full details
	fmt.Println("  ┌─ Components ─────────────────────────")
	printComponentVerbose("AGENTS.md", status.Components.AgentsMD)
	printComponentVerbose("opencode.json", status.Components.OpenCodeJSON)
	printComponentVerbose("Skills", status.Components.Skills)
	printComponentVerbose("Plugins", status.Components.Plugins)
	printComponentVerbose("Personality", status.Components.Personality)
	printComponentVerbose("Theme", status.Components.Theme)
	printComponentVerbose("Commands", status.Components.Commands)
	fmt.Println("  └───────────────────────────────────────")
	fmt.Println()

	// Show binaries with full details
	fmt.Println("  ┌─ Binaries ────────────────────────────")
	printBinaryVerbose("engram", status.Binaries.Engram)
	printBinaryVerbose("opencode", status.Binaries.OpenCode)
	fmt.Println("  └───────────────────────────────────────")

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// printComponentVerbose prints a component status with verbose formatting.
func printComponentVerbose(name string, detail ComponentDetail) {
	icon := "❌"
	status := "Missing"
	if detail.Present {
		icon = "✅"
		status = detail.Detail
	}
	fmt.Printf("  │ %-15s %s %s\n", name, icon, status)
}

// printBinaryVerbose prints a binary status with verbose formatting.
func printBinaryVerbose(name string, detail BinaryDetail) {
	if detail.Installed {
		pathInfo := detail.Path
		if detail.Version != "" {
			pathInfo = fmt.Sprintf("%s (%s)", detail.Path, detail.Version)
		}
		fmt.Printf("  │ %-15s ✅ %s\n", name, pathInfo)
	} else {
		fmt.Printf("  │ %-15s ❌ Not installed\n", name)
	}
}

// printComponent prints a component status line.
func printComponent(name string, detail ComponentDetail) {
	icon := "❌"
	status := "Missing"
	if detail.Present {
		icon = "✅"
		status = detail.Detail
	}
	fmt.Printf("    %-15s %s %s\n", name, icon, status)
}

// printBinary prints a binary status line.
func printBinary(name string, detail BinaryDetail) {
	if detail.Installed {
		pathInfo := detail.Path
		if detail.Version != "" {
			pathInfo = fmt.Sprintf("%s (%s)", detail.Path, detail.Version)
		}
		fmt.Printf("    %-15s ✅ %s\n", name, pathInfo)
	} else {
		fmt.Printf("    %-15s ❌ Not installed\n", name)
	}
}

// formatPath formats a path, replacing home directory with ~.
func formatPath(path string) string {
	homeDir, err := getUserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, homeDir) {
		return "~" + strings.TrimPrefix(path, homeDir)
	}
	return path
}

// ============================================
// JSON output types
// ============================================

// StatusJSON is the JSON representation of the installation status.
type StatusJSON struct {
	Installed    bool       `json:"installed"`
	Version      string     `json:"version,omitempty"`
	ConfigPath   string     `json:"config_path,omitempty"`
	AgentsMD     bool       `json:"agents_md"`
	OpenCodeJSON bool       `json:"opencode_json"`
	Theme        ThemeJSON  `json:"theme"`
	Engram       EngramJSON `json:"engram"`
	Plugins      bool       `json:"plugins"`
	SkillsCount  int        `json:"skills_count"`
}

// ThemeJSON is the JSON representation of theme status.
type ThemeJSON struct {
	Installed bool   `json:"installed"`
	Name      string `json:"name,omitempty"`
	Path      string `json:"path,omitempty"`
}

// EngramJSON is the JSON representation of engram binary status.
type EngramJSON struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
	Path      string `json:"path,omitempty"`
}

// PrintStatusJSON outputs the status information as JSON to stdout.
func PrintStatusJSON(status *StatusInfo) error {
	sj := statusToJSON(status)
	output, err := json.MarshalIndent(sj, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

// statusToJSON converts a StatusInfo to a StatusJSON struct.
func statusToJSON(status *StatusInfo) StatusJSON {
	sj := StatusJSON{
		Installed:    status.Installed,
		Version:      status.Version,
		ConfigPath:   status.ConfigPath,
		AgentsMD:     status.Components.AgentsMD.Present,
		OpenCodeJSON: status.Components.OpenCodeJSON.Present,
		Plugins:      status.Components.Plugins.Present,
	}

	// Theme
	sj.Theme = ThemeJSON{
		Installed: status.Components.Theme.Present,
	}
	if status.Components.Theme.Present {
		// Detail format is "filename (size)" — extract the filename
		detail := status.Components.Theme.Detail
		if idx := strings.Index(detail, " ("); idx > 0 {
			sj.Theme.Name = detail[:idx]
		}
		sj.Theme.Path = filepath.Join(status.ConfigPath, "themes")
	}

	// Engram binary
	sj.Engram = EngramJSON{
		Installed: status.Binaries.Engram.Installed,
		Version:   status.Binaries.Engram.Version,
		Path:      status.Binaries.Engram.Path,
	}

	// Skills count — parse from detail string like "27 directories"
	sj.SkillsCount = parseCountFromDetail(status.Components.Skills.Detail)

	return sj
}

// parseCountFromDetail extracts a number from a detail string like "27 directories" or "5 files".
func parseCountFromDetail(detail string) int {
	var count int
	if _, err := fmt.Sscanf(detail, "%d", &count); err != nil {
		return 0
	}
	return count
}
