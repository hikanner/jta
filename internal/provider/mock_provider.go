package provider

import (
	"context"
	"fmt"
	"sync"
)

// MockProvider is a mock implementation of AIProvider for testing
type MockProvider struct {
	mu            sync.Mutex
	modelName     string
	responses     []string
	responseIndex int
	shouldError   bool
	errorMessage  string
	callCount     int
}

// NewMockProvider creates a new mock provider
func NewMockProvider(modelName string) *MockProvider {
	return &MockProvider{
		modelName:     modelName,
		responses:     []string{},
		responseIndex: 0,
		shouldError:   false,
		callCount:     0,
	}
}

// AddResponse adds a canned response for the next Complete call
func (m *MockProvider) AddResponse(response string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = append(m.responses, response)
}

// SetError configures the provider to return an error
func (m *MockProvider) SetError(errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = true
	m.errorMessage = errMsg
}

// GetCallCount returns the number of times Complete was called
func (m *MockProvider) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// Complete implements AIProvider interface
func (m *MockProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCount++

	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMessage)
	}

	if m.responseIndex >= len(m.responses) {
		return nil, fmt.Errorf("mock provider: no more responses available (called %d times)", m.callCount)
	}

	response := m.responses[m.responseIndex]
	m.responseIndex++

	return &CompletionResponse{
		Content: response,
		Usage: Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}, nil
}

// GetModelName implements AIProvider interface
func (m *MockProvider) GetModelName() string {
	return m.modelName
}

// Name implements AIProvider interface
func (m *MockProvider) Name() string {
	return "mock"
}

// ValidateConfig implements AIProvider interface
func (m *MockProvider) ValidateConfig() error {
	return nil
}

// Reset resets the mock provider state
func (m *MockProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseIndex = 0
	m.shouldError = false
	m.errorMessage = ""
	m.callCount = 0
}
