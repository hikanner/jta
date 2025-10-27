package provider

import (
	"context"
	"strings"

	"github.com/hikanner/jta/internal/domain"
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
		return nil, domain.NewValidationError("Gemini API key is required", nil)
	}

	if modelName == "" {
		modelName = "gemini-2.0-flash-exp" // default model
	}

	// Create Gemini client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, domain.NewProviderError("failed to create Gemini client", err).
			WithContext("provider", "gemini")
	}

	return &GeminiProvider{
		client:    client,
		apiKey:    apiKey,
		modelName: modelName,
	}, nil
}

// Complete executes a text completion
func (p *GeminiProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Build generation config
	config := &genai.GenerateContentConfig{}

	// Add system instruction if provided
	if req.SystemMsg != "" {
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{
				genai.NewPartFromText(req.SystemMsg),
			},
		}
	}

	// Build content with user prompt
	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				genai.NewPartFromText(req.Prompt),
			},
		},
	}

	// Generate content with timeout handling
	type result struct {
		resp *genai.GenerateContentResponse
		err  error
	}

	resultChan := make(chan result, 1)

	go func() {
		apiResp, apiErr := p.client.Models.GenerateContent(ctx, p.modelName, contents, config)
		resultChan <- result{apiResp, apiErr}
	}()

	// Wait for result or timeout
	var resp *genai.GenerateContentResponse
	select {
	case res := <-resultChan:
		if res.err != nil {
			// Check if error is due to invalid model name
			errMsg := res.err.Error()
			if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "models/") {
				return nil, domain.NewProviderError("invalid model name or model not available", res.err).
					WithContext("model", p.modelName).
					WithContext("provider", "gemini")
			}
			return nil, domain.NewProviderError("Gemini API error", res.err).
				WithContext("model", p.modelName).
				WithContext("provider", "gemini")
		}
		resp = res.resp
	case <-ctx.Done():
		return nil, domain.NewProviderError("Gemini API call timeout or cancelled", ctx.Err()).
			WithContext("model", p.modelName).
			WithContext("provider", "gemini")
	}

	// Extract response text
	if len(resp.Candidates) == 0 {
		return nil, domain.NewProviderError("no response candidates from Gemini", nil).
			WithContext("model", p.modelName).
			WithContext("prompt_length", len(req.Prompt))
	}

	candidate := resp.Candidates[0]

	// Check finish reason for safety or other issues
	finishReason := string(candidate.FinishReason)
	if finishReason == "SAFETY" || finishReason == "RECITATION" {
		return nil, domain.NewProviderError("Gemini blocked response due to safety/recitation", nil).
			WithContext("model", p.modelName).
			WithContext("finish_reason", finishReason).
			WithContext("prompt_length", len(req.Prompt))
	}

	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, domain.NewProviderError("empty response from Gemini", nil).
			WithContext("model", p.modelName).
			WithContext("finish_reason", finishReason).
			WithContext("prompt_length", len(req.Prompt))
	}

	// Concatenate all text parts
	var textBuilder strings.Builder
	for _, part := range candidate.Content.Parts {
		// Extract text from part
		if part.Text != "" {
			textBuilder.WriteString(part.Text)
		}
	}

	responseText := textBuilder.String()
	if responseText == "" {
		return nil, domain.NewProviderError("empty text in Gemini response", nil).
			WithContext("model", p.modelName).
			WithContext("finish_reason", finishReason).
			WithContext("parts_count", len(candidate.Content.Parts)).
			WithContext("prompt_length", len(req.Prompt))
	}

	// Extract usage information
	usage := Usage{
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
	}

	if resp.UsageMetadata != nil {
		usage.PromptTokens = int(resp.UsageMetadata.PromptTokenCount)
		usage.CompletionTokens = int(resp.UsageMetadata.CandidatesTokenCount)
		usage.TotalTokens = int(resp.UsageMetadata.TotalTokenCount)
	}

	return &CompletionResponse{
		Content:      responseText,
		FinishReason: finishReason,
		Usage:        usage,
	}, nil
}

// Name returns the provider name
func (p *GeminiProvider) Name() string {
	return "gemini"
}

// GetModelName returns the current model name
func (p *GeminiProvider) GetModelName() string {
	return p.modelName
}

// ValidateConfig validates the provider configuration
func (p *GeminiProvider) ValidateConfig() error {
	if p.apiKey == "" {
		return domain.NewValidationError("Gemini API key is required", nil)
	}
	return nil
}

// Close closes the Gemini client
func (p *GeminiProvider) Close() error {
	// The GenAI client doesn't have a Close method in current version
	// This is a no-op for now
	return nil
}
