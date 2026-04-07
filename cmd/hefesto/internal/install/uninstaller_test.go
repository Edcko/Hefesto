// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestListBackups(t *testing.T) {
	t.Run("no backups returns empty", func(t *testing.T) {
		// Create temp home directory with no backups
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)

		// Create .config directory
		configDir := filepath.Join(tmpHome, ".config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		backups, err := ListBackups()
		if err != nil {
			t.Fatalf("ListBackups failed: %v", err)
		}

		if len(backups) != 0 {
			t.Errorf("Expected 0 backups, got %d", len(backups))
		}
	})

	t.Run("multiple backups sorted newest first", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)

		configDir := filepath.Join(tmpHome, ".config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		// Create multiple backups with different timestamps
		backupNames := []string{
			"opencode-backup-20240101-100000",
			"opencode-backup-20240102-120000",
			"opencode-backup-20240103-140000",
		}

		for _, name := range backupNames {
			backupPath := filepath.Join(configDir, name)
			if err := os.MkdirAll(backupPath, 0755); err != nil {
				t.Fatalf("Failed to create backup %s: %v", name, err)
			}
			// Small delay to ensure different timestamps
			time.Sleep(10 * time.Millisecond)
		}

		backups, err := ListBackups()
		if err != nil {
			t.Fatalf("ListBackups failed: %v", err)
		}

		if len(backups) != 3 {
			t.Errorf("Expected 3 backups, got %d", len(backups))
		}

		// Verify newest first
		for i := 0; i < len(backups)-1; i++ {
			if !backups[i].Timestamp.After(backups[i+1].Timestamp) {
				t.Errorf("Backups not sorted newest first: %s should be after %s",
					backups[i].Name, backups[i+1].Name)
			}
		}
	})

	t.Run("non-matching dirs ignored", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)

		configDir := filepath.Join(tmpHome, ".config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		// Create valid backup
		validBackup := filepath.Join(configDir, "opencode-backup-20240101-100000")
		if err := os.MkdirAll(validBackup, 0755); err != nil {
			t.Fatalf("Failed to create valid backup: %v", err)
		}

		// Create non-matching directories
		invalidDirs := []string{
			"opencode-backup-invalid",
			"other-dir",
			"backup-20240101-100000",
			"opencode-backup-",
		}

		for _, name := range invalidDirs {
			dirPath := filepath.Join(configDir, name)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				t.Fatalf("Failed to create dir %s: %v", name, err)
			}
		}

		backups, err := ListBackups()
		if err != nil {
			t.Fatalf("ListBackups failed: %v", err)
		}

		if len(backups) != 1 {
			t.Errorf("Expected 1 valid backup, got %d", len(backups))
		}

		if len(backups) > 0 && backups[0].Name != "opencode-backup-20240101-100000" {
			t.Errorf("Expected valid backup name, got %s", backups[0].Name)
		}
	})
}

func TestNewUninstaller(t *testing.T) {
	t.Run("creates uninstaller with correct flags", func(t *testing.T) {
		uninstaller := NewUninstaller(true, true)

		if uninstaller == nil {
			t.Fatal("Uninstaller should not be nil")
		}

		if !uninstaller.purge {
			t.Error("purge should be true")
		}

		if !uninstaller.skipConfirm {
			t.Error("skipConfirm should be true")
		}

		if uninstaller.Progress == nil {
			t.Error("Progress channel should be initialized")
		}
	})

	t.Run("creates uninstaller with false flags", func(t *testing.T) {
		uninstaller := NewUninstaller(false, false)

		if uninstaller.purge {
			t.Error("purge should be false")
		}

		if uninstaller.skipConfirm {
			t.Error("skipConfirm should be false")
		}
	})

	t.Run("progress channel is buffered", func(t *testing.T) {
		uninstaller := NewUninstaller(false, false)

		// Channel should have capacity of 10
		select {
		case uninstaller.Progress <- UninstallProgress{Step: "test"}:
			// Successfully sent to channel
		default:
			t.Error("Progress channel should accept at least one message")
		}
	})
}

func TestUninstallerRun(t *testing.T) {
	t.Run("not installed returns error", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		// Don't create any config - should error
		uninstaller := NewUninstaller(false, true)
		err := uninstaller.Run()

		if err == nil {
			t.Error("Expected error when not installed")
		}

		expectedMsg := "no Hefesto configuration found"
		if err != nil && !containsString(err.Error(), expectedMsg) {
			t.Errorf("Error should contain '%s', got: %v", expectedMsg, err)
		}
	})

	t.Run("uninstall with backup", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)

		// Create fake installation
		configPath := filepath.Join(tmpHome, ".config", "opencode")
		if err := os.MkdirAll(configPath, 0755); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		// Create AGENTS.md and opencode.json (required for "installed")
		if err := os.WriteFile(filepath.Join(configPath, "AGENTS.md"), []byte("# Hefesto"), 0644); err != nil {
			t.Fatalf("Failed to create AGENTS.md: %v", err)
		}
		if err := os.WriteFile(filepath.Join(configPath, "opencode.json"), []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create opencode.json: %v", err)
		}

		// Create a backup
		backupPath := filepath.Join(tmpHome, ".config", "opencode-backup-20240101-100000")
		if err := os.MkdirAll(backupPath, 0755); err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}
		// Add a file to the backup
		if err := os.WriteFile(filepath.Join(backupPath, "old-file.txt"), []byte("old content"), 0644); err != nil {
			t.Fatalf("Failed to create backup file: %v", err)
		}

		uninstaller := NewUninstaller(false, true) // skipConfirm = true
		err := uninstaller.Run()

		if err != nil {
			t.Errorf("Uninstall should succeed: %v", err)
		}

		// Verify config was removed and backup restored
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config path should still exist (restored from backup)")
		}

		// Verify backup file exists in restored config
		if _, err := os.Stat(filepath.Join(configPath, "old-file.txt")); os.IsNotExist(err) {
			t.Error("Old file should exist after restore")
		}
	})

	t.Run("uninstall with purge", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)

		// Create fake installation
		configPath := filepath.Join(tmpHome, ".config", "opencode")
		if err := os.MkdirAll(configPath, 0755); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		if err := os.WriteFile(filepath.Join(configPath, "AGENTS.md"), []byte("# Hefesto"), 0644); err != nil {
			t.Fatalf("Failed to create AGENTS.md: %v", err)
		}
		if err := os.WriteFile(filepath.Join(configPath, "opencode.json"), []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create opencode.json: %v", err)
		}

		// Create a backup (should be ignored in purge mode)
		backupPath := filepath.Join(tmpHome, ".config", "opencode-backup-20240101-100000")
		if err := os.MkdirAll(backupPath, 0755); err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		uninstaller := NewUninstaller(true, true) // purge = true, skipConfirm = true
		err := uninstaller.Run()

		if err != nil {
			t.Errorf("Uninstall should succeed: %v", err)
		}

		// Verify config was removed
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Error("Config path should be removed in purge mode")
		}

		// Verify backup still exists (should not be touched in purge)
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Error("Backup should still exist after purge")
		}
	})

	t.Run("uninstall without backup", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)

		// Create fake installation
		configPath := filepath.Join(tmpHome, ".config", "opencode")
		if err := os.MkdirAll(configPath, 0755); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		if err := os.WriteFile(filepath.Join(configPath, "AGENTS.md"), []byte("# Hefesto"), 0644); err != nil {
			t.Fatalf("Failed to create AGENTS.md: %v", err)
		}
		if err := os.WriteFile(filepath.Join(configPath, "opencode.json"), []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create opencode.json: %v", err)
		}

		// No backup created
		uninstaller := NewUninstaller(false, true) // purge = false, skipConfirm = true
		err := uninstaller.Run()

		if err != nil {
			t.Errorf("Uninstall should succeed: %v", err)
		}

		// Verify config was removed
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Error("Config path should be removed when no backup exists")
		}
	})
}

func TestUninstallProgress(t *testing.T) {
	t.Run("progress structure", func(t *testing.T) {
		progress := UninstallProgress{
			Step:    "remove",
			Message: "Removing configuration...",
			Done:    false,
			Error:   nil,
		}

		if progress.Step != "remove" {
			t.Errorf("Expected step 'remove', got %s", progress.Step)
		}

		if progress.Done {
			t.Error("Should not be done")
		}
	})
}

func TestCountDirectories(t *testing.T) {
	t.Run("counts directories correctly", func(t *testing.T) {
		dir := t.TempDir()

		// Create directories and files
		if err := os.MkdirAll(filepath.Join(dir, "dir1"), 0755); err != nil {
			t.Fatalf("Failed to create dir1: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(dir, "dir2"), 0755); err != nil {
			t.Fatalf("Failed to create dir2: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file1: %v", err)
		}

		count := countDirectories(dir)
		if count != 2 {
			t.Errorf("Expected 2 directories, got %d", count)
		}
	})

	t.Run("returns zero for non-existent path", func(t *testing.T) {
		count := countDirectories("/nonexistent/path")
		if count != 0 {
			t.Errorf("Expected 0 for non-existent path, got %d", count)
		}
	})

	t.Run("empty directory returns zero", func(t *testing.T) {
		dir := t.TempDir()
		count := countDirectories(dir)
		if count != 0 {
			t.Errorf("Expected 0 for empty directory, got %d", count)
		}
	})
}

func TestCountFiles(t *testing.T) {
	t.Run("counts files correctly", func(t *testing.T) {
		dir := t.TempDir()

		// Create files and directories
		if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file1: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file2: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(dir, "dir1"), 0755); err != nil {
			t.Fatalf("Failed to create dir1: %v", err)
		}

		count := countFiles(dir)
		if count != 2 {
			t.Errorf("Expected 2 files, got %d", count)
		}
	})

	t.Run("returns zero for non-existent path", func(t *testing.T) {
		count := countFiles("/nonexistent/path")
		if count != 0 {
			t.Errorf("Expected 0 for non-existent path, got %d", count)
		}
	})

	t.Run("empty directory returns zero", func(t *testing.T) {
		dir := t.TempDir()
		count := countFiles(dir)
		if count != 0 {
			t.Errorf("Expected 0 for empty directory, got %d", count)
		}
	})
}

func TestBackupInfo(t *testing.T) {
	t.Run("backup info structure", func(t *testing.T) {
		now := time.Now()
		info := BackupInfo{
			Path:      "/path/to/backup",
			Name:      "opencode-backup-20240101-100000",
			Timestamp: now,
		}

		if info.Path != "/path/to/backup" {
			t.Errorf("Expected path, got %s", info.Path)
		}

		if info.Name != "opencode-backup-20240101-100000" {
			t.Errorf("Expected name, got %s", info.Name)
		}

		if !info.Timestamp.Equal(now) {
			t.Error("Timestamp mismatch")
		}
	})
}

// Note: FormatBackupDate is already tested in rollback_test.go

func TestCopyDirectory(t *testing.T) {
	t.Run("copies directory structure", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		// Create source structure
		if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
			t.Fatalf("Failed to create subdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644); err != nil {
			t.Fatalf("Failed to create file1: %v", err)
		}
		if err := os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0644); err != nil {
			t.Fatalf("Failed to create file2: %v", err)
		}

		// Create node_modules (should be skipped)
		if err := os.MkdirAll(filepath.Join(srcDir, "node_modules"), 0755); err != nil {
			t.Fatalf("Failed to create node_modules: %v", err)
		}
		if err := os.WriteFile(filepath.Join(srcDir, "node_modules", "skip.txt"), []byte("skip"), 0644); err != nil {
			t.Fatalf("Failed to create skip file: %v", err)
		}

		dstPath := filepath.Join(dstDir, "copy")
		err := copyDirectory(srcDir, dstPath)
		if err != nil {
			t.Fatalf("copyDirectory failed: %v", err)
		}

		// Verify files were copied
		if _, err := os.Stat(filepath.Join(dstPath, "file1.txt")); os.IsNotExist(err) {
			t.Error("file1.txt should be copied")
		}
		if _, err := os.Stat(filepath.Join(dstPath, "subdir", "file2.txt")); os.IsNotExist(err) {
			t.Error("subdir/file2.txt should be copied")
		}

		// Verify node_modules was skipped
		if _, err := os.Stat(filepath.Join(dstPath, "node_modules")); !os.IsNotExist(err) {
			t.Error("node_modules should be skipped")
		}
	})

	t.Run("returns error for non-existent source", func(t *testing.T) {
		dstDir := t.TempDir()
		err := copyDirectory("/nonexistent", filepath.Join(dstDir, "copy"))
		if err == nil {
			t.Error("Should return error for non-existent source")
		}
	})
}

func TestRunUninstall(t *testing.T) {
	t.Run("creates uninstaller and runs", func(t *testing.T) {
		// This test verifies the function creates an uninstaller
		// Actual uninstall logic is tested in TestUninstallerRun
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		configDir := filepath.Join(tmpDir, ".config", "opencode")
		os.MkdirAll(configDir, 0755)
		os.WriteFile(filepath.Join(configDir, "AGENTS.md"), []byte("# Test"), 0644)
		os.WriteFile(filepath.Join(configDir, "opencode.json"), []byte("{}"), 0644)

		// With fake config installed, uninstall should succeed
		err := RunUninstall(false, true)
		if err != nil {
			t.Errorf("Uninstall should succeed: %v", err)
		}
	})
}

// Note: Testing print functions (printRestoreSummary, printPurgeSummary, etc.)
// would require capturing stdout. For unit tests, we focus on the logic functions.
// The print functions are tested implicitly through integration tests.
