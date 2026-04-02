// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
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

	// Run npm install
	cmd := exec.CommandContext(ctx, "npm", "install")
	cmd.Dir = configPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("npm install timed out after 60 seconds")
		}
		return fmt.Errorf("npm install failed: %w", err)
	}

	return nil
}
