// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Environment holds information about the current system environment.
type Environment struct {
	OpenCodeInstalled bool   // Whether opencode CLI is installed
	OpenCodeVersion   string // Version of opencode if installed
	OpenCodePath      string // Path to opencode binary
	EngramInstalled   bool   // Whether engram CLI is installed
	EngramVersion     string // Version of engram if installed
	EngramPath        string // Path to engram binary
	ConfigExists      bool   // Whether ~/.config/opencode/ exists
	ConfigPath        string // Path to opencode config directory
	ExistingConfig    string // "gentleman-dots", "hefesto", "custom", "none"
	Platform          string // "darwin", "linux"
	Arch              string // "arm64", "amd64"
	IsAppleSilicon    bool   // True if macOS on ARM
}

// Detect analyzes the current environment and returns an Environment struct.
func Detect() (*Environment, error) {
	env := &Environment{
		Platform: runtime.GOOS,
		Arch:     runtime.GOARCH,
	}

	// Detect Apple Silicon
	env.IsAppleSilicon = env.Platform == "darwin" && env.Arch == "arm64"

	// Get config path using home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	env.ConfigPath = filepath.Join(homeDir, ".config", "opencode")

	// Check if opencode is installed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	path, err := exec.LookPath("opencode")
	if err == nil {
		env.OpenCodeInstalled = true
		env.OpenCodePath = path

		// Get version
		version, err := getOpenCodeVersion(ctx)
		if err == nil {
			env.OpenCodeVersion = version
		}
	}

	// Check if engram is installed
	if p, err := exec.LookPath("engram"); err == nil {
		env.EngramInstalled = true
		env.EngramPath = p
		if out, err := exec.Command("engram", "version").Output(); err == nil {
			env.EngramVersion = strings.TrimSpace(string(out))
		}
	}

	// Check if config directory exists
	if _, err := os.Stat(env.ConfigPath); err == nil {
		env.ConfigExists = true
		env.ExistingConfig = detectExistingConfig(env.ConfigPath)
	} else {
		env.ExistingConfig = "none"
	}

	return env, nil
}

// getOpenCodeVersion runs opencode --version and returns the version string.
func getOpenCodeVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "opencode", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get opencode version: %w", err)
	}

	// Parse version from output (usually "opencode version X.Y.Z" or just "X.Y.Z")
	version := strings.TrimSpace(string(output))
	lines := strings.Split(version, "\n")
	if len(lines) > 0 {
		version = strings.TrimSpace(lines[0])
	}

	// Extract version number if it contains "version" prefix
	if idx := strings.Index(strings.ToLower(version), "version"); idx != -1 {
		version = strings.TrimSpace(version[idx+7:])
	}

	return version, nil
}

// detectExistingConfig analyzes the existing config to determine its type.
func detectExistingConfig(configPath string) string {
	// Check for AGENTS.md
	agentsPath := filepath.Join(configPath, "AGENTS.md")
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		return "custom"
	}

	agentsContent := string(content)

	// Check for Hefesto markers
	if strings.Contains(agentsContent, "Hefesto") {
		return "hefesto"
	}

	// Check for Gentleman.Dots markers
	if strings.Contains(agentsContent, "Gentleman") {
		return "gentleman-dots"
	}

	return "custom"
}
