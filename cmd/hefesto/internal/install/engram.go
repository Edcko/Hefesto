// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const engramVersion = "1.3.1"

func getEngramDownloadURL() (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	var binary string
	switch {
	case osName == "darwin" && arch == "arm64":
		binary = fmt.Sprintf("engram_%s_darwin_arm64.tar.gz", engramVersion)
	case osName == "darwin" && arch == "amd64":
		binary = fmt.Sprintf("engram_%s_darwin_amd64.tar.gz", engramVersion)
	case osName == "linux" && arch == "arm64":
		binary = fmt.Sprintf("engram_%s_linux_arm64.tar.gz", engramVersion)
	case osName == "linux" && arch == "amd64":
		binary = fmt.Sprintf("engram_%s_linux_amd64.tar.gz", engramVersion)
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", osName, arch)
	}

	return fmt.Sprintf("https://github.com/Gentleman-Programming/engram/releases/download/v%s/%s", engramVersion, binary), nil
}

func InstallEngram(ctx context.Context) error {
	// Check if already installed
	if _, err := exec.LookPath("engram"); err == nil {
		// Already installed, verify version
		out, err := exec.Command("engram", "version").Output()
		if err == nil && strings.Contains(string(out), engramVersion) {
			return nil // Already installed with correct version
		}
	}

	// Get download URL
	url, err := getEngramDownloadURL()
	if err != nil {
		return fmt.Errorf("engram install: %w", err)
	}

	// Download and extract
	tmpDir, err := os.MkdirTemp("", "engram-install-*")
	if err != nil {
		return fmt.Errorf("engram install temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download
	tarPath := tmpDir + "/engram.tar.gz"
	curlCmd := exec.CommandContext(ctx, "curl", "-fsSL", url, "-o", tarPath)
	if out, err := curlCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("engram download failed: %w\n%s", err, out)
	}

	// Extract
	tarCmd := exec.Command("tar", "-xzf", tarPath, "-C", tmpDir)
	if out, err := tarCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("engram extract failed: %w\n%s", err, out)
	}

	// Find the binary (it's just called "engram")
	binaryPath := tmpDir + "/engram"
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("engram binary not found in archive")
	}

	// Move to /usr/local/bin (or user-writable location)
	targetPath := "/usr/local/bin/engram"
	if err := os.Rename(binaryPath, targetPath); err != nil {
		// Try with sudo or user-local bin
		targetPath = os.Getenv("HOME") + "/.local/bin/engram"
		os.MkdirAll(os.Getenv("HOME")+"/.local/bin", 0755)
		if err := copyFile(binaryPath, targetPath); err != nil {
			return fmt.Errorf("engram install to %s: %w", targetPath, err)
		}
	}

	// Make executable
	os.Chmod(targetPath, 0755)

	// Verify
	out, err := exec.Command(targetPath, "version").Output()
	if err != nil {
		return fmt.Errorf("engram verify failed: %w", err)
	}

	if !strings.Contains(string(out), engramVersion) {
		return fmt.Errorf("engram version mismatch: got %s", strings.TrimSpace(string(out)))
	}

	return nil
}
