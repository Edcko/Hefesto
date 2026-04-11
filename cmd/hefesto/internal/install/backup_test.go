package install

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestBackup(t *testing.T) {
	// Create a fake config directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Backup it
	backupPath, err := Backup(configDir)
	if err != nil {
		t.Fatalf("Backup() error: %v", err)
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup path %q does not exist", backupPath)
	}

	// Verify content was copied
	content, err := os.ReadFile(filepath.Join(backupPath, "test.txt"))
	if err != nil {
		t.Fatalf("Read backup file error: %v", err)
	}
	if string(content) != "hello" {
		t.Errorf("Backup content = %q, want %q", content, "hello")
	}
}

func TestBackupNonExistent(t *testing.T) {
	_, err := Backup("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Backup() should return error for nonexistent path")
	}
}

func TestBackupWithSubdirectories(t *testing.T) {
	// Create a fake config directory with subdirectories
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(filepath.Join(configDir, "skills", "test"), 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "skills", "test", "SKILL.md"), []byte("# Test Skill"), 0644); err != nil {
		t.Fatalf("Failed to write SKILL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "opencode.json"), []byte(`{"test": true}`), 0644); err != nil {
		t.Fatalf("Failed to write opencode.json: %v", err)
	}

	// Backup it
	backupPath, err := Backup(configDir)
	if err != nil {
		t.Fatalf("Backup() error: %v", err)
	}

	// Verify nested file was copied
	content, err := os.ReadFile(filepath.Join(backupPath, "skills", "test", "SKILL.md"))
	if err != nil {
		t.Fatalf("Read nested backup file error: %v", err)
	}
	if string(content) != "# Test Skill" {
		t.Errorf("Nested file content = %q, want %q", content, "# Test Skill")
	}
}

func TestBackupSkipsNodeModules(t *testing.T) {
	// Create a fake config directory with node_modules
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(filepath.Join(configDir, "node_modules", "package"), 0755); err != nil {
		t.Fatalf("Failed to create node_modules dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "node_modules", "package", "index.js"), []byte("module.exports = {}"), 0644); err != nil {
		t.Fatalf("Failed to write index.js: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("Failed to write config.json: %v", err)
	}

	// Backup it
	backupPath, err := Backup(configDir)
	if err != nil {
		t.Fatalf("Backup() error: %v", err)
	}

	// Verify node_modules was skipped
	nodeModulesPath := filepath.Join(backupPath, "node_modules")
	if _, err := os.Stat(nodeModulesPath); !os.IsNotExist(err) {
		t.Error("node_modules should have been skipped during backup")
	}

	// Verify regular file was copied
	if _, err := os.Stat(filepath.Join(backupPath, "config.json")); os.IsNotExist(err) {
		t.Error("config.json should have been copied")
	}
}

func TestCleanOldBackupsInDir(t *testing.T) {
	// Create a fake config directory with 7 backups (limit is 5)
	tmpDir := t.TempDir()

	// Create backup directories with valid timestamp names
	timestamps := []string{
		"20260101-100000",
		"20260102-100000",
		"20260103-100000",
		"20260104-100000",
		"20260105-100000",
		"20260106-100000",
		"20260107-100000",
	}
	for _, ts := range timestamps {
		backupDir := filepath.Join(tmpDir, "opencode-backup-"+ts)
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			t.Fatalf("Failed to create backup dir: %v", err)
		}
		// Add a marker file so the directory isn't empty
		if err := os.WriteFile(filepath.Join(backupDir, "marker.txt"), []byte(ts), 0644); err != nil {
			t.Fatalf("Failed to write marker file: %v", err)
		}
	}

	// Run cleanup
	if err := CleanOldBackupsInDir(tmpDir); err != nil {
		t.Fatalf("CleanOldBackupsInDir() error: %v", err)
	}

	// Count remaining backups
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read dir: %v", err)
	}

	var remaining int
	for _, e := range entries {
		if e.IsDir() && len(e.Name()) > len("opencode-backup-") {
			remaining++
		}
	}

	if remaining != maxBackups {
		t.Errorf("Expected %d remaining backups, got %d", maxBackups, remaining)
	}

	// Verify the 5 newest survived (dates 03-07)
	for _, ts := range timestamps[2:] { // skip the 2 oldest
		path := filepath.Join(tmpDir, "opencode-backup-"+ts)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected backup %s to exist but it was removed", ts)
		}
	}

	// Verify the 2 oldest were removed (dates 01-02)
	for _, ts := range timestamps[:2] {
		path := filepath.Join(tmpDir, "opencode-backup-"+ts)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Expected backup %s to be removed but it still exists", ts)
		}
	}
}

func TestCleanOldBackupsInDirBelowLimit(t *testing.T) {
	// Create fewer backups than the limit — nothing should be removed
	tmpDir := t.TempDir()

	for _, ts := range []string{"20260101-100000", "20260102-100000", "20260103-100000"} {
		backupDir := filepath.Join(tmpDir, "opencode-backup-"+ts)
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			t.Fatalf("Failed to create backup dir: %v", err)
		}
	}

	if err := CleanOldBackupsInDir(tmpDir); err != nil {
		t.Fatalf("CleanOldBackupsInDir() error: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read dir: %v", err)
	}

	var remaining int
	for _, e := range entries {
		if e.IsDir() {
			remaining++
		}
	}

	if remaining != 3 {
		t.Errorf("Expected 3 remaining backups (below limit), got %d", remaining)
	}
}

func TestCleanOldBackupsInDirEmptyDir(t *testing.T) {
	// Empty directory — should not error
	tmpDir := t.TempDir()

	if err := CleanOldBackupsInDir(tmpDir); err != nil {
		t.Fatalf("CleanOldBackupsInDir() on empty dir error: %v", err)
	}
}

func TestCleanOldBackupsInDirExactLimit(t *testing.T) {
	// Exactly maxBackups — nothing should be removed
	tmpDir := t.TempDir()

	for i := 0; i < maxBackups; i++ {
		ts := fmt.Sprintf("202601%02d-100000", i+1)
		backupDir := filepath.Join(tmpDir, "opencode-backup-"+ts)
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			t.Fatalf("Failed to create backup dir: %v", err)
		}
	}

	if err := CleanOldBackupsInDir(tmpDir); err != nil {
		t.Fatalf("CleanOldBackupsInDir() error: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read dir: %v", err)
	}

	var remaining int
	for _, e := range entries {
		if e.IsDir() {
			remaining++
		}
	}

	if remaining != maxBackups {
		t.Errorf("Expected %d remaining backups (exact limit), got %d", maxBackups, remaining)
	}
}
