// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// VerifyResult holds the results of post-install verification.
type VerifyResult struct {
	ConfigCopied    bool     // Whether opencode.json exists and is valid
	NpmInstalled    bool     // Whether node_modules/@opencode-ai/plugin exists
	OpenCodeWorks   bool     // Whether opencode --version runs successfully
	EngramInstalled bool     // Whether engram CLI is installed
	Errors          []string // List of errors encountered during verification
}

// Verify performs post-installation verification checks.
func Verify(configPath string) (*VerifyResult, error) {
	result := &VerifyResult{
		Errors: []string{},
	}

	// Check that opencode.json exists and is valid JSON
	result.ConfigCopied = verifyConfigJSON(configPath, result)

	// Check that AGENTS.md exists
	if !verifyAgentsMD(configPath, result) {
		// AGENTS.md is critical, note if missing
		result.Errors = append(result.Errors, "AGENTS.md is missing")
	}

	// Check that node_modules/@opencode-ai/plugin exists
	result.NpmInstalled = verifyPluginModule(configPath, result)

	// Try opencode --version to verify OpenCode still works
	result.OpenCodeWorks = verifyOpenCodeWorks(result)

	// Check engram installation
	result.EngramInstalled = verifyEngramWorks(result)

	return result, nil
}

// verifyConfigJSON checks that opencode.json exists and is valid JSON.
func verifyConfigJSON(configPath string, result *VerifyResult) bool {
	configFile := filepath.Join(configPath, "opencode.json")

	content, err := os.ReadFile(configFile) //nolint:gosec // G304: configFile built from known config directory
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("opencode.json not found: %v", err))
		return false
	}

	// Validate JSON
	var js map[string]interface{}
	if err := json.Unmarshal(content, &js); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("opencode.json is not valid JSON: %v", err))
		return false
	}

	return true
}

// verifyAgentsMD checks that AGENTS.md exists.
func verifyAgentsMD(configPath string, result *VerifyResult) bool {
	agentsFile := filepath.Join(configPath, "AGENTS.md")

	if _, err := os.Stat(agentsFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// verifyPluginModule checks that the OpenCode plugin module exists.
func verifyPluginModule(configPath string, result *VerifyResult) bool {
	pluginPath := filepath.Join(configPath, "node_modules", "@opencode-ai", "plugin")

	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		result.Errors = append(result.Errors, "@opencode-ai/plugin not found in node_modules")
		return false
	}

	return true
}

// verifyOpenCodeWorks checks that opencode CLI is still functional.
func verifyOpenCodeWorks(result *VerifyResult) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "opencode", "--version")
	if err := cmd.Run(); err != nil {
		// Try fallback location: ~/.opencode/bin/opencode
		installPath := GetOpenCodeInstallPath()
		if _, statErr := os.Stat(installPath); statErr == nil {
			fallbackCmd := exec.CommandContext(ctx, installPath, "--version") //nolint:gosec // G702: controlled binary path
			if fallbackErr := fallbackCmd.Run(); fallbackErr == nil {
				// Binary exists at fallback but not in PATH — suggest adding it
				shellRC := getShellRCFile(func() string {
					h, _ := getUserHomeDir()
					return h
				}())
				result.Errors = append(result.Errors,
					fmt.Sprintf("opencode installed at %s but not in PATH. Add it by running: echo 'export PATH=\"$HOME/.opencode/bin:$PATH\"' >> %s", installPath, shellRC))
				return false
			}
		}

		result.Errors = append(result.Errors, fmt.Sprintf("opencode CLI not working: %v", err))
		return false
	}

	return true
}

// verifyEngramWorks checks that engram CLI is installed and working.
func verifyEngramWorks(result *VerifyResult) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "engram", "version")
	if err := cmd.Run(); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("engram CLI not working: %v", err))
		return false
	}

	return true
}
