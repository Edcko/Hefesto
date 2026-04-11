// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UninstallProgress represents progress updates during uninstallation.
type UninstallProgress struct {
	Step    string
	Message string
	Done    bool
	Error   error
}

// Uninstaller orchestrates the Hefesto uninstallation process.
type Uninstaller struct {
	configPath  string
	homeDir     string
	purge       bool
	skipConfirm bool
	Progress    chan UninstallProgress
}

// NewUninstaller creates a new Uninstaller instance.
func NewUninstaller(purge, skipConfirm bool) *Uninstaller {
	return &Uninstaller{
		purge:       purge,
		skipConfirm: skipConfirm,
		Progress:    make(chan UninstallProgress, 10),
	}
}

// Run executes the full uninstallation process.
func (u *Uninstaller) Run() error {
	defer close(u.Progress)

	// Get home directory
	homeDir, err := getUserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	u.homeDir = homeDir
	u.configPath = filepath.Join(homeDir, ".config", "opencode")

	// Step 1: Detect installation
	u.Progress <- UninstallProgress{
		Step:    "detect",
		Message: "Detecting installation...",
		Done:    false,
	}

	env, err := Detect()
	if err != nil {
		u.Progress <- UninstallProgress{
			Step:    "detect",
			Message: fmt.Sprintf("Detection failed: %v", err),
			Done:    true,
			Error:   err,
		}
		return fmt.Errorf("detection failed: %w", err)
	}

	if !env.ConfigExists {
		u.Progress <- UninstallProgress{
			Step:    "detect",
			Message: "No Hefesto configuration found",
			Done:    true,
		}
		return fmt.Errorf("no Hefesto configuration found at %s", formatPath(u.configPath))
	}

	u.Progress <- UninstallProgress{
		Step:    "detect",
		Message: fmt.Sprintf("Found Hefesto config at %s", formatPath(u.configPath)),
		Done:    true,
	}

	// Step 2: Find backups using existing function (skip in purge mode)
	var backups []BackupInfo
	if !u.purge {
		backups, err = ListBackups()
		if err != nil {
			// Non-fatal - just warn
			u.Progress <- UninstallProgress{
				Step:    "backup",
				Message: fmt.Sprintf("Warning: Could not scan for backups: %v", err),
				Done:    true,
			}
		}
	}

	// Print summary and get confirmation
	if u.purge {
		u.printPurgeSummary()
	} else {
		u.printRestoreSummary(backups)
	}

	if !u.skipConfirm {
		if !u.confirmUninstall() {
			fmt.Println()
			fmt.Println("  Uninstall cancelled.")
			fmt.Println()
			return nil
		}
	}

	// Step 3: Execute uninstall
	if u.purge {
		return u.executePurge()
	}
	return u.executeRestore(backups)
}

// printRestoreSummary prints the summary for restore mode.
func (u *Uninstaller) printRestoreSummary(backups []BackupInfo) {
	fmt.Println()
	fmt.Println("🔥 Hefesto Uninstall")
	fmt.Println()
	fmt.Printf("  Config found at: %s\n", formatPath(u.configPath))
	fmt.Println()

	if len(backups) == 0 {
		fmt.Println("  ⚠️  No backups found - configuration will be removed without restore")
		fmt.Println()
		fmt.Println("  Action: Remove Hefesto configuration (no backup available)")
	} else {
		fmt.Println("  Available backups:")
		for i, backup := range backups {
			relativeTime := formatRelativeTime(backup.Timestamp)
			fmt.Printf("    %d. %s (%s)\n", i+1, backup.Name, relativeTime)
		}
		fmt.Println()
		fmt.Printf("  Action: Restore backup #1 (%s) and remove Hefesto config\n", backups[0].Name)
	}
	fmt.Println()
}

// printPurgeSummary prints the summary for purge mode.
func (u *Uninstaller) printPurgeSummary() {
	fmt.Println()
	fmt.Println("🔥 Hefesto Uninstall")
	fmt.Println()
	fmt.Println("  ⚠️  PURGE MODE — This will delete ALL OpenCode configuration")
	fmt.Println()

	// Count what will be removed
	skillsDir := filepath.Join(u.configPath, "skills")
	pluginsDir := filepath.Join(u.configPath, "plugins")

	skillsCount := countDirectories(skillsDir)
	pluginsCount := countFiles(pluginsDir)

	fmt.Printf("  Removing:\n")
	fmt.Printf("    %s", formatPath(u.configPath))
	details := []string{}
	if skillsCount > 0 {
		details = append(details, fmt.Sprintf("%d skills", skillsCount))
	}
	if pluginsCount > 0 {
		details = append(details, fmt.Sprintf("%d plugins", pluginsCount))
	}
	// Check for AGENTS.md
	if _, err := os.Stat(filepath.Join(u.configPath, "AGENTS.md")); err == nil {
		details = append(details, "AGENTS.md")
	}
	// Check for opencode.json
	if _, err := os.Stat(filepath.Join(u.configPath, "opencode.json")); err == nil {
		details = append(details, "opencode.json")
	}
	if len(details) > 0 {
		fmt.Printf(" (%s)\n", strings.Join(details, ", "))
	} else {
		fmt.Println()
	}
	fmt.Println()
}

// confirmUninstall asks for user confirmation.
func (u *Uninstaller) confirmUninstall() bool {
	fmt.Print("  Continue? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// executeRestore removes Hefesto config and restores the most recent backup.
func (u *Uninstaller) executeRestore(backups []BackupInfo) error {
	// If no backups, just remove
	if len(backups) == 0 {
		return u.removeConfig()
	}

	// Restore the most recent backup
	backup := backups[0]

	u.Progress <- UninstallProgress{
		Step:    "restore",
		Message: fmt.Sprintf("Restoring backup from %s...", backup.Name),
		Done:    false,
	}

	// Remove current config first
	if err := os.RemoveAll(u.configPath); err != nil {
		u.Progress <- UninstallProgress{
			Step:    "restore",
			Message: fmt.Sprintf("Failed to remove current config: %v", err),
			Done:    true,
			Error:   err,
		}
		return fmt.Errorf("failed to remove current config: %w", err)
	}

	// Copy backup to config location
	if err := CopyDirectory(backup.Path, u.configPath); err != nil {
		u.Progress <- UninstallProgress{
			Step:    "restore",
			Message: fmt.Sprintf("Failed to restore backup: %v", err),
			Done:    true,
			Error:   err,
		}
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	u.Progress <- UninstallProgress{
		Step:    "restore",
		Message: "Backup restored successfully",
		Done:    true,
	}

	// Remove the backup directory after successful restore
	if err := os.RemoveAll(backup.Path); err != nil {
		// Non-fatal - just warn
		u.Progress <- UninstallProgress{
			Step:    "cleanup",
			Message: fmt.Sprintf("Warning: Could not remove backup directory: %v", err),
			Done:    true,
		}
	}

	u.printSuccessMessage(false)
	return nil
}

// executePurge removes everything without restoring backup.
func (u *Uninstaller) executePurge() error {
	u.Progress <- UninstallProgress{
		Step:    "remove",
		Message: "Removing configuration...",
		Done:    false,
	}

	if err := os.RemoveAll(u.configPath); err != nil {
		u.Progress <- UninstallProgress{
			Step:    "remove",
			Message: fmt.Sprintf("Failed to remove configuration: %v", err),
			Done:    true,
			Error:   err,
		}
		return fmt.Errorf("failed to remove configuration: %w", err)
	}

	u.Progress <- UninstallProgress{
		Step:    "remove",
		Message: "All configuration removed",
		Done:    true,
	}

	u.printSuccessMessage(true)
	return nil
}

// removeConfig removes the config without restore.
func (u *Uninstaller) removeConfig() error {
	u.Progress <- UninstallProgress{
		Step:    "remove",
		Message: "Removing configuration...",
		Done:    false,
	}

	if err := os.RemoveAll(u.configPath); err != nil {
		u.Progress <- UninstallProgress{
			Step:    "remove",
			Message: fmt.Sprintf("Failed to remove configuration: %v", err),
			Done:    true,
			Error:   err,
		}
		return fmt.Errorf("failed to remove configuration: %w", err)
	}

	u.Progress <- UninstallProgress{
		Step:    "remove",
		Message: "Configuration removed",
		Done:    true,
	}

	fmt.Println()
	fmt.Println("  ✅ Hefesto configuration removed")
	fmt.Println()
	fmt.Println("  No backup was available. Goodbye! 👋")
	fmt.Println()
	return nil
}

// printSuccessMessage prints the final success message.
func (u *Uninstaller) printSuccessMessage(purge bool) {
	fmt.Println()
	if purge {
		fmt.Println("  ✅ All configuration removed")
		fmt.Println()
		fmt.Println("  No backup was restored. Goodbye! 👋")
	} else {
		fmt.Println("  ✅ Hefesto configuration removed")
		fmt.Println()
		fmt.Println("  Your previous OpenCode config has been restored. Goodbye! 👋")
	}
	fmt.Println()
}

// formatRelativeTime returns a human-readable relative time.
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	if days < 7 {
		return fmt.Sprintf("%d days ago", days)
	}
	weeks := days / 7
	if weeks == 1 {
		return "1 week ago"
	}
	return fmt.Sprintf("%d weeks ago", weeks)
}

// countDirectories counts the number of directories in a path.
func countDirectories(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			count++
		}
	}
	return count
}

// countFiles counts the number of files in a path.
func countFiles(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}
	return count
}

// RunUninstall is a convenience function that creates and runs an uninstaller.
func RunUninstall(purge, skipConfirm bool) error {
	uninstaller := NewUninstaller(purge, skipConfirm)

	// Run uninstallation in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- uninstaller.Run()
	}()

	// Print progress updates
	for progress := range uninstaller.Progress {
		if progress.Error != nil {
			fmt.Printf("❌ [%s] %s\n", progress.Step, progress.Message)
		} else if progress.Done {
			fmt.Printf("✅ [%s] %s\n", progress.Step, progress.Message)
		} else {
			fmt.Printf("⏳ [%s] %s\n", progress.Step, progress.Message)
		}
	}

	// Wait for completion
	return <-errChan
}
