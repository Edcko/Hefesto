package main

import (
	"fmt"
	"os"

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
	installYes    bool
	installDryRun bool
)

func runInstall(cmd *cobra.Command, args []string) error {
	if installDryRun {
		fmt.Println("🔍 Dry run mode - no changes will be made")
		fmt.Println()
		// TODO: Show what would be installed
		return nil
	}

	if installYes {
		fmt.Println("🚀 Non-interactive installation mode")
		// TODO: Implement non-interactive installation
		return nil
	}

	// Launch interactive TUI
	return tui.Run()
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove Hefesto configuration files",
	Long: `Remove Hefesto configuration files from your system.

If a backup exists, it will be restored. Otherwise, the
configuration files will be removed.`,
	RunE: runUninstall,
}

func runUninstall(cmd *cobra.Command, args []string) error {
	fmt.Println("🗑️  Uninstalling Hefesto configuration...")
	// TODO: Implement uninstall logic
	fmt.Println("Uninstall not yet implemented")
	return nil
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
	fmt.Println("🔄 Updating Hefesto configuration...")
	// TODO: Implement update logic
	fmt.Println("Update not yet implemented")
	return nil
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
	fmt.Println("📊 Hefesto Installation Status")
	fmt.Println()
	fmt.Printf("  Version:    %s\n", version)
	fmt.Printf("  Commit:     %s\n", commit)
	fmt.Printf("  Built:      %s\n", date)
	fmt.Println()
	// TODO: Check actual installation status
	fmt.Println("  Status:     Not installed")
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

	// Add commands to root
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(versionCmd)
}
