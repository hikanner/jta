package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hikanner/jta/internal/domain"
)

// ProviderType represents the type of AI provider
type ProviderType string

const (
	ProviderTypeOpenAI    ProviderType = "openai"
	ProviderTypeAnthropic ProviderType = "anthropic"
	ProviderTypeGoogle    ProviderType = "google"
)

// ProviderConfig holds the configuration for creating a provider
type ProviderConfig struct {
	Type   ProviderType
	APIKey string
	Model  string
}

// NewProvider creates a new AI provider based on the configuration
func NewProvider(ctx context.Context, config *ProviderConfig) (AIProvider, error) {
	// If no model specified, use default model
	modelName := config.Model
	if modelName == "" {
		modelName = GetDefaultModel(config.Type)
	}

	switch config.Type {
	case ProviderTypeOpenAI:
		return NewOpenAIProvider(config.APIKey, modelName)

	case ProviderTypeAnthropic:
		return NewAnthropicProvider(config.APIKey, modelName)

	case ProviderTypeGoogle:
		return NewGeminiProvider(ctx, config.APIKey, modelName)

	default:
		return nil, domain.NewValidationError(fmt.Sprintf("unsupported provider type: %s", config.Type), nil).
			WithContext("provider_type", string(config.Type))
	}
}

// GetDefaultModel returns the default model for a provider type
func GetDefaultModel(providerType ProviderType) string {
	switch providerType {
	case ProviderTypeOpenAI:
		return "gpt-4o"
	case ProviderTypeAnthropic:
		return "claude-3-5-sonnet-20250116"
	case ProviderTypeGoogle:
		return "gemini-2.0-flash-exp"
	default:
		return ""
	}
}

// GetContextWindowSize returns the context window size for a provider type
func GetContextWindowSize(providerType ProviderType) int {
	switch providerType {
	case ProviderTypeOpenAI:
		return 128000 // GPT-4o: 128K tokens
	case ProviderTypeAnthropic:
		return 200000 // Claude 3.5 Sonnet: 200K tokens
	case ProviderTypeGoogle:
		return 1000000 // Gemini 2.0 Flash: 1M tokens
	default:
		return 100000 // conservative estimate
	}
}

// NewProviderFromEnv creates a provider from environment variables
func NewProviderFromEnv(ctx context.Context, providerType ProviderType, modelName string) (AIProvider, error) {
	var apiKey string

	switch providerType {
	case ProviderTypeOpenAI:
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, domain.NewConfigError("OPENAI_API_KEY environment variable not set", nil).
				WithContext("provider", "openai")
		}

	case ProviderTypeAnthropic:
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, domain.NewConfigError("ANTHROPIC_API_KEY environment variable not set", nil).
				WithContext("provider", "anthropic")
		}

	case ProviderTypeGoogle:
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			// Fallback to GOOGLE_API_KEY
			apiKey = os.Getenv("GOOGLE_API_KEY")
		}
		if apiKey == "" {
			return nil, domain.NewConfigError("GEMINI_API_KEY or GOOGLE_API_KEY environment variable not set", nil).
				WithContext("provider", "google")
		}

	default:
		return nil, domain.NewValidationError(fmt.Sprintf("unsupported provider type: %s", providerType), nil).
			WithContext("provider_type", string(providerType))
	}

	return NewProvider(ctx, &ProviderConfig{
		Type:   providerType,
		APIKey: apiKey,
		Model:  modelName, // will be handled in NewProvider
	})
}
