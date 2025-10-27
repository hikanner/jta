package provider

import (
	"context"

	"github.com/hikanner/jta/internal/domain"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// OpenAIProvider implements AIProvider for OpenAI
type OpenAIProvider struct {
	client    *openai.Client
	apiKey    string
	modelName string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string, modelName string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, domain.NewValidationError("OpenAI API key is required", nil)
	}

	if modelName == "" {
		modelName = "gpt-4o" // default model
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &OpenAIProvider{
		client:    &client,
		apiKey:    apiKey,
		modelName: modelName,
	}, nil
}

// Complete executes a text completion
func (p *OpenAIProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Build messages
	messages := []openai.ChatCompletionMessageParamUnion{}

	// Add system message if provided
	if req.SystemMsg != "" {
		messages = append(messages, openai.SystemMessage(req.SystemMsg))
	}

	// Add user message
	messages = append(messages, openai.UserMessage(req.Prompt))

	// Determine model to use
	model := req.Model
	if model == "" {
		model = p.modelName
	}

	// Call API
	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    openai.ChatModel(model),
	}

	chatCompletion, err := p.client.Chat.Completions.New(ctx, params)

	if err != nil {
		return nil, domain.NewProviderError("OpenAI API call failed", err).
			WithContext("model", model).
			WithContext("provider", "openai")
	}

	// Parse response
	if len(chatCompletion.Choices) == 0 {
		return nil, domain.NewProviderError("no response from OpenAI", nil).
			WithContext("model", model)
	}

	return &CompletionResponse{
		Content:      chatCompletion.Choices[0].Message.Content,
		FinishReason: string(chatCompletion.Choices[0].FinishReason),
		Usage: Usage{
			PromptTokens:     int(chatCompletion.Usage.PromptTokens),
			CompletionTokens: int(chatCompletion.Usage.CompletionTokens),
			TotalTokens:      int(chatCompletion.Usage.TotalTokens),
		},
	}, nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// GetModelName returns the current model name
func (p *OpenAIProvider) GetModelName() string {
	return p.modelName
}

// ValidateConfig validates the provider configuration
func (p *OpenAIProvider) ValidateConfig() error {
	if p.apiKey == "" {
		return domain.NewValidationError("OpenAI API key is required", nil)
	}
	return nil
}
