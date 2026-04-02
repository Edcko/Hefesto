package install

import (
	"os"
	"path/filepath"
	"testing"

	embedpkg "github.com/Edcko/Hefesto/cmd/hefesto/internal/embed"
)

func TestVerify(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "opencode")

	// Copy config first
	CopyConfig(embedpkg.ConfigFiles, targetPath)

	result, err := Verify(targetPath)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}

	if !result.ConfigCopied {
		t.Error("ConfigCopied = false, want true")
	}
}

func TestVerifyMissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "opencode")

	// Create directory but no config file
	os.MkdirAll(targetPath, 0755)

	result, err := Verify(targetPath)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}

	if result.ConfigCopied {
		t.Error("ConfigCopied = true for missing opencode.json, want false")
	}

	// Should have error about missing config
	if len(result.Errors) == 0 {
		t.Error("Expected errors when config is missing")
	}
}

func TestVerifyMissingAgents(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "opencode")

	// Create only opencode.json
	os.MkdirAll(targetPath, 0755)
	os.WriteFile(filepath.Join(targetPath, "opencode.json"), []byte(`{}`), 0644)

	result, err := Verify(targetPath)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}

	// Should have error about missing AGENTS.md
	foundAgentsError := false
	for _, errMsg := range result.Errors {
		if errMsg == "AGENTS.md is missing" {
			foundAgentsError = true
			break
		}
	}
	if !foundAgentsError {
		t.Error("Expected error about missing AGENTS.md")
	}
}

func TestVerifyInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "opencode")

	// Create invalid JSON
	os.MkdirAll(targetPath, 0755)
	os.WriteFile(filepath.Join(targetPath, "opencode.json"), []byte(`{invalid json}`), 0644)

	result, err := Verify(targetPath)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}

	if result.ConfigCopied {
		t.Error("ConfigCopied = true for invalid JSON, want false")
	}

	// Should have error about invalid JSON
	if len(result.Errors) == 0 {
		t.Error("Expected errors when JSON is invalid")
	}
}

func TestVerifyNpmInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "opencode")

	// Copy config
	CopyConfig(embedpkg.ConfigFiles, targetPath)

	// Create fake node_modules/@opencode-ai/plugin
	pluginPath := filepath.Join(targetPath, "node_modules", "@opencode-ai", "plugin")
	os.MkdirAll(pluginPath, 0755)

	result, err := Verify(targetPath)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}

	if !result.NpmInstalled {
		t.Error("NpmInstalled = false when plugin exists, want true")
	}
}

func TestVerifyNpmNotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "opencode")

	// Copy config (without node_modules)
	CopyConfig(embedpkg.ConfigFiles, targetPath)

	result, err := Verify(targetPath)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}

	if result.NpmInstalled {
		t.Error("NpmInstalled = true when plugin doesn't exist, want false")
	}
}
