package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/anans9/ai-git/internal/ai"
	"github.com/anans9/ai-git/internal/config"
	"github.com/anans9/ai-git/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage AI-Git configuration",
	Long: `Manage AI-Git configuration including AI providers, templates, and preferences.

The config command allows you to:
• Initialize default configuration
• View current settings
• Set individual configuration values
• Manage AI provider credentials
• Customize commit message templates
• Configure Git and UI preferences`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AI-Git configuration",
	Long: `Initialize AI-Git configuration with default settings.
This creates a configuration file in ~/.config/ai-git/config.yaml with sensible defaults.`,
	RunE: runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show [key]",
	Short: "Show configuration values",
	Long: `Show current configuration values.
If no key is specified, shows all configuration.
Use dot notation to access nested values (e.g., ai.provider, git.auto_stage).`,
	RunE: runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set configuration value",
	Long: `Set a configuration value using dot notation.

Examples:
  ai-git config set ai.provider openai
  ai-git config set ai.temperature 0.7
  ai-git config set git.auto_stage true
  ai-git config set ui.color false`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get configuration value",
	Long: `Get a specific configuration value using dot notation.

Examples:
  ai-git config get ai.provider
  ai-git config get git.auto_stage`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	Long: `Open the configuration file in your default editor.
Uses $EDITOR environment variable or falls back to common editors.`,
	RunE: runConfigEdit,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  `Validate the current configuration and test AI provider connections.`,
	RunE:  runConfigValidate,
}

var configProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Manage AI providers",
	Long:  `Manage AI provider configurations including API keys and settings.`,
}

var configProvidersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List AI providers",
	Long:  `List all configured AI providers and their status.`,
	RunE:  runConfigProvidersList,
}

var configProvidersSetCmd = &cobra.Command{
	Use:   "set <provider> <key> <value>",
	Short: "Set AI provider configuration",
	Long: `Set configuration for a specific AI provider.

Examples:
  ai-git config providers set openai api_key sk-...
  ai-git config providers set anthropic api_key sk-ant-...
  ai-git config providers set local base_url http://localhost:11434
  ai-git config providers set openai model gpt-4`,
	Args: cobra.ExactArgs(3),
	RunE: runConfigProvidersSet,
}

var configProvidersTestCmd = &cobra.Command{
	Use:   "test [provider]",
	Short: "Test AI provider connection",
	Long: `Test connection to AI provider(s).
If no provider is specified, tests the current default provider.`,
	RunE: runConfigProvidersTest,
}

var configResetCmd = &cobra.Command{
	Use:   "reset [key]",
	Short: "Reset configuration to defaults",
	Long: `Reset configuration to default values.
If no key is specified, resets entire configuration.
Use with caution as this will overwrite your settings.`,
	RunE: runConfigReset,
}

func init() {
	// Add subcommands
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configResetCmd)

	// Provider management
	configProvidersCmd.AddCommand(configProvidersListCmd)
	configProvidersCmd.AddCommand(configProvidersSetCmd)
	configProvidersCmd.AddCommand(configProvidersTestCmd)
	configCmd.AddCommand(configProvidersCmd)

	// Flags
	configShowCmd.Flags().BoolP("yaml", "y", false, "Output in YAML format")
	configShowCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	configSetCmd.Flags().Bool("global", false, "Set global configuration")
	configResetCmd.Flags().BoolP("force", "f", false, "Force reset without confirmation")
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	ui := ui.NewUI(viper.GetBool("ui.color"), viper.GetBool("ui.interactive"))

	configPath := config.GetConfigPath()

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		ui.Warning("Configuration file already exists at: %s", configPath)

		confirmed, err := ui.Confirm("Overwrite existing configuration?")
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Info("Configuration initialization cancelled")
			return nil
		}
	}

	ui.StartSpinner("Initializing configuration...")

	if err := config.InitConfig(); err != nil {
		ui.StopSpinner()
		ui.Error("Failed to initialize configuration: %v", err)
		return err
	}

	ui.StopSpinner()
	ui.Success("Configuration initialized at: %s", configPath)
	ui.Info("Edit the configuration file to add your AI provider API keys")
	ui.Info("Use 'ai-git config providers set <provider> api_key <key>' to set API keys")

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(viper.GetBool("ui.color"), viper.GetBool("ui.interactive"))

	// If specific key requested
	if len(args) > 0 {
		key := args[0]
		value := viper.Get(key)
		if value == nil {
			ui.Error("Configuration key '%s' not found", key)
			return fmt.Errorf("key not found: %s", key)
		}

		yamlFlag, _ := cmd.Flags().GetBool("yaml")
		jsonFlag, _ := cmd.Flags().GetBool("json")

		if yamlFlag {
			data, err := yaml.Marshal(map[string]interface{}{key: value})
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		} else if jsonFlag {
			fmt.Printf("{\"%s\": %v}\n", key, value)
		} else {
			ui.Printf("%s: %v", key, value)
		}
		return nil
	}

	// Show full configuration
	ui.Header("AI-Git Configuration")

	// AI Configuration
	ui.Highlight("AI Settings:")
	ui.Printf("  Provider: %s", cfg.AI.Provider)
	ui.Printf("  Model: %s", cfg.AI.Model)
	ui.Printf("  Temperature: %.1f", cfg.AI.Temperature)
	ui.Printf("  Max Tokens: %d", cfg.AI.MaxTokens)
	ui.Print("")

	// Git Configuration
	ui.Highlight("Git Settings:")
	ui.Printf("  Auto Stage: %t", cfg.Git.AutoStage)
	ui.Printf("  Auto Push: %t", cfg.Git.AutoPush)
	ui.Printf("  Max Diff Lines: %d", cfg.Git.MaxDiffLines)
	ui.Printf("  Default Branch: %s", cfg.Git.DefaultBranch)
	ui.Print("")

	// UI Configuration
	ui.Highlight("UI Settings:")
	ui.Printf("  Color: %t", cfg.UI.Color)
	ui.Printf("  Interactive: %t", cfg.UI.Interactive)
	ui.Printf("  Show Diff: %t", cfg.UI.ShowDiff)
	ui.Printf("  Confirm Actions: %t", cfg.UI.ConfirmActions)
	ui.Printf("  Theme: %s", cfg.UI.Theme)
	ui.Print("")

	// Providers (without API keys for security)
	ui.Highlight("AI Providers:")
	for name, provider := range cfg.AI.Providers {
		status := "disabled"
		if provider.Enabled {
			status = "enabled"
		}
		hasKey := "no"
		if provider.APIKey != "" {
			hasKey = "yes"
		}
		ui.Printf("  %s: %s (API Key: %s, Model: %s)", name, status, hasKey, provider.Model)
	}
	ui.Print("")

	// Templates
	ui.Highlight("Templates:")
	ui.Printf("  Default: %s", cfg.Templates.Default)
	ui.Printf("  Conventional Commits: %t", cfg.Templates.Patterns.Conventional)
	ui.Print("")

	ui.Info("Configuration file: %s", config.GetConfigPath())

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	ui := ui.NewUI(viper.GetBool("ui.color"), viper.GetBool("ui.interactive"))

	// Convert string value to appropriate type
	var convertedValue interface{}
	switch strings.ToLower(value) {
	case "true":
		convertedValue = true
	case "false":
		convertedValue = false
	default:
		// Try to convert to number
		if strings.Contains(value, ".") {
			if f, err := parseFloat(value); err == nil {
				convertedValue = f
			} else {
				convertedValue = value
			}
		} else {
			if i, err := parseInt(value); err == nil {
				convertedValue = i
			} else {
				convertedValue = value
			}
		}
	}

	viper.Set(key, convertedValue)

	if err := viper.WriteConfig(); err != nil {
		ui.Error("Failed to save configuration: %v", err)
		return err
	}

	ui.Success("Configuration updated: %s = %v", key, convertedValue)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := viper.Get(key)

	if value == nil {
		return fmt.Errorf("configuration key '%s' not found", key)
	}

	fmt.Printf("%v\n", value)
	return nil
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	configPath := config.GetConfigPath()
	editor := os.Getenv("EDITOR")

	if editor == "" {
		// Try common editors
		editors := []string{"nano", "vim", "vi", "code", "subl"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}

	if editor == "" {
		return fmt.Errorf("no editor found. Set $EDITOR environment variable")
	}

	execCmd := exec.Command(editor, configPath)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	ui := ui.NewUI(viper.GetBool("ui.color"), viper.GetBool("ui.interactive"))

	ui.StartSpinner("Validating configuration...")

	cfg, err := config.Load()
	if err != nil {
		ui.StopSpinner()
		ui.Error("Failed to load configuration: %v", err)
		return err
	}

	if err := cfg.Validate(); err != nil {
		ui.StopSpinner()
		ui.Error("Configuration validation failed: %v", err)
		return err
	}

	ui.StopSpinner()
	ui.Success("Configuration is valid")

	// Test AI provider connection
	ui.StartSpinner("Testing AI provider connection...")

	aiClient, err := ai.NewClient(cfg)
	if err != nil {
		ui.StopSpinner()
		ui.Warning("Failed to create AI client: %v", err)
	} else {
		if err := aiClient.TestConnection(context.Background()); err != nil {
			ui.StopSpinner()
			ui.Warning("AI provider connection test failed: %v", err)
		} else {
			ui.StopSpinner()
			ui.Success("AI provider connection test passed")
		}
	}

	return nil
}

func runConfigProvidersList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(viper.GetBool("ui.color"), viper.GetBool("ui.interactive"))

	ui.Header("AI Providers")

	headers := []string{"Provider", "Status", "Model", "API Key", "Base URL"}
	rows := [][]string{}

	for name, provider := range cfg.AI.Providers {
		status := "Disabled"
		if provider.Enabled {
			status = "Enabled"
		}
		if name == cfg.AI.Provider {
			status += " (Current)"
		}

		apiKeyStatus := "Not Set"
		if provider.APIKey != "" {
			apiKeyStatus = "Set"
		}

		baseURL := provider.BaseURL
		if baseURL == "" {
			baseURL = "Default"
		}

		rows = append(rows, []string{
			name,
			status,
			provider.Model,
			apiKeyStatus,
			baseURL,
		})
	}

	ui.PrintTable(headers, rows)
	return nil
}

func runConfigProvidersSet(cmd *cobra.Command, args []string) error {
	providerName := args[0]
	key := args[1]
	value := args[2]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(viper.GetBool("ui.color"), viper.GetBool("ui.interactive"))

	provider, exists := cfg.AI.Providers[providerName]
	if !exists {
		provider = config.AIProvider{}
	}

	// Update provider configuration
	switch key {
	case "api_key":
		provider.APIKey = value
		ui.Success("API key set for provider: %s", providerName)
	case "base_url":
		provider.BaseURL = value
		ui.Success("Base URL set for provider %s: %s", providerName, value)
	case "model":
		provider.Model = value
		ui.Success("Model set for provider %s: %s", providerName, value)
	case "enabled":
		enabled := strings.ToLower(value) == "true"
		provider.Enabled = enabled
		ui.Success("Provider %s %s", providerName, map[bool]string{true: "enabled", false: "disabled"}[enabled])
	default:
		ui.Error("Unknown provider configuration key: %s", key)
		return fmt.Errorf("unknown key: %s", key)
	}

	cfg.SetProvider(providerName, provider)

	if err := config.Save(cfg); err != nil {
		ui.Error("Failed to save configuration: %v", err)
		return err
	}

	return nil
}

func runConfigProvidersTest(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ui := ui.NewUI(viper.GetBool("ui.color"), viper.GetBool("ui.interactive"))

	var providersToTest []string
	if len(args) > 0 {
		providersToTest = args
	} else {
		providersToTest = []string{cfg.AI.Provider}
	}

	for _, providerName := range providersToTest {
		ui.Printf("Testing provider: %s", providerName)

		// Temporarily switch to this provider for testing
		originalProvider := cfg.AI.Provider
		cfg.AI.Provider = providerName

		aiClient, err := ai.NewClient(cfg)
		if err != nil {
			ui.Error("Failed to create client for %s: %v", providerName, err)
			continue
		}

		ui.StartSpinner(fmt.Sprintf("Testing %s connection...", providerName))

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err = aiClient.TestConnection(ctx)
		cancel()

		ui.StopSpinner()

		if err != nil {
			ui.Error("Provider %s test failed: %v", providerName, err)
		} else {
			ui.Success("Provider %s test passed", providerName)
		}

		// Restore original provider
		cfg.AI.Provider = originalProvider
	}

	return nil
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	ui := ui.NewUI(viper.GetBool("ui.color"), viper.GetBool("ui.interactive"))

	force, _ := cmd.Flags().GetBool("force")

	if !force {
		confirmed, err := ui.Confirm("This will reset your configuration to defaults. Continue?")
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Info("Reset cancelled")
			return nil
		}
	}

	if len(args) > 0 {
		// Reset specific key
		key := args[0]
		// This would require implementing default value lookup
		ui.Warning("Resetting specific keys not yet implemented")
		ui.Info("Use 'ai-git config set %s <default_value>' instead", key)
		return nil
	}

	// Reset entire configuration
	if err := config.InitConfig(); err != nil {
		ui.Error("Failed to reset configuration: %v", err)
		return err
	}

	ui.Success("Configuration reset to defaults")
	ui.Info("Don't forget to set your AI provider API keys")

	return nil
}

// Helper functions
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
