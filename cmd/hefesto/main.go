package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	embedpkg "github.com/Edcko/Hefesto/cmd/hefesto/internal/embed"
	"github.com/Edcko/Hefesto/cmd/hefesto/internal/install"
	"github.com/Edcko/Hefesto/cmd/hefesto/internal/logger"
	"github.com/Edcko/Hefesto/cmd/hefesto/internal/tui"
	"github.com/spf13/cobra"
)

// Build-time version info (can be set via ldflags)
var (
	version = "v0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

// doctorExitError carries the exit code from the doctor command
// without calling os.Exit, so deferred cleanup runs normally.
type doctorExitError struct {
	code int
}

func (e *doctorExitError) Error() string {
	return fmt.Sprintf("doctor exit code: %d", e.code)
}

func main() {
	// Logger init is handled in rootCmd.PersistentPreRunE (after flag parsing).
	defer logger.Close()

	if err := rootCmd.Execute(); err != nil {
		// If it's a doctor exit, respect its exit code for the process.
		if docErr, ok := err.(*doctorExitError); ok {
			os.Exit(docErr.code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "hefesto",
	Short:   "Hefesto - Configuration installer for OpenCode",
	Version: version,
	Long: `Hefesto is a TUI-based installer that embeds and deploys
OpenCode configuration files to your system.

It includes skills, themes, commands, and personality configurations
for an enhanced AI-assisted development experience.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger.Init(verbose)
		return nil
	},
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
	verbose        bool // global --verbose flag
	installYes     bool
	installDryRun  bool
	installTest    bool
	rollbackYes    bool
	rollbackList   bool
	uninstallYes   bool
	uninstallPurge bool
	updateYes      bool
	updateDryRun   bool
	statusVerbose  bool
	statusJSON     bool
	doctorJSON     bool
	listJSON       bool
)

func runInstall(cmd *cobra.Command, args []string) error {
	// --dry-run: print summary of what would be installed, no changes made
	if installDryRun {
		return runDryRun()
	}

	// --test mode: install into a temp directory
	if installTest {
		tmpDir, err := os.MkdirTemp("", "hefesto-test-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		if err := os.Setenv("HOME", tmpDir); err != nil {
			return fmt.Errorf("failed to set HOME: %w", err)
		}
		fmt.Fprintf(os.Stderr, "🧪 Test mode: installing to %s\n", tmpDir)
		defer fmt.Fprintf(os.Stderr, "🧪 Test install complete. Files at: %s\n", tmpDir)
	}

	if installYes || installTest {
		fmt.Println("🚀 Non-interactive installation mode")
		fmt.Println()
		installer := install.NewInstaller(false)
		return runInstallerWithProgress(installer)
	}

	// Launch interactive TUI
	tui.SetVersion(version)
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

// dryRunFileInfo holds metadata about a file that would be installed.
type dryRunFileInfo struct {
	name string
	size int64
}

// dryRunCategory groups files by category for display.
type dryRunCategory struct {
	label string
	icon  string
	items []dryRunFileInfo
}

// runDryRun prints a detailed summary of what would be installed without making changes.
func runDryRun() error {
	// Detect environment to get the config path
	env, err := install.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect environment: %w", err)
	}

	// Resolve config path for display
	configDisplayPath := formatPathRel(env.ConfigPath)

	// Walk the embedded filesystem to collect file info
	configFS, err := fs.Sub(embedpkg.ConfigFiles, "config")
	if err != nil {
		return fmt.Errorf("failed to access embedded config: %w", err)
	}

	var categories []dryRunCategory
	var totalSize int64

	// Collect root config files (AGENTS.md, opencode.json, etc.)
	var rootFiles []dryRunFileInfo
	rootEntries, _ := fs.ReadDir(configFS, ".")
	if rootEntries != nil {
		for _, entry := range rootEntries {
			if entry.IsDir() {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			rootFiles = append(rootFiles, dryRunFileInfo{name: entry.Name(), size: info.Size()})
			totalSize += info.Size()
		}
		sort.Slice(rootFiles, func(i, j int) bool { return rootFiles[i].name < rootFiles[j].name })
	}

	// Collect skills
	skillsCat := collectCategory(configFS, "skills", true)
	totalSize += skillsCat.totalSize
	categories = append(categories, dryRunCategory{
		label: fmt.Sprintf("%d skills (skills/)", len(skillsCat.items)),
		icon:  "✅",
		items: skillsCat.items,
	})

	// Collect themes
	themesCat := collectCategory(configFS, "themes", false)
	totalSize += themesCat.totalSize
	categories = append(categories, dryRunCategory{
		label: fmt.Sprintf("%d theme (themes/)", len(themesCat.items)),
		icon:  "✅",
		items: themesCat.items,
	})

	// Collect plugins
	pluginsCat := collectCategory(configFS, "plugins", false)
	totalSize += pluginsCat.totalSize
	categories = append(categories, dryRunCategory{
		label: fmt.Sprintf("%d plugins (plugins/)", len(pluginsCat.items)),
		icon:  "✅",
		items: pluginsCat.items,
	})

	// Collect personality
	personalityCat := collectCategory(configFS, "personality", false)
	totalSize += personalityCat.totalSize
	categories = append(categories, dryRunCategory{
		label: fmt.Sprintf("%d personality (personality/)", len(personalityCat.items)),
		icon:  "✅",
		items: personalityCat.items,
	})

	// Collect commands
	commandsCat := collectCategory(configFS, "commands", false)
	totalSize += commandsCat.totalSize
	categories = append(categories, dryRunCategory{
		label: fmt.Sprintf("%d commands (commands/)", len(commandsCat.items)),
		icon:  "✅",
		items: commandsCat.items,
	})

	// Print the dry-run summary
	fmt.Println()
	fmt.Println("🔍 Dry Run — What would be installed:")
	fmt.Println()
	fmt.Printf("  Config Directory: %s\n", configDisplayPath)
	fmt.Println()
	fmt.Println("  Files to install:")
	for _, rf := range rootFiles {
		fmt.Printf("    ✅ %s (%s)\n", rf.name, formatSize(rf.size))
	}
	for _, cat := range categories {
		fmt.Printf("    %s %s\n", cat.icon, cat.label)
	}
	fmt.Println()
	fmt.Printf("  Total size estimate: %s\n", formatSize(totalSize))
	fmt.Println()
	fmt.Println("  Post-install:")

	// Check npm
	if _, npmErr := exec.LookPath("npm"); npmErr == nil {
		fmt.Println("    - npm install in plugins/")
	} else {
		fmt.Println("    - npm install in plugins/ (npm not found, would be skipped)")
	}

	// Engram status
	if env.EngramInstalled {
		fmt.Printf("    - Engram already installed (%s)\n", env.EngramVersion)
	} else {
		fmt.Println("    - Download engram binary (latest version)")
	}

	fmt.Println()

	// Backup info
	if env.ConfigExists && env.ExistingConfig != "none" {
		backupTimestamp := time.Now().Format("20060102-150405")
		backupDisplay := filepath.Join("~", ".config", fmt.Sprintf("opencode-backup-%s", backupTimestamp))
		fmt.Printf("  Backup: Would create backup of existing %s config at %s/\n", env.ExistingConfig, backupDisplay)
	} else {
		fmt.Println("  Backup: No existing config to back up")
	}

	fmt.Println()
	fmt.Println("  No changes were made. Run without --dry-run to install.")
	fmt.Println()
	return nil
}

// categoryResult holds the result of collecting files from a category directory.
type categoryResult struct {
	items     []dryRunFileInfo
	totalSize int64
}

// collectCategory walks a subdirectory of the embedded FS and collects file info.
// If countDirs is true, directories are counted as items instead of individual files.
func collectCategory(configFS fs.FS, subDir string, countDirs bool) categoryResult {
	var result categoryResult

	entries, err := fs.ReadDir(configFS, subDir)
	if err != nil {
		return result
	}

	if countDirs {
		// Count directories (like skills)
		for _, entry := range entries {
			if entry.IsDir() {
				// Walk the directory to get its total size
				dirSize := walkDirSize(configFS, subDir+"/"+entry.Name())
				result.items = append(result.items, dryRunFileInfo{
					name: entry.Name(),
					size: dirSize,
				})
				result.totalSize += dirSize
			}
		}
	} else {
		// Count individual files
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			result.items = append(result.items, dryRunFileInfo{
				name: entry.Name(),
				size: info.Size(),
			})
			result.totalSize += info.Size()
		}
	}

	sort.Slice(result.items, func(i, j int) bool { return result.items[i].name < result.items[j].name })
	return result
}

// walkDirSize recursively computes the total size of all files in a directory.
func walkDirSize(fsys fs.FS, dir string) int64 {
	var total int64
	_ = fs.WalkDir(fsys, dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, infoErr := d.Info()
			if infoErr == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
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
	Short: "Update Hefesto configuration to the latest version",
	Long: `Update your OpenCode configuration to the latest version bundled with Hefesto.

This creates a timestamped backup of your current configuration and overlays
the latest embedded config files. Your customizations are preserved where possible.

Note: This updates the configuration files, not the Hefesto binary itself.
Use 'brew upgrade hefesto' to update the binary.`,
	RunE: runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	return install.RunUpdate(updateDryRun, updateYes, version)
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
	if statusJSON {
		return install.PrintStatusJSON(status)
	}
	if statusVerbose {
		install.PrintStatusVerbose(status)
	} else {
		install.PrintStatus(status)
	}
	return nil
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostic checks on Hefesto installation",
	Long: `Run comprehensive diagnostic checks on your Hefesto installation.

Checks:
- Configuration directory structure
- AGENTS.md file validity
- opencode.json configuration
- Skills directory and structure
- Plugins directory
- Engram binary and configuration
- OpenCode binary
- Theme configuration
- Personality settings
- Custom commands`,
	RunE: runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	result, exitCode := install.RunDoctor()
	if doctorJSON {
		if err := install.PrintDoctorJSON(result); err != nil {
			return &doctorExitError{code: 1}
		}
		return &doctorExitError{code: exitCode}
	}
	install.PrintDoctor(result)
	return &doctorExitError{code: exitCode}
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
	// Check if stdin is a terminal (not piped/redirected)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		fmt.Println()
		fmt.Println("  ❌ Interactive terminal required for rollback confirmation.")
		fmt.Println("  Use --yes flag to skip confirmation in non-interactive environments.")
		fmt.Println()
		return fmt.Errorf("non-interactive terminal: use --yes to skip confirmation")
	}

	fmt.Println()
	fmt.Println("🔥 Hefesto Rollback")
	fmt.Println()
	fmt.Printf("  Most recent backup: %s (%s)\n", selectedBackup.Name, install.FormatBackupDate(selectedBackup.Timestamp))
	fmt.Println()
	fmt.Print("  Restore this backup? [y/N]: ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

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

// ============================================
// config command (Issue 18)
// ============================================

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View Hefesto configuration",
	Long: `View and inspect Hefesto configuration settings.

Subcommands:
  show — Display current config paths and key settings
  path — Print the config directory path (useful for scripting)`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  `Display the current Hefesto configuration including paths and installed state.`,
	RunE:  runConfigShow,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print config directory path",
	Long:  `Print the Hefesto configuration directory path. Useful for scripting.`,
	RunE:  runConfigPath,
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	env, err := install.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect environment: %w", err)
	}

	homeDir, err := getUserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	backupDir := filepath.Join(homeDir, ".config")

	fmt.Println()
	fmt.Println("🔥 Hefesto Configuration")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("  Config Dir:   %s\n", formatPathRel(env.ConfigPath))
	fmt.Printf("  Backup Dir:   %s\n", formatPathRel(backupDir))
	fmt.Printf("  Version:      %s\n", version)
	fmt.Println()

	if env.ConfigExists {
		fmt.Println("  Status:       ✅ Installed")
	} else {
		fmt.Println("  Status:       ❌ Not installed")
	}

	fmt.Println()

	if env.OpenCodeInstalled {
		fmt.Printf("  OpenCode:     ✅ %s (%s)\n", env.OpenCodeVersion, env.OpenCodePath)
	} else {
		fmt.Println("  OpenCode:     ❌ Not found")
	}

	if env.EngramInstalled {
		fmt.Printf("  Engram:       ✅ %s (%s)\n", env.EngramVersion, env.EngramPath)
	} else {
		fmt.Println("  Engram:       ❌ Not found")
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	return nil
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	env, err := install.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect environment: %w", err)
	}
	fmt.Println(env.ConfigPath)
	return nil
}

// ============================================
// list command (Issue 19)
// ============================================

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Hefesto resources",
	Long: `List available Hefesto resources.

Subcommands:
  skills  — List all embedded skills
  themes  — List available themes
  backups — List timestamped backups`,
}

var listSkillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "List available skills",
	Long:  `List all skills bundled with Hefesto from the embedded configuration.`,
	RunE:  runListSkills,
}

var listThemesCmd = &cobra.Command{
	Use:   "themes",
	Short: "List available themes",
	Long:  `List all themes bundled with Hefesto from the embedded configuration.`,
	RunE:  runListThemes,
}

var listBackupsCmd = &cobra.Command{
	Use:   "backups",
	Short: "List timestamped backups",
	Long:  `List all timestamped configuration backups with dates and sizes.`,
	RunE:  runListBackups,
}

func runListSkills(cmd *cobra.Command, args []string) error {
	skills, err := listEmbeddedSkills()
	if err != nil {
		return fmt.Errorf("failed to list skills: %w", err)
	}

	if listJSON {
		return printListJSON("skills", skills)
	}

	fmt.Println()
	fmt.Println("🔥 Hefesto Skills")
	fmt.Println()
	if len(skills) == 0 {
		fmt.Println("  No skills found.")
		fmt.Println()
		return nil
	}
	for _, s := range skills {
		fmt.Printf("  • %s\n", s)
	}
	fmt.Println()
	fmt.Printf("  Total: %d skills\n", len(skills))
	fmt.Println()
	return nil
}

func runListThemes(cmd *cobra.Command, args []string) error {
	themes, err := listEmbeddedThemes()
	if err != nil {
		return fmt.Errorf("failed to list themes: %w", err)
	}

	if listJSON {
		return printListJSON("themes", themes)
	}

	fmt.Println()
	fmt.Println("🔥 Hefesto Themes")
	fmt.Println()
	if len(themes) == 0 {
		fmt.Println("  No themes found.")
		fmt.Println()
		return nil
	}
	for _, t := range themes {
		fmt.Printf("  • %s\n", t)
	}
	fmt.Println()
	fmt.Printf("  Total: %d themes\n", len(themes))
	fmt.Println()
	return nil
}

func runListBackups(cmd *cobra.Command, args []string) error {
	backups, err := install.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if listJSON {
		return printBackupsJSON(backups)
	}

	fmt.Println()
	fmt.Println("🔥 Hefesto Backups")
	fmt.Println()
	if len(backups) == 0 {
		fmt.Println("  No backups found.")
		fmt.Println()
		return nil
	}
	for i, b := range backups {
		dateStr := install.FormatBackupDate(b.Timestamp)
		sizeStr := dirSize(b.Path)
		fmt.Printf("  #%d  %-35s  %s  (%s)\n", i+1, b.Name, dateStr, sizeStr)
	}
	fmt.Println()
	fmt.Printf("  Total: %d backups\n", len(backups))
	fmt.Println()
	return nil
}

// listEmbeddedSkills reads skill directory names from the embedded filesystem.
func listEmbeddedSkills() ([]string, error) {
	configFS, err := fs.Sub(embedpkg.ConfigFiles, "config")
	if err != nil {
		return nil, fmt.Errorf("failed to access embedded config: %w", err)
	}

	skillsDir, err := fs.ReadDir(configFS, "skills")
	if err != nil {
		return nil, fmt.Errorf("failed to read skills directory: %w", err)
	}

	var skills []string
	for _, entry := range skillsDir {
		if entry.IsDir() {
			skills = append(skills, entry.Name())
		}
	}

	sort.Strings(skills)
	return skills, nil
}

// listEmbeddedThemes reads theme file names from the embedded filesystem.
func listEmbeddedThemes() ([]string, error) {
	configFS, err := fs.Sub(embedpkg.ConfigFiles, "config")
	if err != nil {
		return nil, fmt.Errorf("failed to access embedded config: %w", err)
	}

	themesDir, err := fs.ReadDir(configFS, "themes")
	if err != nil {
		return nil, fmt.Errorf("failed to read themes directory: %w", err)
	}

	var themes []string
	for _, entry := range themesDir {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			name := strings.TrimSuffix(entry.Name(), ".json")
			themes = append(themes, name)
		}
	}

	sort.Strings(themes)
	return themes, nil
}

// printListJSON outputs a list of items as JSON.
func printListJSON(resource string, items []string) error {
	output := map[string]interface{}{
		resource: items,
		"count":  len(items),
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// printBackupsJSON outputs backups as JSON.
func printBackupsJSON(backups []install.BackupInfo) error {
	type backupJSON struct {
		Name      string `json:"name"`
		Path      string `json:"path"`
		Timestamp string `json:"timestamp"`
		Size      string `json:"size"`
	}

	var items []backupJSON
	for _, b := range backups {
		items = append(items, backupJSON{
			Name:      b.Name,
			Path:      b.Path,
			Timestamp: b.Timestamp.Format(time.RFC3339),
			Size:      dirSize(b.Path),
		})
	}

	output := map[string]interface{}{
		"backups": items,
		"count":   len(items),
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// dirSize returns a human-readable total size of a directory.
func dirSize(path string) string {
	var total int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error { //nolint:errcheck // best-effort size calculation
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return formatSize(total)
}

// formatSize formats bytes into a human-readable string.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatPathRel formats a path, replacing home directory with ~.
func formatPathRel(path string) string {
	homeDir, err := getUserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, homeDir) {
		return "~" + strings.TrimPrefix(path, homeDir)
	}
	return path
}

// getUserHomeDir returns the user's home directory (re-exported from install package).
func getUserHomeDir() (string, error) {
	return install.DetectHomeDir()
}

func init() {
	// Global persistent flag: --verbose / -V
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "enable debug logging")

	// Install command flags
	installCmd.Flags().BoolVarP(&installYes, "yes", "y", false, "non-interactive mode, accept all defaults")
	installCmd.Flags().BoolVarP(&installDryRun, "dry-run", "d", false, "show what would happen without making changes")
	installCmd.Flags().BoolVar(&installTest, "test", false, "install into a temp directory for safe testing")

	// Rollback command flags
	rollbackCmd.Flags().BoolVarP(&rollbackYes, "yes", "y", false, "Restore most recent backup without prompting")
	rollbackCmd.Flags().BoolVar(&rollbackList, "list", false, "List available backups")

	// Uninstall command flags
	uninstallCmd.Flags().BoolVarP(&uninstallYes, "yes", "y", false, "Skip confirmation")
	uninstallCmd.Flags().BoolVar(&uninstallPurge, "purge", false, "Delete everything without restoring backup")

	// Update command flags
	updateCmd.Flags().BoolVarP(&updateYes, "yes", "y", false, "Skip confirmation")
	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "Show what would change")

	// Status command flags
	statusCmd.Flags().BoolVarP(&statusVerbose, "verbose", "v", false, "Show detailed status information")
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output status in JSON format")

	// Doctor command flags
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Output doctor results in JSON format")

	// List command flags (persistent so all subcommands inherit --json)
	listCmd.PersistentFlags().BoolVar(&listJSON, "json", false, "Output in JSON format")

	// Config subcommands
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)

	// List subcommands
	listCmd.AddCommand(listSkillsCmd)
	listCmd.AddCommand(listThemesCmd)
	listCmd.AddCommand(listBackupsCmd)

	// Add commands to root
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listCmd)
}
