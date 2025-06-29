package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anans9/ai-git/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "1.0.0"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ai-git",
	Short: "AI-powered Git CLI for automated workflows and commit messages",
	Long: `🤖 AI-Git - AI-Powered Git Workflow Automation

AI-Git is a powerful CLI tool that leverages AI to automate your Git workflows.
It can generate intelligent commit messages, automate common Git operations,
and help you maintain better commit history with minimal effort.

✨ Features:
   • AI-powered commit message generation
   • Automated Git workflows
   • Multiple AI provider support (OpenAI, Anthropic, etc.)
   • Interactive and non-interactive modes
   • Customizable templates and prompts
   • Smart diff analysis

🚀 Quick Start:
   ai-git init                                    # Initialize repository
   ai-git config providers set openai api_key "your-key"
   ai-git commit --auto-stage                     # AI-powered commits

📚 Learn More:
   ai-git [command] --help                       # Get help for any command
   ai-git config show                            # View current configuration
   ai-git template list                          # See available templates`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ai-git.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
	rootCmd.PersistentFlags().String("provider", "", "AI provider to use (openai, anthropic, local)")
	rootCmd.PersistentFlags().String("model", "", "AI model to use")
	rootCmd.PersistentFlags().Bool("dry-run", false, "show what would be done without executing")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("provider", rootCmd.PersistentFlags().Lookup("provider"))
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))

	// Add subcommands
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(workflowCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(uninstallCmd)
}

// initConfig reads in config file and ENV variables
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ai-git" (without extension)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ai-git")
	}

	// Environment variables
	viper.SetEnvPrefix("AI_GIT")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}

	// Initialize default configuration
	config.SetDefaults()
}

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "ai-git")
}
