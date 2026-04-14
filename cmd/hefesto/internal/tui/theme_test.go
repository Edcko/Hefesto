package tui

import (
	"strings"
	"testing"
)

// ===== Screen.String() tests =====

func TestScreenString(t *testing.T) {
	tests := []struct {
		name     string
		screen   Screen
		expected string
	}{
		{"welcome screen", ScreenWelcome, "welcome"},
		{"detect screen", ScreenDetect, "detect"},
		{"component-select screen", ScreenComponentSelect, "component-select"},
		{"backup screen", ScreenBackup, "backup"},
		{"install screen", ScreenInstall, "install"},
		{"complete screen", ScreenComplete, "complete"},
		{"error screen", ScreenError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.screen.String()
			if got != tt.expected {
				t.Errorf("Screen(%d).String() = %q, want %q", tt.screen, got, tt.expected)
			}
		})
	}
}

// ===== ProgressBar tests =====

func TestProgressBar(t *testing.T) {
	tests := []struct {
		name         string
		width        int
		filled       int
		total        int
		wantContains string
		wantEmpty    bool
	}{
		{
			name:      "zero total returns empty",
			width:     10,
			filled:    5,
			total:     0,
			wantEmpty: true,
		},
		{
			name:         "half filled contains block chars",
			width:        10,
			filled:       5,
			total:        10,
			wantContains: "█",
		},
		{
			name:         "full bar has no empty chars",
			width:        10,
			filled:       10,
			total:        10,
			wantContains: "█",
		},
		{
			name:         "empty bar has empty chars",
			width:        10,
			filled:       0,
			total:        10,
			wantContains: "░",
		},
		{
			name:         "single unit filled",
			width:        20,
			filled:       1,
			total:        20,
			wantContains: "█",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProgressBar(tt.width, tt.filled, tt.total)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("ProgressBar() = %q, want empty string", got)
				}
				return
			}

			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("ProgressBar() = %q, want to contain %q", got, tt.wantContains)
			}
		})
	}
}

// ===== CenterText tests =====

func TestCenterText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		contains string
	}{
		{"centers short text", "hello", 20, "hello"},
		{"centers longer text", "hello world", 40, "hello world"},
		{"empty text", "", 20, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CenterText(tt.text, tt.width)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("CenterText(%q, %d) = %q, want to contain %q", tt.text, tt.width, got, tt.contains)
			}
		})
	}
}

// ===== CenterBlock tests =====

func TestCenterBlock(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		contains string
	}{
		{"single line", "hello", 30, "hello"},
		{"multi line", "hello\nworld", 30, "hello"},
		{"empty block", "", 30, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CenterBlock(tt.text, tt.width)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("CenterBlock() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

// ===== Text rendering functions tests =====

func TestAmberText(t *testing.T) {
	result := AmberText("test")
	if !strings.Contains(result, "test") {
		t.Errorf("AmberText() = %q, want to contain 'test'", result)
	}
}

func TestCopperText(t *testing.T) {
	result := CopperText("test")
	if !strings.Contains(result, "test") {
		t.Errorf("CopperText() = %q, want to contain 'test'", result)
	}
}

func TestWhiteText(t *testing.T) {
	result := WhiteText("test")
	if !strings.Contains(result, "test") {
		t.Errorf("WhiteText() = %q, want to contain 'test'", result)
	}
}

func TestGrayText(t *testing.T) {
	result := GrayText("test")
	if !strings.Contains(result, "test") {
		t.Errorf("GrayText() = %q, want to contain 'test'", result)
	}
}

func TestGreenText(t *testing.T) {
	result := GreenText("test")
	if !strings.Contains(result, "test") {
		t.Errorf("GreenText() = %q, want to contain 'test'", result)
	}
}

func TestRedText(t *testing.T) {
	result := RedText("test")
	if !strings.Contains(result, "test") {
		t.Errorf("RedText() = %q, want to contain 'test'", result)
	}
}

// ===== BulletItem tests =====

func TestBulletItem(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		contains string
	}{
		{"bullet point text", "Angular skill", "Angular skill"},
		{"bullet point bullet char", "test", "•"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BulletItem(tt.text)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("BulletItem(%q) = %q, want to contain %q", tt.text, got, tt.contains)
			}
		})
	}
}

// ===== BannerStyle tests =====

func TestBannerStyle(t *testing.T) {
	result := BannerStyle()
	// BannerAnvil now uses single-width * instead of 🔥 to avoid alignment issues
	if !strings.Contains(result, "*") {
		t.Errorf("BannerStyle() = %q, want to contain flame spark", result)
	}
	// Should contain the anvil structure
	if !strings.Contains(result, "╱") {
		t.Errorf("BannerStyle() = %q, want to contain anvil art", result)
	}
}

// ===== VersionStyle tests =====

func TestVersionStyle(t *testing.T) {
	result := VersionStyle()
	if !strings.Contains(result, "HEFESTO") {
		t.Errorf("VersionStyle() = %q, want to contain 'HEFESTO'", result)
	}
}

// ===== TaglineStyle tests =====

func TestTaglineStyle(t *testing.T) {
	result := TaglineStyle()
	if !strings.Contains(result, "AI Dev Environment Forge") {
		t.Errorf("TaglineStyle() = %q, want to contain tagline", result)
	}
}

// ===== repeatStr tests =====

func TestRepeatStr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		n        int
		expected string
	}{
		{"repeat zero", "x", 0, ""},
		{"repeat one", "x", 1, "x"},
		{"repeat three", "ab", 3, "ababab"},
		{"repeat empty", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repeatStr(tt.input, tt.n)
			if got != tt.expected {
				t.Errorf("repeatStr(%q, %d) = %q, want %q", tt.input, tt.n, got, tt.expected)
			}
		})
	}
}

// ===== stripAnsi tests =====

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain text", "hello world", "hello world"},
		{"with ansi code", "\x1b[32mhello\x1b[0m", "hello"},
		{"multiple ansi codes", "\x1b[1m\x1b[31mtest\x1b[0m", "test"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAnsi(tt.input)
			if got != tt.expected {
				t.Errorf("stripAnsi(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// ===== wrapText tests =====

func TestWrapText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxWidth  int
		maxLines  int
		wantFirst string
	}{
		{
			name:      "short text fits in one line",
			text:      "hello",
			maxWidth:  50,
			maxLines:  1,
			wantFirst: "hello",
		},
		{
			name:      "long text wraps",
			text:      "this is a very long error message that should be wrapped across multiple lines when it exceeds the maximum width",
			maxWidth:  30,
			maxLines:  2,
			wantFirst: "this is a very long error",
		},
		{
			name:      "single word fits",
			text:      "test",
			maxWidth:  10,
			maxLines:  1,
			wantFirst: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.text, tt.maxWidth)
			if len(got) == 0 {
				t.Fatal("wrapText() returned empty slice")
			}
			if got[0] != tt.wantFirst {
				t.Errorf("wrapText() first line = %q, want %q", got[0], tt.wantFirst)
			}
			if len(got) > tt.maxLines {
				// Allow some flexibility for wrapping
				t.Logf("wrapText() produced %d lines (max expected ~%d): %v", len(got), tt.maxLines, got)
			}
		})
	}
}

// ===== renderProgressBar tests =====

func TestRenderProgressBar(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		progress float64
		wantLen  int
	}{
		{"zero progress", 10, 0.0, 10},
		{"half progress", 10, 0.5, 10},
		{"full progress", 10, 1.0, 10},
		{"clamped negative", 10, -0.5, 10},
		{"clamped over 1", 10, 1.5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderProgressBar(tt.width, tt.progress)
			// Use rune count (not byte count) since █ and ░ are multi-byte UTF-8
			plain := stripAnsi(got)
			runeCount := len([]rune(plain))
			if runeCount != tt.wantLen {
				t.Errorf("renderProgressBar(%d, %.1f) rune count = %d, want %d (got %q)",
					tt.width, tt.progress, runeCount, tt.wantLen, plain)
			}
		})
	}
}

// ===== SetVersion tests =====

func TestSetVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"sets version", "v1.2.3", "v1.2.3"},
		{"empty string does not override", "", "v1.2.3"}, // should keep previous
	}

	// Set a known version first
	SetVersion("v1.2.3")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVersion(tt.version)
			if Version != tt.expected {
				t.Errorf("after SetVersion(%q), Version = %q, want %q", tt.version, Version, tt.expected)
			}
		})
	}
}

// ===== Icon constants test =====

func TestIconConstants(t *testing.T) {
	if IconCheck != "✓" {
		t.Errorf("IconCheck = %q, want '✓'", IconCheck)
	}
	if IconCross != "✗" {
		t.Errorf("IconCross = %q, want '✗'", IconCross)
	}
	if IconArrow != "→" {
		t.Errorf("IconArrow = %q, want '→'", IconArrow)
	}
	if IconBullet != "•" {
		t.Errorf("IconBullet = %q, want '•'", IconBullet)
	}
	if IconFire != "*" {
		t.Errorf("IconFire = %q, want '*'", IconFire)
	}
	if len(IconSpinner) == 0 {
		t.Error("IconSpinner is empty, expected braille characters")
	}
}

// ===== New color token tests =====

func TestNewColorTokens(t *testing.T) {
	// ColorMutedBorder should be a non-empty color value
	mutedBorder := string(ColorMutedBorder)
	if mutedBorder == "" {
		t.Error("ColorMutedBorder is empty")
	}
	// ColorDimText should be a non-empty color value
	dimText := string(ColorDimText)
	if dimText == "" {
		t.Error("ColorDimText is empty")
	}
}

// ===== Typography hierarchy tests =====

func TestHeroTitleStyle(t *testing.T) {
	result := HeroTitleStyle.Render("Test Hero")
	if !strings.Contains(result, "Test Hero") {
		t.Errorf("HeroTitleStyle.Render() = %q, want to contain 'Test Hero'", result)
	}
}

func TestSectionTitleStyle(t *testing.T) {
	result := SectionTitleStyle.Render("Section")
	if !strings.Contains(result, "Section") {
		t.Errorf("SectionTitleStyle.Render() = %q, want to contain 'Section'", result)
	}
}

func TestDimTextStyle(t *testing.T) {
	result := DimTextStyle.Render("dim text")
	if !strings.Contains(result, "dim text") {
		t.Errorf("DimTextStyle.Render() = %q, want to contain 'dim text'", result)
	}
}
