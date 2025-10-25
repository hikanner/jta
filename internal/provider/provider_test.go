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
		Prompt:      "Test prompt",
		Temperature: 0.7,
		MaxTokens:   100,
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
	provider.Complete(ctx, &CompletionRequest{Prompt: "Test"})

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
			name:         "google provider",
			providerType: ProviderTypeGoogle,
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
		Prompt:      "Test prompt",
		Model:       "test-model",
		Temperature: 0.7,
		MaxTokens:   100,
		SystemMsg:   "System message",
	}

	if req.Prompt != "Test prompt" {
		t.Errorf("Prompt = %s, want 'Test prompt'", req.Prompt)
	}

	if req.Model != "test-model" {
		t.Errorf("Model = %s, want 'test-model'", req.Model)
	}

	if req.Temperature != 0.7 {
		t.Errorf("Temperature = %f, want 0.7", req.Temperature)
	}

	if req.MaxTokens != 100 {
		t.Errorf("MaxTokens = %d, want 100", req.MaxTokens)
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
