// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Backup creates a timestamped backup of the opencode config directory.
// Returns the path to the backup directory or an error.
func Backup(configPath string) (string, error) {
	// Check if config directory exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("config directory does not exist: %s", configPath)
	}

	// Get home directory for backup location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create timestamped backup directory name
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(homeDir, fmt.Sprintf(".config/opencode-backup-%s", timestamp))

	// Create backup directory parent if needed
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy the entire config directory
	err = copyDirectory(configPath, backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to copy config to backup: %w", err)
	}

	return backupPath, nil
}

// copyDirectory recursively copies a directory and all its contents.
func copyDirectory(src, dst string) error {
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
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Get source file info for permissions
	info, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	destFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
