package cmd

import (
	"fmt"
	"strings"

	"github.com/anans9/ai-git/internal/config"
	"github.com/anans9/ai-git/internal/ui"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage commit message templates",
	Long: `Manage commit message templates for consistent commit formatting.

Templates define the structure and format of commit messages, supporting:
• Conventional commit format (type(scope): description)
• Custom message templates with variables
• Type-specific templates (feat, fix, docs, etc.)
• Scope definitions for different areas of the project
• Validation rules and patterns

Examples:
  ai-git template list                    # List all templates
  ai-git template show conventional       # Show a specific template
  ai-git template create my-template      # Create a new template
  ai-git template set-default feat        # Set default template type`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long:  `List all available commit message templates and their current status.`,
	RunE:  runTemplateList,
}

var templateShowCmd = &cobra.Command{
	Use:   "show <template-name>",
	Short: "Show template details",
	Long: `Show detailed information about a specific template including format,
variables, and example usage.`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateShow,
}

var templateCreateCmd = &cobra.Command{
	Use:   "create <template-name>",
	Short: "Create a new template",
	Long: `Create a new commit message template with interactive configuration.

This will guide you through defining the template format, variables,
and validation rules.`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateCreate,
}

var templateEditCmd = &cobra.Command{
	Use:   "edit <template-name>",
	Short: "Edit an existing template",
	Long:  `Edit an existing template configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateEdit,
}

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <template-name>",
	Short: "Delete a template",
	Long:  `Delete a custom template. Built-in templates cannot be deleted.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateDelete,
}

var templateSetDefaultCmd = &cobra.Command{
	Use:   "set-default <template-name>",
	Short: "Set default template",
	Long:  `Set the default template to use for commit message generation.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateSetDefault,
}

var templateValidateCmd = &cobra.Command{
	Use:   "validate <message>",
	Short: "Validate a commit message",
	Long: `Validate a commit message against the current template rules.

Examples:
  ai-git template validate "feat: add new feature"
  ai-git template validate "fix(auth): resolve login issue"`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateValidate,
}

var templateTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "Manage commit types",
	Long:  `Manage available commit types (feat, fix, docs, etc.).`,
}

var templateTypesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List commit types",
	Long:  `List all available commit types and their descriptions.`,
	RunE:  runTemplateTypesList,
}

var templateTypesAddCmd = &cobra.Command{
	Use:   "add <type> <description>",
	Short: "Add a new commit type",
	Long:  `Add a new commit type with description.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runTemplateTypesAdd,
}

var templateTypesRemoveCmd = &cobra.Command{
	Use:   "remove <type>",
	Short: "Remove a commit type",
	Long:  `Remove a custom commit type. Built-in types cannot be removed.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateTypesRemove,
}

var templateScopesCmd = &cobra.Command{
	Use:   "scopes",
	Short: "Manage commit scopes",
	Long:  `Manage available commit scopes for different areas of the project.`,
}

var templateScopesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List commit scopes",
	Long:  `List all available commit scopes.`,
	RunE:  runTemplateScopesList,
}

var templateScopesAddCmd = &cobra.Command{
	Use:   "add <scope>",
	Short: "Add a new commit scope",
	Long:  `Add a new commit scope.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateScopesAdd,
}

var templateScopesRemoveCmd = &cobra.Command{
	Use:   "remove <scope>",
	Short: "Remove a commit scope",
	Long:  `Remove a commit scope.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateScopesRemove,
}

func init() {
	// Add subcommands
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateShowCmd)
	templateCmd.AddCommand(templateCreateCmd)
	templateCmd.AddCommand(templateEditCmd)
	templateCmd.AddCommand(templateDeleteCmd)
	templateCmd.AddCommand(templateSetDefaultCmd)
	templateCmd.AddCommand(templateValidateCmd)

	// Types management
	templateTypesCmd.AddCommand(templateTypesListCmd)
	templateTypesCmd.AddCommand(templateTypesAddCmd)
	templateTypesCmd.AddCommand(templateTypesRemoveCmd)
	templateCmd.AddCommand(templateTypesCmd)

	// Scopes management
	templateScopesCmd.AddCommand(templateScopesListCmd)
	templateScopesCmd.AddCommand(templateScopesAddCmd)
	templateScopesCmd.AddCommand(templateScopesRemoveCmd)
	templateCmd.AddCommand(templateScopesCmd)

	// Flags
	templateListCmd.Flags().BoolP("builtin", "b", false, "Show only built-in templates")
	templateListCmd.Flags().BoolP("custom", "c", false, "Show only custom templates")
	templateShowCmd.Flags().BoolP("example", "e", false, "Show example usage")
	templateCreateCmd.Flags().StringP("format", "f", "", "Template format string")
	templateCreateCmd.Flags().StringP("description", "d", "", "Template description")
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	builtinOnly, _ := cmd.Flags().GetBool("builtin")
	customOnly, _ := cmd.Flags().GetBool("custom")

	ui.Header("Commit Message Templates")

	// Built-in templates
	if !customOnly {
		ui.Highlight("Built-in Templates:")

		builtinTemplates := map[string]string{
			"conventional": "type(scope): description",
			"feat":         "feat: {description}",
			"fix":          "fix: {description}",
			"docs":         "docs: {description}",
			"style":        "style: {description}",
			"refactor":     "refactor: {description}",
			"test":         "test: {description}",
			"chore":        "chore: {description}",
		}

		for name, format := range builtinTemplates {
			status := ""
			if name == cfg.Templates.Default {
				status = " (default)"
			}
			ui.Printf("  %s: %s%s", name, format, status)
		}
		ui.Print("")
	}

	// Custom templates
	if !builtinOnly && len(cfg.Templates.Custom) > 0 {
		ui.Highlight("Custom Templates:")
		for name, format := range cfg.Templates.Custom {
			status := ""
			if name == cfg.Templates.Default {
				status = " (default)"
			}
			ui.Printf("  %s: %s%s", name, format, status)
		}
		ui.Print("")
	}

	// Current settings
	ui.Highlight("Current Settings:")
	ui.Printf("  Default template: %s", cfg.Templates.Default)
	ui.Printf("  Conventional commits: %t", cfg.Templates.Patterns.Conventional)
	ui.Printf("  Available types: %s", strings.Join(cfg.Templates.Patterns.Types, ", "))
	if len(cfg.Templates.Patterns.Scopes) > 0 {
		ui.Printf("  Available scopes: %s", strings.Join(cfg.Templates.Patterns.Scopes, ", "))
	}

	return nil
}

func runTemplateShow(cmd *cobra.Command, args []string) error {
	templateName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)
	showExample, _ := cmd.Flags().GetBool("example")

	ui.Header(fmt.Sprintf("Template: %s", templateName))

	// Check built-in templates first
	builtinTemplates := map[string]TemplateInfo{
		"conventional": {
			Format:      "type(scope): description",
			Description: "Conventional commit format with optional scope",
			Variables:   []string{"type", "scope", "description"},
			Example:     "feat(auth): add user authentication",
		},
		"feat": {
			Format:      "feat: {description}",
			Description: "Feature addition template",
			Variables:   []string{"description"},
			Example:     "feat: add user authentication",
		},
		"fix": {
			Format:      "fix: {description}",
			Description: "Bug fix template",
			Variables:   []string{"description"},
			Example:     "fix: resolve login validation issue",
		},
		"docs": {
			Format:      "docs: {description}",
			Description: "Documentation changes template",
			Variables:   []string{"description"},
			Example:     "docs: update API documentation",
		},
		"style": {
			Format:      "style: {description}",
			Description: "Code style changes template",
			Variables:   []string{"description"},
			Example:     "style: fix code formatting",
		},
		"refactor": {
			Format:      "refactor: {description}",
			Description: "Code refactoring template",
			Variables:   []string{"description"},
			Example:     "refactor: simplify user service",
		},
		"test": {
			Format:      "test: {description}",
			Description: "Test-related changes template",
			Variables:   []string{"description"},
			Example:     "test: add user authentication tests",
		},
		"chore": {
			Format:      "chore: {description}",
			Description: "Maintenance tasks template",
			Variables:   []string{"description"},
			Example:     "chore: update dependencies",
		},
	}

	var templateInfo TemplateInfo
	var found bool

	// Check built-in templates
	if info, exists := builtinTemplates[templateName]; exists {
		templateInfo = info
		found = true
		ui.Info("Type: Built-in")
	} else if format, exists := cfg.Templates.Custom[templateName]; exists {
		// Check custom templates
		templateInfo = TemplateInfo{
			Format:      format,
			Description: "Custom template",
			Variables:   extractVariables(format),
		}
		found = true
		ui.Info("Type: Custom")
	}

	if !found {
		ui.Error("Template '%s' not found", templateName)
		return fmt.Errorf("template not found: %s", templateName)
	}

	ui.Print("")
	ui.Highlight("Details:")
	ui.Printf("  Format: %s", templateInfo.Format)
	ui.Printf("  Description: %s", templateInfo.Description)

	if len(templateInfo.Variables) > 0 {
		ui.Printf("  Variables: %s", strings.Join(templateInfo.Variables, ", "))
	}

	if templateName == cfg.Templates.Default {
		ui.Printf("  Status: Default template")
	}

	if showExample && templateInfo.Example != "" {
		ui.Print("")
		ui.Highlight("Example:")
		ui.Printf("  %s", templateInfo.Example)
	}

	return nil
}

func runTemplateCreate(cmd *cobra.Command, args []string) error {
	templateName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Check if template already exists
	if _, exists := cfg.Templates.Custom[templateName]; exists {
		ui.Error("Template '%s' already exists", templateName)
		return fmt.Errorf("template already exists: %s", templateName)
	}

	ui.Header(fmt.Sprintf("Creating Template: %s", templateName))

	// Get format from flag or interactively
	format, _ := cmd.Flags().GetString("format")
	if format == "" {
		var err error
		format, err = ui.Input("Template format (use {variable} for placeholders)", "")
		if err != nil {
			return err
		}
	}

	if strings.TrimSpace(format) == "" {
		ui.Error("Template format cannot be empty")
		return fmt.Errorf("empty template format")
	}

	// Get description
	description, _ := cmd.Flags().GetString("description")
	if description == "" {
		var err error
		description, err = ui.Input("Template description", "")
		if err != nil {
			return err
		}
	}

	// Initialize custom templates map if nil
	if cfg.Templates.Custom == nil {
		cfg.Templates.Custom = make(map[string]string)
	}

	// Add template
	cfg.Templates.Custom[templateName] = format

	// Save configuration
	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save template: %v", err)
		return err
	}

	ui.Success("Template '%s' created successfully", templateName)
	ui.Info("Format: %s", format)

	// Show variables if any
	variables := extractVariables(format)
	if len(variables) > 0 {
		ui.Info("Variables: %s", strings.Join(variables, ", "))
	}

	ui.Print("")
	ui.Info("Use 'ai-git template set-default %s' to make this the default template", templateName)

	return nil
}

func runTemplateEdit(cmd *cobra.Command, args []string) error {
	templateName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Check if it's a custom template
	currentFormat, exists := cfg.Templates.Custom[templateName]
	if !exists {
		ui.Error("Template '%s' not found or is built-in (cannot edit built-in templates)", templateName)
		return fmt.Errorf("template not found or not editable: %s", templateName)
	}

	ui.Header(fmt.Sprintf("Editing Template: %s", templateName))
	ui.Info("Current format: %s", currentFormat)

	// Get new format
	newFormat, err := ui.Input("New template format", currentFormat)
	if err != nil {
		return err
	}

	if strings.TrimSpace(newFormat) == "" {
		ui.Error("Template format cannot be empty")
		return fmt.Errorf("empty template format")
	}

	// Update template
	cfg.Templates.Custom[templateName] = newFormat

	// Save configuration
	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save template: %v", err)
		return err
	}

	ui.Success("Template '%s' updated successfully", templateName)
	ui.Info("New format: %s", newFormat)

	return nil
}

func runTemplateDelete(cmd *cobra.Command, args []string) error {
	templateName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Check if it's a custom template
	if _, exists := cfg.Templates.Custom[templateName]; !exists {
		ui.Error("Template '%s' not found or is built-in (cannot delete built-in templates)", templateName)
		return fmt.Errorf("template not found or not deletable: %s", templateName)
	}

	// Confirm deletion
	confirmed, err := ui.Confirm(fmt.Sprintf("Delete template '%s'?", templateName))
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Info("Deletion cancelled")
		return nil
	}

	// Remove template
	delete(cfg.Templates.Custom, templateName)

	// If this was the default template, reset to conventional
	if cfg.Templates.Default == templateName {
		cfg.Templates.Default = "conventional"
		ui.Warning("Default template reset to 'conventional'")
	}

	// Save configuration
	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save configuration: %v", err)
		return err
	}

	ui.Success("Template '%s' deleted", templateName)
	return nil
}

func runTemplateSetDefault(cmd *cobra.Command, args []string) error {
	templateName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Check if template exists
	builtinTemplates := []string{"conventional", "feat", "fix", "docs", "style", "refactor", "test", "chore"}
	isBuiltin := false
	for _, builtin := range builtinTemplates {
		if builtin == templateName {
			isBuiltin = true
			break
		}
	}

	isCustom := false
	if _, exists := cfg.Templates.Custom[templateName]; exists {
		isCustom = true
	}

	if !isBuiltin && !isCustom {
		ui.Error("Template '%s' not found", templateName)
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Set as default
	cfg.Templates.Default = templateName

	// Save configuration
	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save configuration: %v", err)
		return err
	}

	ui.Success("Default template set to '%s'", templateName)
	return nil
}

func runTemplateValidate(cmd *cobra.Command, args []string) error {
	message := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	ui.Header("Validating Commit Message")
	ui.Info("Message: %s", message)
	ui.Print("")

	// Validate against conventional commit pattern if enabled
	if cfg.Templates.Patterns.Conventional {
		if err := validateConventionalCommit(message, cfg.Templates.Patterns.Types, cfg.Templates.Patterns.Scopes); err != nil {
			ui.Error("Validation failed: %v", err)
			return err
		}
	}

	// Additional validation rules
	if len(message) > 72 {
		ui.Warning("Message is longer than 72 characters (current: %d)", len(message))
	}

	if len(message) > 50 {
		ui.Warning("First line is longer than 50 characters (recommended for subject line)")
	}

	if strings.HasSuffix(message, ".") {
		ui.Warning("Message ends with a period (not recommended for commit subjects)")
	}

	ui.Success("Commit message validation passed")
	return nil
}

func runTemplateTypesList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	ui.Header("Commit Types")

	typeDescriptions := map[string]string{
		"feat":     "New features",
		"fix":      "Bug fixes",
		"docs":     "Documentation changes",
		"style":    "Code style changes (formatting, etc.)",
		"refactor": "Code refactoring",
		"test":     "Test-related changes",
		"chore":    "Maintenance tasks",
		"ci":       "CI/CD changes",
		"build":    "Build system changes",
		"perf":     "Performance improvements",
	}

	for _, commitType := range cfg.Templates.Patterns.Types {
		description := typeDescriptions[commitType]
		if description == "" {
			description = "Custom type"
		}
		ui.Printf("  %s: %s", commitType, description)
	}

	return nil
}

func runTemplateTypesAdd(cmd *cobra.Command, args []string) error {
	commitType := args[0]
	description := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Check if type already exists
	for _, existingType := range cfg.Templates.Patterns.Types {
		if existingType == commitType {
			ui.Error("Commit type '%s' already exists", commitType)
			return fmt.Errorf("type already exists: %s", commitType)
		}
	}

	// Add type
	cfg.Templates.Patterns.Types = append(cfg.Templates.Patterns.Types, commitType)

	// Save configuration
	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save configuration: %v", err)
		return err
	}

	ui.Success("Commit type '%s' added: %s", commitType, description)
	return nil
}

func runTemplateTypesRemove(cmd *cobra.Command, args []string) error {
	commitType := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Built-in types that cannot be removed
	builtinTypes := []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"}
	for _, builtin := range builtinTypes {
		if builtin == commitType {
			ui.Error("Cannot remove built-in commit type: %s", commitType)
			return fmt.Errorf("cannot remove built-in type: %s", commitType)
		}
	}

	// Find and remove type
	newTypes := []string{}
	found := false
	for _, existingType := range cfg.Templates.Patterns.Types {
		if existingType == commitType {
			found = true
			continue
		}
		newTypes = append(newTypes, existingType)
	}

	if !found {
		ui.Error("Commit type '%s' not found", commitType)
		return fmt.Errorf("type not found: %s", commitType)
	}

	cfg.Templates.Patterns.Types = newTypes

	// Save configuration
	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save configuration: %v", err)
		return err
	}

	ui.Success("Commit type '%s' removed", commitType)
	return nil
}

func runTemplateScopesList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	ui.Header("Commit Scopes")

	if len(cfg.Templates.Patterns.Scopes) == 0 {
		ui.Info("No scopes configured")
		ui.Print("Use 'ai-git template scopes add <scope>' to add scopes")
		return nil
	}

	for _, scope := range cfg.Templates.Patterns.Scopes {
		ui.Printf("  %s", scope)
	}

	return nil
}

func runTemplateScopesAdd(cmd *cobra.Command, args []string) error {
	scope := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Check if scope already exists
	for _, existingScope := range cfg.Templates.Patterns.Scopes {
		if existingScope == scope {
			ui.Error("Scope '%s' already exists", scope)
			return fmt.Errorf("scope already exists: %s", scope)
		}
	}

	// Add scope
	cfg.Templates.Patterns.Scopes = append(cfg.Templates.Patterns.Scopes, scope)

	// Save configuration
	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save configuration: %v", err)
		return err
	}

	ui.Success("Scope '%s' added", scope)
	return nil
}

func runTemplateScopesRemove(cmd *cobra.Command, args []string) error {
	scope := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(cfg.UI.Color, cfg.UI.Interactive)

	// Find and remove scope
	newScopes := []string{}
	found := false
	for _, existingScope := range cfg.Templates.Patterns.Scopes {
		if existingScope == scope {
			found = true
			continue
		}
		newScopes = append(newScopes, existingScope)
	}

	if !found {
		ui.Error("Scope '%s' not found", scope)
		return fmt.Errorf("scope not found: %s", scope)
	}

	cfg.Templates.Patterns.Scopes = newScopes

	// Save configuration
	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save configuration: %v", err)
		return err
	}

	ui.Success("Scope '%s' removed", scope)
	return nil
}

// Helper types and functions

type TemplateInfo struct {
	Format      string
	Description string
	Variables   []string
	Example     string
}

func extractVariables(format string) []string {
	var variables []string
	parts := strings.Split(format, "{")
	for i := 1; i < len(parts); i++ {
		if closeBrace := strings.Index(parts[i], "}"); closeBrace != -1 {
			variable := parts[i][:closeBrace]
			variables = append(variables, variable)
		}
	}
	return variables
}

func validateConventionalCommit(message string, types []string, scopes []string) error {
	// Basic format: type(scope): description
	parts := strings.SplitN(message, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("message must be in format 'type(scope): description' or 'type: description'")
	}

	typeAndScope := strings.TrimSpace(parts[0])
	description := strings.TrimSpace(parts[1])

	if description == "" {
		return fmt.Errorf("description cannot be empty")
	}

	// Extract type and scope
	var commitType, scope string
	if strings.Contains(typeAndScope, "(") && strings.Contains(typeAndScope, ")") {
		// Has scope
		openParen := strings.Index(typeAndScope, "(")
		closeParen := strings.Index(typeAndScope, ")")
		if closeParen <= openParen {
			return fmt.Errorf("invalid scope format")
		}
		commitType = typeAndScope[:openParen]
		scope = typeAndScope[openParen+1 : closeParen]
	} else {
		// No scope
		commitType = typeAndScope
	}

	// Validate type
	validType := false
	for _, t := range types {
		if t == commitType {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid commit type '%s'. Valid types: %s", commitType, strings.Join(types, ", "))
	}

	// Validate scope if present and scopes are configured
	if scope != "" && len(scopes) > 0 {
		validScope := false
		for _, s := range scopes {
			if s == scope {
				validScope = true
				break
			}
		}
		if !validScope {
			return fmt.Errorf("invalid scope '%s'. Valid scopes: %s", scope, strings.Join(scopes, ", "))
		}
	}

	return nil
}
