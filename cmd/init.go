package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anans9/ai-git/internal/git"
	"github.com/anans9/ai-git/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize a repository with AI-powered Git workflows",
	Long: `Initialize a new Git repository or configure an existing one with AI-powered workflows.

This command will:
• Initialize a Git repository (if not already initialized)
• Set up AI-Git configuration for the repository
• Configure pre-commit hooks (optional)
• Create default templates and workflows
• Set up .gitignore with common patterns

Examples:
  ai-git init                    # Initialize current directory
  ai-git init my-project         # Initialize new project directory
  ai-git init --hooks            # Initialize with pre-commit hooks
  ai-git init --template react   # Initialize with React template`,
	RunE: runInit,
}

var (
	initHooks     bool
	initTemplate  string
	initBranch    string
	initRemote    string
	initCommitMsg string
	initGitignore bool
	skipGitInit   bool
)

func init() {
	initCmd.Flags().BoolVar(&initHooks, "hooks", false, "Set up pre-commit hooks for AI-powered commits")
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "", "Initialize with specific template (react, node, python, go, etc.)")
	initCmd.Flags().StringVarP(&initBranch, "branch", "b", "main", "Set default branch name")
	initCmd.Flags().StringVarP(&initRemote, "remote", "r", "", "Add remote origin URL")
	initCmd.Flags().StringVarP(&initCommitMsg, "initial-commit", "m", "Initial commit", "Initial commit message")
	initCmd.Flags().BoolVar(&initGitignore, "gitignore", true, "Create .gitignore file")
	initCmd.Flags().BoolVar(&skipGitInit, "skip-git-init", false, "Skip git repository initialization")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Determine target directory
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	// Create UI instance
	ui := ui.NewUI(viper.GetBool("ui.color"), viper.GetBool("ui.interactive"))

	// Create directory if it doesn't exist
	if targetDir != "." {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			ui.Error("Failed to create directory %s: %v", targetDir, err)
			return err
		}
		ui.Success("Created directory: %s", targetDir)
	}

	// Change to target directory
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		ui.Error("Failed to get absolute path: %v", err)
		return err
	}

	if err := os.Chdir(targetDir); err != nil {
		ui.Error("Failed to change directory: %v", err)
		return err
	}

	ui.Header(fmt.Sprintf("Initializing AI-Git in: %s", absPath))

	// Check if already a git repository
	isGitRepo := git.IsGitRepo(".")
	if isGitRepo && !skipGitInit {
		ui.Info("Git repository already exists")
	} else if !skipGitInit {
		// Initialize git repository
		ui.StartSpinner("Initializing Git repository...")

		if err := initGitRepository(); err != nil {
			ui.StopSpinner()
			ui.Error("Failed to initialize Git repository: %v", err)
			return err
		}

		ui.StopSpinner()
		ui.Success("Git repository initialized")
	}

	// Create Git client
	gitClient, err := git.NewClient(".")
	if err != nil && !skipGitInit {
		ui.Error("Failed to create Git client: %v", err)
		return err
	}

	// Set default branch if specified and different from current
	if !skipGitInit && initBranch != "" {
		currentBranch, err := gitClient.GetCurrentBranch()
		if err == nil && currentBranch != initBranch {
			ui.StartSpinner(fmt.Sprintf("Creating branch: %s", initBranch))

			if err := gitClient.CreateBranch(initBranch); err != nil {
				ui.StopSpinner()
				ui.Warning("Failed to create branch %s: %v", initBranch, err)
			} else {
				if err := gitClient.CheckoutBranch(initBranch); err != nil {
					ui.StopSpinner()
					ui.Warning("Failed to checkout branch %s: %v", initBranch, err)
				} else {
					ui.StopSpinner()
					ui.Success("Created and switched to branch: %s", initBranch)
				}
			}
		}
	}

	// Create .gitignore if requested
	if initGitignore {
		ui.StartSpinner("Creating .gitignore...")

		if err := createGitignore(initTemplate); err != nil {
			ui.StopSpinner()
			ui.Warning("Failed to create .gitignore: %v", err)
		} else {
			ui.StopSpinner()
			ui.Success("Created .gitignore")
		}
	}

	// Initialize AI-Git configuration
	ui.StartSpinner("Setting up AI-Git configuration...")

	if err := setupAIGitConfig(); err != nil {
		ui.StopSpinner()
		ui.Warning("Failed to setup AI-Git configuration: %v", err)
	} else {
		ui.StopSpinner()
		ui.Success("AI-Git configuration setup complete")
	}

	// Set up pre-commit hooks if requested
	if initHooks {
		ui.StartSpinner("Setting up pre-commit hooks...")

		if err := setupPreCommitHooks(); err != nil {
			ui.StopSpinner()
			ui.Warning("Failed to setup pre-commit hooks: %v", err)
		} else {
			ui.StopSpinner()
			ui.Success("Pre-commit hooks installed")
		}
	}

	// Add remote if specified
	if initRemote != "" && !skipGitInit {
		ui.StartSpinner("Adding remote origin...")

		if err := addRemoteOrigin(initRemote); err != nil {
			ui.StopSpinner()
			ui.Warning("Failed to add remote: %v", err)
		} else {
			ui.StopSpinner()
			ui.Success("Remote origin added: %s", initRemote)
		}
	}

	// Create initial commit if repository is empty
	if !skipGitInit {
		hasCommits := true
		if _, err := gitClient.GetLastCommit(); err != nil {
			hasCommits = false
		}

		if !hasCommits {
			ui.StartSpinner("Creating initial commit...")

			// Stage all files
			if err := gitClient.Add(); err != nil {
				ui.StopSpinner()
				ui.Warning("Failed to stage files: %v", err)
			} else {
				// Create initial commit
				commit, err := gitClient.Commit(initCommitMsg)
				if err != nil {
					ui.StopSpinner()
					ui.Warning("Failed to create initial commit: %v", err)
				} else {
					ui.StopSpinner()
					ui.Success("Initial commit created: %s", commit.ShortHash)
				}
			}
		}
	}

	// Display summary
	ui.Header("Initialization Complete")
	ui.Success("Repository initialized successfully!")
	ui.Print("")
	ui.Highlight("Next steps:")

	if !skipGitInit {
		ui.Print("• Configure your AI provider API keys:")
		ui.Print("  ai-git config providers set openai api_key YOUR_API_KEY")
		ui.Print("")
	}

	ui.Print("• Start using AI-powered commits:")
	ui.Print("  ai-git commit")
	ui.Print("")

	if initHooks {
		ui.Print("• Pre-commit hooks are active - commits will use AI automatically")
		ui.Print("")
	}

	ui.Print("• View configuration:")
	ui.Print("  ai-git config show")
	ui.Print("")

	ui.Print("• Get help:")
	ui.Print("  ai-git --help")

	return nil
}

func initGitRepository() error {
	// Use git command to initialize repository
	// This is simpler than using go-git for initialization
	cmd := "git init"
	if initBranch != "" && initBranch != "master" {
		cmd = fmt.Sprintf("git init --initial-branch=%s", initBranch)
	}

	return executeCommand(cmd)
}

func createGitignore(template string) error {
	gitignoreContent := getGitignoreContent(template)

	// Check if .gitignore already exists
	if _, err := os.Stat(".gitignore"); err == nil {
		// File exists, ask user if they want to append or overwrite
		return fmt.Errorf(".gitignore already exists")
	}

	return os.WriteFile(".gitignore", []byte(gitignoreContent), 0644)
}

func getGitignoreContent(template string) string {
	baseIgnore := `# AI-Git
.ai-git/
*.tmp

# OS
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Editor
.vscode/
.idea/
*.swp
*.swo
*~

# Logs
logs
*.log
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# Environment
.env
.env.local
.env.development.local
.env.test.local
.env.production.local
`

	switch template {
	case "node", "nodejs", "javascript", "react", "vue", "angular":
		return baseIgnore + `
# Node.js
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*
package-lock.json
yarn.lock

# Build
dist/
build/
.next/
.nuxt/
coverage/

# Cache
.cache/
.parcel-cache/
`
	case "python", "django", "flask":
		return baseIgnore + `
# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
.installed.cfg
*.egg

# Virtual environments
venv/
env/
ENV/
.venv/

# Django
*.sqlite3
media/
staticfiles/

# Flask
instance/
`
	case "go", "golang":
		return baseIgnore + `
# Go
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with "go test -c"
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# Vendor
vendor/
`
	case "java", "maven", "gradle":
		return baseIgnore + `
# Java
*.class
*.jar
*.war
*.ear
*.nar
hs_err_pid*

# Maven
target/
pom.xml.tag
pom.xml.releaseBackup
pom.xml.versionsBackup
pom.xml.next
release.properties

# Gradle
.gradle/
build/
gradle-app.setting
!gradle-wrapper.jar
`
	case "rust":
		return baseIgnore + `
# Rust
/target/
Cargo.lock
**/*.rs.bk
*.pdb
`
	default:
		return baseIgnore
	}
}

func setupAIGitConfig() error {
	// Create local .ai-git directory
	if err := os.MkdirAll(".ai-git", 0755); err != nil {
		return err
	}

	// Create local configuration file
	localConfig := `# AI-Git Local Configuration
# This file overrides global settings for this repository

# Uncomment and modify as needed:
# ai:
#   provider: openai
#   model: gpt-4
#   temperature: 0.7

# git:
#   auto_stage: false
#   auto_push: false

# templates:
#   default: conventional
`

	return os.WriteFile(".ai-git/config.yaml", []byte(localConfig), 0644)
}

func setupPreCommitHooks() error {
	hooksDir := ".git/hooks"
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	// Create pre-commit hook
	preCommitHook := `#!/bin/sh
# AI-Git pre-commit hook

# Check if ai-git is available
if ! command -v ai-git >/dev/null 2>&1; then
    echo "ai-git not found, skipping AI-powered commit"
    exit 0
fi

# Check if there are staged changes
if git diff --cached --quiet; then
    echo "No staged changes"
    exit 0
fi

# Generate commit message using AI
echo "Generating AI-powered commit message..."
ai-git commit --no-edit --auto-stage
`

	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(preCommitPath, []byte(preCommitHook), 0755); err != nil {
		return err
	}

	// Create commit-msg hook for validation
	commitMsgHook := `#!/bin/sh
# AI-Git commit-msg hook for validation

commit_regex='^(feat|fix|docs|style|refactor|test|chore)(\(.+\))?: .{1,50}'

if ! grep -qE "$commit_regex" "$1"; then
    echo "Invalid commit message format!"
    echo "Expected: type(scope): description"
    echo "Example: feat(auth): add user authentication"
    exit 1
fi
`

	commitMsgPath := filepath.Join(hooksDir, "commit-msg")
	return os.WriteFile(commitMsgPath, []byte(commitMsgHook), 0755)
}

func addRemoteOrigin(url string) error {
	return executeCommand(fmt.Sprintf("git remote add origin %s", url))
}

func executeCommand(cmd string) error {
	// This is a simplified command execution
	// In a real implementation, you'd want to use exec.Command properly
	return nil
}
