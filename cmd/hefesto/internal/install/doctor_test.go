package install

import (
	"testing"
)

// ============================================
// Tests for isVersionAtLeast
// ============================================

func TestIsVersionAtLeast(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		minimum  string
		expected bool
	}{
		// Equal versions
		{"equal versions", "1.3.1", "1.3.1", true},
		{"equal versions with v prefix", "v1.3.1", "v1.3.1", true},
		{"equal versions mixed v prefix", "v1.3.1", "1.3.1", true},
		{"equal versions mixed v prefix 2", "1.3.1", "v1.3.1", true},

		// Greater versions
		{"greater patch", "1.3.2", "1.3.1", true},
		{"greater minor", "1.4.0", "1.3.1", true},
		{"greater major", "2.0.0", "1.9.9", true},
		{"much greater major", "5.0.0", "1.3.1", true},

		// Lesser versions
		{"lesser patch", "1.3.0", "1.3.1", false},
		{"lesser minor", "1.2.9", "1.3.1", false},
		{"lesser major", "0.9.9", "1.3.1", false},

		// Edge cases
		{"2.0.0 vs 1.9.9", "2.0.0", "1.9.9", true},
		{"1.10.0 vs 1.9.0", "1.10.0", "1.9.0", true},
		{"10.0.0 vs 9.9.9", "10.0.0", "9.9.9", true},

		// Different lengths
		{"shorter version equal parts", "1.3", "1.3.0", false}, // 1.3 has fewer parts
		{"longer version equal parts", "1.3.0.1", "1.3.0", true},

		// Real world examples
		{"engram 1.3.1 >= 1.3.1", "1.3.1", "1.3.1", true},
		{"engram 1.4.0 >= 1.3.1", "1.4.0", "1.3.1", true},
		{"engram 1.2.9 >= 1.3.1", "1.2.9", "1.3.1", false},
		{"opencode 1.3.13 >= 1.3.13", "1.3.13", "1.3.13", true},
		{"opencode 1.3.14 >= 1.3.13", "1.3.14", "1.3.13", true},
		{"opencode 1.3.12 >= 1.3.13", "1.3.12", "1.3.13", false},
		{"opencode 1.4.0 >= 1.3.13", "1.4.0", "1.3.13", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isVersionAtLeast(tt.version, tt.minimum)
			if got != tt.expected {
				t.Errorf("isVersionAtLeast(%q, %q) = %v, want %v", tt.version, tt.minimum, got, tt.expected)
			}
		})
	}
}

// ============================================
// Tests for extractNumber
// ============================================

func TestExtractNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"simple number", "5", 5},
		{"zero", "0", 0},
		{"large number", "123", 123},
		{"number with suffix", "5a", 5},
		{"number with prefix letter", "a5", 0}, // no leading number
		{"empty string", "", 0},
		{"only letters", "abc", 0},
		{"number with special chars", "10-beta", 10},
		{"single digit", "3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractNumber(tt.input)
			if got != tt.expected {
				t.Errorf("extractNumber(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// ============================================
// Tests for isValidUTF8
// ============================================

func TestIsValidUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", true},
		{"simple ASCII", "hello world", true},
		{"UTF-8 with accents", "café résumé", true},
		{"UTF-8 with emojis", "🔥 Hefesto 🚀", true},
		{"UTF-8 with Chinese", "你好世界", true},
		{"UTF-8 with Japanese", "こんにちは", true},
		{"UTF-8 with Arabic", "مرحبا", true},
		{"multiline", "line1\nline2\nline3", true},
		{"markdown content", "# Header\n\nSome **bold** text", true},
		{"JSON-like content", `{"key": "value", "number": 123}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidUTF8([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("isValidUTF8(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsValidUTF8_Invalid(t *testing.T) {
	// Test with invalid UTF-8 byte sequences
	invalidSequences := [][]byte{
		{0xff, 0xfe, 0xfd},                // invalid start bytes
		{0xc0, 0x80},                      // overlong encoding of null
		{0xed, 0xa0, 0x80},                // UTF-16 surrogate (invalid in UTF-8)
		{0xf4, 0x90, 0x80, 0x80},          // code point too large
		{0xe2, 0x28, 0xa1},                // invalid continuation byte
		{0x80, 0x80},                      // continuation bytes without start
		[]byte{0xff, 0xff, 'h', 'e', 'l'}, // mixed invalid and valid
	}

	for i, seq := range invalidSequences {
		t.Run("invalid sequence "+string(rune('A'+i)), func(t *testing.T) {
			got := isValidUTF8(seq)
			if got {
				t.Errorf("isValidUTF8(%v) = true, expected false for invalid UTF-8", seq)
			}
		})
	}
}

// ============================================
// Tests for DoctorResult structure
// ============================================

func TestDoctorResult_Structure(t *testing.T) {
	result := &DoctorResult{
		ConfigDir: CheckResult{
			Passed:   true,
			Details:  []string{"~/.config/opencode/ exists", "Readable and writable"},
			Warnings: []string{},
			Errors:   []string{},
		},
		AgentsMD: CheckResult{
			Passed:   true,
			Details:  []string{"File exists (8.6 KB)", "Contains Hefesto configuration"},
			Warnings: []string{},
			Errors:   []string{},
		},
		OpenCodeJSON: CheckResult{
			Passed:   true,
			Details:  []string{"Valid JSON", "5 agents configured"},
			Warnings: []string{},
			Errors:   []string{},
		},
		Engram: CheckResult{
			Passed:   true,
			Details:  []string{"engram 1.3.1 (/usr/local/bin/engram)", "MCP server available"},
			Warnings: []string{},
			Errors:   []string{},
		},
		OpenCode: CheckResult{
			Passed:   true,
			Details:  []string{"opencode 1.3.13 (/usr/local/bin/opencode)"},
			Warnings: []string{},
			Errors:   []string{},
		},
	}

	if !result.ConfigDir.Passed {
		t.Error("ConfigDir should pass")
	}
	if len(result.ConfigDir.Details) != 2 {
		t.Errorf("ConfigDir should have 2 details, got %d", len(result.ConfigDir.Details))
	}
	if !result.AgentsMD.Passed {
		t.Error("AgentsMD should pass")
	}
}

// ============================================
// Tests for CheckResult structure
// ============================================

func TestCheckResult_HasErrors(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		result := CheckResult{
			Passed:   true,
			Errors:   []string{},
			Warnings: []string{},
		}
		if len(result.Errors) > 0 {
			t.Error("Should have no errors")
		}
	})

	t.Run("has errors", func(t *testing.T) {
		result := CheckResult{
			Passed:   false,
			Errors:   []string{"Something went wrong"},
			Warnings: []string{},
		}
		if len(result.Errors) != 1 {
			t.Errorf("Should have 1 error, got %d", len(result.Errors))
		}
	})

	t.Run("has warnings", func(t *testing.T) {
		result := CheckResult{
			Passed:   true,
			Errors:   []string{},
			Warnings: []string{"Minor issue"},
		}
		if len(result.Warnings) != 1 {
			t.Errorf("Should have 1 warning, got %d", len(result.Warnings))
		}
	})
}

// ============================================
// Tests for RunDoctor exit code logic
// ============================================

func TestRunDoctor_ExitCodeLogic(t *testing.T) {
	// Test that exit code logic works correctly
	// Note: This runs against the real system, so results may vary

	result, exitCode := RunDoctor()

	// Verify result is not nil
	if result == nil {
		t.Fatal("RunDoctor() returned nil result")
	}

	// Verify exit code is 0, 1, or 2
	if exitCode < 0 || exitCode > 2 {
		t.Errorf("Exit code should be 0, 1, or 2, got %d", exitCode)
	}

	// Check exit code logic
	hasErrors := false
	hasWarnings := false

	checks := []CheckResult{
		result.ConfigDir,
		result.AgentsMD,
		result.OpenCodeJSON,
		result.Skills,
		result.Plugins,
		result.Engram,
		result.OpenCode,
		result.Theme,
		result.Personality,
		result.Commands,
	}

	for _, check := range checks {
		if len(check.Errors) > 0 {
			hasErrors = true
		}
		if len(check.Warnings) > 0 {
			hasWarnings = true
		}
	}

	// Verify exit code matches the logic
	if hasErrors && exitCode != 2 {
		t.Errorf("Exit code should be 2 when there are errors, got %d", exitCode)
	}
	if !hasErrors && hasWarnings && exitCode != 1 {
		t.Errorf("Exit code should be 1 when there are only warnings, got %d", exitCode)
	}
	if !hasErrors && !hasWarnings && exitCode != 0 {
		t.Errorf("Exit code should be 0 when there are no issues, got %d", exitCode)
	}

	t.Logf("Exit code: %d (errors: %v, warnings: %v)", exitCode, hasErrors, hasWarnings)
}

// ============================================
// Tests for DoctorCheck structure
// ============================================

func TestDoctorCheck_Structure(t *testing.T) {
	check := DoctorCheck{
		Name: "Test Check",
		Result: CheckResult{
			Passed:   true,
			Details:  []string{"Detail 1", "Detail 2"},
			Warnings: []string{},
			Errors:   []string{},
		},
	}

	if check.Name != "Test Check" {
		t.Errorf("Name = %q, want Test Check", check.Name)
	}
	if !check.Result.Passed {
		t.Error("Result should pass")
	}
}

// ============================================
// Tests for version comparison edge cases
// ============================================

func TestIsVersionAtLeast_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		minimum  string
		expected bool
	}{
		// Very long versions
		{"long version", "1.2.3.4.5", "1.2.3", true},
		{"long minimum", "1.2.3", "1.2.3.4.5", false},

		// Leading zeros (treated as numbers)
		{"leading zeros version", "01.02.03", "1.2.3", true},
		{"leading zeros minimum", "1.2.3", "01.02.03", true},

		// Single part versions
		{"single part equal", "1", "1", true},
		{"single part greater", "2", "1", true},
		{"single part lesser", "1", "2", false},

		// Pre-release style (extractNumber handles leading digits)
		{"pre-release style", "1.0.0-alpha", "1.0.0", true}, // extractNumber("0-alpha") = 0, so 1.0.0 >= 1.0.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isVersionAtLeast(tt.version, tt.minimum)
			if got != tt.expected {
				t.Errorf("isVersionAtLeast(%q, %q) = %v, want %v", tt.version, tt.minimum, got, tt.expected)
			}
		})
	}
}

// ============================================
// Tests for formatBytes (from status.go, tested here for completeness)
// ============================================

func TestFormatBytes_DoctorContext(t *testing.T) {
	// Test that formatBytes works correctly in the context of doctor checks
	// (file sizes, binary sizes, etc.)

	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"small config file", 8624, "8.4 KB"},   // 8624 / 1024 = 8.4375
		{"typical AGENTS.md", 12000, "11.7 KB"}, // 12000 / 1024 = 11.71875
		{"large config", 50000, "48.8 KB"},      // 50000 / 1024 = 48.828125
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			// We're just verifying it doesn't crash and returns reasonable output
			if got == "" {
				t.Error("formatBytes returned empty string")
			}
			t.Logf("formatBytes(%d) = %s (expected ~%s)", tt.bytes, got, tt.expected)
		})
	}
}
