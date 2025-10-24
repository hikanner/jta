package provider

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// GeminiProvider implements AIProvider for Google Gemini
type GeminiProvider struct {
	client    *genai.Client
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

	// Initialize client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client:    client,
		apiKey:    apiKey,
		modelName: modelName,
	}, nil
}

// Complete executes a text completion
func (p *GeminiProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Determine model to use
	model := req.Model
	if model == "" {
		model = p.modelName
	}

	// Build contents
	var contents []*genai.Content

	// User message
	contents = append(contents, &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			genai.NewTextPart(req.Prompt),
		},
	})

	// Build configuration
	temperature := float32(req.Temperature)
	maxTokens := int32(req.MaxTokens)

	config := &genai.GenerateContentConfig{
		GenerationConfig: &genai.GenerationConfig{
			Temperature:     &temperature,
			MaxOutputTokens: &maxTokens,
		},
	}

	// Add system instruction if provided
	if req.SystemMsg != "" {
		config.SystemInstruction = &genai.Content{
			Role: "system",
			Parts: []*genai.Part{
				genai.NewTextPart(req.SystemMsg),
			},
		}
	}

	// Call API
	result, err := p.client.Models.GenerateContent(
		ctx,
		model,
		contents,
		config,
	)

	if err != nil {
		return nil, fmt.Errorf("Gemini API call failed: %w", err)
	}

	// Extract text content
	text, err := result.Text()
	if err != nil {
		return nil, fmt.Errorf("failed to extract text from Gemini response: %w", err)
	}

	// Extract token usage information
	var promptTokens, completionTokens int
	if result.UsageMetadata != nil {
		promptTokens = int(result.UsageMetadata.PromptTokenCount)
		completionTokens = int(result.UsageMetadata.CandidatesTokenCount)
	}

	return &CompletionResponse{
		Content:      text,
		FinishReason: "stop", // Gemini's finish reason mapping
		Usage: Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}, nil
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
	return p.client.Close()
}
