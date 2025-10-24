package provider

import "context"

// CompletionRequest represents a completion request to AI provider
type CompletionRequest struct {
	Prompt      string
	Model       string
	Temperature float32
	MaxTokens   int
	SystemMsg   string
}

// CompletionResponse represents the response from AI provider
type CompletionResponse struct {
	Content      string
	FinishReason string
	Usage        Usage
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// AIProvider defines the interface for AI provider implementations
type AIProvider interface {
	// Complete executes a text completion
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Name returns the provider name
	Name() string

	// GetModelName returns the current model name
	GetModelName() string

	// ValidateConfig validates the provider configuration
	ValidateConfig() error
}
