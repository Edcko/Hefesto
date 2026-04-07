package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectWithTempHome(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	
	// First check - what does os.UserHomeDir() return?
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() failed: %v", err)
	}
	t.Logf("Step 1: os.UserHomeDir() = %s", homeDir)
	t.Logf("Step 1: Expected = %s", tmpHome)
	
	if homeDir != tmpHome {
		t.Errorf("os.UserHomeDir() returned %s but expected %s", homeDir, tmpHome)
	}
	
	// Now call Detect()
	env, err := Detect()
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}
	
	t.Logf("Step 2: env.ConfigPath = %s", env.ConfigPath)
	t.Logf("Step 2: Expected = %s", filepath.Join(tmpHome, ".config", "opencode"))
	
	// Check what Detect() found
	if env.ConfigPath != filepath.Join(tmpHome, ".config", "opencode") {
		t.Errorf("Detect() found config at %s but expected %s", env.ConfigPath, filepath.Join(tmpHome, ".config", "opencode"))
	}
	
	// Check if config exists
	t.Logf("Step 3: env.ConfigExists = %v", env.ConfigExists)
	if env.ConfigExists {
		// Let's check if it file really exists
		if _, err := os.Stat(env.ConfigPath); err == nil {
			t.Logf("Step 3a: File DOES exist at %s", env.ConfigPath)
		} else {
			t.Logf("Step 3b: File stat error: %v", err)
		}
	}
}
