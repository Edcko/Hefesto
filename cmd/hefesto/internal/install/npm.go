// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Edcko/Hefesto/cmd/hefesto/internal/logger"
)

// NpmInstall runs npm install in the config directory.
// It has a 60-second timeout and returns an error if npm is not installed.
// Note: npm install failure is not fatal - config works without npm deps if plugin is already in node_modules.
func NpmInstall(configPath string) error {
	// Check if npm is installed
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm is not installed: %w", err)
	}

	// Check if package.json exists
	packageJSONPath := configPath + "/package.json"
	if _, err := os.Stat(packageJSONPath); os.IsNotExist(err) {
		return fmt.Errorf("package.json not found in %s", configPath)
	}

	// Create context with 60-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Run npm install — capture output so it doesn't clutter hefesto's output.
	// Only print on failure for debugging.
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "npm", "install")
	cmd.Dir = configPath
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("npm install timed out after 60 seconds")
		}
		// Print captured output so the user can debug the failure
		if stderr.Len() > 0 {
			fmt.Fprintln(os.Stderr, stderr.String())
		}
		return fmt.Errorf("npm install failed: %w", err)
	}

	logger.Debug("npm install completed: %s", stdout.String())
	return nil
}
