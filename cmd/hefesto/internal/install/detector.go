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

	"github.com/Edcko/Hefesto/cmd/hefesto/internal/logger"
)

// getUserHomeDir returns the user's home directory, prioritizing the $HOME
// environment variable. This is necessary because os.UserHomeDir() on macOS
// does not respect $HOME when changed via t.Setenv() in tests.
func getUserHomeDir() (string, error) {
	// First check $HOME environment variable (respects t.Setenv in tests)
	if homeDir := os.Getenv("HOME"); homeDir != "" {
		return homeDir, nil
	}
	// Fallback to os.UserHomeDir() for systems where $HOME is not set
	return os.UserHomeDir()
}

// DetectHomeDir is the exported version of getUserHomeDir for use by commands
// in the main package.
func DetectHomeDir() (string, error) {
	return getUserHomeDir()
}

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
	homeDir, err := getUserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	env.ConfigPath = filepath.Join(homeDir, ".config", "opencode")
	logger.Debug("detect: config path resolved to %s", env.ConfigPath)

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
		logger.Debug("detect: opencode found at %s (version %s)", env.OpenCodePath, env.OpenCodeVersion)
	} else {
		// Fallback: check ~/.opencode/bin/opencode — the official installer puts it there
		// but it may not be in PATH yet.
		fallbackPath := GetOpenCodeInstallPath()
		if info, statErr := os.Stat(fallbackPath); statErr == nil && !info.IsDir() {
			env.OpenCodeInstalled = true
			env.OpenCodePath = fallbackPath

			// Get version from the fallback binary
			cmd := exec.CommandContext(ctx, fallbackPath, "--version") //nolint:gosec // G702: controlled binary path
			if out, cmdErr := cmd.Output(); cmdErr == nil {
				env.OpenCodeVersion = parseOpenCodeVersionOutput(string(out))
			}
			logger.Debug("detect: opencode found at fallback %s (version %s)", env.OpenCodePath, env.OpenCodeVersion)
		} else {
			logger.Debug("detect: opencode not found in PATH or ~/.opencode/bin/")
		}
	}

	// Check if engram is installed
	if p, err := exec.LookPath("engram"); err == nil {
		env.EngramInstalled = true
		env.EngramPath = p
		if out, err := exec.Command("engram", "version").Output(); err == nil {
			env.EngramVersion = strings.TrimSpace(string(out))
		}
		logger.Debug("detect: engram found at %s (version %s)", env.EngramPath, env.EngramVersion)
	} else {
		logger.Debug("detect: engram not found in PATH")
	}

	// Check if config directory exists AND has actual config files.
	// We can't rely on directory existence alone because `opencode --version`
	// (called above in getOpenCodeVersion) auto-creates the directory as a
	// side-effect. Only treat as "existing" if AGENTS.md or opencode.json
	// are present — those indicate a real installation.
	if _, err := os.Stat(env.ConfigPath); err == nil {
		agentsExists := fileExists(filepath.Join(env.ConfigPath, "AGENTS.md"))
		jsonExists := fileExists(filepath.Join(env.ConfigPath, "opencode.json"))
		if agentsExists || jsonExists {
			env.ConfigExists = true
			env.ExistingConfig = detectExistingConfig(env.ConfigPath)
		} else {
			env.ExistingConfig = "none"
		}
	} else {
		env.ExistingConfig = "none"
	}
	logger.Debug("detect: config exists=%v, type=%s, path=%s", env.ConfigExists, env.ExistingConfig, env.ConfigPath)

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

// fileExists returns true if the file at path exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// detectExistingConfig analyzes the existing config to determine its type.
func detectExistingConfig(configPath string) string {
	// Check for AGENTS.md
	agentsPath := filepath.Join(configPath, "AGENTS.md")
	content, err := os.ReadFile(agentsPath) //nolint:gosec // G304: agentsPath is built from known config path
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
