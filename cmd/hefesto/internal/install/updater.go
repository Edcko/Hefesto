// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	embedpkg "github.com/Edcko/Hefesto/cmd/hefesto/internal/embed"
)

// ChangeType represents the type of change detected for a file.
type ChangeType string

const (
	ChangeAdded     ChangeType = "added"     // File is new in embedded config
	ChangeModified  ChangeType = "modified"  // File exists but content differs
	ChangeUnchanged ChangeType = "unchanged" // File exists and content is same
	ChangeRemoved   ChangeType = "removed"   // File exists locally but not in embed (user-added)
)

// FileDiff represents the diff status of a single file.
type FileDiff struct {
	Path        string
	ChangeType  ChangeType
	OldSize     int64
	NewSize     int64
	OldHash     string
	NewHash     string
	IsDirectory bool
}

// DiffResult holds the complete diff analysis between embedded and installed config.
type DiffResult struct {
	Files         []FileDiff
	Summary       DiffSummary
	ConfigPath    string
	BackupPath    string
	SkillsAdded   int
	SkillsRemoved int
}

// DiffSummary holds aggregated counts of changes.
type DiffSummary struct {
	Added     int
	Modified  int
	Unchanged int
}

// UpdateProgress represents progress updates during update.
type UpdateProgress struct {
	Step    string
	Message string
	Done    bool
	Error   error
}

// Updater handles the Hefesto update process.
type Updater struct {
	configPath  string
	backupPath  string
	dryRun      bool
	skipConfirm bool
	Progress    chan UpdateProgress
}

// NewUpdater creates a new Updater instance.
func NewUpdater(dryRun, skipConfirm bool) *Updater {
	return &Updater{
		dryRun:      dryRun,
		skipConfirm: skipConfirm,
		Progress:    make(chan UpdateProgress, 20),
	}
}

// ComputeDiff analyzes differences between embedded and installed config.
func ComputeDiff(fsys embed.FS, configPath string) (*DiffResult, error) {
	result := &DiffResult{
		ConfigPath: configPath,
		Files:      []FileDiff{},
	}

	// Get the config subdirectory from the embedded filesystem
	configFS, err := fs.Sub(fsys, "config")
	if err != nil {
		return nil, fmt.Errorf("failed to get config subdirectory: %w", err)
	}

	// Track skills for special counting
	skillsInEmbed := make(map[string]bool)
	skillsInstalled := make(map[string]bool)

	// First, get list of installed skills
	installedSkillsPath := filepath.Join(configPath, "skills")
	if entries, err := os.ReadDir(installedSkillsPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				skillsInstalled[entry.Name()] = true
			}
		}
	}

	// Walk embedded files and compare with installed
	err = fs.WalkDir(configFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip node_modules
		if d.IsDir() && d.Name() == "node_modules" {
			return fs.SkipDir
		}

		installedPath := filepath.Join(configPath, path)

		// Track skills directories
		if d.IsDir() && strings.HasPrefix(path, "skills/") && path != "skills" {
			skillName := strings.TrimPrefix(path, "skills/")
			if !strings.Contains(skillName, "/") {
				skillsInEmbed[skillName] = true
			}
		}

		if d.IsDir() {
			// For directories, check if they exist
			info, err := os.Stat(installedPath)
			if os.IsNotExist(err) {
				result.Files = append(result.Files, FileDiff{
					Path:        path,
					ChangeType:  ChangeAdded,
					IsDirectory: true,
				})
				result.Summary.Added++
			} else if err == nil && info.IsDir() {
				result.Files = append(result.Files, FileDiff{
					Path:        path,
					ChangeType:  ChangeUnchanged,
					IsDirectory: true,
				})
				result.Summary.Unchanged++
			}
			return nil
		}

		// For files, compare content
		embeddedContent, err := fs.ReadFile(configFS, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}
		embeddedHash := hashContent(embeddedContent)
		embeddedSize := int64(len(embeddedContent))

		installedContent, err := os.ReadFile(installedPath)
		if os.IsNotExist(err) {
			// File doesn't exist - it's new
			diff := FileDiff{
				Path:       path,
				ChangeType: ChangeAdded,
				NewSize:    embeddedSize,
				NewHash:    embeddedHash,
			}
			result.Files = append(result.Files, diff)
			result.Summary.Added++
			return nil
		} else if err != nil {
			return fmt.Errorf("failed to read installed file %s: %w", path, err)
		}

		installedHash := hashContent(installedContent)
		installedSize := int64(len(installedContent))

		changeType := ChangeUnchanged
		if embeddedHash != installedHash {
			changeType = ChangeModified
			result.Summary.Modified++
		} else {
			result.Summary.Unchanged++
		}

		diff := FileDiff{
			Path:       path,
			ChangeType: changeType,
			OldSize:    installedSize,
			NewSize:    embeddedSize,
			OldHash:    installedHash,
			NewHash:    embeddedHash,
		}
		result.Files = append(result.Files, diff)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to compute diff: %w", err)
	}

	// Count skills changes
	for skill := range skillsInEmbed {
		if !skillsInstalled[skill] {
			result.SkillsAdded++
		}
	}
	for skill := range skillsInstalled {
		if !skillsInEmbed[skill] {
			result.SkillsRemoved++
		}
	}

	return result, nil
}

// hashContent returns a SHA256 hash of the content.
func hashContent(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes for shorter hash
}

// Run executes the full update process.
func (u *Updater) Run() error {
	defer close(u.Progress)

	// Step 1: Check if Hefesto is installed
	u.Progress <- UpdateProgress{
		Step:    "check",
		Message: "Checking installation status...",
		Done:    false,
	}

	env, err := Detect()
	if err != nil {
		u.Progress <- UpdateProgress{
			Step:    "check",
			Message: fmt.Sprintf("Detection failed: %v", err),
			Done:    true,
			Error:   err,
		}
		return fmt.Errorf("detection failed: %w", err)
	}

	if !env.ConfigExists {
		u.Progress <- UpdateProgress{
			Step:    "check",
			Message: "Hefesto is not installed",
			Done:    true,
			Error:   fmt.Errorf("not installed"),
		}
		return fmt.Errorf("hefesto is not installed. Run `hefesto install` first")
	}

	u.configPath = env.ConfigPath

	u.Progress <- UpdateProgress{
		Step:    "check",
		Message: fmt.Sprintf("Found %s at %s", env.ExistingConfig, formatPath(u.configPath)),
		Done:    true,
	}

	// Step 2: Compute diff
	u.Progress <- UpdateProgress{
		Step:    "diff",
		Message: "Analyzing changes...",
		Done:    false,
	}

	diff, err := ComputeDiff(embedpkg.ConfigFiles, u.configPath)
	if err != nil {
		u.Progress <- UpdateProgress{
			Step:    "diff",
			Message: fmt.Sprintf("Diff analysis failed: %v", err),
			Done:    true,
			Error:   err,
		}
		return fmt.Errorf("failed to compute diff: %w", err)
	}

	u.Progress <- UpdateProgress{
		Step:    "diff",
		Message: fmt.Sprintf("Found %d added, %d modified, %d unchanged", diff.Summary.Added, diff.Summary.Modified, diff.Summary.Unchanged),
		Done:    true,
	}

	// Step 3: Create backup (skip in dry-run)
	if !u.dryRun {
		u.Progress <- UpdateProgress{
			Step:    "backup",
			Message: "Creating backup...",
			Done:    false,
		}

		backupPath, err := Backup(u.configPath)
		if err != nil {
			u.Progress <- UpdateProgress{
				Step:    "backup",
				Message: fmt.Sprintf("Backup failed: %v", err),
				Done:    true,
				Error:   err,
			}
			return fmt.Errorf("backup failed: %w", err)
		}
		u.backupPath = backupPath
		diff.BackupPath = backupPath

		u.Progress <- UpdateProgress{
			Step:    "backup",
			Message: fmt.Sprintf("Backup created: %s", formatBackupPath(backupPath)),
			Done:    true,
		}
	} else {
		u.Progress <- UpdateProgress{
			Step:    "backup",
			Message: "Backup skipped (dry run)",
			Done:    true,
		}
	}

	// Step 4: Copy embedded config (overlay, not delete)
	u.Progress <- UpdateProgress{
		Step:    "update",
		Message: "Updating configuration...",
		Done:    false,
	}

	if !u.dryRun {
		if err := CopyConfig(embedpkg.ConfigFiles, u.configPath); err != nil {
			u.Progress <- UpdateProgress{
				Step:    "update",
				Message: fmt.Sprintf("Update failed: %v", err),
				Done:    true,
				Error:   err,
			}
			return fmt.Errorf("failed to update config: %w", err)
		}
	}

	u.Progress <- UpdateProgress{
		Step:    "update",
		Message: "Configuration updated",
		Done:    true,
	}

	// Step 5: Run npm install
	u.Progress <- UpdateProgress{
		Step:    "npm",
		Message: "Updating npm dependencies...",
		Done:    false,
	}

	if !u.dryRun {
		if err := NpmInstall(u.configPath); err != nil {
			// npm install failure is not fatal
			u.Progress <- UpdateProgress{
				Step:    "npm",
				Message: fmt.Sprintf("npm install skipped (non-fatal): %v", err),
				Done:    true,
			}
		} else {
			u.Progress <- UpdateProgress{
				Step:    "npm",
				Message: "Dependencies updated",
				Done:    true,
			}
		}
	} else {
		u.Progress <- UpdateProgress{
			Step:    "npm",
			Message: "npm install skipped (dry run)",
			Done:    true,
		}
	}

	// Step 6: Verify
	u.Progress <- UpdateProgress{
		Step:    "verify",
		Message: "Verifying update...",
		Done:    false,
	}

	if !u.dryRun {
		result, err := Verify(u.configPath)
		if err != nil {
			u.Progress <- UpdateProgress{
				Step:    "verify",
				Message: fmt.Sprintf("Verification error: %v", err),
				Done:    true,
				Error:   err,
			}
			return fmt.Errorf("verification failed: %w", err)
		}

		if !result.ConfigCopied {
			err := fmt.Errorf("config verification failed")
			u.Progress <- UpdateProgress{
				Step:    "verify",
				Message: "Config verification failed",
				Done:    true,
				Error:   err,
			}
			return err
		}

		u.Progress <- UpdateProgress{
			Step:    "verify",
			Message: "Update verified successfully",
			Done:    true,
		}
	} else {
		u.Progress <- UpdateProgress{
			Step:    "verify",
			Message: "Verification skipped (dry run)",
			Done:    true,
		}
	}

	return nil
}

// formatBackupPath shortens the backup path for display.
func formatBackupPath(path string) string {
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" && strings.HasPrefix(path, homeDir) {
		return "~" + strings.TrimPrefix(path, homeDir)
	}
	return path
}

// PrintUpdateHeader prints the update header.
func PrintUpdateHeader(version string, dryRun bool) {
	if dryRun {
		fmt.Println()
		fmt.Println("🔥 Hefesto Update (dry-run)")
	} else {
		fmt.Println()
		fmt.Println("🔥 Hefesto Update")
	}
	fmt.Println()
}

// PrintUpdateStatus prints the current vs target version info.
func PrintUpdateStatus(currentConfig string, embeddedVersion string) {
	fmt.Printf("  Current: %s (installed)\n", currentConfig)
	fmt.Printf("  Updating to: %s (embedded)\n", embeddedVersion)
	fmt.Println()
}

// PrintDiffSummary prints the diff analysis in a formatted way.
func PrintDiffSummary(diff *DiffResult, dryRun bool) {
	if dryRun {
		fmt.Println("  Would update:")
	} else {
		fmt.Println("  Changes:")
	}

	// Group files by top-level directory
	fileGroups := make(map[string][]FileDiff)
	for _, f := range diff.Files {
		if f.IsDirectory {
			continue // Skip directories in output
		}
		topDir := getTopLevelPath(f.Path)
		fileGroups[topDir] = append(fileGroups[topDir], f)
	}

	// Define order of display
	order := []string{"AGENTS.md", "opencode.json", "skills/", "plugins/", "commands/", "personality", "themes/"}

	for _, key := range order {
		files, exists := fileGroups[key]
		if !exists {
			continue
		}

		// Special handling for skills
		if key == "skills/" {
			printSkillsSummary(diff)
			continue
		}

		// For single files
		if len(files) == 1 {
			f := files[0]
			printFileChange(f)
		} else {
			// Multiple files in a directory - show summary
			modified := 0
			added := 0
			for _, f := range files {
				switch f.ChangeType {
				case ChangeModified:
					modified++
				case ChangeAdded:
					added++
				}
			}

			if added > 0 || modified > 0 {
				status := ""
				if added > 0 && modified > 0 {
					status = fmt.Sprintf("(%d new, %d modified)", added, modified)
				} else if added > 0 {
					status = fmt.Sprintf("(%d new)", added)
				} else {
					status = fmt.Sprintf("(%d modified)", modified)
				}
				fmt.Printf("    ✏️  %-14s %s\n", key, status)
			} else {
				fmt.Printf("    ➡️  %-14s (unchanged)\n", key)
			}
		}
	}

	fmt.Println()
}

// getTopLevelPath extracts the top-level path component.
func getTopLevelPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		// For directories like skills/, plugins/, etc.
		if parts[0] == "skills" || parts[0] == "plugins" || parts[0] == "commands" || parts[0] == "themes" {
			return parts[0] + "/"
		}
	}
	return parts[0]
}

// printSkillsSummary prints a summary of skills changes.
func printSkillsSummary(diff *DiffResult) {
	added := diff.SkillsAdded
	removed := diff.SkillsRemoved

	if added > 0 || removed > 0 {
		if added > 0 && removed > 0 {
			fmt.Printf("    ✏️  skills/         (%d added, %d removed)\n", added, removed)
		} else if added > 0 {
			fmt.Printf("    ✏️  skills/         (%d new)\n", added)
		} else {
			fmt.Printf("    ✏️  skills/         (%d removed)\n", removed)
		}
	} else if diff.Summary.Modified > 0 {
		fmt.Printf("    ✏️  skills/         (modified)\n")
	} else {
		fmt.Printf("    ➡️  skills/         (unchanged)\n")
	}
}

// printFileChange prints a single file change.
func printFileChange(f FileDiff) {
	switch f.ChangeType {
	case ChangeAdded:
		fmt.Printf("    ✏️  %-14s (new)\n", f.Path)
	case ChangeModified:
		oldSize := formatBytes(f.OldSize)
		newSize := formatBytes(f.NewSize)
		fmt.Printf("    ✏️  %-14s (modified — %s → %s)\n", f.Path, oldSize, newSize)
	case ChangeUnchanged:
		fmt.Printf("    ➡️  %-14s (unchanged)\n", f.Path)
	}
}

// PrintUpdateResult prints the final update result.
func PrintUpdateResult(diff *DiffResult, backupPath string, dryRun bool) {
	if dryRun {
		fmt.Println("  No changes made. Run without --dry-run to update.")
		fmt.Println()
		return
	}

	changes := diff.Summary.Added + diff.Summary.Modified
	fmt.Printf("  Updated %d files. Backup saved. 🛠️\n", changes)
	fmt.Println()
}

// RunUpdate is a convenience function that creates and runs an updater.
func RunUpdate(dryRun, skipConfirm bool) error {
	// First, check status and compute diff
	status, err := CheckStatus()
	if err != nil {
		return fmt.Errorf("failed to check status: %w", err)
	}

	if !status.Installed {
		fmt.Println()
		fmt.Println("❌ Hefesto is not installed. Run `hefesto install` first.")
		fmt.Println()
		return nil
	}

	// Compute diff for preview
	diff, err := ComputeDiff(embedpkg.ConfigFiles, status.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to compute diff: %w", err)
	}

	// Print header
	PrintUpdateHeader("v0.1.0", dryRun)

	// Print current status
	PrintUpdateStatus(status.Version, "v0.1.0")

	// Print diff summary
	PrintDiffSummary(diff, dryRun)

	// In dry-run mode, stop here
	if dryRun {
		PrintUpdateResult(diff, "", true)
		return nil
	}

	// Check if there are any changes
	changes := diff.Summary.Added + diff.Summary.Modified
	if changes == 0 {
		fmt.Println("  Already up to date. No changes needed.")
		fmt.Println()
		return nil
	}

	// Ask for confirmation unless skipped
	if !skipConfirm {
		fmt.Printf("  Continue with update? [y/N] ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println()
			fmt.Println("  Update cancelled.")
			fmt.Println()
			return nil
		}
		fmt.Println()
	}

	// Run the updater
	updater := NewUpdater(false, true)

	// Start update in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- updater.Run()
	}()

	// Print progress updates
	for progress := range updater.Progress {
		if progress.Error != nil {
			// Only print errors for non-"not installed" cases
			if progress.Step != "check" || progress.Error.Error() != "not installed" {
				fmt.Printf("❌ [%s] %s\n", progress.Step, progress.Message)
			}
		} else if progress.Done {
			fmt.Printf("✅ [%s] %s\n", progress.Step, progress.Message)
		} else {
			fmt.Printf("⏳ [%s] %s\n", progress.Step, progress.Message)
		}
	}

	// Wait for completion
	if err := <-errChan; err != nil {
		fmt.Println()
		fmt.Printf("❌ Update failed: %v\n", err)
		return err
	}

	// Print final result
	PrintUpdateResult(diff, updater.backupPath, false)

	return nil
}

// GetCurrentBackupPath returns the path to the backup created during update.
func (u *Updater) GetCurrentBackupPath() string {
	return u.backupPath
}

// CompareFiles compares two files and returns true if they differ.
func CompareFiles(path1, path2 string) (bool, error) {
	content1, err := os.ReadFile(path1)
	if err != nil {
		return false, err
	}

	content2, err := os.ReadFile(path2)
	if err != nil {
		return false, err
	}

	return !bytes.Equal(content1, content2), nil
}
