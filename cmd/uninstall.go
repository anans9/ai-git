package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anans9/ai-git/internal/ui"
	"github.com/spf13/cobra"
)

var (
	uninstallForce bool
	uninstallAll   bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall AI-Git CLI and remove all data",
	Long: `Completely remove AI-Git CLI from your system including:
• Binary file from /usr/local/bin
• Configuration files and directories
• All templates and settings
• Cache and temporary files

This action cannot be undone.`,
	RunE: runUninstall,
}

func init() {
	uninstallCmd.Flags().BoolVarP(&uninstallForce, "force", "f", false, "Skip confirmation prompts")
	uninstallCmd.Flags().BoolVar(&uninstallAll, "all", false, "Remove all traces including config files")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	ui := ui.NewUI(true, true) // Force color and interactive for uninstall

	ui.Header("AI-Git CLI Uninstaller")
	ui.Warning("This will completely remove AI-Git CLI from your system")

	if !uninstallForce {
		confirmed, err := ui.Confirm("Are you sure you want to continue?")
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Info("Uninstall cancelled")
			return nil
		}
	}

	ui.Print("")
	ui.Info("Starting uninstall process...")

	var errors []string

	// 1. Remove binary
	if err := removeBinary(ui); err != nil {
		errors = append(errors, fmt.Sprintf("Failed to remove binary: %v", err))
	}

	// 2. Remove config files
	if uninstallAll {
		if err := removeConfigFiles(ui); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to remove config files: %v", err))
		}
	}

	// 3. Remove cache and temp files
	if err := removeCacheFiles(ui); err != nil {
		errors = append(errors, fmt.Sprintf("Failed to remove cache files: %v", err))
	}

	ui.Print("")

	if len(errors) > 0 {
		ui.Error("Uninstall completed with errors:")
		for _, err := range errors {
			ui.Printf("  • %s", err)
		}
		ui.Print("")
		ui.Warning("You may need to manually remove some files with administrator privileges")
		return fmt.Errorf("uninstall completed with %d errors", len(errors))
	}

	ui.Success("AI-Git CLI has been completely removed from your system")
	ui.Print("")
	ui.Info("Thank you for using AI-Git CLI! 👋")

	return nil
}

func removeBinary(ui *ui.UI) error {
	ui.StartSpinner("Removing binary...")

	// Common binary locations
	binaryPaths := []string{
		"/usr/local/bin/ai-git",
		"/usr/bin/ai-git",
		filepath.Join(os.Getenv("HOME"), "bin", "ai-git"),
		filepath.Join(os.Getenv("HOME"), ".local", "bin", "ai-git"),
	}

	// Find where the binary is installed
	var installedPath string
	for _, path := range binaryPaths {
		if _, err := os.Stat(path); err == nil {
			installedPath = path
			break
		}
	}

	ui.StopSpinner()

	if installedPath == "" {
		ui.Warning("Binary not found in common locations")
		return nil
	}

	ui.Info("Found binary at: %s", installedPath)

	// Try to remove without sudo first
	if err := os.Remove(installedPath); err != nil {
		// If permission denied, try with sudo
		if os.IsPermission(err) {
			ui.Info("Administrator privileges required...")
			if err := runCommand("sudo", "rm", installedPath); err != nil {
				return fmt.Errorf("failed to remove binary: %w", err)
			}
		} else {
			return fmt.Errorf("failed to remove binary: %w", err)
		}
	}

	ui.Success("Binary removed")
	return nil
}

func removeConfigFiles(ui *ui.UI) error {
	ui.StartSpinner("Removing configuration files...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		ui.StopSpinner()
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPaths := []string{
		filepath.Join(homeDir, ".config", "ai-git"),
		filepath.Join(homeDir, ".ai-git.yaml"),
		filepath.Join(homeDir, ".ai-git"),
	}

	ui.StopSpinner()

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			ui.Info("Removing: %s", path)
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}
		}
	}

	ui.Success("Configuration files removed")
	return nil
}

func removeCacheFiles(ui *ui.UI) error {
	ui.StartSpinner("Removing cache files...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		ui.StopSpinner()
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	cachePaths := []string{
		filepath.Join(homeDir, ".cache", "ai-git"),
		filepath.Join(os.TempDir(), "ai-git-*"),
	}

	ui.StopSpinner()

	for _, path := range cachePaths {
		// Handle glob patterns for temp files
		if filepath.Base(path) == "ai-git-*" {
			matches, err := filepath.Glob(path)
			if err != nil {
				continue
			}
			for _, match := range matches {
				if err := os.RemoveAll(match); err != nil {
					ui.Warning("Failed to remove temp file: %s", match)
				}
			}
		} else {
			if _, err := os.Stat(path); err == nil {
				ui.Info("Removing cache: %s", path)
				if err := os.RemoveAll(path); err != nil {
					return fmt.Errorf("failed to remove %s: %w", path, err)
				}
			}
		}
	}

	ui.Success("Cache files removed")
	return nil
}

func runCommand(name string, args ...string) error {
	cmd := fmt.Sprintf("%s %s", name, filepath.Join(args...))
	return executeCommand(cmd)
}
