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

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage and execute automated Git workflows",
	Long: `Manage and execute automated Git workflows that combine AI-powered commit messages
with common Git operations.

Workflows can automate repetitive tasks like:
• Commit + Push workflows
• Feature branch workflows
• Release preparation workflows
• Code review workflows
• Automated testing and deployment triggers

Examples:
  ai-git workflow list                    # List available workflows
  ai-git workflow run auto-commit-push    # Run a specific workflow
  ai-git workflow create my-workflow      # Create a custom workflow
  ai-git workflow enable feature-branch   # Enable a workflow`,
}

var workflowListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available workflows",
	Long:  `List all configured workflows and their current status.`,
	RunE:  runWorkflowList,
}

var workflowRunCmd = &cobra.Command{
	Use:   "run <workflow-name>",
	Short: "Execute a specific workflow",
	Long: `Execute a specific workflow by name.

Examples:
  ai-git workflow run auto-commit-push
  ai-git workflow run feature-branch --branch feature/new-feature`,
	Args: cobra.ExactArgs(1),
	RunE: runWorkflowRun,
}

var workflowCreateCmd = &cobra.Command{
	Use:   "create <workflow-name>",
	Short: "Create a new custom workflow",
	Long: `Create a new custom workflow with interactive configuration.

This will guide you through defining workflow steps, triggers, and conditions.`,
	Args: cobra.ExactArgs(1),
	RunE: runWorkflowCreate,
}

var workflowEditCmd = &cobra.Command{
	Use:   "edit <workflow-name>",
	Short: "Edit an existing workflow",
	Long:  `Edit an existing workflow configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowEdit,
}

var workflowDeleteCmd = &cobra.Command{
	Use:   "delete <workflow-name>",
	Short: "Delete a workflow",
	Long:  `Delete a workflow configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowDelete,
}

var workflowEnableCmd = &cobra.Command{
	Use:   "enable <workflow-name>",
	Short: "Enable a workflow",
	Long:  `Enable a workflow to make it available for execution.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowEnable,
}

var workflowDisableCmd = &cobra.Command{
	Use:   "disable <workflow-name>",
	Short: "Disable a workflow",
	Long:  `Disable a workflow to prevent it from being executed.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowDisable,
}

var workflowShowCmd = &cobra.Command{
	Use:   "show <workflow-name>",
	Short: "Show workflow details",
	Long:  `Show detailed information about a specific workflow.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkflowShow,
}

var (
	workflowBranch    string
	workflowMessage   string
	workflowSkipSteps []string
	workflowDryRun    bool
)

func init() {
	// Add subcommands
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowRunCmd)
	workflowCmd.AddCommand(workflowCreateCmd)
	workflowCmd.AddCommand(workflowEditCmd)
	workflowCmd.AddCommand(workflowDeleteCmd)
	workflowCmd.AddCommand(workflowEnableCmd)
	workflowCmd.AddCommand(workflowDisableCmd)
	workflowCmd.AddCommand(workflowShowCmd)

	// Flags
	workflowRunCmd.Flags().StringVarP(&workflowBranch, "branch", "b", "", "Target branch for workflow")
	workflowRunCmd.Flags().StringVarP(&workflowMessage, "message", "m", "", "Custom message for workflow steps")
	workflowRunCmd.Flags().StringSliceVar(&workflowSkipSteps, "skip", []string{}, "Steps to skip during execution")
	workflowRunCmd.Flags().BoolVar(&workflowDryRun, "dry-run", false, "Show what would be done without executing")

	workflowListCmd.Flags().BoolP("enabled-only", "e", false, "Show only enabled workflows")
	workflowShowCmd.Flags().BoolP("yaml", "y", false, "Output in YAML format")
}

func runWorkflowList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	enabledOnly, _ := cmd.Flags().GetBool("enabled-only")

	if len(cfg.Workflows) == 0 {
		ui.Info("No workflows configured")
		ui.Print("Use 'ai-git workflow create <name>' to create a new workflow")
		return nil
	}

	ui.Header("Available Workflows")

	headers := []string{"Name", "Status", "Trigger", "Steps", "Description"}
	rows := [][]string{}

	for _, workflow := range cfg.Workflows {
		if enabledOnly && !workflow.Enabled {
			continue
		}

		status := "Disabled"
		if workflow.Enabled {
			status = "Enabled"
		}

		stepCount := fmt.Sprintf("%d", len(workflow.Steps))
		description := workflow.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		rows = append(rows, []string{
			workflow.Name,
			status,
			workflow.Trigger.Event,
			stepCount,
			description,
		})
	}

	if len(rows) == 0 {
		ui.Info("No enabled workflows found")
		return nil
	}

	ui.PrintTable(headers, rows)
	return nil
}

func runWorkflowRun(cmd *cobra.Command, args []string) error {
	workflowName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Find workflow
	var targetWorkflow *config.WorkflowConfig
	for _, workflow := range cfg.Workflows {
		if workflow.Name == workflowName {
			targetWorkflow = &workflow
			break
		}
	}

	if targetWorkflow == nil {
		ui.Error("Workflow '%s' not found", workflowName)
		return fmt.Errorf("workflow not found: %s", workflowName)
	}

	if !targetWorkflow.Enabled {
		ui.Error("Workflow '%s' is disabled", workflowName)
		return fmt.Errorf("workflow disabled: %s", workflowName)
	}

	// Create Git client
	gitClient, err := git.NewClient("")
	if err != nil {
		ui.Error("Failed to initialize Git client: %v", err)
		return err
	}

	// Create AI client
	aiClient, err := ai.NewClient(cfg)
	if err != nil {
		ui.Warning("Failed to initialize AI client: %v", err)
		// Continue without AI for workflows that don't need it
	}

	// Execute workflow
	executor := &WorkflowExecutor{
		workflow:  *targetWorkflow,
		config:    cfg,
		ui:        ui,
		gitClient: gitClient,
		aiClient:  aiClient,
		dryRun:    workflowDryRun,
		skipSteps: workflowSkipSteps,
		context: WorkflowContext{
			Branch:  workflowBranch,
			Message: workflowMessage,
		},
	}

	return executor.Execute()
}

func runWorkflowCreate(cmd *cobra.Command, args []string) error {
	workflowName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Check if workflow already exists
	for _, workflow := range cfg.Workflows {
		if workflow.Name == workflowName {
			ui.Error("Workflow '%s' already exists", workflowName)
			return fmt.Errorf("workflow already exists: %s", workflowName)
		}
	}

	ui.Header(fmt.Sprintf("Creating Workflow: %s", workflowName))

	// Interactive workflow creation
	description, err := ui.Input("Description", "")
	if err != nil {
		return err
	}

	triggerEvent, err := ui.Input("Trigger event (manual, pre-commit, post-commit)", "manual")
	if err != nil {
		return err
	}

	// Create basic workflow structure
	newWorkflow := config.WorkflowConfig{
		Name:        workflowName,
		Description: description,
		Trigger: config.WorkflowTrigger{
			Event: triggerEvent,
		},
		Steps:   []config.WorkflowStep{},
		Enabled: true,
	}

	// Add steps interactively
	ui.Info("Add workflow steps (press Enter with empty name to finish):")

	for {
		stepName, err := ui.Input("Step name", "")
		if err != nil {
			return err
		}
		if stepName == "" {
			break
		}

		stepAction, err := ui.Input("Step action", "")
		if err != nil {
			return err
		}

		step := config.WorkflowStep{
			Name:   stepName,
			Action: stepAction,
		}

		newWorkflow.Steps = append(newWorkflow.Steps, step)
		ui.Success("Added step: %s", stepName)
	}

	if len(newWorkflow.Steps) == 0 {
		ui.Error("Workflow must have at least one step")
		return fmt.Errorf("no steps defined")
	}

	// Add to configuration
	cfg.Workflows = append(cfg.Workflows, newWorkflow)

	// Save configuration
	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save workflow: %v", err)
		return err
	}

	ui.Success("Workflow '%s' created successfully", workflowName)
	ui.Info("Use 'ai-git workflow run %s' to execute it", workflowName)

	return nil
}

func runWorkflowEdit(cmd *cobra.Command, args []string) error {
	workflowName := args[0]
	ui := ui.NewUI(viper.GetBool("ui.color"), viper.GetBool("ui.interactive"))

	ui.Warning("Workflow editing not yet implemented")
	ui.Info("Use 'ai-git config edit' to manually edit workflow configurations")
	ui.Info("Workflow name: %s", workflowName)

	return nil
}

func runWorkflowDelete(cmd *cobra.Command, args []string) error {
	workflowName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Find and remove workflow
	found := false
	newWorkflows := []config.WorkflowConfig{}

	for _, workflow := range cfg.Workflows {
		if workflow.Name == workflowName {
			found = true
			continue
		}
		newWorkflows = append(newWorkflows, workflow)
	}

	if !found {
		ui.Error("Workflow '%s' not found", workflowName)
		return fmt.Errorf("workflow not found: %s", workflowName)
	}

	// Confirm deletion
	confirmed, err := ui.Confirm(fmt.Sprintf("Delete workflow '%s'?", workflowName))
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Info("Deletion cancelled")
		return nil
	}

	cfg.Workflows = newWorkflows

	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save configuration: %v", err)
		return err
	}

	ui.Success("Workflow '%s' deleted", workflowName)
	return nil
}

func runWorkflowEnable(cmd *cobra.Command, args []string) error {
	return setWorkflowStatus(args[0], true)
}

func runWorkflowDisable(cmd *cobra.Command, args []string) error {
	return setWorkflowStatus(args[0], false)
}

func setWorkflowStatus(workflowName string, enabled bool) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Find and update workflow
	found := false
	for i, workflow := range cfg.Workflows {
		if workflow.Name == workflowName {
			cfg.Workflows[i].Enabled = enabled
			found = true
			break
		}
	}

	if !found {
		ui.Error("Workflow '%s' not found", workflowName)
		return fmt.Errorf("workflow not found: %s", workflowName)
	}

	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save configuration: %v", err)
		return err
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}
	ui.Success("Workflow '%s' %s", workflowName, status)

	return nil
}

func runWorkflowShow(cmd *cobra.Command, args []string) error {
	workflowName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Find workflow
	var targetWorkflow *config.WorkflowConfig
	for _, workflow := range cfg.Workflows {
		if workflow.Name == workflowName {
			targetWorkflow = &workflow
			break
		}
	}

	if targetWorkflow == nil {
		ui.Error("Workflow '%s' not found", workflowName)
		return fmt.Errorf("workflow not found: %s", workflowName)
	}

	ui.Header(fmt.Sprintf("Workflow: %s", targetWorkflow.Name))

	ui.Highlight("Details:")
	ui.Printf("  Name: %s", targetWorkflow.Name)
	ui.Printf("  Description: %s", targetWorkflow.Description)
	ui.Printf("  Status: %s", map[bool]string{true: "Enabled", false: "Disabled"}[targetWorkflow.Enabled])
	ui.Print("")

	ui.Highlight("Trigger:")
	ui.Printf("  Event: %s", targetWorkflow.Trigger.Event)
	if len(targetWorkflow.Trigger.Branches) > 0 {
		ui.Printf("  Branches: %s", strings.Join(targetWorkflow.Trigger.Branches, ", "))
	}
	if len(targetWorkflow.Trigger.Files) > 0 {
		ui.Printf("  Files: %s", strings.Join(targetWorkflow.Trigger.Files, ", "))
	}
	ui.Print("")

	ui.Highlight("Steps:")
	for i, step := range targetWorkflow.Steps {
		ui.Printf("  %d. %s", i+1, step.Name)
		ui.Printf("     Action: %s", step.Action)
		if step.Condition != "" {
			ui.Printf("     Condition: %s", step.Condition)
		}
		if step.ContinueOnError {
			ui.Printf("     Continue on error: true")
		}
		if len(step.Parameters) > 0 {
			ui.Printf("     Parameters:")
			for key, value := range step.Parameters {
				ui.Printf("       %s: %s", key, value)
			}
		}
	}

	return nil
}

// WorkflowExecutor executes workflows
type WorkflowExecutor struct {
	workflow  config.WorkflowConfig
	config    *config.Config
	ui        *ui.UI
	gitClient *git.Client
	aiClient  *ai.Client
	dryRun    bool
	skipSteps []string
	context   WorkflowContext
}

// WorkflowContext holds context information for workflow execution
type WorkflowContext struct {
	Branch  string
	Message string
	Data    map[string]interface{}
}

// Execute runs the workflow
func (e *WorkflowExecutor) Execute() error {
	e.ui.Header(fmt.Sprintf("Executing Workflow: %s", e.workflow.Name))

	if e.dryRun {
		e.ui.Warning("DRY RUN MODE - No changes will be made")
	}

	// Initialize context data
	if e.context.Data == nil {
		e.context.Data = make(map[string]interface{})
	}

	// Execute each step
	for i, step := range e.workflow.Steps {
		// Check if step should be skipped
		shouldSkip := false
		for _, skipStep := range e.skipSteps {
			if skipStep == step.Name || skipStep == fmt.Sprintf("%d", i+1) {
				shouldSkip = true
				break
			}
		}

		if shouldSkip {
			e.ui.Info("Skipping step %d: %s", i+1, step.Name)
			continue
		}

		e.ui.Info("Executing step %d: %s", i+1, step.Name)

		if err := e.executeStep(step); err != nil {
			if step.ContinueOnError {
				e.ui.Warning("Step failed but continuing: %v", err)
				continue
			}
			return fmt.Errorf("step '%s' failed: %w", step.Name, err)
		}

		e.ui.Success("Step completed: %s", step.Name)
	}

	e.ui.Success("Workflow '%s' completed successfully", e.workflow.Name)
	return nil
}

func (e *WorkflowExecutor) executeStep(step config.WorkflowStep) error {
	if e.dryRun {
		e.ui.Printf("  Would execute: %s", step.Action)
		return nil
	}

	switch step.Action {
	case "ai-commit":
		return e.executeAICommit(step)
	case "git-add":
		return e.executeGitAdd(step)
	case "git-commit":
		return e.executeGitCommit(step)
	case "git-push":
		return e.executeGitPush(step)
	case "create-branch":
		return e.executeCreateBranch(step)
	case "checkout-branch":
		return e.executeCheckoutBranch(step)
	case "create-pr":
		return e.executeCreatePR(step)
	default:
		return fmt.Errorf("unknown action: %s", step.Action)
	}
}

func (e *WorkflowExecutor) executeAICommit(step config.WorkflowStep) error {
	if e.aiClient == nil {
		return fmt.Errorf("AI client not available")
	}

	// Get diff for AI analysis
	diff, err := e.gitClient.GetStagedDiff()
	if err != nil {
		// Try unstaged diff if no staged changes
		diff, err = e.gitClient.GetDiff()
		if err != nil {
			return fmt.Errorf("failed to get diff: %w", err)
		}
	}

	if len(diff.Files) == 0 {
		e.ui.Warning("No changes to commit")
		return nil
	}

	// Generate commit message
	e.ui.StartSpinner("Generating AI commit message...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Format diff for AI
	diffContent := formatDiffForAI(diff, e.config.Git.MaxDiffLines)
	message, err := e.aiClient.GenerateCommitMessage(ctx, diffContent)
	if err != nil {
		e.ui.StopSpinner()
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	e.ui.StopSpinner()

	// Store message in context for later steps
	e.context.Data["commit_message"] = message
	e.context.Message = message

	e.ui.Info("Generated commit message: %s", message)
	return nil
}

func (e *WorkflowExecutor) executeGitAdd(step config.WorkflowStep) error {
	e.ui.StartSpinner("Staging changes...")
	err := e.gitClient.Add()
	e.ui.StopSpinner()
	return err
}

func (e *WorkflowExecutor) executeGitCommit(step config.WorkflowStep) error {
	message := e.context.Message
	if message == "" {
		if msg, ok := e.context.Data["commit_message"].(string); ok {
			message = msg
		} else {
			message = "Automated commit"
		}
	}

	e.ui.StartSpinner("Creating commit...")
	commit, err := e.gitClient.Commit(message)
	e.ui.StopSpinner()

	if err != nil {
		return err
	}

	e.ui.Success("Commit created: %s", commit.ShortHash)
	return nil
}

func (e *WorkflowExecutor) executeGitPush(step config.WorkflowStep) error {
	e.ui.StartSpinner("Pushing to remote...")
	err := e.gitClient.Push()
	e.ui.StopSpinner()
	return err
}

func (e *WorkflowExecutor) executeCreateBranch(step config.WorkflowStep) error {
	branchName := e.context.Branch
	if branchName == "" {
		if name, ok := step.Parameters["name"]; ok {
			branchName = name
		} else {
			return fmt.Errorf("branch name not specified")
		}
	}

	e.ui.StartSpinner(fmt.Sprintf("Creating branch: %s", branchName))
	err := e.gitClient.CreateBranch(branchName)
	e.ui.StopSpinner()

	if err != nil {
		return err
	}

	e.context.Branch = branchName
	return nil
}

func (e *WorkflowExecutor) executeCheckoutBranch(step config.WorkflowStep) error {
	branchName := e.context.Branch
	if branchName == "" {
		if name, ok := step.Parameters["name"]; ok {
			branchName = name
		} else {
			return fmt.Errorf("branch name not specified")
		}
	}

	e.ui.StartSpinner(fmt.Sprintf("Switching to branch: %s", branchName))
	err := e.gitClient.CheckoutBranch(branchName)
	e.ui.StopSpinner()
	return err
}

func (e *WorkflowExecutor) executeCreatePR(step config.WorkflowStep) error {
	e.ui.Info("PR creation not yet implemented")
	e.ui.Info("This would create a pull request with the current changes")
	return nil
}
