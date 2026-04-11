// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Edcko/Hefesto/cmd/hefesto/internal/logger"
)

// BackupInfo holds information about a backup directory.
type BackupInfo struct {
	Path      string
	Name      string
	Timestamp time.Time
}

// ListBackups finds all backup directories in ~/.config/
func ListBackups() ([]BackupInfo, error) {
	homeDir, err := getUserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config")
	return ListBackupsInDir(configDir)
}

// ListBackupsInDir finds all backup directories in the specified directory.
// This is a testable version of ListBackups.
func ListBackupsInDir(dir string) ([]BackupInfo, error) {
	pattern := filepath.Join(dir, "opencode-backup-*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search for backups: %w", err)
	}

	var backups []BackupInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil || !info.IsDir() {
			continue
		}

		name := filepath.Base(match)
		timestamp, err := parseBackupTimestamp(name)
		if err != nil {
			continue // Skip invalid backup names
		}

		backups = append(backups, BackupInfo{
			Path:      match,
			Name:      name,
			Timestamp: timestamp,
		})
	}

	// Sort by timestamp, newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// FindMostRecentBackup returns the most recent backup from the list.
// Returns an error if no backups are available.
func FindMostRecentBackup(backups []BackupInfo) (*BackupInfo, error) {
	if len(backups) == 0 {
		return nil, fmt.Errorf("no backups available")
	}

	// Backups are already sorted newest-first by ListBackups
	return &backups[0], nil
}

// ParseBackupTimestamp extracts the timestamp from a backup directory name.
// Expected format: opencode-backup-YYYYMMDD-HHMMSS
func ParseBackupTimestamp(name string) (time.Time, error) {
	prefix := "opencode-backup-"
	if !strings.HasPrefix(name, prefix) {
		return time.Time{}, fmt.Errorf("invalid backup name format: %s", name)
	}

	timestampStr := strings.TrimPrefix(name, prefix)
	t, err := time.Parse("20060102-150405", timestampStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return t, nil
}

// parseBackupTimestamp is an alias for ParseBackupTimestamp for backward compatibility.
var parseBackupTimestamp = ParseBackupTimestamp

// Rollback restores a backup to the opencode config directory.
// It creates a safety backup of the current config before restoring.
func Rollback(backupPath string) (string, error) {
	homeDir, err := getUserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "opencode")

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return "", fmt.Errorf("backup does not exist: %s", backupPath)
	}

	// Create safety backup of current config (if it exists)
	var safetyBackupPath string
	if _, err := os.Stat(configPath); err == nil {
		safetyBackupPath, err = Backup(configPath)
		if err != nil {
			return "", fmt.Errorf("failed to create safety backup: %w", err)
		}
	}

	// Remove current config
	if err := os.RemoveAll(configPath); err != nil {
		return "", fmt.Errorf("failed to remove current config: %w", err)
	}

	// Copy backup to config location
	if err := CopyDirectory(backupPath, configPath); err != nil {
		return "", fmt.Errorf("failed to restore backup: %w", err)
	}

	logger.Debug("rollback: restored backup from %s to %s, safety=%s", backupPath, configPath, safetyBackupPath)
	return safetyBackupPath, nil
}

// PrintBackups prints the list of available backups.
func PrintBackups(backups []BackupInfo) {
	fmt.Println()
	fmt.Println("🔥 Hefesto Backups")
	fmt.Println()

	if len(backups) == 0 {
		homeDir, _ := getUserHomeDir()
		configDir := filepath.Join(homeDir, ".config")
		tildePath := "~/.config/"
		if homeDir != "" {
			tildePath = strings.Replace(configDir, homeDir, "~", 1) + "/"
		}

		fmt.Println("  ❌ No backups found in " + tildePath)
		fmt.Println()
		fmt.Println("  Backups are created automatically during install and update.")
		fmt.Println()
		return
	}

	fmt.Println("  Available backups:")
	fmt.Println()
	for i, backup := range backups {
		dateStr := FormatBackupDate(backup.Timestamp)
		fmt.Printf("    #%d  %-35s  (%s)\n", i+1, backup.Name, dateStr)
	}
	fmt.Println()
	fmt.Println("  Run `hefesto rollback` to restore the most recent backup.")
	fmt.Println("  Run `hefesto rollback --yes` to skip confirmation.")
	fmt.Println()
}

// FormatBackupDate formats a timestamp for display.
func FormatBackupDate(t time.Time) string {
	return fmt.Sprintf("%s — %s", t.Format("Jan 2, 2006"), t.Format("15:04"))
}

// PrintRollbackResult prints the result of a rollback operation.
func PrintRollbackResult(backup BackupInfo, safetyBackupPath string) {
	fmt.Println()
	fmt.Println("🔥 Hefesto Rollback")
	fmt.Println()
	fmt.Printf("  Found backup: %s\n", backup.Name)
	fmt.Println()

	if safetyBackupPath != "" {
		safetyName := filepath.Base(safetyBackupPath)
		fmt.Println("  ⏳ Creating safety backup of current config...")
		fmt.Printf("  ✅ Safety backup: %s\n", safetyName)
		fmt.Println()
	}

	fmt.Println("  ⏳ Restoring backup...")
	fmt.Println("  ✅ Backup restored successfully")
	fmt.Println()

	dateStr := FormatBackupDate(backup.Timestamp)
	fmt.Printf("  Restored config from %s. 🛠️\n", dateStr)
	fmt.Println()
}

// PromptRollback prompts the user to select a backup to restore.
// Returns the selected backup or nil if cancelled.
func PromptRollback(backups []BackupInfo) *BackupInfo {
	if len(backups) == 0 {
		return nil
	}

	// Default to most recent backup
	return &backups[0]
}
