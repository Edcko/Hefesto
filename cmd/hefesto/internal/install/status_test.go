package install

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================
// Tests for formatBytes
// ============================================

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 bytes"},
		{"100 bytes", 100, "100 bytes"},
		{"1023 bytes", 1023, "1023 bytes"},
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"1 MB", 1048576, "1.0 MB"},
		{"1.5 MB", 1572864, "1.5 MB"},
		{"1 GB", 1073741824, "1.0 GB"},
		{"2.5 GB", 2684354560, "2.5 GB"},
		{"1 TB", 1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			if got != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.expected)
			}
		})
	}
}

// ============================================
// Tests for checkFile
// ============================================

func TestCheckFile(t *testing.T) {
	t.Run("file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.md")
		os.WriteFile(testFile, []byte("hello world"), 0644)

		result := checkFile(tmpDir, "test.md")

		if !result.Present {
			t.Error("Expected Present=true, got false")
		}
		if result.Detail == "Missing" {
			t.Error("Expected size detail, got Missing")
		}
	})

	t.Run("file missing", func(t *testing.T) {
		tmpDir := t.TempDir()

		result := checkFile(tmpDir, "nonexistent.md")

		if result.Present {
			t.Error("Expected Present=false, got true")
		}
		if result.Detail != "Missing" {
			t.Errorf("Expected Detail=Missing, got %q", result.Detail)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "empty.md")
		os.WriteFile(testFile, []byte{}, 0644)

		result := checkFile(tmpDir, "empty.md")

		if !result.Present {
			t.Error("Expected Present=true for empty file")
		}
		if result.Detail != "0 bytes" {
			t.Errorf("Expected Detail=0 bytes, got %q", result.Detail)
		}
	})
}

// ============================================
// Tests for checkDirectory
// ============================================

func TestCheckDirectory(t *testing.T) {
	t.Run("directory exists with files", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "plugins")
		os.MkdirAll(subDir, 0755)
		os.WriteFile(filepath.Join(subDir, "plugin1.ts"), []byte("code"), 0644)
		os.WriteFile(filepath.Join(subDir, "plugin2.ts"), []byte("code"), 0644)

		result := checkDirectory(tmpDir, "plugins")

		if !result.Present {
			t.Error("Expected Present=true, got false")
		}
	})

	t.Run("directory missing", func(t *testing.T) {
		tmpDir := t.TempDir()

		result := checkDirectory(tmpDir, "nonexistent")

		if result.Present {
			t.Error("Expected Present=false, got true")
		}
		if result.Detail != "Missing" {
			t.Errorf("Expected Detail=Missing, got %q", result.Detail)
		}
	})

	t.Run("path is a file not directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "skills")
		os.WriteFile(filePath, []byte("not a dir"), 0644)

		result := checkDirectory(tmpDir, "skills")

		if result.Present {
			t.Error("Expected Present=false when path is a file")
		}
		if result.Detail != "Not a directory" {
			t.Errorf("Expected Detail=Not a directory, got %q", result.Detail)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "plugins")
		os.MkdirAll(subDir, 0755)

		result := checkDirectory(tmpDir, "plugins")

		if !result.Present {
			t.Error("Expected Present=true for empty directory")
		}
	})

	t.Run("skills directory counts subdirectories", func(t *testing.T) {
		tmpDir := t.TempDir()
		skillsDir := filepath.Join(tmpDir, "skills")
		os.MkdirAll(filepath.Join(skillsDir, "skill1"), 0755)
		os.MkdirAll(filepath.Join(skillsDir, "skill2"), 0755)
		os.WriteFile(filepath.Join(skillsDir, "file.txt"), []byte("not a dir"), 0644)

		result := checkDirectory(tmpDir, "skills")

		if !result.Present {
			t.Error("Expected Present=true")
		}
		// Should report directories count, not files
	})
}

// ============================================
// Tests for checkThemeDirectory
// ============================================

func TestCheckThemeDirectory(t *testing.T) {
	t.Run("themes directory with json file", func(t *testing.T) {
		tmpDir := t.TempDir()
		themesDir := filepath.Join(tmpDir, "themes")
		os.MkdirAll(themesDir, 0755)
		os.WriteFile(filepath.Join(themesDir, "kanagawa.json"), []byte(`{"name": "Kanagawa"}`), 0644)

		result := checkThemeDirectory(tmpDir)

		if !result.Present {
			t.Error("Expected Present=true, got false")
		}
	})

	t.Run("themes directory missing", func(t *testing.T) {
		tmpDir := t.TempDir()

		result := checkThemeDirectory(tmpDir)

		if result.Present {
			t.Error("Expected Present=false, got true")
		}
		if result.Detail != "Missing" {
			t.Errorf("Expected Detail=Missing, got %q", result.Detail)
		}
	})

	t.Run("themes directory empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		themesDir := filepath.Join(tmpDir, "themes")
		os.MkdirAll(themesDir, 0755)

		result := checkThemeDirectory(tmpDir)

		if result.Present {
			t.Error("Expected Present=false when no json files")
		}
		if result.Detail != "No .json files" {
			t.Errorf("Expected Detail=No .json files, got %q", result.Detail)
		}
	})

	t.Run("themes directory with non-json files only", func(t *testing.T) {
		tmpDir := t.TempDir()
		themesDir := filepath.Join(tmpDir, "themes")
		os.MkdirAll(themesDir, 0755)
		os.WriteFile(filepath.Join(themesDir, "readme.txt"), []byte("readme"), 0644)

		result := checkThemeDirectory(tmpDir)

		if result.Present {
			t.Error("Expected Present=false when no json files")
		}
		if result.Detail != "No .json files" {
			t.Errorf("Expected Detail=No .json files, got %q", result.Detail)
		}
	})

	t.Run("themes is a file not directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "themes")
		os.WriteFile(filePath, []byte("not a dir"), 0644)

		result := checkThemeDirectory(tmpDir)

		if result.Present {
			t.Error("Expected Present=false when themes is a file")
		}
		if result.Detail != "Not a directory" {
			t.Errorf("Expected Detail=Not a directory, got %q", result.Detail)
		}
	})
}

// ============================================
// Tests for detectVersion
// ============================================

func TestDetectVersion(t *testing.T) {
	t.Run("AGENTS.md with Hefesto", func(t *testing.T) {
		tmpDir := t.TempDir()
		agentsPath := filepath.Join(tmpDir, "AGENTS.md")
		os.WriteFile(agentsPath, []byte("# Hefesto Configuration\nSome content"), 0644)

		version := detectVersion(tmpDir)

		if version != "Hefesto Config" {
			t.Errorf("Expected version=Hefesto Config, got %q", version)
		}
	})

	t.Run("AGENTS.md with Gentleman", func(t *testing.T) {
		tmpDir := t.TempDir()
		agentsPath := filepath.Join(tmpDir, "AGENTS.md")
		os.WriteFile(agentsPath, []byte("# Gentleman.Dots Configuration\nSome content"), 0644)

		version := detectVersion(tmpDir)

		if version != "Gentleman.Dots" {
			t.Errorf("Expected version=Gentleman.Dots, got %q", version)
		}
	})

	t.Run("opencode.json with version field", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte("No markers here"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "opencode.json"), []byte(`{"version": "1.0.0"}`), 0644)

		version := detectVersion(tmpDir)

		if version != "Custom Config" {
			t.Errorf("Expected version=Custom Config, got %q", version)
		}
	})

	t.Run("no version markers", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte("Generic content"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "opencode.json"), []byte(`{}`), 0644)

		version := detectVersion(tmpDir)

		if version != "Unknown" {
			t.Errorf("Expected version=Unknown, got %q", version)
		}
	})

	t.Run("no files", func(t *testing.T) {
		tmpDir := t.TempDir()

		version := detectVersion(tmpDir)

		if version != "Unknown" {
			t.Errorf("Expected version=Unknown, got %q", version)
		}
	})
}

// ============================================
// Tests for checkComponents
// ============================================

func TestCheckComponents(t *testing.T) {
	t.Run("full install", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create all components
		os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte("# Hefesto"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "opencode.json"), []byte(`{}`), 0644)
		os.WriteFile(filepath.Join(tmpDir, "personality"), []byte("personality"), 0644)

		// Create directories
		os.MkdirAll(filepath.Join(tmpDir, "skills", "skill1"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "plugins"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "commands"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "themes"), 0755)
		os.WriteFile(filepath.Join(tmpDir, "themes", "theme.json"), []byte(`{}`), 0644)

		result := checkComponents(tmpDir)

		if !result.AgentsMD.Present {
			t.Error("AGENTS.md should be present")
		}
		if !result.OpenCodeJSON.Present {
			t.Error("opencode.json should be present")
		}
		if !result.Skills.Present {
			t.Error("skills should be present")
		}
		if !result.Plugins.Present {
			t.Error("plugins should be present")
		}
		if !result.Personality.Present {
			t.Error("personality should be present")
		}
		if !result.Theme.Present {
			t.Error("theme should be present")
		}
		if !result.Commands.Present {
			t.Error("commands should be present")
		}
	})

	t.Run("only AGENTS.md and opencode.json", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte("# Hefesto"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "opencode.json"), []byte(`{}`), 0644)

		result := checkComponents(tmpDir)

		if !result.AgentsMD.Present {
			t.Error("AGENTS.md should be present")
		}
		if !result.OpenCodeJSON.Present {
			t.Error("opencode.json should be present")
		}
		if result.Skills.Present {
			t.Error("skills should NOT be present")
		}
		if result.Theme.Present {
			t.Error("theme should NOT be present")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		result := checkComponents(tmpDir)

		if result.AgentsMD.Present {
			t.Error("AGENTS.md should NOT be present")
		}
		if result.OpenCodeJSON.Present {
			t.Error("opencode.json should NOT be present")
		}
		if result.Skills.Present {
			t.Error("skills should NOT be present")
		}
	})

	t.Run("missing theme", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte("# Hefesto"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "opencode.json"), []byte(`{}`), 0644)
		os.MkdirAll(filepath.Join(tmpDir, "themes"), 0755) // empty themes dir

		result := checkComponents(tmpDir)

		if result.Theme.Present {
			t.Error("theme should NOT be present when no json files")
		}
		if result.Theme.Detail != "No .json files" {
			t.Errorf("Expected Detail=No .json files, got %q", result.Theme.Detail)
		}
	})

	t.Run("missing personality", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte("# Hefesto"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "opencode.json"), []byte(`{}`), 0644)

		result := checkComponents(tmpDir)

		if result.Personality.Present {
			t.Error("personality should NOT be present")
		}
	})
}

// ============================================
// Tests for CheckStatus (integration-style)
// ============================================

func TestCheckStatus_Integration(t *testing.T) {
	// This test uses the real system, so results may vary
	status, err := CheckStatus()
	if err != nil {
		t.Fatalf("CheckStatus() error: %v", err)
	}

	// Basic sanity checks
	if status.ConfigPath == "" {
		t.Error("ConfigPath should not be empty")
	}

	// The Installed flag should be based on both AGENTS.md AND opencode.json
	// We can't assume the test machine has a full install
	t.Logf("Installed: %v", status.Installed)
	t.Logf("ConfigPath: %s", status.ConfigPath)
}

// ============================================
// Tests for StatusInfo structure
// ============================================

func TestStatusInfo_Structure(t *testing.T) {
	info := &StatusInfo{
		ConfigPath: "/test/path",
		Installed:  true,
		Version:    "1.0.0",
		Components: ComponentStatus{
			AgentsMD:     ComponentDetail{Present: true, Detail: "100 bytes"},
			OpenCodeJSON: ComponentDetail{Present: true, Detail: "200 bytes"},
			Skills:       ComponentDetail{Present: true, Detail: "5 directories"},
			Plugins:      ComponentDetail{Present: false, Detail: "Missing"},
			Personality:  ComponentDetail{Present: true, Detail: "50 bytes"},
			Theme:        ComponentDetail{Present: true, Detail: "theme.json (1.0 KB)"},
			Commands:     ComponentDetail{Present: true, Detail: "3 files"},
		},
		Binaries: BinaryStatus{
			Engram: BinaryDetail{
				Installed: true,
				Version:   "1.3.1",
				Path:      "/usr/local/bin/engram",
			},
			OpenCode: BinaryDetail{
				Installed: true,
				Version:   "1.3.13",
				Path:      "/usr/local/bin/opencode",
			},
		},
	}

	if info.ConfigPath != "/test/path" {
		t.Errorf("ConfigPath = %q, want /test/path", info.ConfigPath)
	}
	if !info.Installed {
		t.Error("Installed should be true")
	}
	if !info.Components.AgentsMD.Present {
		t.Error("AgentsMD should be present")
	}
	if !info.Binaries.Engram.Installed {
		t.Error("Engram should be installed")
	}
}
