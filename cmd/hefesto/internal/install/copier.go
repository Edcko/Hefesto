// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CopyConfig copies all files from the embedded filesystem to the target path.
// The embedded files are expected to be under a "config/" prefix.
func CopyConfig(fsys embed.FS, targetPath string) error {
	// Get the config subdirectory from the embedded filesystem
	configFS, err := fs.Sub(fsys, "config")
	if err != nil {
		return fmt.Errorf("failed to get config subdirectory from embedded FS: %w", err)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetPath, 0750); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Walk through the embedded filesystem and copy all files
	err = fs.WalkDir(configFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip node_modules directory if it exists
		if d.IsDir() && d.Name() == "node_modules" {
			return fs.SkipDir
		}

		targetFilePath := filepath.Join(targetPath, path)

		if d.IsDir() {
			// Create directory with 0750 permissions
			if err := os.MkdirAll(targetFilePath, 0750); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetFilePath, err)
			}
			return nil
		}

		// Copy file
		return copyEmbeddedFile(configFS, path, targetFilePath)
	})

	if err != nil {
		return fmt.Errorf("failed to copy embedded config: %w", err)
	}

	return nil
}

// copyEmbeddedFile copies a single file from the embedded filesystem to the target path.
func copyEmbeddedFile(fsys fs.FS, srcPath, targetPath string) error {
	// Read file content from embedded FS
	content, err := fs.ReadFile(fsys, srcPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded file %s: %w", srcPath, err)
	}

	// Create parent directories if needed
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return fmt.Errorf("failed to create parent directory %s: %w", targetDir, err)
	}

	// Determine file permissions based on file type
	perm := fs.FileMode(0644)
	// Shell scripts and executable files get execute permission
	if strings.HasSuffix(srcPath, ".sh") {
		perm = 0755
	}

	// Write file to target location
	if err := os.WriteFile(targetPath, content, perm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", targetPath, err)
	}

	return nil
}
