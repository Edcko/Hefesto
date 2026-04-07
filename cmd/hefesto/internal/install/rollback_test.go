// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestListBackupsInDir(t *testing.T) {
	t.Run("no backups exist", func(t *testing.T) {
		dir := t.TempDir()
		backups, err := ListBackupsInDir(dir)
		if err != nil {
			t.Fatalf("ListBackupsInDir() error = %v", err)
		}
		if len(backups) != 0 {
			t.Errorf("ListBackupsInDir() = %d backups, want 0", len(backups))
		}
	})

	t.Run("one backup exists", func(t *testing.T) {
		dir := t.TempDir()
		backupPath := filepath.Join(dir, "opencode-backup-20260406-160143")
		if err := os.MkdirAll(backupPath, 0755); err != nil {
			t.Fatalf("Failed to create backup directory: %v", err)
		}

		backups, err := ListBackupsInDir(dir)
		if err != nil {
			t.Fatalf("ListBackupsInDir() error = %v", err)
		}
		if len(backups) != 1 {
			t.Errorf("ListBackupsInDir() = %d backups, want 1", len(backups))
		}
		if backups[0].Name != "opencode-backup-20260406-160143" {
			t.Errorf("ListBackupsInDir() backup name = %s, want opencode-backup-20260406-160143", backups[0].Name)
		}
	})

	t.Run("multiple backups sorted newest first", func(t *testing.T) {
		dir := t.TempDir()
		// Create backups in random order
		backup1 := filepath.Join(dir, "opencode-backup-20260401-120000")
		backup2 := filepath.Join(dir, "opencode-backup-20260406-160143")
		backup3 := filepath.Join(dir, "opencode-backup-20260403-173508")

		for _, b := range []string{backup1, backup2, backup3} {
			if err := os.MkdirAll(b, 0755); err != nil {
				t.Fatalf("Failed to create backup directory: %v", err)
			}
		}

		backups, err := ListBackupsInDir(dir)
		if err != nil {
			t.Fatalf("ListBackupsInDir() error = %v", err)
		}
		if len(backups) != 3 {
			t.Errorf("ListBackupsInDir() = %d backups, want 3", len(backups))
		}

		// Verify order: newest first
		expected := []string{
			"opencode-backup-20260406-160143",
			"opencode-backup-20260403-173508",
			"opencode-backup-20260401-120000",
		}
		for i, exp := range expected {
			if backups[i].Name != exp {
				t.Errorf("backups[%d].Name = %s, want %s", i, backups[i].Name, exp)
			}
		}
	})

	t.Run("backups with different timestamps correct ordering", func(t *testing.T) {
		dir := t.TempDir()
		// Create backups from different years and months
		backup1 := filepath.Join(dir, "opencode-backup-20250115-090000")
		backup2 := filepath.Join(dir, "opencode-backup-20260406-160143")
		backup3 := filepath.Join(dir, "opencode-backup-20251231-235959")

		for _, b := range []string{backup1, backup2, backup3} {
			if err := os.MkdirAll(b, 0755); err != nil {
				t.Fatalf("Failed to create backup directory: %v", err)
			}
		}

		backups, err := ListBackupsInDir(dir)
		if err != nil {
			t.Fatalf("ListBackupsInDir() error = %v", err)
		}

		// Should be sorted: 2026 > 2025-12 > 2025-01
		expected := []string{
			"opencode-backup-20260406-160143",
			"opencode-backup-20251231-235959",
			"opencode-backup-20250115-090000",
		}
		for i, exp := range expected {
			if backups[i].Name != exp {
				t.Errorf("backups[%d].Name = %s, want %s", i, backups[i].Name, exp)
			}
		}
	})

	t.Run("non-matching directories are ignored", func(t *testing.T) {
		dir := t.TempDir()
		// Create valid backup
		validBackup := filepath.Join(dir, "opencode-backup-20260406-160143")
		if err := os.MkdirAll(validBackup, 0755); err != nil {
			t.Fatalf("Failed to create backup directory: %v", err)
		}

		// Create non-matching directories
		invalidDirs := []string{
			"opencode-20260406-160143",         // missing "backup" prefix
			"other-dir",                        // completely different name
			"opencode-backup-invalid",          // invalid timestamp
			"opencode-backup-2026-04-06",       // wrong timestamp format
			"opencode-backup-20260406-1601430", // extra digit
		}
		for _, d := range invalidDirs {
			if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}
		}

		backups, err := ListBackupsInDir(dir)
		if err != nil {
			t.Fatalf("ListBackupsInDir() error = %v", err)
		}
		if len(backups) != 1 {
			t.Errorf("ListBackupsInDir() = %d backups, want 1 (only valid backup)", len(backups))
		}
		if len(backups) > 0 && backups[0].Name != "opencode-backup-20260406-160143" {
			t.Errorf("ListBackupsInDir() = %s, want opencode-backup-20260406-160143", backups[0].Name)
		}
	})

	t.Run("files are ignored only directories", func(t *testing.T) {
		dir := t.TempDir()
		// Create a file with backup pattern name
		filePath := filepath.Join(dir, "opencode-backup-20260406-160143")
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		// Create a valid backup directory
		backupDir := filepath.Join(dir, "opencode-backup-20260403-173508")
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			t.Fatalf("Failed to create backup directory: %v", err)
		}

		backups, err := ListBackupsInDir(dir)
		if err != nil {
			t.Fatalf("ListBackupsInDir() error = %v", err)
		}
		if len(backups) != 1 {
			t.Errorf("ListBackupsInDir() = %d backups, want 1 (files should be ignored)", len(backups))
		}
		if len(backups) > 0 && backups[0].Name != "opencode-backup-20260403-173508" {
			t.Errorf("ListBackupsInDir() = %s, want opencode-backup-20260403-173508", backups[0].Name)
		}
	})
}

func TestParseBackupTimestamp(t *testing.T) {
	t.Run("valid format", func(t *testing.T) {
		name := "opencode-backup-20260406-160143"
		got, err := ParseBackupTimestamp(name)
		if err != nil {
			t.Fatalf("ParseBackupTimestamp() error = %v", err)
		}

		// Expected: April 6, 2026 at 16:01:43
		want := time.Date(2026, 4, 6, 16, 1, 43, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("ParseBackupTimestamp() = %v, want %v", got, want)
		}
	})

	t.Run("invalid format - missing prefix", func(t *testing.T) {
		name := "backup-20260406-160143"
		_, err := ParseBackupTimestamp(name)
		if err == nil {
			t.Error("ParseBackupTimestamp() expected error for invalid prefix, got nil")
		}
	})

	t.Run("invalid format - wrong timestamp", func(t *testing.T) {
		name := "opencode-backup-2026-04-06"
		_, err := ParseBackupTimestamp(name)
		if err == nil {
			t.Error("ParseBackupTimestamp() expected error for invalid timestamp format, got nil")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := ParseBackupTimestamp("")
		if err == nil {
			t.Error("ParseBackupTimestamp() expected error for empty string, got nil")
		}
	})

	t.Run("different valid dates", func(t *testing.T) {
		tests := []struct {
			name     string
			expected time.Time
		}{
			{
				name:     "opencode-backup-20260101-000000",
				expected: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				name:     "opencode-backup-20251231-235959",
				expected: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			},
			{
				name:     "opencode-backup-20240229-120000",
				expected: time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := ParseBackupTimestamp(tt.name)
				if err != nil {
					t.Fatalf("ParseBackupTimestamp() error = %v", err)
				}
				if !got.Equal(tt.expected) {
					t.Errorf("ParseBackupTimestamp() = %v, want %v", got, tt.expected)
				}
			})
		}
	})

	t.Run("invalid timestamp - extra digits", func(t *testing.T) {
		name := "opencode-backup-20260406-1601430"
		_, err := ParseBackupTimestamp(name)
		if err == nil {
			t.Error("ParseBackupTimestamp() expected error for extra digits, got nil")
		}
	})
}

func TestFormatBackupDate(t *testing.T) {
	t.Run("known timestamp", func(t *testing.T) {
		// April 6, 2026 at 16:01:43
		tm := time.Date(2026, 4, 6, 16, 1, 43, 0, time.UTC)
		got := FormatBackupDate(tm)
		expected := "Apr 6, 2026 — 16:01"
		if got != expected {
			t.Errorf("FormatBackupDate() = %q, want %q", got, expected)
		}
	})

	t.Run("zero time", func(t *testing.T) {
		got := FormatBackupDate(time.Time{})
		// Zero time should format to "Jan 1, 0001 — 00:00"
		expected := "Jan 1, 0001 — 00:00"
		if got != expected {
			t.Errorf("FormatBackupDate() = %q, want %q", got, expected)
		}
	})

	t.Run("different dates", func(t *testing.T) {
		tests := []struct {
			time     time.Time
			expected string
		}{
			{
				time:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				expected: "Jan 1, 2026 — 00:00",
			},
			{
				time:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
				expected: "Dec 31, 2025 — 23:59",
			},
			{
				time:     time.Date(2024, 7, 4, 14, 30, 0, 0, time.UTC),
				expected: "Jul 4, 2024 — 14:30",
			},
		}

		for _, tt := range tests {
			t.Run(tt.expected, func(t *testing.T) {
				got := FormatBackupDate(tt.time)
				if got != tt.expected {
					t.Errorf("FormatBackupDate() = %q, want %q", got, tt.expected)
				}
			})
		}
	})
}

func TestFindMostRecentBackup(t *testing.T) {
	t.Run("multiple backups returns newest", func(t *testing.T) {
		backups := []BackupInfo{
			{Name: "opencode-backup-20260406-160143", Timestamp: time.Date(2026, 4, 6, 16, 1, 43, 0, time.UTC)},
			{Name: "opencode-backup-20260403-173508", Timestamp: time.Date(2026, 4, 3, 17, 35, 8, 0, time.UTC)},
			{Name: "opencode-backup-20260401-120000", Timestamp: time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)},
		}

		got, err := FindMostRecentBackup(backups)
		if err != nil {
			t.Fatalf("FindMostRecentBackup() error = %v", err)
		}
		if got.Name != "opencode-backup-20260406-160143" {
			t.Errorf("FindMostRecentBackup() = %s, want opencode-backup-20260406-160143", got.Name)
		}
	})

	t.Run("no backups returns error", func(t *testing.T) {
		backups := []BackupInfo{}
		_, err := FindMostRecentBackup(backups)
		if err == nil {
			t.Error("FindMostRecentBackup() expected error for empty list, got nil")
		}
	})

	t.Run("single backup returns that one", func(t *testing.T) {
		backups := []BackupInfo{
			{Name: "opencode-backup-20260406-160143", Timestamp: time.Date(2026, 4, 6, 16, 1, 43, 0, time.UTC)},
		}

		got, err := FindMostRecentBackup(backups)
		if err != nil {
			t.Fatalf("FindMostRecentBackup() error = %v", err)
		}
		if got.Name != "opencode-backup-20260406-160143" {
			t.Errorf("FindMostRecentBackup() = %s, want opencode-backup-20260406-160143", got.Name)
		}
	})
}

func TestPromptRollback(t *testing.T) {
	t.Run("no backups returns nil", func(t *testing.T) {
		backups := []BackupInfo{}
		got := PromptRollback(backups)
		if got != nil {
			t.Error("PromptRollback() expected nil for empty list, got non-nil")
		}
	})

	t.Run("with backups returns most recent", func(t *testing.T) {
		backups := []BackupInfo{
			{Name: "opencode-backup-20260406-160143", Timestamp: time.Date(2026, 4, 6, 16, 1, 43, 0, time.UTC)},
			{Name: "opencode-backup-20260403-173508", Timestamp: time.Date(2026, 4, 3, 17, 35, 8, 0, time.UTC)},
		}

		got := PromptRollback(backups)
		if got == nil {
			t.Fatal("PromptRollback() returned nil, expected non-nil")
		}
		if got.Name != "opencode-backup-20260406-160143" {
			t.Errorf("PromptRollback() = %s, want opencode-backup-20260406-160143", got.Name)
		}
	})

	t.Run("single backup returns that one", func(t *testing.T) {
		backups := []BackupInfo{
			{Name: "opencode-backup-20260406-160143", Timestamp: time.Date(2026, 4, 6, 16, 1, 43, 0, time.UTC)},
		}

		got := PromptRollback(backups)
		if got == nil {
			t.Fatal("PromptRollback() returned nil, expected non-nil")
		}
		if got.Name != "opencode-backup-20260406-160143" {
			t.Errorf("PromptRollback() = %s, want opencode-backup-20260406-160143", got.Name)
		}
	})
}
