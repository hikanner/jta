package provider

import (
	"context"
	"fmt"
)

// GeminiProvider implements AIProvider for Google Gemini
// Note: Simplified implementation - full SDK integration pending
type GeminiProvider struct {
	apiKey    string
	modelName string
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(ctx context.Context, apiKey string, modelName string) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	if modelName == "" {
		modelName = "gemini-2.0-flash-exp" // default model
	}

	return &GeminiProvider{
		apiKey:    apiKey,
		modelName: modelName,
	}, nil
}

// Complete executes a text completion
func (p *GeminiProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// TODO: Implement full Gemini SDK integration
	// For now, return an error indicating it's not yet implemented
	return nil, fmt.Errorf("Gemini provider not yet fully implemented - please use OpenAI or Anthropic for now")
}

// Name returns the provider name
func (p *GeminiProvider) Name() string {
	return "google"
}

// GetModelName returns the current model name
func (p *GeminiProvider) GetModelName() string {
	return p.modelName
}

// ValidateConfig validates the provider configuration
func (p *GeminiProvider) ValidateConfig() error {
	if p.apiKey == "" {
		return fmt.Errorf("Gemini API key is required")
	}
	return nil
}

// Close closes the Gemini client
func (p *GeminiProvider) Close() error {
	return nil
}
