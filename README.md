# AI-Git CLI

AI-powered Git CLI that generates intelligent commit messages and automates Git workflows.

## 🚀 Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/anans9/ai-git/main/install.sh | bash
```

## ⚡ Quick Start

```bash
# Set up your OpenAI API key
ai-git config providers set openai api_key "your-openai-api-key"

# Initialize a repository
ai-git init

# Make some changes and commit with AI-generated message
echo "# My Project" > README.md
ai-git commit --auto-stage
```

## 📖 Usage

### Basic Commands

```bash
ai-git commit                    # Generate AI commit message for staged changes
ai-git commit --auto-stage       # Stage all changes and generate commit message
ai-git commit --type feat        # Generate commit with specific type
ai-git commit --push             # Commit and push to remote
ai-git init                      # Initialize repository with AI-Git
ai-git config show              # Show current configuration
```

### Configuration

```bash
# Set up AI provider
ai-git config providers set openai api_key "your-key"
ai-git config providers set anthropic api_key "your-key"

# Configure templates
ai-git template list             # List available templates
ai-git template create my-template --format "feat: {description}"
ai-git template set-default conventional

# Show all settings
ai-git config show
```

### Custom Templates

```bash
# Create custom commit templates
ai-git template create feature --format "✨ feat({scope}): {description}"
ai-git template create bugfix --format "🐛 fix({scope}): {description}"
ai-git template set-default feature
```

## 🔧 Configuration File

Config is stored at `~/.config/ai-git/config.yaml`:

```yaml
ai:
  provider: openai
  model: gpt-4
  temperature: 0.7
  providers:
    openai:
      api_key: "your-openai-key"

git:
  auto_stage: false
  auto_push: false
  default_branch: main

ui:
  color: true
  interactive: true
  show_diff: true

templates:
  default: conventional
  custom:
    feature: "feat({scope}): {description}"
```

## 🚀 Deployment

### Building from Source

```bash
# Build the binary
make build

# Install globally
make install

# Create release build
make release
```

### Using Install Script

```bash
# Install directly from repository
curl -sSL https://raw.githubusercontent.com/anans9/ai-git/main/install.sh | bash
```

### GitHub Releases

The project uses GoReleaser for automated releases:

```bash
# Create and push a new tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GoReleaser will automatically create:
# - Cross-platform binaries (Linux, macOS, Windows)
# - GitHub release with changelog
# - Checksums and signatures
```

### Manual Installation

```bash
# Download binary for your platform from GitHub releases
# Make it executable and move to PATH
chmod +x ai-git
sudo mv ai-git /usr/local/bin/
```

## 🛠️ Development

```bash
# Clone and build
git clone https://github.com/anans9/ai-git
cd ai-git
make build

# Install locally
make install

# Run tests
make test
```

## 🗑️ Uninstall

```bash
ai-git uninstall --all
```

This removes:
- Binary from `/usr/local/bin`
- All configuration files
- Templates and cache

## 📋 Requirements

- Git
- Internet connection (for AI providers)
- OpenAI/Anthropic API key or local AI model

## 🐛 Issues

Report issues at: https://github.com/anans9/ai-git/issues

## 📄 License

MIT