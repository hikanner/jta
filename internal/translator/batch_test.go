package translator

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/provider"
)

func TestNewBatchProcessor(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	reflectionEngine := NewReflectionEngine(mockProvider)

	bp := NewBatchProcessor(mockProvider, reflectionEngine)

	if bp == nil {
		t.Fatal("NewBatchProcessor() returned nil")
	}
	if bp.provider == nil {
		t.Error("BatchProcessor.provider is nil")
	}
	if bp.formatProtector == nil {
		t.Error("BatchProcessor.formatProtector is nil")
	}
	if bp.reflectionEngine == nil {
		t.Error("BatchProcessor.reflectionEngine is nil")
	}
}

func TestBatchProcessor_SetProgressCallback(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	reflectionEngine := NewReflectionEngine(mockProvider)
	bp := NewBatchProcessor(mockProvider, reflectionEngine)

	callbackCalled := false
	callback := func(event BatchProgressEvent) {
		callbackCalled = true
	}

	bp.SetProgressCallback(callback)

	if bp.progressCallback == nil {
		t.Error("SetProgressCallback() did not set callback")
	}

	// Test that callback can be called
	if bp.progressCallback != nil {
		bp.progressCallback(BatchProgressEvent{Type: "test"})
		if !callbackCalled {
			t.Error("Progress callback was not called")
		}
	}
}

func TestBatchProcessor_ProcessBatches_SingleBatch(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")

	// Add responses for: initial translation + reflection evaluation + improvement
	// Use numeric indices [1], [2], etc. as expected by parseBatchResponse
	mockProvider.AddResponse("[1] 你好\n[2] 世界")
	mockProvider.AddResponse("[1] Translation is good\n[2] Translation is good")
	mockProvider.AddResponse("[1] 你好\n[2] 世界")

	reflectionEngine := NewReflectionEngine(mockProvider)
	bp := NewBatchProcessor(mockProvider, reflectionEngine)

	ctx := context.Background()
	batches := [][]domain.BatchItem{
		{
			{Key: "hello", Text: "Hello", Context: "greeting"},
			{Key: "world", Text: "World", Context: "noun"},
		},
	}

	// Track progress events
	var events []BatchProgressEvent
	bp.SetProgressCallback(func(event BatchProgressEvent) {
		events = append(events, event)
	})

	results, stats, err := bp.ProcessBatches(
		ctx,
		batches,
		"en",
		"zh",
		"",
		nil,
		nil,
		1,
	)

	if err != nil {
		t.Fatalf("ProcessBatches() error = %v", err)
	}

	// Check results
	if len(results) != 2 {
		t.Errorf("ProcessBatches() got %d results, want 2", len(results))
	}

	if results["hello"] == "" {
		t.Error("ProcessBatches() missing translation for 'hello'")
	}
	if results["world"] == "" {
		t.Error("ProcessBatches() missing translation for 'world'")
	}

	// Check stats
	if stats.APICallsCount == 0 {
		t.Error("ProcessBatches() APICallsCount is 0")
	}

	// Check progress events
	if len(events) == 0 {
		t.Error("ProcessBatches() did not emit progress events")
	}

	// Should have at least start and complete events
	hasStart := false
	hasComplete := false
	for _, event := range events {
		if event.Type == "start" {
			hasStart = true
		}
		if event.Type == "complete" {
			hasComplete = true
		}
	}

	if !hasStart {
		t.Error("ProcessBatches() did not emit 'start' event")
	}
	if !hasComplete {
		t.Error("ProcessBatches() did not emit 'complete' event")
	}
}

func TestBatchProcessor_ProcessBatches_MultipleBatches(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")

	// Add responses for 3 batches, each with 3 calls (translate + reflect + improve)
	// Use numeric indices [1], [2], etc.
	// Batch 1
	mockProvider.AddResponse("[1] 文本1\n[2] 文本2")
	mockProvider.AddResponse("[1] Good\n[2] Good")
	mockProvider.AddResponse("[1] 文本1\n[2] 文本2")
	// Batch 2
	mockProvider.AddResponse("[1] 文本3\n[2] 文本4")
	mockProvider.AddResponse("[1] Good\n[2] Good")
	mockProvider.AddResponse("[1] 文本3\n[2] 文本4")
	// Batch 3
	mockProvider.AddResponse("[1] 文本5")
	mockProvider.AddResponse("[1] Good")
	mockProvider.AddResponse("[1] 文本5")

	reflectionEngine := NewReflectionEngine(mockProvider)
	bp := NewBatchProcessor(mockProvider, reflectionEngine)

	ctx := context.Background()
	batches := [][]domain.BatchItem{
		{
			{Key: "key1", Text: "Text 1", Context: ""},
			{Key: "key2", Text: "Text 2", Context: ""},
		},
		{
			{Key: "key3", Text: "Text 3", Context: ""},
			{Key: "key4", Text: "Text 4", Context: ""},
		},
		{
			{Key: "key5", Text: "Text 5", Context: ""},
		},
	}

	results, stats, err := bp.ProcessBatches(
		ctx,
		batches,
		"en",
		"zh",
		"",
		nil,
		nil,
		2, // concurrency = 2
	)

	if err != nil {
		t.Fatalf("ProcessBatches() error = %v", err)
	}

	// Should have 5 translations
	if len(results) != 5 {
		t.Errorf("ProcessBatches() got %d results, want 5", len(results))
	}

	// Check all keys exist
	for i := 1; i <= 5; i++ {
		key := "key" + string(rune('0'+i))
		if _, exists := results[key]; !exists {
			t.Errorf("ProcessBatches() missing translation for '%s'", key)
		}
	}

	// Should have multiple API calls (one per batch + reflection)
	if stats.APICallsCount < 3 {
		t.Errorf("ProcessBatches() APICallsCount = %d, want at least 3", stats.APICallsCount)
	}
}

func TestBatchProcessor_ProcessBatches_WithTerminology(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")

	// Add responses for terminology test
	mockProvider.AddResponse("[1] 你好 API")
	mockProvider.AddResponse("[1] Good translation, API preserved")
	mockProvider.AddResponse("[1] 你好 API")

	reflectionEngine := NewReflectionEngine(mockProvider)
	bp := NewBatchProcessor(mockProvider, reflectionEngine)

	ctx := context.Background()
	batches := [][]domain.BatchItem{
		{
			{Key: "greeting", Text: "Hello API", Context: ""},
		},
	}

	terminology := &domain.Terminology{
		SourceLanguage:  "en",
		PreserveTerms:   []string{"API"},
		ConsistentTerms: []string{"Hello"},
	}

	termTranslation := &domain.TerminologyTranslation{
		SourceLanguage: "en",
		TargetLanguage: "zh",
		Translations: map[string]string{
			"Hello": "你好",
		},
	}

	results, _, err := bp.ProcessBatches(
		ctx,
		batches,
		"en",
		"zh",
		"API must not be translated\nHello = 你好",
		terminology,
		termTranslation,
		1,
	)

	if err != nil {
		t.Fatalf("ProcessBatches() error = %v", err)
	}

	translation := results["greeting"]
	if translation == "" {
		t.Error("ProcessBatches() missing translation")
	}

	// API should be preserved (mock will uppercase it)
	if !strings.Contains(strings.ToUpper(translation), "API") {
		t.Errorf("ProcessBatches() translation '%s' should contain 'API'", translation)
	}
}

func TestBatchProcessor_ProcessBatches_DefaultConcurrency(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")

	// Add responses for default concurrency test
	mockProvider.AddResponse("[1] 测试")
	mockProvider.AddResponse("[1] Good")
	mockProvider.AddResponse("[1] 测试")

	reflectionEngine := NewReflectionEngine(mockProvider)
	bp := NewBatchProcessor(mockProvider, reflectionEngine)

	ctx := context.Background()
	batches := [][]domain.BatchItem{
		{{Key: "test", Text: "Test", Context: ""}},
	}

	// Pass 0 concurrency to test default
	_, _, err := bp.ProcessBatches(
		ctx,
		batches,
		"en",
		"zh",
		"",
		nil,
		nil,
		0, // should default to 3
	)

	if err != nil {
		t.Fatalf("ProcessBatches() error = %v", err)
	}
}

func TestBatchProcessor_ProcessBatches_EmptyBatches(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	reflectionEngine := NewReflectionEngine(mockProvider)
	bp := NewBatchProcessor(mockProvider, reflectionEngine)

	ctx := context.Background()
	batches := [][]domain.BatchItem{}

	results, stats, err := bp.ProcessBatches(
		ctx,
		batches,
		"en",
		"zh",
		"",
		nil,
		nil,
		1,
	)

	if err != nil {
		t.Fatalf("ProcessBatches() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("ProcessBatches() got %d results, want 0", len(results))
	}

	if stats.APICallsCount != 0 {
		t.Errorf("ProcessBatches() APICallsCount = %d, want 0", stats.APICallsCount)
	}
}

func TestBatchProgressEvent_Types(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{"start event", "start"},
		{"complete event", "complete"},
		{"retry event", "retry"},
		{"error event", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := BatchProgressEvent{
				Type:         tt.eventType,
				BatchIndex:   1,
				TotalBatches: 3,
				BatchSize:    5,
				Concurrency:  2,
				Attempt:      1,
				MaxAttempts:  3,
				Duration:     100 * time.Millisecond,
				Tokens:       500,
			}

			if event.Type != tt.eventType {
				t.Errorf("BatchProgressEvent.Type = %s, want %s", event.Type, tt.eventType)
			}
		})
	}
}

func TestBatchStats(t *testing.T) {
	stats := BatchStats{
		APICallsCount: 5,
		TotalTokens:   1000,
	}

	if stats.APICallsCount != 5 {
		t.Errorf("BatchStats.APICallsCount = %d, want 5", stats.APICallsCount)
	}

	if stats.TotalTokens != 1000 {
		t.Errorf("BatchStats.TotalTokens = %d, want 1000", stats.TotalTokens)
	}
}
