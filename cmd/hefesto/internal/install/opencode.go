// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Edcko/Hefesto/cmd/hefesto/internal/logger"
)

const (
	// openCodeInstallScript is the official OpenCode installer URL.
	openCodeInstallScript = "https://opencode.ai/install"

	// openCodeBinDir is the directory where the OpenCode installer puts the binary.
	openCodeBinDir = ".opencode/bin"
)

// InstallOpenCode downloads and installs the OpenCode CLI using the official
// installer. It runs: curl -fsSL https://opencode.ai/install | bash
// Returns the installed version or an error.
func InstallOpenCode() (string, error) {
	// Check prerequisites: curl and bash must be available
	if _, err := exec.LookPath("curl"); err != nil {
		return "", fmt.Errorf("curl is required but not found in PATH: %w", err)
	}
	if _, err := exec.LookPath("bash"); err != nil {
		return "", fmt.Errorf("bash is required but not found in PATH: %w", err)
	}

	// Run the official installer with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("curl -fsSL %s | bash", openCodeInstallScript)) //nolint:gosec // G204: controlled install URL
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logger.Debug("opencode: running official installer")

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("opencode install timed out after 120 seconds")
		}
		return "", fmt.Errorf("opencode install failed: %w\n%s", err, stderr.String())
	}

	// Ensure ~/.opencode/bin is in PATH for future sessions
	if err := ensureOpenCodeInPath(); err != nil {
		// Non-fatal: the binary is installed, just PATH may not be updated
		logger.Debug("opencode: PATH update warning (non-fatal): %v", err)
	}

	// Update the current process PATH so this session can find opencode.
	// The curl|bash installer puts the binary at ~/.opencode/bin/opencode,
	// but the current shell's PATH hasn't been refreshed yet.
	updateCurrentProcessPath()

	// Verify the binary actually exists at the expected location
	installPath := GetOpenCodeInstallPath()
	if _, statErr := os.Stat(installPath); statErr != nil {
		return "", fmt.Errorf("opencode installer completed but binary not found at %s: %w", installPath, statErr)
	}
	logger.Debug("opencode: binary verified at %s", installPath)

	// Verify installation by checking the binary path or running --version
	version, err := verifyOpenCodeInstall()
	if err != nil {
		return "", fmt.Errorf("opencode installed but verification failed: %w", err)
	}

	logger.Debug("opencode: installed version %s", version)
	return version, nil
}

// IsOpenCodeInstalled checks if opencode binary exists in PATH or ~/.opencode/bin/.
func IsOpenCodeInstalled() bool {
	// Check PATH first
	if _, err := exec.LookPath("opencode"); err == nil {
		return true
	}

	// Fallback: check ~/.opencode/bin/opencode
	installPath := GetOpenCodeInstallPath()
	if _, err := os.Stat(installPath); err == nil {
		return true
	}

	return false
}

// GetOpenCodeInstallPath returns the path where opencode is (or would be) installed.
// This returns the fallback path (~/.opencode/bin/opencode) which is where the
// official installer puts the binary.
func GetOpenCodeInstallPath() string {
	homeDir, err := getUserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, openCodeBinDir, "opencode")
}

// verifyOpenCodeInstall attempts to get the installed version of OpenCode.
// It first tries the binary in PATH, then falls back to ~/.opencode/bin/opencode.
func verifyOpenCodeInstall() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try opencode in PATH first (might work if installer updated current session)
	version, err := getOpenCodeVersion(ctx)
	if err == nil {
		return version, nil
	}

	// Fallback: try the known install location directly
	installPath := GetOpenCodeInstallPath()
	if _, statErr := os.Stat(installPath); statErr == nil {
		cmd := exec.CommandContext(ctx, installPath, "--version") //nolint:gosec // G702: controlled binary path
		output, cmdErr := cmd.Output()
		if cmdErr == nil {
			return parseOpenCodeVersionOutput(string(output)), nil
		}
	}

	return "", fmt.Errorf("could not verify opencode installation: %w", err)
}

// parseOpenCodeVersionOutput extracts the version from opencode --version output.
func parseOpenCodeVersionOutput(output string) string {
	version := strings.TrimSpace(output)
	lines := strings.Split(version, "\n")
	if len(lines) > 0 {
		version = strings.TrimSpace(lines[0])
	}

	// Extract version number if it contains "version" prefix
	if idx := strings.Index(strings.ToLower(version), "version"); idx != -1 {
		version = strings.TrimSpace(version[idx+7:])
	}

	return version
}

// ensureOpenCodeInPath adds ~/.opencode/bin/ to the user's shell RC file
// if it's not already there. This ensures opencode is available in future sessions.
func ensureOpenCodeInPath() error {
	homeDir, err := getUserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	pathExport := fmt.Sprintf(`export PATH="$HOME/%s:$PATH"`, openCodeBinDir)

	// Detect shell RC file
	rcFile := getShellRCFile(homeDir)
	if rcFile == "" {
		return fmt.Errorf("could not detect shell RC file")
	}

	// Read existing RC file
	content, err := os.ReadFile(rcFile) //nolint:gosec // G304: rcFile is a known shell config path
	if err != nil {
		if os.IsNotExist(err) {
			// Create the file with the export
			return os.WriteFile(rcFile, []byte(pathExport+"\n"), 0644) //nolint:gosec // G306: shell RC file
		}
		return fmt.Errorf("cannot read %s: %w", rcFile, err)
	}

	// Check if the path is already exported
	if strings.Contains(string(content), openCodeBinDir) {
		logger.Debug("opencode: %s already in PATH configuration", openCodeBinDir)
		return nil
	}

	// Append the export
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY, 0644) //nolint:gosec // G302,G304: appending to known shell RC file
	if err != nil {
		return fmt.Errorf("cannot write to %s: %w", rcFile, err)
	}
	defer func() { _ = f.Close() }()

	line := fmt.Sprintf("\n# Added by Hefesto - OpenCode CLI\n%s\n", pathExport)
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("cannot write to %s: %w", rcFile, err)
	}

	logger.Debug("opencode: added %s to %s", openCodeBinDir, rcFile)
	return nil
}

// updateCurrentProcessPath adds ~/.opencode/bin to the PATH of the current
// process environment. This ensures that subsequent exec.LookPath and
// exec.Command calls within this process can find the opencode binary
// without requiring the user to open a new terminal.
func updateCurrentProcessPath() {
	homeDir, err := getUserHomeDir()
	if err != nil {
		logger.Debug("opencode: cannot update process PATH: %v", err)
		return
	}

	opencodeBinDir := filepath.Join(homeDir, openCodeBinDir)
	currentPath := os.Getenv("PATH")

	// Check if already in PATH
	for _, dir := range filepath.SplitList(currentPath) {
		if dir == opencodeBinDir {
			logger.Debug("opencode: %s already in process PATH", opencodeBinDir)
			return
		}
	}

	newPath := opencodeBinDir + string(filepath.ListSeparator) + currentPath
	if err := os.Setenv("PATH", newPath); err != nil {
		logger.Debug("opencode: failed to update process PATH: %v", err)
		return
	}

	logger.Debug("opencode: added %s to current process PATH", opencodeBinDir)
}

// getShellRCFile returns the path to the user's shell RC file.
// Detects bash vs zsh and returns the appropriate file.
func getShellRCFile(homeDir string) string {
	shell := os.Getenv("SHELL")

	if strings.Contains(shell, "zsh") {
		return filepath.Join(homeDir, ".zshrc")
	}

	// Default to bashrc for bash and any other shell
	return filepath.Join(homeDir, ".bashrc")
}
