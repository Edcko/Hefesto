package install

import (
	"os"
	"path/filepath"
	"testing"

	embedpkg "github.com/Edcko/Hefesto/cmd/hefesto/internal/embed"
)

func TestCopyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "opencode")

	err := CopyConfig(embedpkg.ConfigFiles, targetPath)
	if err != nil {
		t.Fatalf("CopyConfig() error: %v", err)
	}

	// Verify key files exist
	mustExist := []string{
		"opencode.json",
		"AGENTS.md",
		"plugins/engram.ts",
		"skills/sdd-apply/SKILL.md",
	}

	for _, f := range mustExist {
		full := filepath.Join(targetPath, f)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("Expected file %q to exist", f)
		}
	}
}

func TestCopyConfigIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "opencode")

	// Copy twice - should not error
	err := CopyConfig(embedpkg.ConfigFiles, targetPath)
	if err != nil {
		t.Fatalf("First CopyConfig() error: %v", err)
	}

	err = CopyConfig(embedpkg.ConfigFiles, targetPath)
	if err != nil {
		t.Fatalf("Second CopyConfig() error: %v", err)
	}

	// Verify files still exist
	if _, err := os.Stat(filepath.Join(targetPath, "opencode.json")); os.IsNotExist(err) {
		t.Error("opencode.json should exist after second copy")
	}
}

func TestCopyConfigCreatesTargetDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Use a nested path that doesn't exist
	targetPath := filepath.Join(tmpDir, "deeply", "nested", "opencode")

	err := CopyConfig(embedpkg.ConfigFiles, targetPath)
	if err != nil {
		t.Fatalf("CopyConfig() error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("Target directory should have been created")
	}
}

func TestCopyConfigPreservesFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "opencode")

	err := CopyConfig(embedpkg.ConfigFiles, targetPath)
	if err != nil {
		t.Fatalf("CopyConfig() error: %v", err)
	}

	// Read and verify opencode.json is valid
	content, err := os.ReadFile(filepath.Join(targetPath, "opencode.json"))
	if err != nil {
		t.Fatalf("Read opencode.json error: %v", err)
	}

	if len(content) == 0 {
		t.Error("opencode.json should not be empty")
	}
}
