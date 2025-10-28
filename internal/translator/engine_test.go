package translator

import (
	"testing"

	"github.com/hikanner/jta/internal/provider"
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
