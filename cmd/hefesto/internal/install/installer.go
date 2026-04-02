// Package install provides installation logic for Hefesto TUI installer.
package install

import (
	"fmt"
	"os"

	embedpkg "github.com/Edcko/Hefesto/cmd/hefesto/internal/embed"
)

// Step represents a step in the installation process.
type Step string

const (
	StepDetect Step = "detect"
	StepBackup Step = "backup"
	StepCopy   Step = "copy"
	StepNpm    Step = "npm"
	StepVerify Step = "verify"
)

// InstallProgress represents progress updates during installation.
type InstallProgress struct {
	Step    Step
	Message string
	Done    bool
	Error   error
}

// Installer orchestrates the Hefesto installation process.
type Installer struct {
	env        *Environment
	configPath string
	backupPath string
	dryRun     bool
	Progress   chan InstallProgress
}

// NewInstaller creates a new Installer instance.
func NewInstaller(dryRun bool) *Installer {
	return &Installer{
		dryRun:   dryRun,
		Progress: make(chan InstallProgress, 10),
	}
}

// Run executes the full installation process.
// It sends progress updates through the Progress channel.
func (i *Installer) Run() error {
	defer close(i.Progress)

	// Step 1: Detect environment
	i.Progress <- InstallProgress{
		Step:    StepDetect,
		Message: "Detecting environment...",
		Done:    false,
	}

	env, err := Detect()
	if err != nil {
		i.Progress <- InstallProgress{
			Step:    StepDetect,
			Message: fmt.Sprintf("Detection failed: %v", err),
			Done:    true,
			Error:   err,
		}
		return fmt.Errorf("detection failed: %w", err)
	}
	i.env = env
	i.configPath = env.ConfigPath

	i.Progress <- InstallProgress{
		Step:    StepDetect,
		Message: fmt.Sprintf("Detected: %s/%s, OpenCode: %s", env.Platform, env.Arch, env.OpenCodeVersion),
		Done:    true,
	}

	// Step 2: Backup existing config if present
	if env.ConfigExists && env.ExistingConfig != "none" {
		i.Progress <- InstallProgress{
			Step:    StepBackup,
			Message: fmt.Sprintf("Backing up existing %s config...", env.ExistingConfig),
			Done:    false,
		}

		if !i.dryRun {
			backupPath, err := Backup(i.configPath)
			if err != nil {
				i.Progress <- InstallProgress{
					Step:    StepBackup,
					Message: fmt.Sprintf("Backup failed: %v", err),
					Done:    true,
					Error:   err,
				}
				return fmt.Errorf("backup failed: %w", err)
			}
			i.backupPath = backupPath
		}

		i.Progress <- InstallProgress{
			Step:    StepBackup,
			Message: fmt.Sprintf("Backup created: %s", i.backupPath),
			Done:    true,
		}
	}

	// Step 3: Copy embedded config
	i.Progress <- InstallProgress{
		Step:    StepCopy,
		Message: "Copying Hefesto configuration...",
		Done:    false,
	}

	if !i.dryRun {
		if err := CopyConfig(embedpkg.ConfigFiles, i.configPath); err != nil {
			i.Progress <- InstallProgress{
				Step:    StepCopy,
				Message: fmt.Sprintf("Copy failed: %v", err),
				Done:    true,
				Error:   err,
			}
			return fmt.Errorf("failed to copy config: %w", err)
		}
	}

	i.Progress <- InstallProgress{
		Step:    StepCopy,
		Message: "Configuration copied successfully",
		Done:    true,
	}

	// Step 4: Run npm install
	i.Progress <- InstallProgress{
		Step:    StepNpm,
		Message: "Running npm install...",
		Done:    false,
	}

	if !i.dryRun {
		if err := NpmInstall(i.configPath); err != nil {
			// npm install failure is not fatal - config works without npm deps
			i.Progress <- InstallProgress{
				Step:    StepNpm,
				Message: fmt.Sprintf("npm install skipped or failed (non-fatal): %v", err),
				Done:    true,
				// Don't set Error here - this is not a fatal error
			}
		} else {
			i.Progress <- InstallProgress{
				Step:    StepNpm,
				Message: "npm install completed successfully",
				Done:    true,
			}
		}
	} else {
		i.Progress <- InstallProgress{
			Step:    StepNpm,
			Message: "npm install skipped (dry run)",
			Done:    true,
		}
	}

	// Step 5: Verify installation
	i.Progress <- InstallProgress{
		Step:    StepVerify,
		Message: "Verifying installation...",
		Done:    false,
	}

	if !i.dryRun {
		result, err := Verify(i.configPath)
		if err != nil {
			i.Progress <- InstallProgress{
				Step:    StepVerify,
				Message: fmt.Sprintf("Verification error: %v", err),
				Done:    true,
				Error:   err,
			}
			return fmt.Errorf("verification failed: %w", err)
		}

		// Check critical verifications
		if !result.ConfigCopied {
			err := fmt.Errorf("config verification failed")
			i.Progress <- InstallProgress{
				Step:    StepVerify,
				Message: "Config verification failed",
				Done:    true,
				Error:   err,
			}
			return err
		}

		i.Progress <- InstallProgress{
			Step: StepVerify,
			Message: fmt.Sprintf("Verification complete (Config: %v, NPM: %v, OpenCode: %v)",
				result.ConfigCopied, result.NpmInstalled, result.OpenCodeWorks),
			Done: true,
		}
	} else {
		i.Progress <- InstallProgress{
			Step:    StepVerify,
			Message: "Verification skipped (dry run)",
			Done:    true,
		}
	}

	return nil
}

// Uninstall removes the Hefesto configuration.
func (i *Installer) Uninstall() error {
	if i.configPath == "" {
		env, err := Detect()
		if err != nil {
			return fmt.Errorf("failed to detect environment: %w", err)
		}
		i.configPath = env.ConfigPath
	}

	// Check if config directory exists on disk
	if _, err := os.Stat(i.configPath); os.IsNotExist(err) {
		return fmt.Errorf("no config found to uninstall at %s", i.configPath)
	}

	// Remove the config directory
	if err := os.RemoveAll(i.configPath); err != nil {
		return fmt.Errorf("failed to remove config directory: %w", err)
	}

	return nil
}

// Status returns the current installation status.
func (i *Installer) Status() (*Environment, error) {
	if i.env == nil {
		env, err := Detect()
		if err != nil {
			return nil, err
		}
		i.env = env
	}
	return i.env, nil
}
