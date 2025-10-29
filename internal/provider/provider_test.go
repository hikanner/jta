package provider

import (
	"context"
	"testing"
)

func TestNewMockProvider(t *testing.T) {
	provider := NewMockProvider("test-model")

	if provider.GetModelName() != "test-model" {
		t.Errorf("GetModelName() = %s, want test-model", provider.GetModelName())
	}

	if provider.GetCallCount() != 0 {
		t.Errorf("GetCallCount() = %d, want 0", provider.GetCallCount())
	}
}

func TestMockProviderComplete(t *testing.T) {
	ctx := context.Background()
	provider := NewMockProvider("test-model")

	// Add canned responses
	provider.AddResponse("Hello, world!")
	provider.AddResponse("Second response")

	// Test first call
	req := &CompletionRequest{
		Prompt: "Test prompt",
	}

	resp, err := provider.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete() error = %v, want nil", err)
	}

	if resp.Content != "Hello, world!" {
		t.Errorf("Complete() content = %s, want 'Hello, world!'", resp.Content)
	}

	if resp.Usage.TotalTokens != 150 {
		t.Errorf("Complete() usage.TotalTokens = %d, want 150", resp.Usage.TotalTokens)
	}

	if provider.GetCallCount() != 1 {
		t.Errorf("GetCallCount() = %d, want 1", provider.GetCallCount())
	}

	// Test second call
	resp, err = provider.Complete(ctx, req)
	if err != nil {
		t.Fatalf("Complete() error = %v, want nil", err)
	}

	if resp.Content != "Second response" {
		t.Errorf("Complete() content = %s, want 'Second response'", resp.Content)
	}

	if provider.GetCallCount() != 2 {
		t.Errorf("GetCallCount() = %d, want 2", provider.GetCallCount())
	}
}

func TestMockProviderError(t *testing.T) {
	ctx := context.Background()
	provider := NewMockProvider("test-model")

	provider.SetError("simulated error")

	req := &CompletionRequest{
		Prompt: "Test prompt",
	}

	_, err := provider.Complete(ctx, req)
	if err == nil {
		t.Fatal("Complete() error = nil, want error")
	}

	if err.Error() != "simulated error" {
		t.Errorf("Complete() error = %v, want 'simulated error'", err)
	}
}

func TestMockProviderNoResponses(t *testing.T) {
	ctx := context.Background()
	provider := NewMockProvider("test-model")

	// Don't add any responses
	req := &CompletionRequest{
		Prompt: "Test prompt",
	}

	_, err := provider.Complete(ctx, req)
	if err == nil {
		t.Fatal("Complete() error = nil, want error")
	}
}

func TestMockProviderReset(t *testing.T) {
	ctx := context.Background()
	provider := NewMockProvider("test-model")

	provider.AddResponse("Response 1")
	_, _ = provider.Complete(ctx, &CompletionRequest{Prompt: "Test"})

	if provider.GetCallCount() != 1 {
		t.Errorf("GetCallCount() = %d, want 1", provider.GetCallCount())
	}

	// Reset
	provider.Reset()

	if provider.GetCallCount() != 0 {
		t.Errorf("GetCallCount() after reset = %d, want 0", provider.GetCallCount())
	}

	// Should be able to reuse the same responses
	resp, err := provider.Complete(ctx, &CompletionRequest{Prompt: "Test"})
	if err != nil {
		t.Fatalf("Complete() after reset error = %v, want nil", err)
	}

	if resp.Content != "Response 1" {
		t.Errorf("Complete() after reset content = %s, want 'Response 1'", resp.Content)
	}
}

func TestProviderFactory(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		providerType ProviderType
		model        string
		apiKey       string
		wantErr      bool
		wantNil      bool
	}{
		{
			name:         "openai provider",
			providerType: ProviderTypeOpenAI,
			model:        "gpt-4",
			apiKey:       "test-key",
			wantErr:      false,
			wantNil:      false,
		},
		{
			name:         "anthropic provider",
			providerType: ProviderTypeAnthropic,
			model:        "claude-3",
			apiKey:       "test-key",
			wantErr:      false,
			wantNil:      false,
		},
		{
			name:         "gemini provider",
			providerType: ProviderTypeGemini,
			model:        "gemini-pro",
			apiKey:       "test-key",
			wantErr:      false,
			wantNil:      false,
		},
		{
			name:         "unknown provider",
			providerType: ProviderType("unknown"),
			model:        "test",
			apiKey:       "test-key",
			wantErr:      true,
			wantNil:      true,
		},
		{
			name:         "missing api key",
			providerType: ProviderTypeOpenAI,
			model:        "gpt-4",
			apiKey:       "",
			wantErr:      true, // Providers validate API keys at creation time
			wantNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ProviderConfig{
				Type:   tt.providerType,
				Model:  tt.model,
				APIKey: tt.apiKey,
			}
			provider, err := NewProvider(ctx, config)

			if tt.wantErr && err == nil {
				t.Errorf("NewProvider() error = nil, want error")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("NewProvider() error = %v, want nil", err)
				return
			}

			// Only check provider if there's no error
			if !tt.wantErr {
				if tt.wantNil && provider != nil {
					t.Errorf("NewProvider() provider = %v, want nil", provider)
				}

				if !tt.wantNil && provider == nil {
					t.Errorf("NewProvider() provider = nil, want non-nil")
				}
			}
		})
	}
}

func TestCompletionRequest(t *testing.T) {
	req := &CompletionRequest{
		Prompt:    "Test prompt",
		Model:     "test-model",
		SystemMsg: "System message",
	}

	if req.Prompt != "Test prompt" {
		t.Errorf("Prompt = %s, want 'Test prompt'", req.Prompt)
	}

	if req.Model != "test-model" {
		t.Errorf("Model = %s, want 'test-model'", req.Model)
	}

	if req.SystemMsg != "System message" {
		t.Errorf("SystemMsg = %s, want 'System message'", req.SystemMsg)
	}
}

func TestCompletionResponse(t *testing.T) {
	resp := &CompletionResponse{
		Content: "Response content",
		Usage: Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}

	if resp.Content != "Response content" {
		t.Errorf("Content = %s, want 'Response content'", resp.Content)
	}

	if resp.Usage.PromptTokens != 100 {
		t.Errorf("Usage.PromptTokens = %d, want 100", resp.Usage.PromptTokens)
	}

	if resp.Usage.CompletionTokens != 50 {
		t.Errorf("Usage.CompletionTokens = %d, want 50", resp.Usage.CompletionTokens)
	}

	if resp.Usage.TotalTokens != 150 {
		t.Errorf("Usage.TotalTokens = %d, want 150", resp.Usage.TotalTokens)
	}
}

func TestGetDefaultModel(t *testing.T) {
	tests := []struct {
		name         string
		providerType ProviderType
		expected     string
	}{
		{"openai default", ProviderTypeOpenAI, "gpt-5"},
		{"anthropic default", ProviderTypeAnthropic, "claude-sonnet-4-5"},
		{"gemini default", ProviderTypeGemini, "gemini-2.5-flash"},
		{"unknown default", ProviderType("unknown"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDefaultModel(tt.providerType)
			if result != tt.expected {
				t.Errorf("GetDefaultModel(%v) = %q, want %q", tt.providerType, result, tt.expected)
			}
		})
	}
}

func TestGetContextWindowSize(t *testing.T) {
	tests := []struct {
		name         string
		providerType ProviderType
		expected     int
	}{
		{"openai context", ProviderTypeOpenAI, 128000},
		{"anthropic context", ProviderTypeAnthropic, 200000},
		{"gemini context", ProviderTypeGemini, 1048576},
		{"unknown provider", ProviderType("unknown"), 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetContextWindowSize(tt.providerType)
			if result != tt.expected {
				t.Errorf("GetContextWindowSize(%v) = %d, want %d", tt.providerType, result, tt.expected)
			}
		})
	}
}

func TestGetSupportedModels(t *testing.T) {
	tests := []struct {
		name         string
		providerType ProviderType
		minExpected  int
	}{
		{"openai models", ProviderTypeOpenAI, 3},
		{"anthropic models", ProviderTypeAnthropic, 3},
		{"gemini models", ProviderTypeGemini, 2},
		{"unknown provider", ProviderType("unknown"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSupportedModels(tt.providerType)
			if len(result) < tt.minExpected {
				t.Errorf("GetSupportedModels(%v) returned %d models, want at least %d", tt.providerType, len(result), tt.minExpected)
			}
		})
	}
}

func TestMockProviderName(t *testing.T) {
	provider := NewMockProvider("test-model")
	name := provider.Name()
	if name != "mock" {
		t.Errorf("Name() = %q, want %q", name, "mock")
	}
}

func TestMockProviderValidateConfig(t *testing.T) {
	provider := NewMockProvider("test-model")
	err := provider.ValidateConfig()
	if err != nil {
		t.Errorf("ValidateConfig() = %v, want nil", err)
	}
}

func TestProviderTypes(t *testing.T) {
	// Test that provider type constants are defined
	types := []ProviderType{
		ProviderTypeOpenAI,
		ProviderTypeAnthropic,
		ProviderTypeGemini,
	}

	expectedValues := []string{
		"openai",
		"anthropic",
		"gemini",
	}

	for i, providerType := range types {
		if string(providerType) != expectedValues[i] {
			t.Errorf("ProviderType[%d] = %q, want %q", i, providerType, expectedValues[i])
		}
	}
}

func TestProviderConfig(t *testing.T) {
	config := &ProviderConfig{
		Type:   ProviderTypeOpenAI,
		Model:  "gpt-4",
		APIKey: "test-key",
	}

	if config.Type != ProviderTypeOpenAI {
		t.Errorf("Type = %v, want %v", config.Type, ProviderTypeOpenAI)
	}
	if config.Model != "gpt-4" {
		t.Errorf("Model = %s, want gpt-4", config.Model)
	}
	if config.APIKey != "test-key" {
		t.Errorf("APIKey = %s, want test-key", config.APIKey)
	}
}

func TestNewProviderWithEmptyModel(t *testing.T) {
	ctx := context.Background()
	config := &ProviderConfig{
		Type:   ProviderTypeOpenAI,
		Model:  "", // Empty model should use default
		APIKey: "test-key",
	}

	provider, err := NewProvider(ctx, config)
	if err != nil {
		t.Fatalf("NewProvider() with empty model error = %v, want nil", err)
	}

	if provider == nil {
		t.Fatal("NewProvider() returned nil provider")
	}

	// Should have default model
	modelName := provider.GetModelName()
	if modelName == "" {
		t.Error("GetModelName() returned empty string, should have default model")
	}
}

func TestUsageStruct(t *testing.T) {
	usage := Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	if usage.PromptTokens != 100 {
		t.Errorf("PromptTokens = %d, want 100", usage.PromptTokens)
	}
	if usage.CompletionTokens != 50 {
		t.Errorf("CompletionTokens = %d, want 50", usage.CompletionTokens)
	}
	if usage.TotalTokens != 150 {
		t.Errorf("TotalTokens = %d, want 150", usage.TotalTokens)
	}
}

// Additional constructor error tests
func TestNewOpenAIProvider_EmptyAPIKey(t *testing.T) {
	_, err := NewOpenAIProvider("", "gpt-4")
	if err == nil {
		t.Error("NewOpenAIProvider() with empty API key should return error")
	}
}



func TestOpenAIProvider_Name(t *testing.T) {
	provider, err := NewOpenAIProvider("sk-test", "gpt-4")
	if err != nil {
		t.Fatalf("NewOpenAIProvider() error = %v", err)
	}
	if provider.Name() != "openai" {
		t.Errorf("Name() = %s, want openai", provider.Name())
	}
}

func TestAnthropicProvider_Name(t *testing.T) {
	provider, err := NewAnthropicProvider("sk-ant-test", "claude-3-5-sonnet-20241022")
	if err != nil {
		t.Fatalf("NewAnthropicProvider() error = %v", err)
	}
	if provider.Name() != "anthropic" {
		t.Errorf("Name() = %s, want anthropic", provider.Name())
	}
}

func TestGeminiProvider_Name(t *testing.T) {
	ctx := context.Background()
	provider, err := NewGeminiProvider(ctx, "test-key", "gemini-1.5-pro")
	if err != nil {
		t.Fatalf("NewGeminiProvider() error = %v", err)
	}
	if provider.Name() != "gemini" {
		t.Errorf("Name() = %s, want gemini", provider.Name())
	}
}

func TestOpenAIProvider_ValidateConfig(t *testing.T) {
	provider, err := NewOpenAIProvider("sk-test", "gpt-4")
	if err != nil {
		t.Fatalf("NewOpenAIProvider() error = %v", err)
	}
	if err := provider.ValidateConfig(); err != nil {
		t.Errorf("ValidateConfig() error = %v, want nil", err)
	}
}

func TestAnthropicProvider_ValidateConfig(t *testing.T) {
	provider, err := NewAnthropicProvider("sk-ant-test", "claude-3-5-sonnet-20241022")
	if err != nil {
		t.Fatalf("NewAnthropicProvider() error = %v", err)
	}
	if err := provider.ValidateConfig(); err != nil {
		t.Errorf("ValidateConfig() error = %v, want nil", err)
	}
}

func TestGeminiProvider_ValidateConfig(t *testing.T) {
	ctx := context.Background()
	provider, err := NewGeminiProvider(ctx, "test-key", "gemini-1.5-pro")
	if err != nil {
		t.Fatalf("NewGeminiProvider() error = %v", err)
	}
	if err := provider.ValidateConfig(); err != nil {
		t.Errorf("ValidateConfig() error = %v, want nil", err)
	}
}

func TestGeminiProvider_Close(t *testing.T) {
	ctx := context.Background()
	provider, err := NewGeminiProvider(ctx, "test-key", "gemini-1.5-pro")
	if err != nil {
		t.Fatalf("NewGeminiProvider() error = %v", err)
	}
	if err := provider.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestOpenAIProvider_GetModelName(t *testing.T) {
	provider, err := NewOpenAIProvider("sk-test", "gpt-4-turbo")
	if err != nil {
		t.Fatalf("NewOpenAIProvider() error = %v", err)
	}
	
	modelName := provider.GetModelName()
	if modelName != "gpt-4-turbo" {
		t.Errorf("GetModelName() = %s, want gpt-4-turbo", modelName)
	}
}

func TestAnthropicProvider_GetModelName(t *testing.T) {
	provider, err := NewAnthropicProvider("sk-test", "claude-3-opus")
	if err != nil {
		t.Fatalf("NewAnthropicProvider() error = %v", err)
	}
	
	modelName := provider.GetModelName()
	if modelName != "claude-3-opus" {
		t.Errorf("GetModelName() = %s, want claude-3-opus", modelName)
	}
}

func TestGeminiProvider_GetModelName(t *testing.T) {
	ctx := context.Background()
	provider, err := NewGeminiProvider(ctx, "test-key", "gemini-1.5-pro")
	if err != nil {
		t.Fatalf("NewGeminiProvider() error = %v", err)
	}
	defer provider.Close()
	
	modelName := provider.GetModelName()
	if modelName != "gemini-1.5-pro" {
		t.Errorf("GetModelName() = %s, want gemini-1.5-pro", modelName)
	}
}
