package translator

import (
	"testing"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/provider"
	"github.com/hikanner/jta/internal/terminology"
)

func TestRebuildJSONWithPath(t *testing.T) {
	// Create a simple mock provider (we won't actually use it for this test)
	mockProvider := &provider.OpenAIProvider{}
	engine := &Engine{
		provider: mockProvider,
	}

	tests := []struct {
		name         string
		source       any
		translations map[string]string
		expected     any
	}{
		{
			name: "simple flat object",
			source: map[string]any{
				"title":       "Hello",
				"description": "World",
			},
			translations: map[string]string{
				"title":       "你好",
				"description": "世界",
			},
			expected: map[string]any{
				"title":       "你好",
				"description": "世界",
			},
		},
		{
			name: "nested object",
			source: map[string]any{
				"app": map[string]any{
					"name":    "MyApp",
					"version": "1.0.0",
				},
			},
			translations: map[string]string{
				"app.name": "我的应用",
			},
			expected: map[string]any{
				"app": map[string]any{
					"name":    "我的应用",
					"version": "1.0.0",
				},
			},
		},
		{
			name: "deeply nested object",
			source: map[string]any{
				"settings": map[string]any{
					"theme": map[string]any{
						"dark":  "Dark Mode",
						"light": "Light Mode",
					},
				},
			},
			translations: map[string]string{
				"settings.theme.dark":  "深色模式",
				"settings.theme.light": "浅色模式",
			},
			expected: map[string]any{
				"settings": map[string]any{
					"theme": map[string]any{
						"dark":  "深色模式",
						"light": "浅色模式",
					},
				},
			},
		},
		{
			name: "array values",
			source: map[string]any{
				"items": []any{
					"First",
					"Second",
					"Third",
				},
			},
			translations: map[string]string{
				"items[0]": "第一",
				"items[1]": "第二",
				"items[2]": "第三",
			},
			expected: map[string]any{
				"items": []any{
					"第一",
					"第二",
					"第三",
				},
			},
		},
		{
			name: "mixed types",
			source: map[string]any{
				"text":    "Hello",
				"number":  42,
				"boolean": true,
				"null":    nil,
			},
			translations: map[string]string{
				"text": "你好",
			},
			expected: map[string]any{
				"text":    "你好",
				"number":  42,
				"boolean": true,
				"null":    nil,
			},
		},
		{
			name: "partial translations",
			source: map[string]any{
				"translated":   "Hello",
				"untranslated": "World",
			},
			translations: map[string]string{
				"translated": "你好",
			},
			expected: map[string]any{
				"translated":   "你好",
				"untranslated": "World",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.rebuildJSONWithPath(tt.source, tt.translations, "")

			// Deep comparison
			if !deepEqual(result, tt.expected) {
				t.Errorf("rebuildJSONWithPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// deepEqual compares two interfaces deeply
func deepEqual(a, b any) bool {
	switch va := a.(type) {
	case map[string]any:
		vb, ok := b.(map[string]any)
		if !ok || len(va) != len(vb) {
			return false
		}
		for key, valueA := range va {
			valueB, exists := vb[key]
			if !exists || !deepEqual(valueA, valueB) {
				return false
			}
		}
		return true

	case []any:
		vb, ok := b.([]any)
		if !ok || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !deepEqual(va[i], vb[i]) {
				return false
			}
		}
		return true

	default:
		return a == b
	}
}

func TestExtractTranslatableItems(t *testing.T) {
	mockProvider := &provider.OpenAIProvider{}
	engine := &Engine{
		provider: mockProvider,
	}

	tests := []struct {
		name          string
		data          any
		expectedCount int
	}{
		{
			name: "flat object",
			data: map[string]any{
				"title":       "Hello",
				"description": "World",
			},
			expectedCount: 2,
		},
		{
			name: "nested object",
			data: map[string]any{
				"app": map[string]any{
					"name":    "MyApp",
					"version": "1.0.0",
				},
			},
			expectedCount: 2,
		},
		{
			name: "with non-string values",
			data: map[string]any{
				"text":    "Hello",
				"number":  42,
				"boolean": true,
			},
			expectedCount: 1, // only the text field
		},
		{
			name: "empty strings ignored",
			data: map[string]any{
				"text":  "Hello",
				"empty": "",
			},
			expectedCount: 1, // empty strings are ignored
		},
		{
			name: "array values",
			data: map[string]any{
				"items": []any{
					"First",
					"Second",
				},
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := engine.extractTranslatableItems(tt.data, "")
			if err != nil {
				t.Errorf("extractTranslatableItems() error = %v", err)
				return
			}
			if len(items) != tt.expectedCount {
				t.Errorf("extractTranslatableItems() got %d items, want %d", len(items), tt.expectedCount)
			}
		})
	}
}

func TestNewEngine(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	termManager := terminology.NewManager(mockProvider)

	engine := NewEngine(mockProvider, termManager)

	if engine == nil {
		t.Error("NewEngine() returned nil")
	}

	if engine.provider != mockProvider {
		t.Error("Engine provider not set correctly")
	}

	if engine.batchProcessor == nil {
		t.Error("Engine batchProcessor not initialized")
	}

	if engine.reflectionEngine == nil {
		t.Error("Engine reflectionEngine not initialized")
	}
}

func TestEngine_GetBatchProcessor(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	termManager := terminology.NewManager(mockProvider)
	engine := NewEngine(mockProvider, termManager)

	bp := engine.GetBatchProcessor()
	if bp == nil {
		t.Error("GetBatchProcessor() returned nil")
	}

	if bp != engine.batchProcessor {
		t.Error("GetBatchProcessor() returned wrong processor")
	}
}

func TestEngine_GetReflectionEngine(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	termManager := terminology.NewManager(mockProvider)
	engine := NewEngine(mockProvider, termManager)

	re := engine.GetReflectionEngine()
	if re == nil {
		t.Error("GetReflectionEngine() returned nil")
	}

	if re != engine.reflectionEngine {
		t.Error("GetReflectionEngine() returned wrong engine")
	}
}

func TestEngine_CreateBatches(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	termManager := terminology.NewManager(mockProvider)
	engine := NewEngine(mockProvider, termManager)

	tests := []struct {
		name          string
		items         []domain.BatchItem
		batchSize     int
		expectedCount int
	}{
		{
			name: "single batch",
			items: []domain.BatchItem{
				{Key: "key1", Text: "text1"},
				{Key: "key2", Text: "text2"},
			},
			batchSize:     10,
			expectedCount: 1,
		},
		{
			name: "multiple batches",
			items: []domain.BatchItem{
				{Key: "key1", Text: "text1"},
				{Key: "key2", Text: "text2"},
				{Key: "key3", Text: "text3"},
			},
			batchSize:     2,
			expectedCount: 2,
		},
		{
			name: "default batch size (0)",
			items: []domain.BatchItem{
				{Key: "key1", Text: "text1"},
			},
			batchSize:     0,
			expectedCount: 1,
		},
		{
			name: "negative batch size uses default",
			items: []domain.BatchItem{
				{Key: "key1", Text: "text1"},
			},
			batchSize:     -1,
			expectedCount: 1,
		},
		{
			name:          "empty items",
			items:         []domain.BatchItem{},
			batchSize:     10,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := engine.createBatches(tt.items, tt.batchSize)
			if len(batches) != tt.expectedCount {
				t.Errorf("createBatches() created %d batches, want %d", len(batches), tt.expectedCount)
			}

			// Verify all items are included
			totalItems := 0
			for _, batch := range batches {
				totalItems += len(batch)
			}
			if totalItems != len(tt.items) {
				t.Errorf("Total items in batches = %d, want %d", totalItems, len(tt.items))
			}
		})
	}
}

func TestEngine_ParseKeyPatterns(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	termManager := terminology.NewManager(mockProvider)
	engine := NewEngine(mockProvider, termManager)

	tests := []struct {
		name        string
		patterns    []string
		expectError bool
		expectCount int
	}{
		{
			name:        "single pattern",
			patterns:    []string{"settings.title"},
			expectError: false,
			expectCount: 1,
		},
		{
			name:        "multiple patterns",
			patterns:    []string{"settings.*", "user.name"},
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "empty patterns",
			patterns:    []string{},
			expectError: false,
			expectCount: 0,
		},
		{
			name:        "pattern with comma-separated",
			patterns:    []string{"settings.title,user.name"},
			expectError: false,
			expectCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns, err := engine.parseKeyPatterns(tt.patterns)

			if tt.expectError && err == nil {
				t.Error("parseKeyPatterns() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("parseKeyPatterns() unexpected error = %v", err)
			}

			if len(patterns) != tt.expectCount {
				t.Errorf("parseKeyPatterns() returned %d patterns, want %d", len(patterns), tt.expectCount)
			}
		})
	}
}

func TestEngine_InferContext(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	termManager := terminology.NewManager(mockProvider)
	engine := NewEngine(mockProvider, termManager)
	
	tests := []struct {
		name     string
		keyPath  string
		expected string
	}{
		{
			name:     "title context",
			keyPath:  "app.title",
			expected: "title",
		},
		{
			name:     "name context",
			keyPath:  "user.name",
			expected: "title",
		},
		{
			name:     "label context",
			keyPath:  "form.label",
			expected: "title",
		},
		{
			name:     "description context",
			keyPath:  "product.description",
			expected: "description",
		},
		{
			name:     "desc context",
			keyPath:  "item.desc",
			expected: "description",
		},
		{
			name:     "detail context",
			keyPath:  "page.detail",
			expected: "description",
		},
		{
			name:     "button context",
			keyPath:  "button.submit",
			expected: "action",
		},
		{
			name:     "action context",
			keyPath:  "menu.action",
			expected: "action",
		},
		{
			name:     "cta context",
			keyPath:  "banner.cta",
			expected: "action",
		},
		{
			name:     "error context",
			keyPath:  "validation.error",
			expected: "message",
		},
		{
			name:     "warning context",
			keyPath:  "system.warning",
			expected: "message",
		},
		{
			name:     "alert context",
			keyPath:  "notification.alert",
			expected: "message",
		},
		{
			name:     "general context",
			keyPath:  "some.other.key",
			expected: "general",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.inferContext(tt.keyPath)
			if result != tt.expected {
				t.Errorf("inferContext(%s) = %s, want %s", tt.keyPath, result, tt.expected)
			}
		})
	}
}
