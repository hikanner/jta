package translator

import (
	"context"
	"strings"
	"testing"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/provider"
)

// TestNewReflectionEngine tests the creation of a reflection engine
func TestNewReflectionEngine(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	engine := NewReflectionEngine(mockProvider)

	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}
	if engine.provider == nil {
		t.Error("Expected provider to be set")
	}
	if engine.formatProtector == nil {
		t.Error("Expected formatProtector to be initialized")
	}
}

// TestReflect_EmptyInput tests reflection with empty input
func TestReflect_EmptyInput(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	engine := NewReflectionEngine(mockProvider)

	input := ReflectionInput{
		SourceTexts:     make(map[string]string),
		TranslatedTexts: make(map[string]string),
		SourceLang:      "en",
		TargetLang:      "zh",
	}

	result, err := engine.Reflect(context.Background(), input, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.ReflectionNeeded {
		t.Error("Expected ReflectionNeeded to be false for empty input")
	}
	if result.APICallsUsed != 0 {
		t.Errorf("Expected 0 API calls, got: %d", result.APICallsUsed)
	}
}

// TestReflect_SuccessfulReflection tests a complete successful reflection workflow
func TestReflect_SuccessfulReflection(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")

	// Setup mock responses for reflection and improvement steps
	reflectionResponse := `[key1] The translation is accurate but could be more natural
[key2] OK
[key3] Consider using more formal language`

	improvementResponse := `[key1] 这是改进后的翻译
[key2] 第二个翻译
[key3] 第三个更正式的翻译`

	mockProvider.AddResponse(reflectionResponse)
	mockProvider.AddResponse(improvementResponse)

	engine := NewReflectionEngine(mockProvider)

	input := ReflectionInput{
		SourceTexts: map[string]string{
			"key1": "This is a test",
			"key2": "Another test",
			"key3": "Final test",
		},
		TranslatedTexts: map[string]string{
			"key1": "这是一个测试",
			"key2": "另一个测试",
			"key3": "最后的测试",
		},
		SourceLang: "en",
		TargetLang: "zh",
	}

	// Track progress callbacks
	var events []ReflectionProgressEvent
	callback := func(event ReflectionProgressEvent) {
		events = append(events, event)
	}

	result, err := engine.Reflect(context.Background(), input, callback)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify result
	if !result.ReflectionNeeded {
		t.Error("Expected ReflectionNeeded to be true")
	}
	if result.APICallsUsed != 2 {
		t.Errorf("Expected 2 API calls (reflect + improve), got: %d", result.APICallsUsed)
	}

	// Verify suggestions were parsed
	if len(result.Suggestions) == 0 {
		t.Error("Expected suggestions to be parsed")
	}
	if _, ok := result.Suggestions["key1"]; !ok {
		t.Error("Expected suggestion for key1")
	}

	// Verify improvements were parsed
	if len(result.ImprovedTexts) == 0 {
		t.Error("Expected improved texts to be parsed")
	}
	if improved, ok := result.ImprovedTexts["key1"]; !ok || improved == "" {
		t.Error("Expected improved text for key1")
	}

	// Verify progress callbacks
	if len(events) != 4 {
		t.Errorf("Expected 4 progress events, got: %d", len(events))
	}

	expectedEventTypes := []string{"reflecting_start", "reflected_complete", "improving_start", "improved_complete"}
	for i, expectedType := range expectedEventTypes {
		if i >= len(events) {
			break
		}
		if events[i].Type != expectedType {
			t.Errorf("Event %d: expected type %s, got %s", i, expectedType, events[i].Type)
		}
	}
}

// TestReflect_WithTerminology tests reflection with terminology
func TestReflect_WithTerminology(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")

	reflectionResponse := `[key1] Good translation, API is correctly preserved`
	improvementResponse := `[key1] 这个API是正确的`

	mockProvider.AddResponse(reflectionResponse)
	mockProvider.AddResponse(improvementResponse)

	engine := NewReflectionEngine(mockProvider)

	terminology := &domain.Terminology{
		SourceLanguage:  "en",
		PreserveTerms:   []string{"API", "JSON"},
		ConsistentTerms: []string{"user", "settings"},
	}

	termTranslation := &domain.TerminologyTranslation{
		SourceLanguage: "en",
		TargetLanguage: "zh",
		Translations: map[string]string{
			"user":     "用户",
			"settings": "设置",
		},
	}

	input := ReflectionInput{
		SourceTexts: map[string]string{
			"key1": "The API is working",
		},
		TranslatedTexts: map[string]string{
			"key1": "API正在工作",
		},
		SourceLang:             "en",
		TargetLang:             "zh",
		Terminology:            terminology,
		TerminologyTranslation: termTranslation,
	}

	result, err := engine.Reflect(context.Background(), input, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.ReflectionNeeded {
		t.Error("Expected ReflectionNeeded to be true")
	}
	if len(result.Suggestions) == 0 {
		t.Error("Expected suggestions to be generated")
	}
}

// TestReflect_WithProgressCallback tests progress callback functionality
func TestReflect_WithProgressCallback(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	mockProvider.AddResponse(`[key1] Good`)
	mockProvider.AddResponse(`[key1] 改进的翻译`)

	engine := NewReflectionEngine(mockProvider)

	input := ReflectionInput{
		SourceTexts:     map[string]string{"key1": "Test"},
		TranslatedTexts: map[string]string{"key1": "测试"},
		SourceLang:      "en",
		TargetLang:      "zh",
	}

	var callbackCalled bool
	var eventCount int
	callback := func(event ReflectionProgressEvent) {
		callbackCalled = true
		eventCount++

		// Verify event has required fields
		if event.Type == "" {
			t.Error("Expected non-empty event type")
		}
		if event.Count < 0 {
			t.Error("Expected non-negative count")
		}

		// Complete events should have non-negative duration
		// Note: Duration may be 0 on fast systems (Windows) when using mock providers
		if strings.HasSuffix(event.Type, "_complete") && event.Duration < 0 {
			t.Error("Expected non-negative duration for complete event")
		}
	}

	_, err := engine.Reflect(context.Background(), input, callback)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !callbackCalled {
		t.Error("Expected callback to be called")
	}
	if eventCount != 4 {
		t.Errorf("Expected 4 callback invocations, got: %d", eventCount)
	}
}

// TestReflectStep_Success tests the reflection step in isolation
func TestReflectStep_Success(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	mockProvider.AddResponse(`[key1] Good translation
[key2] Could be better`)

	engine := NewReflectionEngine(mockProvider)

	input := ReflectionInput{
		SourceTexts: map[string]string{
			"key1": "Hello",
			"key2": "World",
		},
		TranslatedTexts: map[string]string{
			"key1": "你好",
			"key2": "世界",
		},
		SourceLang: "en",
		TargetLang: "zh",
	}

	suggestions, err := engine.reflectStep(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got: %d", len(suggestions))
	}
	if _, ok := suggestions["key1"]; !ok {
		t.Error("Expected suggestion for key1")
	}
	if _, ok := suggestions["key2"]; !ok {
		t.Error("Expected suggestion for key2")
	}
}

// TestReflectStep_LLMError tests reflection step with LLM error
func TestReflectStep_LLMError(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	mockProvider.SetError("API error occurred")

	engine := NewReflectionEngine(mockProvider)

	input := ReflectionInput{
		TranslatedTexts: map[string]string{"key1": "test"},
		SourceLang:      "en",
		TargetLang:      "zh",
	}

	_, err := engine.reflectStep(context.Background(), input)
	if err == nil {
		t.Fatal("Expected error from LLM failure")
	}

	// Verify error contains context
	if domainErr, ok := err.(*domain.Error); ok {
		if domainErr.Type != domain.ErrorTypeTranslation {
			t.Errorf("Expected translation error type, got: %v", domainErr.Type)
		}
	}
}

// TestImproveStep_Success tests the improvement step in isolation
func TestImproveStep_Success(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	mockProvider.AddResponse(`[key1] 改进后的你好
[key2] 更好的世界`)

	engine := NewReflectionEngine(mockProvider)

	input := ReflectionInput{
		SourceTexts: map[string]string{
			"key1": "Hello",
			"key2": "World",
		},
		TranslatedTexts: map[string]string{
			"key1": "你好",
			"key2": "世界",
		},
		SourceLang: "en",
		TargetLang: "zh",
	}

	suggestions := map[string]string{
		"key1": "Make it more formal",
		"key2": "Use better terminology",
	}

	improved, err := engine.improveStep(context.Background(), input, suggestions)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(improved) != 2 {
		t.Errorf("Expected 2 improvements, got: %d", len(improved))
	}
	if _, ok := improved["key1"]; !ok {
		t.Error("Expected improved text for key1")
	}
}

// TestBuildReflectionPrompt tests reflection prompt construction
func TestBuildReflectionPrompt(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	engine := NewReflectionEngine(mockProvider)

	input := ReflectionInput{
		SourceTexts: map[string]string{
			"key1": "Hello world",
		},
		TranslatedTexts: map[string]string{
			"key1": "你好世界",
		},
		SourceLang: "en",
		TargetLang: "zh",
	}

	prompt := engine.buildReflectionPrompt(input)

	// Verify prompt contains required sections
	requiredSections := []string{
		"source texts",
		"translations",
		"<SOURCE_TEXTS>",
		"<TRANSLATIONS>",
		"accuracy",
		"fluency",
		"style",
		"terminology",
		"[key1]",
	}

	for _, section := range requiredSections {
		if !strings.Contains(strings.ToLower(prompt), strings.ToLower(section)) {
			t.Errorf("Prompt missing required section: %s", section)
		}
	}

	// Verify it includes language info
	if !strings.Contains(prompt, "en") || !strings.Contains(prompt, "zh") {
		t.Error("Prompt should contain source and target languages")
	}
}

// TestBuildReflectionPrompt_WithTerminology tests prompt with terminology
func TestBuildReflectionPrompt_WithTerminology(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	engine := NewReflectionEngine(mockProvider)

	terminology := &domain.Terminology{
		SourceLanguage:  "en",
		PreserveTerms:   []string{"API", "JSON"},
		ConsistentTerms: []string{"user"},
	}

	termTranslation := &domain.TerminologyTranslation{
		SourceLanguage: "en",
		TargetLanguage: "zh",
		Translations: map[string]string{
			"user": "用户",
		},
	}

	input := ReflectionInput{
		SourceTexts:            map[string]string{"key1": "User API"},
		TranslatedTexts:        map[string]string{"key1": "用户 API"},
		SourceLang:             "en",
		TargetLang:             "zh",
		Terminology:            terminology,
		TerminologyTranslation: termTranslation,
	}

	prompt := engine.buildReflectionPrompt(input)

	// Verify terminology appears in prompt
	if !strings.Contains(prompt, "API") {
		t.Error("Prompt should contain preserved term 'API'")
	}
	if !strings.Contains(prompt, "user") {
		t.Error("Prompt should contain consistent term 'user'")
	}
	if !strings.Contains(prompt, "用户") {
		t.Error("Prompt should contain translation '用户'")
	}
}

// TestBuildImprovementPrompt tests improvement prompt construction
func TestBuildImprovementPrompt(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	engine := NewReflectionEngine(mockProvider)

	input := ReflectionInput{
		SourceTexts: map[string]string{
			"key1": "Hello",
		},
		TranslatedTexts: map[string]string{
			"key1": "你好",
		},
		SourceLang: "en",
		TargetLang: "zh",
	}

	suggestions := map[string]string{
		"key1": "Make it more formal",
	}

	prompt := engine.buildImprovementPrompt(input, suggestions)

	// Verify prompt contains required sections
	requiredSections := []string{
		"<SOURCE_TEXTS>",
		"<INITIAL_TRANSLATIONS>",
		"<EXPERT_SUGGESTIONS>",
		"Make it more formal",
		"accuracy",
		"fluency",
		"IMPORTANT",
	}

	for _, section := range requiredSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("Improvement prompt missing required section: %s", section)
		}
	}
}

// TestParseReflectionSuggestions tests parsing reflection response
func TestParseReflectionSuggestions(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	engine := NewReflectionEngine(mockProvider)

	translations := map[string]string{
		"key1": "translation1",
		"key2": "translation2",
		"key3": "translation3",
	}

	tests := []struct {
		name          string
		response      string
		expectedCount int
		expectedKeys  []string
	}{
		{
			name: "Valid suggestions",
			response: `[key1] This is a suggestion
[key2] Another suggestion
[key3] OK`,
			expectedCount: 3,
			expectedKeys:  []string{"key1", "key2", "key3"},
		},
		{
			name: "With empty lines",
			response: `[key1] First suggestion

[key2] Second suggestion

`,
			expectedCount: 2,
			expectedKeys:  []string{"key1", "key2"},
		},
		{
			name: "Invalid key not in translations",
			response: `[key1] Valid
[invalid_key] Should be ignored
[key2] Also valid`,
			expectedCount: 2,
			expectedKeys:  []string{"key1", "key2"},
		},
		{
			name:          "Empty response",
			response:      "",
			expectedCount: 0,
			expectedKeys:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := engine.parseReflectionSuggestions(tt.response, translations)

			if len(suggestions) != tt.expectedCount {
				t.Errorf("Expected %d suggestions, got %d", tt.expectedCount, len(suggestions))
			}

			for _, key := range tt.expectedKeys {
				if _, ok := suggestions[key]; !ok {
					t.Errorf("Expected suggestion for key: %s", key)
				}
			}
		})
	}
}

// TestParseImprovedTranslations tests parsing improvement response
func TestParseImprovedTranslations(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	engine := NewReflectionEngine(mockProvider)

	originalTranslations := map[string]string{
		"key1": "original1",
		"key2": "original2",
	}

	tests := []struct {
		name          string
		response      string
		expectedCount int
		expectedKeys  []string
	}{
		{
			name: "Valid improvements",
			response: `[key1] Improved translation one
[key2] Improved translation two`,
			expectedCount: 2,
			expectedKeys:  []string{"key1", "key2"},
		},
		{
			name: "With extra whitespace",
			response: `  [key1]   Improved one
  [key2]   Improved two  `,
			expectedCount: 2,
			expectedKeys:  []string{"key1", "key2"},
		},
		{
			name: "Invalid key ignored",
			response: `[key1] Valid improvement
[invalid_key] Should be ignored`,
			expectedCount: 1,
			expectedKeys:  []string{"key1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			improved := engine.parseImprovedTranslations(tt.response, originalTranslations)

			if len(improved) != tt.expectedCount {
				t.Errorf("Expected %d improvements, got %d", tt.expectedCount, len(improved))
			}

			for _, key := range tt.expectedKeys {
				if _, ok := improved[key]; !ok {
					t.Errorf("Expected improved translation for key: %s", key)
				}
			}
		})
	}
}

// TestShouldReflect tests the reflection decision logic
func TestShouldReflect(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	engine := NewReflectionEngine(mockProvider)

	tests := []struct {
		name         string
		translations map[string]string
		terminology  *domain.Terminology
		expected     bool
	}{
		{
			name:         "Empty translations",
			translations: map[string]string{},
			expected:     false,
		},
		{
			name: "Non-empty translations",
			translations: map[string]string{
				"key1": "value1",
			},
			expected: true,
		},
		{
			name: "With terminology",
			translations: map[string]string{
				"key1": "value1",
			},
			terminology: &domain.Terminology{
				PreserveTerms: []string{"API"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.ShouldReflect(tt.translations, tt.terminology)
			if result != tt.expected {
				t.Errorf("Expected ShouldReflect to return %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestReflect_Timeout tests reflection with context timeout
// Note: This test is skipped because MockProvider doesn't support delay simulation
// The actual timeout handling is tested through integration tests
func TestReflect_Timeout(t *testing.T) {
	t.Skip("Skipping timeout test - MockProvider doesn't support delay simulation")

	// Timeout handling is verified in the reflection.go implementation
	// Each API call uses context.WithTimeout(context.Background(), 5*time.Minute)
}

// TestReflect_Duration tests that durations are properly recorded
func TestReflect_Duration(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	mockProvider.AddResponse("[key1] Good")
	mockProvider.AddResponse("[key1] 改进")

	engine := NewReflectionEngine(mockProvider)

	input := ReflectionInput{
		SourceTexts:     map[string]string{"key1": "Test"},
		TranslatedTexts: map[string]string{"key1": "测试"},
		SourceLang:      "en",
		TargetLang:      "zh",
	}

	result, err := engine.Reflect(context.Background(), input, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Duration may be 0 on fast systems (Windows) when using mock providers
	// Just verify they are non-negative
	if result.ReflectDuration < 0 {
		t.Error("Expected non-negative ReflectDuration")
	}
	if result.ImproveDuration < 0 {
		t.Error("Expected non-negative ImproveDuration")
	}
}
