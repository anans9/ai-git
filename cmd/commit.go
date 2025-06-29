package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anans9/ai-git/internal/ai"
	"github.com/anans9/ai-git/internal/config"
	"github.com/anans9/ai-git/internal/git"
	"github.com/anans9/ai-git/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var commitCmd = &cobra.Command{
	Use:   "commit [flags]",
	Short: "Generate AI-powered commit messages and create commits",
	Long: `Generate intelligent commit messages using AI based on your git diff.
The command analyzes your staged changes (or all changes if --auto-stage is used)
and generates a meaningful commit message following best practices.

Examples:
  ai-git commit                    # Generate commit message for staged changes
  ai-git commit --auto-stage       # Stage all changes and generate commit message
  ai-git commit --message "fix: custom message"  # Use custom message
  ai-git commit --type feat        # Generate message with specific type
  ai-git commit --push             # Commit and push to remote
  ai-git commit --dry-run          # Show what would be committed without doing it`,
	RunE: runCommit,
}

var (
	commitMessage string
	commitType    string
	commitScope   string
	autoStage     bool
	autoPush      bool
	skipVerify    bool
	amendCommit   bool
	noEdit        bool
	showDiff      bool
	maxDiffLines  int
)

func init() {
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Use custom commit message instead of AI generation")
	commitCmd.Flags().StringVarP(&commitType, "type", "t", "", "Commit type (feat, fix, docs, style, refactor, test, chore)")
	commitCmd.Flags().StringVarP(&commitScope, "scope", "s", "", "Commit scope (optional)")
	commitCmd.Flags().BoolVar(&autoStage, "auto-stage", false, "Automatically stage all changes before committing")
	commitCmd.Flags().BoolVar(&autoPush, "push", false, "Push to remote after successful commit")
	commitCmd.Flags().BoolVar(&skipVerify, "no-verify", false, "Skip pre-commit and commit-msg hooks")
	commitCmd.Flags().BoolVar(&amendCommit, "amend", false, "Amend the previous commit")
	commitCmd.Flags().BoolVar(&noEdit, "no-edit", false, "Don't open editor for message editing")
	commitCmd.Flags().BoolVar(&showDiff, "show-diff", false, "Show diff before generating commit message")
	commitCmd.Flags().IntVar(&maxDiffLines, "max-diff-lines", 1000, "Maximum number of diff lines to analyze")

	// Bind flags to viper for configuration
	viper.BindPFlag("git.auto_stage", commitCmd.Flags().Lookup("auto-stage"))
	viper.BindPFlag("git.auto_push", commitCmd.Flags().Lookup("push"))
	viper.BindPFlag("git.max_diff_lines", commitCmd.Flags().Lookup("max-diff-lines"))
	viper.BindPFlag("ui.show_diff", commitCmd.Flags().Lookup("show-diff"))
}

func runCommit(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with command line flags
	if autoStage {
		cfg.Git.AutoStage = true
	}
	if autoPush {
		cfg.Git.AutoPush = true
	}
	if maxDiffLines > 0 {
		cfg.Git.MaxDiffLines = maxDiffLines
	}
	if showDiff {
		cfg.UI.ShowDiff = true
	}

	// Create UI instance
	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive && !noEdit)

	// Create Git client
	gitClient, err := git.NewClient("")
	if err != nil {
		ui.Error("Not a git repository or failed to initialize git client: %v", err)
		return err
	}

	// Check if repository is clean when not auto-staging
	if !cfg.Git.AutoStage && !autoStage {
		hasStaged, err := gitClient.HasStagedChanges()
		if err != nil {
			ui.Error("Failed to check staged changes: %v", err)
			return err
		}

		if !hasStaged {
			hasChanges, err := gitClient.HasChanges()
			if err != nil {
				ui.Error("Failed to check for changes: %v", err)
				return err
			}

			if !hasChanges {
				ui.Success("Nothing to commit, working tree clean")
				return nil
			}

			ui.Warning("No staged changes found. Use --auto-stage to stage all changes, or stage files manually.")

			// Show current status
			status, err := gitClient.GetStatus()
			if err != nil {
				ui.Error("Failed to get repository status: %v", err)
				return err
			}
			ui.PrintStatus(status)
			return nil
		}
	}

	// Auto-stage changes if requested
	if cfg.Git.AutoStage || autoStage {
		ui.StartSpinner("Staging changes...")

		if err := gitClient.Add(); err != nil {
			ui.StopSpinner()
			ui.Error("Failed to stage changes: %v", err)
			return err
		}

		ui.StopSpinner()
		ui.Success("All changes staged")
	}

	// Get staged changes for commit message generation
	diff, err := gitClient.GetStagedDiff()
	if err != nil {
		ui.Error("Failed to get staged diff: %v", err)
		return err
	}

	if len(diff.Files) == 0 {
		ui.Warning("No staged changes to commit")
		return nil
	}

	// Show diff if requested
	if cfg.UI.ShowDiff {
		ui.Header("Changes to be committed")
		ui.PrintDiff(diff)
	}

	// Get commit message
	var finalMessage string

	if commitMessage != "" {
		// Use provided message
		finalMessage = commitMessage
	} else {
		// Generate AI-powered commit message
		finalMessage, err = generateCommitMessage(cfg, ui, diff)
		if err != nil {
			ui.Error("Failed to generate commit message: %v", err)
			return err
		}
	}

	// Apply commit type and scope if specified
	if commitType != "" {
		if commitScope != "" {
			finalMessage = fmt.Sprintf("%s(%s): %s", commitType, commitScope, strings.TrimPrefix(finalMessage, commitType+": "))
		} else {
			if !strings.HasPrefix(finalMessage, commitType+":") {
				finalMessage = fmt.Sprintf("%s: %s", commitType, strings.TrimPrefix(finalMessage, commitType+": "))
			}
		}
	}

	// Allow user to edit message if interactive and not disabled
	if cfg.UI.Interactive && !noEdit {
		ui.Header("Generated Commit Message")
		ui.Highlight(finalMessage)

		confirmed, err := ui.Confirm("Use this commit message?")
		if err != nil {
			return err
		}

		if !confirmed {
			editedMessage, err := ui.Input("Enter commit message", finalMessage)
			if err != nil {
				return err
			}
			finalMessage = editedMessage
		}
	}

	// Validate commit message
	if strings.TrimSpace(finalMessage) == "" {
		ui.Error("Commit message cannot be empty")
		return fmt.Errorf("empty commit message")
	}

	// Show final commit message
	ui.Header("Final Commit Message")
	ui.Highlight(finalMessage)

	// Confirm commit in interactive mode
	if cfg.UI.ConfirmActions && cfg.UI.Interactive {
		confirmed, err := ui.Confirm("Proceed with commit?")
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Warning("Commit cancelled")
			return nil
		}
	}

	// Dry run check
	if viper.GetBool("dry-run") {
		ui.Info("DRY RUN: Would commit with message: %s", finalMessage)
		return nil
	}

	// Create commit
	ui.StartSpinner("Creating commit...")

	commit, err := gitClient.Commit(finalMessage)
	if err != nil {
		ui.StopSpinner()
		ui.Error("Failed to create commit: %v", err)
		return err
	}

	ui.StopSpinner()
	ui.Success("Commit created: %s", commit.ShortHash)
	ui.Info("Author: %s <%s>", commit.Author, commit.Email)
	ui.Info("Date: %s", commit.Date.Format(time.RFC3339))

	// Push if requested
	if cfg.Git.AutoPush || autoPush {
		ui.StartSpinner("Pushing to remote...")

		if err := gitClient.Push(); err != nil {
			ui.StopSpinner()
			ui.Warning("Failed to push to remote: %v", err)
			ui.Info("Commit was created successfully but not pushed")
		} else {
			ui.StopSpinner()
			ui.Success("Changes pushed to remote")
		}
	}

	// Show repository status after commit
	if viper.GetBool("verbose") {
		ui.Header("Repository Status After Commit")
		status, err := gitClient.GetStatus()
		if err != nil {
			ui.Warning("Failed to get status: %v", err)
		} else {
			ui.PrintStatus(status)
		}
	}

	return nil
}

func generateCommitMessage(cfg *config.Config, ui *ui.UI, diff *git.Diff) (string, error) {
	// Create AI client
	aiClient, err := ai.NewClient(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to initialize AI client: %w", err)
	}

	// Prepare diff content for AI analysis
	diffContent := formatDiffForAI(diff, cfg.Git.MaxDiffLines)

	if strings.TrimSpace(diffContent) == "" {
		return "", fmt.Errorf("no diff content available for analysis")
	}

	ui.StartSpinner(fmt.Sprintf("Generating commit message using %s...", aiClient.GetProviderName()))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	message, err := aiClient.GenerateCommitMessage(ctx, diffContent)
	if err != nil {
		ui.StopSpinner()
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	ui.StopSpinner()

	// Clean up and validate the generated message
	message = strings.TrimSpace(message)
	if message == "" {
		return "", fmt.Errorf("AI generated empty commit message")
	}

	// Remove any markdown formatting if present
	message = strings.ReplaceAll(message, "```", "")
	message = strings.ReplaceAll(message, "`", "")

	// Split into lines and take the first line as the main message
	lines := strings.Split(message, "\n")
	message = strings.TrimSpace(lines[0])

	return message, nil
}

func formatDiffForAI(diff *git.Diff, maxLines int) string {
	var result strings.Builder

	// Add summary
	result.WriteString(fmt.Sprintf("Files changed: %d, Insertions: %d, Deletions: %d\n\n",
		diff.Stats.Files, diff.Stats.Additions, diff.Stats.Deletions))

	lineCount := 0
	for _, file := range diff.Files {
		if lineCount >= maxLines {
			result.WriteString(fmt.Sprintf("\n... (truncated, %d more files)", len(diff.Files)))
			break
		}

		// Add file header
		result.WriteString(fmt.Sprintf("File: %s (Status: %s)\n", file.Path, file.Status))
		if file.Additions > 0 || file.Deletions > 0 {
			result.WriteString(fmt.Sprintf("Changes: +%d -%d\n", file.Additions, file.Deletions))
		}

		// Add diff content (limited)
		if file.Content != "" {
			lines := strings.Split(file.Content, "\n")
			for i, line := range lines {
				if lineCount >= maxLines {
					result.WriteString("... (truncated)\n")
					break
				}

				// Skip binary files or very long lines
				if len(line) > 200 {
					result.WriteString("... (line too long)\n")
				} else {
					result.WriteString(line + "\n")
				}

				lineCount++

				// Limit lines per file
				if i > 50 {
					result.WriteString("... (file truncated)\n")
					break
				}
			}
		}

		result.WriteString("\n")
	}

	return result.String()
}
