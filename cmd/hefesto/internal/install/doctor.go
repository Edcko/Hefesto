// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// DoctorResult holds the results of the doctor diagnostic.
type DoctorResult struct {
	ConfigDir    CheckResult
	AgentsMD     CheckResult
	OpenCodeJSON CheckResult
	Skills       CheckResult
	Plugins      CheckResult
	Engram       CheckResult
	OpenCode     CheckResult
	Theme        CheckResult
	Personality  CheckResult
	Commands     CheckResult
}

// CheckResult holds the result of a single check category.
type CheckResult struct {
	Passed   bool
	Warnings []string
	Errors   []string
	Details  []string
}

// DoctorCheck represents a single diagnostic check.
type DoctorCheck struct {
	Name   string
	Result CheckResult
}

// RunDoctor performs all diagnostic checks and returns the results.
func RunDoctor() (*DoctorResult, int) {
	result := &DoctorResult{}
	exitCode := 0

	// Run all checks
	result.ConfigDir = checkConfigDir()
	result.AgentsMD = checkAgentsMD()
	result.OpenCodeJSON = checkOpenCodeJSON()
	result.Skills = checkSkills()
	result.Plugins = checkPlugins()
	result.Engram = checkEngram()
	result.OpenCode = checkOpenCode()
	result.Theme = checkTheme()
	result.Personality = checkPersonality()
	result.Commands = checkCommands()

	// Determine exit code
	hasErrors := false
	hasWarnings := false

	checks := []CheckResult{
		result.ConfigDir,
		result.AgentsMD,
		result.OpenCodeJSON,
		result.Skills,
		result.Plugins,
		result.Engram,
		result.OpenCode,
		result.Theme,
		result.Personality,
		result.Commands,
	}

	for _, check := range checks {
		if len(check.Errors) > 0 {
			hasErrors = true
		}
		if len(check.Warnings) > 0 {
			hasWarnings = true
		}
	}

	if hasErrors {
		exitCode = 2
	} else if hasWarnings {
		exitCode = 1
	}

	return result, exitCode
}

// checkConfigDir checks if the config directory exists and is accessible.
func checkConfigDir() CheckResult {
	result := CheckResult{Passed: true}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "Cannot determine home directory")
		return result
	}

	configPath := filepath.Join(homeDir, ".config", "opencode")

	// Check if directory exists
	info, err := os.Stat(configPath)
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "~/.config/opencode/ does not exist")
		return result
	}

	result.Details = append(result.Details, "~/.config/opencode/ exists")

	// Check if it's a directory
	if !info.IsDir() {
		result.Passed = false
		result.Errors = append(result.Errors, "~/.config/opencode/ is not a directory")
		return result
	}

	// Check if readable
	file, err := os.Open(configPath)
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "~/.config/opencode/ is not readable")
		return result
	}
	file.Close()

	// Check if writable by trying to create a temp file
	testFile := filepath.Join(configPath, ".hefesto-write-test")
	f, err := os.Create(testFile)
	if err != nil {
		result.Warnings = append(result.Warnings, "~/.config/opencode/ is not writable")
	} else {
		f.Close()
		os.Remove(testFile)
		result.Details = append(result.Details, "Readable and writable")
	}

	return result
}

// checkAgentsMD checks if AGENTS.md exists and is valid.
func checkAgentsMD() CheckResult {
	result := CheckResult{Passed: true}

	homeDir, _ := os.UserHomeDir()
	agentsPath := filepath.Join(homeDir, ".config", "opencode", "AGENTS.md")

	// Check if file exists
	info, err := os.Stat(agentsPath)
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "AGENTS.md does not exist")
		return result
	}

	size := formatBytes(info.Size())
	result.Details = append(result.Details, fmt.Sprintf("File exists (%s)", size))

	// Check if non-empty
	if info.Size() == 0 {
		result.Passed = false
		result.Errors = append(result.Errors, "AGENTS.md is empty")
		return result
	}

	// Read content
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "Cannot read AGENTS.md")
		return result
	}

	// Check for Hefesto keyword
	if !strings.Contains(string(content), "Hefesto") {
		result.Warnings = append(result.Warnings, "AGENTS.md does not contain Hefesto configuration (may be from another config)")
	} else {
		result.Details = append(result.Details, "Contains Hefesto configuration")
	}

	// Check if valid UTF-8
	if !isValidUTF8(content) {
		result.Passed = false
		result.Errors = append(result.Errors, "AGENTS.md is not valid UTF-8")
		return result
	}
	result.Details = append(result.Details, "Valid UTF-8")

	return result
}

// checkOpenCodeJSON checks if opencode.json exists and is valid.
func checkOpenCodeJSON() CheckResult {
	result := CheckResult{Passed: true}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "opencode", "opencode.json")

	// Check if file exists
	content, err := os.ReadFile(configPath)
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "opencode.json does not exist")
		return result
	}

	// Check if valid JSON
	var js map[string]interface{}
	if err := json.Unmarshal(content, &js); err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf("opencode.json is not valid JSON: %v", err))
		return result
	}
	result.Details = append(result.Details, "Valid JSON")

	// Check for agents key
	agents, ok := js["agent"].(map[string]interface{})
	if !ok {
		result.Passed = false
		result.Errors = append(result.Errors, "opencode.json missing 'agent' key")
		return result
	}

	agentCount := len(agents)
	result.Details = append(result.Details, fmt.Sprintf("%d agents configured", agentCount))

	// Check for required agents
	requiredAgents := []string{"hefesto", "sdd-orchestrator"}
	for _, agent := range requiredAgents {
		if _, exists := agents[agent]; !exists {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Agent %q not configured", agent))
		}
	}

	// Check if agent paths are absolute
	relativePathCount := 0
	for agentName, agentConfig := range agents {
		if ac, ok := agentConfig.(map[string]interface{}); ok {
			if prompt, ok := ac["prompt"].(string); ok {
				// Check if prompt uses relative path (starts with ./) instead of absolute
				if strings.HasPrefix(prompt, "{file:./") {
					relativePathCount++
				} else if strings.HasPrefix(prompt, "{file:~/.config") || strings.HasPrefix(prompt, "{file:~/") {
					// Absolute path is good
				}
			}
			// Check for missing steps limit on subagents
			if mode, ok := ac["mode"].(string); ok && mode == "subagent" {
				if _, hasSteps := ac["steps"]; !hasSteps {
					if agentName != "remote-exec" { // remote-exec doesn't need steps
						result.Warnings = append(result.Warnings, fmt.Sprintf("Agent %q missing steps limit", agentName))
					}
				}
			}
		}
	}

	if relativePathCount > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("%d agents use relative paths (should be absolute)", relativePathCount))
	} else {
		result.Details = append(result.Details, "Agent paths are absolute")
	}

	return result
}

// checkSkills checks if skills directory exists and has all expected skills.
func checkSkills() CheckResult {
	result := CheckResult{Passed: true}

	homeDir, _ := os.UserHomeDir()
	skillsPath := filepath.Join(homeDir, ".config", "opencode", "skills")

	// Check if directory exists
	info, err := os.Stat(skillsPath)
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "skills/ directory does not exist")
		return result
	}

	if !info.IsDir() {
		result.Passed = false
		result.Errors = append(result.Errors, "skills/ is not a directory")
		return result
	}

	// Count skill directories with SKILL.md
	contents, err := ioutil.ReadDir(skillsPath)
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "Cannot read skills directory")
		return result
	}

	skillCount := 0
	missingSkillMD := []string{}

	for _, c := range contents {
		if c.IsDir() && !strings.HasPrefix(c.Name(), ".") && c.Name() != "_shared" {
			skillMDPath := filepath.Join(skillsPath, c.Name(), "SKILL.md")
			if _, err := os.Stat(skillMDPath); err == nil {
				skillCount++
			} else {
				missingSkillMD = append(missingSkillMD, c.Name())
			}
		}
	}

	// Count _shared as well if it exists
	sharedPath := filepath.Join(skillsPath, "_shared")
	if sharedInfo, err := os.Stat(sharedPath); err == nil && sharedInfo.IsDir() {
		// _shared is a special directory, count it
	}

	result.Details = append(result.Details, fmt.Sprintf("%d skills found", skillCount))

	// Expected skill count (26 based on the embed)
	expectedCount := 26
	if skillCount < expectedCount {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Expected %d skills, found %d", expectedCount, skillCount))
	}

	if len(missingSkillMD) > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Skills missing SKILL.md: %s", strings.Join(missingSkillMD, ", ")))
	} else {
		result.Details = append(result.Details, "All skills have SKILL.md")
	}

	return result
}

// checkPlugins checks if plugins are properly installed.
func checkPlugins() CheckResult {
	result := CheckResult{Passed: true}

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "opencode")
	pluginsPath := filepath.Join(configPath, "plugins")

	// Check plugins directory
	if info, err := os.Stat(pluginsPath); err != nil || !info.IsDir() {
		result.Warnings = append(result.Warnings, "plugins/ directory does not exist")
		return result
	}

	// Check engram.ts
	engramPath := filepath.Join(pluginsPath, "engram.ts")
	if info, err := os.Stat(engramPath); err != nil {
		result.Warnings = append(result.Warnings, "engram.ts is missing")
	} else {
		result.Details = append(result.Details, fmt.Sprintf("engram.ts (%s)", formatBytes(info.Size())))
	}

	// Check background-agents.ts
	bgAgentsPath := filepath.Join(pluginsPath, "background-agents.ts")
	if info, err := os.Stat(bgAgentsPath); err != nil {
		result.Warnings = append(result.Warnings, "background-agents.ts is missing")
	} else {
		result.Details = append(result.Details, fmt.Sprintf("background-agents.ts (%s)", formatBytes(info.Size())))
	}

	// Check package.json
	packageJSONPath := filepath.Join(configPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err != nil {
		result.Warnings = append(result.Warnings, "package.json is missing")
	}

	// Check node_modules
	nodeModulesPath := filepath.Join(configPath, "node_modules")
	if info, err := os.Stat(nodeModulesPath); err != nil || !info.IsDir() {
		result.Warnings = append(result.Warnings, "npm dependencies not installed (node_modules/ missing)")
	} else {
		// Check @opencode-ai/plugin
		pluginSDKPath := filepath.Join(nodeModulesPath, "@opencode-ai", "plugin")
		if _, err := os.Stat(pluginSDKPath); err != nil {
			result.Warnings = append(result.Warnings, "@opencode-ai/plugin SDK not installed")
		} else {
			result.Details = append(result.Details, "npm dependencies installed")
		}
	}

	return result
}

// checkEngram checks if engram binary is installed and working.
func checkEngram() CheckResult {
	result := CheckResult{Passed: true}

	// Check if engram is in PATH
	engramPath, err := exec.LookPath("engram")
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "engram binary not found in PATH")
		result.Details = append(result.Details, "Install with: brew install engram")
		return result
	}

	// Get version
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	output, err := exec.CommandContext(ctx, "engram", "version").Output()
	if err != nil {
		result.Warnings = append(result.Warnings, "Cannot determine engram version")
		result.Details = append(result.Details, fmt.Sprintf("engram installed at %s", engramPath))
		return result
	}

	version := strings.TrimSpace(string(output))

	// Parse version and check minimum
	minVersion := "1.3.1"
	if !isVersionAtLeast(version, minVersion) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("engram version %s is below recommended %s", version, minVersion))
	}

	result.Details = append(result.Details, fmt.Sprintf("engram %s (%s)", version, engramPath))

	// Check if MCP command exists
	mcpCmd := exec.CommandContext(ctx, "engram", "mcp", "--help")
	if err := mcpCmd.Run(); err != nil {
		result.Warnings = append(result.Warnings, "engram MCP server mode not available")
	} else {
		result.Details = append(result.Details, "MCP server available")
	}

	return result
}

// checkOpenCode checks if opencode binary is installed and working.
func checkOpenCode() CheckResult {
	result := CheckResult{Passed: true}

	// Check if opencode is in PATH
	opencodePath, err := exec.LookPath("opencode")
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, "opencode binary not found in PATH")
		return result
	}

	// Get version
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	version, err := getOpenCodeVersion(ctx)
	if err != nil {
		result.Warnings = append(result.Warnings, "Cannot determine opencode version")
		result.Details = append(result.Details, fmt.Sprintf("opencode installed at %s", opencodePath))
		return result
	}

	// Check minimum version (1.3.13)
	minVersion := "1.3.13"
	if !isVersionAtLeast(version, minVersion) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("opencode version %s is below recommended %s", version, minVersion))
	}

	result.Details = append(result.Details, fmt.Sprintf("opencode %s (%s)", version, opencodePath))

	return result
}

// checkTheme checks if themes are properly configured.
func checkTheme() CheckResult {
	result := CheckResult{Passed: true}

	homeDir, _ := os.UserHomeDir()
	themesPath := filepath.Join(homeDir, ".config", "opencode", "themes")

	// Check if themes directory exists
	info, err := os.Stat(themesPath)
	if err != nil {
		result.Warnings = append(result.Warnings, "themes/ directory does not exist")
		return result
	}

	if !info.IsDir() {
		result.Warnings = append(result.Warnings, "themes/ is not a directory")
		return result
	}

	// Find JSON files
	contents, err := ioutil.ReadDir(themesPath)
	if err != nil {
		result.Warnings = append(result.Warnings, "Cannot read themes directory")
		return result
	}

	jsonFiles := []string{}
	for _, c := range contents {
		if !c.IsDir() && strings.HasSuffix(c.Name(), ".json") {
			jsonFiles = append(jsonFiles, c.Name())

			// Validate JSON
			themePath := filepath.Join(themesPath, c.Name())
			content, err := os.ReadFile(themePath)
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Cannot read %s", c.Name()))
				continue
			}

			var theme map[string]interface{}
			if err := json.Unmarshal(content, &theme); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s is not valid JSON", c.Name()))
				continue
			}

			// Check for expected keys
			expectedKeys := []string{"name"}
			for _, key := range expectedKeys {
				if _, exists := theme[key]; !exists {
					result.Warnings = append(result.Warnings, fmt.Sprintf("%s missing %q key", c.Name(), key))
				}
			}

			result.Details = append(result.Details, fmt.Sprintf("%s (valid JSON)", c.Name()))
		}
	}

	if len(jsonFiles) == 0 {
		result.Warnings = append(result.Warnings, "No theme files found")
	}

	return result
}

// checkPersonality checks if personality configuration exists.
func checkPersonality() CheckResult {
	result := CheckResult{Passed: true}

	homeDir, _ := os.UserHomeDir()
	personalityPath := filepath.Join(homeDir, ".config", "opencode", "personality")

	// Check if personality directory exists
	info, err := os.Stat(personalityPath)
	if err != nil {
		result.Warnings = append(result.Warnings, "personality/ directory does not exist")
		return result
	}

	if !info.IsDir() {
		result.Warnings = append(result.Warnings, "personality/ is not a directory")
		return result
	}

	// Check for hefesto.md
	hefestoMDPath := filepath.Join(personalityPath, "hefesto.md")
	if info, err := os.Stat(hefestoMDPath); err != nil {
		result.Warnings = append(result.Warnings, "hefesto.md is missing")
	} else {
		if info.Size() == 0 {
			result.Warnings = append(result.Warnings, "hefesto.md is empty")
		} else {
			result.Details = append(result.Details, fmt.Sprintf("hefesto.md (%s)", formatBytes(info.Size())))
		}
	}

	return result
}

// checkCommands checks if slash commands are installed.
func checkCommands() CheckResult {
	result := CheckResult{Passed: true}

	homeDir, _ := os.UserHomeDir()
	commandsPath := filepath.Join(homeDir, ".config", "opencode", "commands")

	// Check if commands directory exists
	info, err := os.Stat(commandsPath)
	if err != nil {
		result.Warnings = append(result.Warnings, "commands/ directory does not exist")
		return result
	}

	if !info.IsDir() {
		result.Warnings = append(result.Warnings, "commands/ is not a directory")
		return result
	}

	// Count .md files
	contents, err := ioutil.ReadDir(commandsPath)
	if err != nil {
		result.Warnings = append(result.Warnings, "Cannot read commands directory")
		return result
	}

	commandCount := 0
	expectedCommands := []string{"sdd-init.md", "sdd-new.md", "sdd-apply.md", "sdd-verify.md", "sdd-ff.md"}
	missingCommands := []string{}

	for _, c := range contents {
		if !c.IsDir() && strings.HasSuffix(c.Name(), ".md") {
			commandCount++
		}
	}

	// Check for expected SDD commands
	for _, expected := range expectedCommands {
		found := false
		for _, c := range contents {
			if c.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			missingCommands = append(missingCommands, expected)
		}
	}

	if len(missingCommands) > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Missing commands: %s", strings.Join(missingCommands, ", ")))
	}

	result.Details = append(result.Details, fmt.Sprintf("%d slash commands", commandCount))

	return result
}

// PrintDoctor prints the doctor results in a formatted way.
func PrintDoctor(result *DoctorResult) {
	fmt.Println()
	fmt.Println("🔥 Hefesto Doctor — Diagnosing installation...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	checks := []struct {
		name   string
		result CheckResult
	}{
		{"Config Directory", result.ConfigDir},
		{"AGENTS.md", result.AgentsMD},
		{"opencode.json", result.OpenCodeJSON},
		{"Skills", result.Skills},
		{"Plugins", result.Plugins},
		{"Engram", result.Engram},
		{"OpenCode", result.OpenCode},
		{"Theme", result.Theme},
		{"Personality", result.Personality},
		{"Commands", result.Commands},
	}

	allErrors := []string{}
	allWarnings := []string{}

	for _, check := range checks {
		fmt.Printf("  Checking: %s\n", check.name)

		// Print details (✅)
		for _, detail := range check.result.Details {
			fmt.Printf("    ✅ %s\n", detail)
		}

		// Print warnings (⚠️)
		for _, warning := range check.result.Warnings {
			fmt.Printf("    ⚠️  %s\n", warning)
			allWarnings = append(allWarnings, warning)
		}

		// Print errors (❌)
		for _, err := range check.result.Errors {
			fmt.Printf("    ❌ %s\n", err)
			allErrors = append(allErrors, err)
		}

		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Summary
	if len(allErrors) == 0 && len(allWarnings) == 0 {
		fmt.Println("  🟢 All checks passed — Hefesto is healthy!")
	} else if len(allErrors) == 0 {
		fmt.Printf("  🟡 %d warning(s) found, but no errors\n", len(allWarnings))
		for i, w := range allWarnings {
			fmt.Printf("    %d. %s\n", i+1, w)
		}
	} else {
		fmt.Printf("  🔴 %d issue(s) found:\n", len(allErrors))
		for i, e := range allErrors {
			fmt.Printf("    %d. %s\n", i+1, e)
		}
		fmt.Println()
		fmt.Println("  Run `hefesto install` to fix these issues.")
	}

	fmt.Println()
}

// Helper functions

// isValidUTF8 checks if a byte slice is valid UTF-8.
func isValidUTF8(b []byte) bool {
	return utf8.Valid(b)
}

// isVersionAtLeast checks if version is >= minVersion.
// Supports semver-like versions (e.g., "1.3.1", "v1.3.1").
func isVersionAtLeast(version, minVersion string) bool {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")
	minVersion = strings.TrimPrefix(minVersion, "v")

	// Parse versions
	vParts := strings.Split(version, ".")
	mParts := strings.Split(minVersion, ".")

	// Compare each part
	for i := 0; i < len(vParts) && i < len(mParts); i++ {
		vNum := extractNumber(vParts[i])
		mNum := extractNumber(mParts[i])

		if vNum > mNum {
			return true
		}
		if vNum < mNum {
			return false
		}
	}

	// If all compared parts are equal, check length
	return len(vParts) >= len(mParts)
}

// extractNumber extracts the leading number from a version part.
func extractNumber(s string) int {
	re := regexp.MustCompile(`^\d+`)
	match := re.FindString(s)
	if match == "" {
		return 0
	}
	var num int
	fmt.Sscanf(match, "%d", &num)
	return num
}
