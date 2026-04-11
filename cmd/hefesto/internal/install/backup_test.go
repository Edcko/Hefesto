package install

import (
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
