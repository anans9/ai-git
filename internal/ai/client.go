package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anans9/ai-git/internal/config"
	"github.com/sashabaranov/go-openai"
)

// Client represents an AI client that can work with multiple providers
type Client struct {
	config   *config.Config
	provider Provider
	client   *http.Client
}

// Provider defines the interface for AI providers
type Provider interface {
	GenerateCommitMessage(ctx context.Context, diff string) (string, error)
	GeneratePRTitle(ctx context.Context, changes string) (string, error)
	GeneratePRDescription(ctx context.Context, changes string) (string, error)
	Name() string
}

// Request represents a generic AI request
type Request struct {
	Prompt       string
	SystemPrompt string
	MaxTokens    int
	Temperature  float64
}

// Response represents a generic AI response
type Response struct {
	Content string
	Usage   Usage
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// NewClient creates a new AI client with the specified configuration
func NewClient(cfg *config.Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &Client{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Initialize the appropriate provider
	switch cfg.AI.Provider {
	case "openai":
		provider, err := NewOpenAIProvider(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
		}
		client.provider = provider
	case "anthropic":
		provider, err := NewAnthropicProvider(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create Anthropic provider: %w", err)
		}
		client.provider = provider
	case "local":
		provider, err := NewLocalProvider(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create local provider: %w", err)
		}
		client.provider = provider
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", cfg.AI.Provider)
	}

	return client, nil
}

// GenerateCommitMessage generates a commit message based on the git diff
func (c *Client) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	return c.provider.GenerateCommitMessage(ctx, diff)
}

// GeneratePRTitle generates a pull request title
func (c *Client) GeneratePRTitle(ctx context.Context, changes string) (string, error) {
	return c.provider.GeneratePRTitle(ctx, changes)
}

// GeneratePRDescription generates a pull request description
func (c *Client) GeneratePRDescription(ctx context.Context, changes string) (string, error) {
	return c.provider.GeneratePRDescription(ctx, changes)
}

// GetProviderName returns the name of the current provider
func (c *Client) GetProviderName() string {
	return c.provider.Name()
}

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	client *openai.Client
	config *config.Config
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(cfg *config.Config) (*OpenAIProvider, error) {
	providerConfig, err := cfg.GetProvider("openai")
	if err != nil {
		return nil, err
	}

	if providerConfig.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	clientConfig := openai.DefaultConfig(providerConfig.APIKey)
	if providerConfig.BaseURL != "" {
		clientConfig.BaseURL = providerConfig.BaseURL
	}

	return &OpenAIProvider{
		client: openai.NewClientWithConfig(clientConfig),
		config: cfg,
	}, nil
}

func (p *OpenAIProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	prompt := strings.ReplaceAll(p.config.Templates.Prompts.CommitMessage, "{diff}", diff)
	return p.generate(ctx, prompt)
}

func (p *OpenAIProvider) GeneratePRTitle(ctx context.Context, changes string) (string, error) {
	prompt := strings.ReplaceAll(p.config.Templates.Prompts.PRTitle, "{changes}", changes)
	return p.generate(ctx, prompt)
}

func (p *OpenAIProvider) GeneratePRDescription(ctx context.Context, changes string) (string, error) {
	prompt := strings.ReplaceAll(p.config.Templates.Prompts.PRDescription, "{changes}", changes)
	return p.generate(ctx, prompt)
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) generate(ctx context.Context, prompt string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:       p.config.AI.Model,
		Temperature: float32(p.config.AI.Temperature),
		MaxTokens:   p.config.AI.MaxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: p.config.AI.SystemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// AnthropicProvider implements the Provider interface for Anthropic Claude
type AnthropicProvider struct {
	apiKey string
	config *config.Config
	client *http.Client
}

// AnthropicRequest represents a request to the Anthropic API
type AnthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature"`
	Messages    []AnthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
}

// AnthropicMessage represents a message in the Anthropic API
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents a response from the Anthropic API
type AnthropicResponse struct {
	Content []AnthropicContent `json:"content"`
	Usage   AnthropicUsage     `json:"usage"`
}

// AnthropicContent represents content in the Anthropic response
type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicUsage represents usage information from Anthropic
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(cfg *config.Config) (*AnthropicProvider, error) {
	providerConfig, err := cfg.GetProvider("anthropic")
	if err != nil {
		return nil, err
	}

	if providerConfig.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	return &AnthropicProvider{
		apiKey: providerConfig.APIKey,
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (p *AnthropicProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	prompt := strings.ReplaceAll(p.config.Templates.Prompts.CommitMessage, "{diff}", diff)
	return p.generate(ctx, prompt)
}

func (p *AnthropicProvider) GeneratePRTitle(ctx context.Context, changes string) (string, error) {
	prompt := strings.ReplaceAll(p.config.Templates.Prompts.PRTitle, "{changes}", changes)
	return p.generate(ctx, prompt)
}

func (p *AnthropicProvider) GeneratePRDescription(ctx context.Context, changes string) (string, error) {
	prompt := strings.ReplaceAll(p.config.Templates.Prompts.PRDescription, "{changes}", changes)
	return p.generate(ctx, prompt)
}

func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

func (p *AnthropicProvider) generate(ctx context.Context, prompt string) (string, error) {
	req := AnthropicRequest{
		Model:       p.config.AI.Model,
		MaxTokens:   p.config.AI.MaxTokens,
		Temperature: p.config.AI.Temperature,
		System:      p.config.AI.SystemPrompt,
		Messages: []AnthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("Anthropic API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, string(body))
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("no content in Anthropic response")
	}

	return strings.TrimSpace(anthropicResp.Content[0].Text), nil
}

// LocalProvider implements the Provider interface for local models (e.g., Ollama)
type LocalProvider struct {
	baseURL string
	model   string
	config  *config.Config
	client  *http.Client
}

// LocalRequest represents a request to a local AI model
type LocalRequest struct {
	Model   string       `json:"model"`
	Prompt  string       `json:"prompt"`
	Stream  bool         `json:"stream"`
	Options LocalOptions `json:"options,omitempty"`
}

// LocalOptions represents options for local AI models
type LocalOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

// LocalResponse represents a response from a local AI model
type LocalResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewLocalProvider creates a new local provider
func NewLocalProvider(cfg *config.Config) (*LocalProvider, error) {
	providerConfig, err := cfg.GetProvider("local")
	if err != nil {
		return nil, err
	}

	baseURL := providerConfig.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &LocalProvider{
		baseURL: baseURL,
		model:   providerConfig.Model,
		config:  cfg,
		client:  &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *LocalProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	prompt := fmt.Sprintf("%s\n\n%s", p.config.AI.SystemPrompt,
		strings.ReplaceAll(p.config.Templates.Prompts.CommitMessage, "{diff}", diff))
	return p.generate(ctx, prompt)
}

func (p *LocalProvider) GeneratePRTitle(ctx context.Context, changes string) (string, error) {
	prompt := fmt.Sprintf("%s\n\n%s", p.config.AI.SystemPrompt,
		strings.ReplaceAll(p.config.Templates.Prompts.PRTitle, "{changes}", changes))
	return p.generate(ctx, prompt)
}

func (p *LocalProvider) GeneratePRDescription(ctx context.Context, changes string) (string, error) {
	prompt := fmt.Sprintf("%s\n\n%s", p.config.AI.SystemPrompt,
		strings.ReplaceAll(p.config.Templates.Prompts.PRDescription, "{changes}", changes))
	return p.generate(ctx, prompt)
}

func (p *LocalProvider) Name() string {
	return "local"
}

func (p *LocalProvider) generate(ctx context.Context, prompt string) (string, error) {
	req := LocalRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: false,
		Options: LocalOptions{
			Temperature: p.config.AI.Temperature,
			NumPredict:  p.config.AI.MaxTokens,
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("local AI API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("local AI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var localResp LocalResponse
	if err := json.Unmarshal(body, &localResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return strings.TrimSpace(localResp.Response), nil
}

// TestConnection tests the connection to the AI provider
func (c *Client) TestConnection(ctx context.Context) error {
	testPrompt := "Hello, please respond with 'OK' to confirm the connection is working."

	switch c.config.AI.Provider {
	case "openai":
		_, err := c.provider.(*OpenAIProvider).generate(ctx, testPrompt)
		return err
	case "anthropic":
		_, err := c.provider.(*AnthropicProvider).generate(ctx, testPrompt)
		return err
	case "local":
		_, err := c.provider.(*LocalProvider).generate(ctx, testPrompt)
		return err
	default:
		return fmt.Errorf("unsupported provider for connection test: %s", c.config.AI.Provider)
	}
}
