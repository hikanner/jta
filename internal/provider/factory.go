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
	ProviderTypeGemini    ProviderType = "gemini"
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

	case ProviderTypeGemini:
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
		return "gpt-5"
	case ProviderTypeAnthropic:
		return "claude-sonnet-4-5"
	case ProviderTypeGemini:
		return "gemini-2.5-flash"
	default:
		return ""
	}
}

// GetContextWindowSize returns the context window size for a provider type
func GetContextWindowSize(providerType ProviderType) int {
	switch providerType {
	case ProviderTypeOpenAI:
		return 128000 // GPT-5: 128K tokens (estimated, update when official specs available)
	case ProviderTypeAnthropic:
		return 200000 // Claude Sonnet 4.5: 200K tokens (1M beta available)
	case ProviderTypeGemini:
		return 1048576 // Gemini 2.5 Flash: ~1M tokens
	default:
		return 100000 // conservative estimate
	}
}

// GetSupportedModels returns all supported models for a provider type
func GetSupportedModels(providerType ProviderType) []string {
	switch providerType {
	case ProviderTypeOpenAI:
		return []string{
			"gpt-5",       // default
			"gpt-5-mini",  // faster, cost-efficient
			"gpt-5-nano",  // fastest, most cost-efficient
			"gpt-5-pro",   // smarter and more precise
			"gpt-4o",      // legacy (still supported)
			"gpt-4o-mini", // legacy mini
		}
	case ProviderTypeAnthropic:
		return []string{
			"claude-sonnet-4-5",          // default
			"claude-haiku-4-5",           // fastest
			"claude-opus-4-1",            // exceptional reasoning
			"claude-sonnet-4-0",          // legacy Sonnet 4
			"claude-3-5-sonnet-20250116", // legacy 3.5
		}
	case ProviderTypeGemini:
		return []string{
			"gemini-2.5-flash",      // default (price-performance)
			"gemini-2.5-pro",        // state-of-the-art thinking
			"gemini-2.5-flash-lite", // fastest, high throughput
			"gemini-2.0-flash-exp",  // legacy 2.0
		}
	default:
		return []string{}
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

	case ProviderTypeGemini:
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			// Fallback to GOOGLE_API_KEY for backward compatibility
			apiKey = os.Getenv("GOOGLE_API_KEY")
		}
		if apiKey == "" {
			return nil, domain.NewConfigError("GEMINI_API_KEY or GOOGLE_API_KEY environment variable not set", nil).
				WithContext("provider", "gemini")
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
