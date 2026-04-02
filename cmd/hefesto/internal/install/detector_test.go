package install

import "testing"

func TestDetect(t *testing.T) {
	env, err := Detect()
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if env.Platform != "darwin" && env.Platform != "linux" {
		t.Errorf("Platform = %q, want darwin or linux", env.Platform)
	}
	if env.Arch != "arm64" && env.Arch != "amd64" {
		t.Errorf("Arch = %q, want arm64 or amd64", env.Arch)
	}
}

func TestDetectOpenCode(t *testing.T) {
	env, err := Detect()
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	// OpenCode should be installed on this machine
	if !env.OpenCodeInstalled {
		t.Log("Warning: OpenCode not installed (expected on dev machine)")
	}
}

func TestDetectConfigPath(t *testing.T) {
	env, err := Detect()
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if env.ConfigPath == "" {
		t.Error("ConfigPath is empty, expected a valid path")
	}
}

func TestDetectAppleSilicon(t *testing.T) {
	env, err := Detect()
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	// On darwin/arm64, IsAppleSilicon should be true
	if env.Platform == "darwin" && env.Arch == "arm64" {
		if !env.IsAppleSilicon {
			t.Error("IsAppleSilicon = false on darwin/arm64, want true")
		}
	}
}
