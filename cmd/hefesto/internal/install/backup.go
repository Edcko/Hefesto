// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Edcko/Hefesto/cmd/hefesto/internal/logger"
)

// maxBackups is the maximum number of backup directories to keep.
// Older backups beyond this limit are removed after install, update, or rollback.
const maxBackups = 5

// Backup creates a timestamped backup of the opencode config directory.
// Returns the path to the backup directory or an error.
func Backup(configPath string) (string, error) {
	// Check if config directory exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("config directory does not exist: %s", configPath)
	}

	// Get home directory for backup location
	homeDir, err := getUserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create timestamped backup directory name
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(homeDir, fmt.Sprintf(".config/opencode-backup-%s", timestamp))

	// Create backup directory parent if needed
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy the entire config directory
	err = CopyDirectory(configPath, backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to copy config to backup: %w", err)
	}

	logger.Debug("backup: created backup at %s from %s", backupPath, configPath)
	return backupPath, nil
}

// CopyDirectory recursively copies a directory and all its contents.
// It skips node_modules directories.
func CopyDirectory(src, dst string) error {
	// Get source directory info
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory with same permissions
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Skip node_modules directory
			if entry.Name() == "node_modules" {
				continue
			}
			// Recursively copy subdirectory
			if err := CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile copies a single file from src to dst.
func CopyFile(src, dst string) error {
	// Open source file
	sourceFile, err := os.Open(src) //nolint:gosec // G304: src is an internal path, not user input
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	// Get source file info for permissions
	info, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	destFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode()) //nolint:gosec // G304: dst is an internal path
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	// Copy contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

// CopyPath copies a file or directory from src to dst.
// If src is a directory, it copies recursively (skipping node_modules).
// If src is a file, it copies the file preserving permissions.
func CopyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return CopyDirectory(src, dst)
	}
	return CopyFile(src, dst)
}

// CleanOldBackups removes old backup directories, keeping only the N most
// recent ones (where N = maxBackups). It lists all opencode-backup-* dirs in
// ~/.config/, sorts them by timestamp (newest first), and removes any beyond
// the retention limit.
//
// This function is safe: if there are maxBackups or fewer backups, nothing is
// deleted. Errors during individual removal are logged but do not stop the
// cleanup of remaining backups.
func CleanOldBackups() error {
	backups, err := ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups for cleanup: %w", err)
	}

	if len(backups) <= maxBackups {
		logger.Debug("backup-cleanup: %d backups found, within limit of %d — nothing to remove", len(backups), maxBackups)
		return nil
	}

	// backups are sorted newest-first by ListBackups, so remove from index maxBackups onward
	toRemove := backups[maxBackups:]
	var removed int
	for _, b := range toRemove {
		if err := os.RemoveAll(b.Path); err != nil {
			logger.Debug("backup-cleanup: failed to remove %s: %v", b.Name, err)
			continue
		}
		logger.Debug("backup-cleanup: removed old backup %s", b.Name)
		removed++
	}

	logger.Debug("backup-cleanup: removed %d/%d old backups (kept %d most recent)", removed, len(toRemove), maxBackups)
	return nil
}

// CleanOldBackupsInDir is the testable version of CleanOldBackups that
// operates on an arbitrary directory instead of ~/.config/.
func CleanOldBackupsInDir(dir string) error {
	backups, err := ListBackupsInDir(dir)
	if err != nil {
		return fmt.Errorf("failed to list backups for cleanup: %w", err)
	}

	if len(backups) <= maxBackups {
		return nil
	}

	toRemove := backups[maxBackups:]
	for _, b := range toRemove {
		if err := os.RemoveAll(b.Path); err != nil {
			logger.Debug("backup-cleanup: failed to remove %s: %v", b.Name, err)
			continue
		}
		logger.Debug("backup-cleanup: removed old backup %s", b.Name)
	}

	return nil
}
