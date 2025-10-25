package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/hikanner/jta/internal/domain"
)

// AnthropicProvider implements AIProvider for Anthropic Claude
type AnthropicProvider struct {
	client    *anthropic.Client
	apiKey    string
	modelName string
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey string, modelName string) (*AnthropicProvider, error) {
	if apiKey == "" {
		return nil, domain.NewValidationError("Anthropic API key is required", nil)
	}

	if modelName == "" {
		modelName = "claude-3-5-sonnet-20250116" // default model
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &AnthropicProvider{
		client:    &client,
		apiKey:    apiKey,
		modelName: modelName,
	}, nil
}

// Complete executes a text completion
func (p *AnthropicProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Determine model to use
	model := req.Model
	if model == "" {
		model = p.modelName
	}

	// Build parameters
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: int64(req.MaxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
		},
	}

	// Add temperature if specified
	if req.Temperature > 0 {
		params.Temperature = anthropic.Float(float64(req.Temperature))
	}

	// Add system prompt if provided
	if req.SystemMsg != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: req.SystemMsg},
		}
	}

	// Call API
	message, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, domain.NewProviderError("Anthropic API call failed", err).
			WithContext("model", model).
			WithContext("provider", "anthropic")
	}

	// Extract text content
	var content string
	for _, block := range message.Content {
		if block.Text != "" {
			content += block.Text
		}
	}

	return &CompletionResponse{
		Content:      content,
		FinishReason: string(message.StopReason),
		Usage: Usage{
			PromptTokens:     int(message.Usage.InputTokens),
			CompletionTokens: int(message.Usage.OutputTokens),
			TotalTokens:      int(message.Usage.InputTokens + message.Usage.OutputTokens),
		},
	}, nil
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// GetModelName returns the current model name
func (p *AnthropicProvider) GetModelName() string {
	return p.modelName
}

// ValidateConfig validates the provider configuration
func (p *AnthropicProvider) ValidateConfig() error {
	if p.apiKey == "" {
		return domain.NewValidationError("Anthropic API key is required", nil)
	}
	return nil
}
