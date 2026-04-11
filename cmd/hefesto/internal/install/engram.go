// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	// engramFallbackVersion is used when the GitHub API is unreachable.
	engramFallbackVersion = "1.3.1"
	engramGitHubAPI       = "https://api.github.com/repos/Gentleman-Programming/engram/releases/latest"
)

// cachedEngramVersion caches the resolved version for the process lifetime,
// so we only hit the GitHub API once per run.
var cachedEngramVersion string

// resolveEngramVersion returns the latest engram version. It fetches the
// latest release from the GitHub API on the first call and caches the result.
// If the API call fails for any reason, it falls back to engramFallbackVersion.
func resolveEngramVersion() string {
	if cachedEngramVersion != "" {
		return cachedEngramVersion
	}

	version, err := fetchLatestEngramVersion()
	if err != nil {
		log.Printf("engram: failed to fetch latest version from GitHub, using fallback %s: %v", engramFallbackVersion, err)
		cachedEngramVersion = engramFallbackVersion
		return cachedEngramVersion
	}

	cachedEngramVersion = version
	return cachedEngramVersion
}

// fetchLatestEngramVersion calls the GitHub releases API and returns the
// latest version string (without the "v" prefix).
func fetchLatestEngramVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, engramGitHubAPI, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decoding GitHub response: %w", err)
	}

	version := strings.TrimPrefix(release.TagName, "v")
	if version == "" {
		return "", fmt.Errorf("empty version tag from GitHub API")
	}

	return version, nil
}

// ResetEngramVersionCache clears the cached version, useful for testing.
func ResetEngramVersionCache() {
	cachedEngramVersion = ""
}

func getEngramDownloadURL(version string) (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	var binary string
	switch {
	case osName == "darwin" && arch == "arm64":
		binary = fmt.Sprintf("engram_%s_darwin_arm64.tar.gz", version)
	case osName == "darwin" && arch == "amd64":
		binary = fmt.Sprintf("engram_%s_darwin_amd64.tar.gz", version)
	case osName == "linux" && arch == "arm64":
		binary = fmt.Sprintf("engram_%s_linux_arm64.tar.gz", version)
	case osName == "linux" && arch == "amd64":
		binary = fmt.Sprintf("engram_%s_linux_amd64.tar.gz", version)
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", osName, arch)
	}

	return fmt.Sprintf("https://github.com/Gentleman-Programming/engram/releases/download/v%s/%s", version, binary), nil
}

func InstallEngram(ctx context.Context) error {
	version := resolveEngramVersion()

	// Check if already installed
	if _, err := exec.LookPath("engram"); err == nil {
		// Already installed, verify version
		out, err := exec.Command("engram", "version").Output()
		if err == nil && strings.Contains(string(out), version) {
			return nil // Already installed with correct version
		}
	}

	// Get download URL
	url, err := getEngramDownloadURL(version)
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
		if err := CopyFile(binaryPath, targetPath); err != nil {
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

	if !strings.Contains(string(out), version) {
		return fmt.Errorf("engram version mismatch: expected %s, got %s", version, strings.TrimSpace(string(out)))
	}

	return nil
}
