package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Edcko/Hefesto/cmd/hefesto/internal/install"
	"github.com/Edcko/Hefesto/cmd/hefesto/internal/tui"
	"github.com/spf13/cobra"
)

// Build-time version info (can be set via ldflags)
var (
	version = "v0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "hefesto",
	Short: "Hefesto - Configuration installer for OpenCode",
	Long: `Hefesto is a TUI-based installer that embeds and deploys
OpenCode configuration files to your system.

It includes skills, themes, commands, and personality configurations
for an enhanced AI-assisted development experience.`,
	SilenceUsage: true,
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Hefesto configuration files",
	Long: `Launch the interactive TUI installer to deploy
Hefesto configuration files to your system.

The installer will:
- Detect your OpenCode configuration directory
- Create backups of existing configurations
- Deploy embedded configuration files
- Set up skills, themes, and commands`,
	RunE: runInstall,
}

var (
	installYes     bool
	installDryRun  bool
	rollbackYes    bool
	rollbackList   bool
	uninstallYes   bool
	uninstallPurge bool
	updateYes      bool
	updateDryRun   bool
)

func runInstall(cmd *cobra.Command, args []string) error {
	if installDryRun {
		fmt.Println("🔍 Dry run mode - no changes will be made")
		fmt.Println()
		// Run the installer in dry-run mode
		installer := install.NewInstaller(true)
		return runInstallerWithProgress(installer)
	}

	if installYes {
		fmt.Println("🚀 Non-interactive installation mode")
		fmt.Println()
		installer := install.NewInstaller(false)
		return runInstallerWithProgress(installer)
	}

	// Launch interactive TUI
	return tui.Run()
}

// runInstallerWithProgress runs the installer and prints progress to stdout.
func runInstallerWithProgress(installer *install.Installer) error {
	// Start installation in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- installer.Run()
	}()

	// Print progress updates
	for progress := range installer.Progress {
		if progress.Error != nil {
			fmt.Printf("❌ [%s] %s\n", progress.Step, progress.Message)
		} else if progress.Done {
			fmt.Printf("✅ [%s] %s\n", progress.Step, progress.Message)
		} else {
			fmt.Printf("⏳ [%s] %s\n", progress.Step, progress.Message)
		}
	}

	// Wait for completion
	if err := <-errChan; err != nil {
		fmt.Println()
		fmt.Printf("❌ Installation failed: %v\n", err)
		return err
	}

	fmt.Println()
	fmt.Println("🎉 Hefesto configuration installed successfully!")
	fmt.Println()
	return nil
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove Hefesto configuration files",
	Long: `Remove Hefesto configuration files from your system.

By default, the most recent backup will be restored.
Use --purge to delete everything without restoring a backup.`,
	RunE: runUninstall,
}

func runUninstall(cmd *cobra.Command, args []string) error {
	return install.RunUninstall(uninstallPurge, uninstallYes)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Hefesto to the latest version",
	Long: `Update Hefesto configuration files to the latest version.

This will:
- Pull the latest configuration files
- Create a backup of current configurations
- Apply the new configurations`,
	RunE: runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	return install.RunUpdate(updateDryRun, updateYes)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installation status",
	Long: `Display the current installation status of Hefesto.

Shows:
- Installation directory
- Installed version
- Available skills
- Configuration status`,
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	status, err := install.CheckStatus()
	if err != nil {
		return fmt.Errorf("failed to check status: %w", err)
	}

	install.PrintStatus(status)
	return nil
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Restore a previous backup of OpenCode configuration",
	Long: `Restore a previous backup of OpenCode configuration.

This will:
- List available backups
- Allow you to select which backup to restore
- Create a safety backup before restoring`,
	RunE: runRollback,
}

func runRollback(cmd *cobra.Command, args []string) error {
	backups, err := install.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	// Just list backups
	if rollbackList {
		install.PrintBackups(backups)
		return nil
	}

	// No backups found
	if len(backups) == 0 {
		install.PrintBackups(backups)
		return nil
	}

	// Select backup to restore (most recent by default)
	selectedBackup := install.PromptRollback(backups)
	if selectedBackup == nil {
		return fmt.Errorf("no backup selected")
	}

	// Non-interactive mode with --yes flag
	if rollbackYes {
		safetyBackup, err := install.Rollback(selectedBackup.Path)
		if err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}
		install.PrintRollbackResult(*selectedBackup, safetyBackup)
		return nil
	}

	// Interactive mode - confirm before rollback
	fmt.Println()
	fmt.Println("🔥 Hefesto Rollback")
	fmt.Println()
	fmt.Printf("  Most recent backup: %s (%s)\n", selectedBackup.Name, install.FormatBackupDate(selectedBackup.Timestamp))
	fmt.Println()
	fmt.Print("  Restore this backup? [y/N]: ")

	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println()
		fmt.Println("  ❌ Rollback cancelled.")
		fmt.Println()
		return nil
	}

	safetyBackup, err := install.Rollback(selectedBackup.Path)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}
	install.PrintRollbackResult(*selectedBackup, safetyBackup)
	return nil
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display the version, commit, and build date for Hefesto.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Hefesto %s\n", version)
		fmt.Printf("  Commit: %s\n", commit)
		fmt.Printf("  Built:  %s\n", date)
	},
}

func init() {
	// Install command flags
	installCmd.Flags().BoolVarP(&installYes, "yes", "y", false, "non-interactive mode, accept all defaults")
	installCmd.Flags().BoolVarP(&installDryRun, "dry-run", "d", false, "show what would happen without making changes")

	// Rollback command flags
	rollbackCmd.Flags().BoolVarP(&rollbackYes, "yes", "y", false, "Restore most recent backup without prompting")
	rollbackCmd.Flags().BoolVar(&rollbackList, "list", false, "List available backups")

	// Uninstall command flags
	uninstallCmd.Flags().BoolVarP(&uninstallYes, "yes", "y", false, "Skip confirmation")
	uninstallCmd.Flags().BoolVar(&uninstallPurge, "purge", false, "Delete everything without restoring backup")

	// Update command flags
	updateCmd.Flags().BoolVarP(&updateYes, "yes", "y", false, "Skip confirmation")
	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "Show what would change")

	// Add commands to root
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(versionCmd)
}
