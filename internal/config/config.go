package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	AI        AIConfig         `yaml:"ai" mapstructure:"ai"`
	Git       GitConfig        `yaml:"git" mapstructure:"git"`
	UI        UIConfig         `yaml:"ui" mapstructure:"ui"`
	Templates TemplateConfig   `yaml:"templates" mapstructure:"templates"`
	Workflows []WorkflowConfig `yaml:"workflows" mapstructure:"workflows"`
}

// AIConfig holds AI provider configurations
type AIConfig struct {
	Provider     string                `yaml:"provider" mapstructure:"provider"`
	Model        string                `yaml:"model" mapstructure:"model"`
	Temperature  float64               `yaml:"temperature" mapstructure:"temperature"`
	MaxTokens    int                   `yaml:"max_tokens" mapstructure:"max_tokens"`
	SystemPrompt string                `yaml:"system_prompt" mapstructure:"system_prompt"`
	Providers    map[string]AIProvider `yaml:"providers" mapstructure:"providers"`
}

// AIProvider represents configuration for a specific AI provider
type AIProvider struct {
	APIKey  string `yaml:"api_key,omitempty" mapstructure:"api_key"`
	BaseURL string `yaml:"base_url,omitempty" mapstructure:"base_url"`
	Model   string `yaml:"model" mapstructure:"model"`
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
}

// GitConfig holds Git-related configuration
type GitConfig struct {
	AutoStage     bool     `yaml:"auto_stage" mapstructure:"auto_stage"`
	AutoPush      bool     `yaml:"auto_push" mapstructure:"auto_push"`
	IgnoreFiles   []string `yaml:"ignore_files" mapstructure:"ignore_files"`
	MaxDiffLines  int      `yaml:"max_diff_lines" mapstructure:"max_diff_lines"`
	DefaultBranch string   `yaml:"default_branch" mapstructure:"default_branch"`
}

// UIConfig holds user interface preferences
type UIConfig struct {
	Color          bool   `yaml:"color" mapstructure:"color"`
	Interactive    bool   `yaml:"interactive" mapstructure:"interactive"`
	ShowDiff       bool   `yaml:"show_diff" mapstructure:"show_diff"`
	ConfirmActions bool   `yaml:"confirm_actions" mapstructure:"confirm_actions"`
	Theme          string `yaml:"theme" mapstructure:"theme"`
}

// TemplateConfig holds commit message templates
type TemplateConfig struct {
	Default  string            `yaml:"default" mapstructure:"default"`
	Custom   map[string]string `yaml:"custom" mapstructure:"custom"`
	Prompts  PromptConfig      `yaml:"prompts" mapstructure:"prompts"`
	Patterns CommitPatterns    `yaml:"patterns" mapstructure:"patterns"`
}

// PromptConfig holds AI prompt configurations
type PromptConfig struct {
	CommitMessage string `yaml:"commit_message" mapstructure:"commit_message"`
	PRTitle       string `yaml:"pr_title" mapstructure:"pr_title"`
	PRDescription string `yaml:"pr_description" mapstructure:"pr_description"`
	CodeReview    string `yaml:"code_review" mapstructure:"code_review"`
}

// CommitPatterns holds commit message patterns
type CommitPatterns struct {
	Conventional bool              `yaml:"conventional" mapstructure:"conventional"`
	Types        []string          `yaml:"types" mapstructure:"types"`
	Scopes       []string          `yaml:"scopes" mapstructure:"scopes"`
	Custom       map[string]string `yaml:"custom" mapstructure:"custom"`
}

// WorkflowConfig represents an automated workflow
type WorkflowConfig struct {
	Name        string            `yaml:"name" mapstructure:"name"`
	Description string            `yaml:"description" mapstructure:"description"`
	Trigger     WorkflowTrigger   `yaml:"trigger" mapstructure:"trigger"`
	Steps       []WorkflowStep    `yaml:"steps" mapstructure:"steps"`
	Conditions  map[string]string `yaml:"conditions" mapstructure:"conditions"`
	Enabled     bool              `yaml:"enabled" mapstructure:"enabled"`
}

// WorkflowTrigger defines when a workflow should run
type WorkflowTrigger struct {
	Event      string            `yaml:"event" mapstructure:"event"`
	Branches   []string          `yaml:"branches" mapstructure:"branches"`
	Files      []string          `yaml:"files" mapstructure:"files"`
	Conditions map[string]string `yaml:"conditions" mapstructure:"conditions"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Name            string            `yaml:"name" mapstructure:"name"`
	Action          string            `yaml:"action" mapstructure:"action"`
	Parameters      map[string]string `yaml:"parameters" mapstructure:"parameters"`
	Condition       string            `yaml:"condition,omitempty" mapstructure:"condition"`
	ContinueOnError bool              `yaml:"continue_on_error" mapstructure:"continue_on_error"`
}

var defaultConfig = Config{
	AI: AIConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   150,
		SystemPrompt: `You are an expert software engineer helping to write commit messages.
Generate concise, descriptive commit messages that follow conventional commit format.
Focus on what changed and why. Be specific but brief.`,
		Providers: map[string]AIProvider{
			"openai": {
				Model:   "gpt-4",
				Enabled: true,
			},
			"anthropic": {
				Model:   "claude-3-sonnet-20240229",
				Enabled: false,
			},
			"local": {
				BaseURL: "http://localhost:11434",
				Model:   "codellama",
				Enabled: false,
			},
		},
	},
	Git: GitConfig{
		AutoStage:     false,
		AutoPush:      false,
		IgnoreFiles:   []string{".env", "*.log", "node_modules/", ".DS_Store"},
		MaxDiffLines:  1000,
		DefaultBranch: "main",
	},
	UI: UIConfig{
		Color:          true,
		Interactive:    true,
		ShowDiff:       true,
		ConfirmActions: true,
		Theme:          "default",
	},
	Templates: TemplateConfig{
		Default: "conventional",
		Custom: map[string]string{
			"fix":      "fix: {description}",
			"feat":     "feat: {description}",
			"docs":     "docs: {description}",
			"style":    "style: {description}",
			"refactor": "refactor: {description}",
			"test":     "test: {description}",
			"chore":    "chore: {description}",
		},
		Prompts: PromptConfig{
			CommitMessage: `Analyze the following git diff and generate a concise commit message.
Follow conventional commit format: type(scope): description

Rules:
- Use present tense ("add" not "added")
- Don't capitalize first letter of description
- No period at the end
- Maximum 50 characters for the first line
- Focus on what and why, not how

Git diff:
{diff}

Commit message:`,
			PRTitle: `Generate a clear and descriptive pull request title based on the changes:

{changes}

Title:`,
			PRDescription: `Generate a detailed pull request description based on the changes:

{changes}

Include:
- What was changed
- Why it was changed
- Any breaking changes
- Testing information

Description:`,
		},
		Patterns: CommitPatterns{
			Conventional: true,
			Types:        []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"},
			Scopes:       []string{"api", "ui", "db", "auth", "config", "ci"},
		},
	},
	Workflows: []WorkflowConfig{
		{
			Name:        "auto-commit-push",
			Description: "Automatically commit and push changes",
			Trigger: WorkflowTrigger{
				Event: "pre-commit",
			},
			Steps: []WorkflowStep{
				{
					Name:   "Generate commit message",
					Action: "ai-commit",
				},
				{
					Name:   "Stage files",
					Action: "git-add",
				},
				{
					Name:   "Commit changes",
					Action: "git-commit",
				},
			},
			Enabled: false,
		},
		{
			Name:        "feature-branch-workflow",
			Description: "Complete feature branch workflow",
			Trigger: WorkflowTrigger{
				Event:    "manual",
				Branches: []string{"feature/*"},
			},
			Steps: []WorkflowStep{
				{
					Name:   "Generate commit message",
					Action: "ai-commit",
				},
				{
					Name:   "Commit changes",
					Action: "git-commit",
				},
				{
					Name:   "Push to origin",
					Action: "git-push",
				},
				{
					Name:   "Create pull request",
					Action: "create-pr",
				},
			},
			Enabled: false,
		},
	},
}

// SetDefaults sets default values in viper
func SetDefaults() {
	// AI defaults
	viper.SetDefault("ai.provider", defaultConfig.AI.Provider)
	viper.SetDefault("ai.model", defaultConfig.AI.Model)
	viper.SetDefault("ai.temperature", defaultConfig.AI.Temperature)
	viper.SetDefault("ai.max_tokens", defaultConfig.AI.MaxTokens)
	viper.SetDefault("ai.system_prompt", defaultConfig.AI.SystemPrompt)

	// Git defaults
	viper.SetDefault("git.auto_stage", defaultConfig.Git.AutoStage)
	viper.SetDefault("git.auto_push", defaultConfig.Git.AutoPush)
	viper.SetDefault("git.ignore_files", defaultConfig.Git.IgnoreFiles)
	viper.SetDefault("git.max_diff_lines", defaultConfig.Git.MaxDiffLines)
	viper.SetDefault("git.default_branch", defaultConfig.Git.DefaultBranch)

	// UI defaults
	viper.SetDefault("ui.color", defaultConfig.UI.Color)
	viper.SetDefault("ui.interactive", defaultConfig.UI.Interactive)
	viper.SetDefault("ui.show_diff", defaultConfig.UI.ShowDiff)
	viper.SetDefault("ui.confirm_actions", defaultConfig.UI.ConfirmActions)
	viper.SetDefault("ui.theme", defaultConfig.UI.Theme)

	// Template defaults
	viper.SetDefault("templates.default", defaultConfig.Templates.Default)
	viper.SetDefault("templates.patterns.conventional", defaultConfig.Templates.Patterns.Conventional)
	viper.SetDefault("templates.patterns.types", defaultConfig.Templates.Patterns.Types)
	viper.SetDefault("templates.patterns.scopes", defaultConfig.Templates.Patterns.Scopes)
}

// Load loads the configuration from viper
func Load() (*Config, error) {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}

// Save saves the configuration to file
func Save(config *Config) error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// InitConfig creates a default configuration file
func InitConfig() error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// Check if config already exists
	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("config file already exists at %s", configFile)
	}

	return Save(&defaultConfig)
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	return filepath.Join(getConfigDir(), "config.yaml")
}

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "ai-git")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate AI provider
	if c.AI.Provider == "" {
		return fmt.Errorf("AI provider is required")
	}

	provider, exists := c.AI.Providers[c.AI.Provider]
	if !exists {
		return fmt.Errorf("unknown AI provider: %s", c.AI.Provider)
	}

	if !provider.Enabled {
		return fmt.Errorf("AI provider %s is disabled", c.AI.Provider)
	}

	// Validate temperature range
	if c.AI.Temperature < 0 || c.AI.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	// Validate max tokens
	if c.AI.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	return nil
}

// GetProvider returns the configuration for the specified provider
func (c *Config) GetProvider(name string) (AIProvider, error) {
	provider, exists := c.AI.Providers[name]
	if !exists {
		return AIProvider{}, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// SetProvider updates the configuration for a provider
func (c *Config) SetProvider(name string, provider AIProvider) {
	if c.AI.Providers == nil {
		c.AI.Providers = make(map[string]AIProvider)
	}
	c.AI.Providers[name] = provider
}
