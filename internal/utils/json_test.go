package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewJSONUtil(t *testing.T) {
	util := NewJSONUtil()
	if util == nil {
		t.Fatal("NewJSONUtil returned nil")
	}
}

func TestLoadJSON(t *testing.T) {
	util := NewJSONUtil()

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "jta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // Cleanup, error not critical

	tests := []struct {
		name      string
		content   string
		expectErr bool
		validate  func(t *testing.T, data map[string]any)
	}{
		{
			name:      "valid simple JSON",
			content:   `{"name": "test", "count": 42}`,
			expectErr: false,
			validate: func(t *testing.T, data map[string]any) {
				if data["name"] != "test" {
					t.Errorf("Expected name=test, got %v", data["name"])
				}
				if data["count"].(float64) != 42 {
					t.Errorf("Expected count=42, got %v", data["count"])
				}
			},
		},
		{
			name: "valid nested JSON",
			content: `{
				"user": {
					"name": "Alice",
					"age": 30
				},
				"settings": {
					"theme": "dark"
				}
			}`,
			expectErr: false,
			validate: func(t *testing.T, data map[string]any) {
				user, ok := data["user"].(map[string]any)
				if !ok {
					t.Error("Expected user to be a map")
					return
				}
				if user["name"] != "Alice" {
					t.Errorf("Expected user.name=Alice, got %v", user["name"])
				}
			},
		},
		{
			name: "valid array JSON",
			content: `{
				"items": [1, 2, 3],
				"names": ["a", "b", "c"]
			}`,
			expectErr: false,
			validate: func(t *testing.T, data map[string]any) {
				items, ok := data["items"].([]any)
				if !ok {
					t.Error("Expected items to be an array")
					return
				}
				if len(items) != 3 {
					t.Errorf("Expected 3 items, got %d", len(items))
				}
			},
		},
		{
			name:      "invalid JSON",
			content:   `{"name": invalid}`,
			expectErr: true,
			validate:  nil,
		},
		{
			name:      "empty JSON object",
			content:   `{}`,
			expectErr: false,
			validate: func(t *testing.T, data map[string]any) {
				if len(data) != 0 {
					t.Errorf("Expected empty map, got %d items", len(data))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test file
			testFile := filepath.Join(tmpDir, "test.json")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Load JSON
			data, err := util.LoadJSON(testFile)
			if (err != nil) != tt.expectErr {
				t.Errorf("LoadJSON() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr && tt.validate != nil {
				tt.validate(t, data)
			}
		})
	}
}

func TestLoadJSON_NonExistentFile(t *testing.T) {
	util := NewJSONUtil()
	_, err := util.LoadJSON("/non/existent/file.json")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestSaveJSON(t *testing.T) {
	util := NewJSONUtil()

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "jta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // Cleanup, error not critical

	tests := []struct {
		name      string
		data      map[string]any
		expectErr bool
	}{
		{
			name: "simple data",
			data: map[string]any{
				"name":  "test",
				"count": float64(42),
			},
			expectErr: false,
		},
		{
			name: "nested data",
			data: map[string]any{
				"user": map[string]any{
					"name": "Alice",
					"age":  float64(30),
				},
				"settings": map[string]any{
					"theme": "dark",
				},
			},
			expectErr: false,
		},
		{
			name: "with arrays",
			data: map[string]any{
				"items": []any{float64(1), float64(2), float64(3)},
				"names": []any{"a", "b", "c"},
			},
			expectErr: false,
		},
		{
			name:      "empty data",
			data:      map[string]any{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "output.json")

			// Save JSON
			err := util.SaveJSON(testFile, tt.data)
			if (err != nil) != tt.expectErr {
				t.Errorf("SaveJSON() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr {
				// Verify file exists and is readable
				if _, err := os.Stat(testFile); os.IsNotExist(err) {
					t.Error("Saved file does not exist")
					return
				}

				// Load back and compare
				loaded, err := util.LoadJSON(testFile)
				if err != nil {
					t.Errorf("Failed to load saved file: %v", err)
					return
				}

				if !deepEqual(tt.data, loaded) {
					t.Error("Saved and loaded data do not match")
				}
			}
		})
	}
}

func TestSaveJSON_InvalidPath(t *testing.T) {
	util := NewJSONUtil()
	data := map[string]any{"test": "value"}
	err := util.SaveJSON("/invalid/path/to/file.json", data)
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}

func TestDeepCopy(t *testing.T) {
	util := NewJSONUtil()

	tests := []struct {
		name string
		data map[string]any
	}{
		{
			name: "simple data",
			data: map[string]any{
				"name":  "test",
				"count": float64(42),
			},
		},
		{
			name: "nested data",
			data: map[string]any{
				"user": map[string]any{
					"name": "Alice",
					"age":  float64(30),
				},
				"settings": map[string]any{
					"theme": "dark",
				},
			},
		},
		{
			name: "with arrays",
			data: map[string]any{
				"items": []any{float64(1), float64(2), float64(3)},
				"names": []any{"a", "b", "c"},
			},
		},
		{
			name: "empty data",
			data: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copied := util.DeepCopy(tt.data)
			if copied == nil {
				t.Fatal("DeepCopy returned nil")
			}

			// Verify content is the same
			if !util.CompareJSON(tt.data, copied) {
				t.Error("Copied data does not match original")
			}

			// Modify copied data to verify it's a deep copy
			if len(copied) > 0 {
				copied["__modified__"] = true
				if _, exists := tt.data["__modified__"]; exists {
					t.Error("Modifying copy affected original - not a deep copy")
				}
			}
		})
	}
}

func TestCompareJSON(t *testing.T) {
	util := NewJSONUtil()

	tests := []struct {
		name     string
		a        any
		b        any
		expected bool
	}{
		{
			name:     "identical simple objects",
			a:        map[string]any{"name": "test", "count": float64(42)},
			b:        map[string]any{"name": "test", "count": float64(42)},
			expected: true,
		},
		{
			name:     "different simple objects",
			a:        map[string]any{"name": "test1"},
			b:        map[string]any{"name": "test2"},
			expected: false,
		},
		{
			name: "identical nested objects",
			a: map[string]any{
				"user": map[string]any{"name": "Alice"},
			},
			b: map[string]any{
				"user": map[string]any{"name": "Alice"},
			},
			expected: true,
		},
		{
			name: "different nested objects",
			a: map[string]any{
				"user": map[string]any{"name": "Alice"},
			},
			b: map[string]any{
				"user": map[string]any{"name": "Bob"},
			},
			expected: false,
		},
		{
			name:     "identical arrays",
			a:        []any{float64(1), float64(2), float64(3)},
			b:        []any{float64(1), float64(2), float64(3)},
			expected: true,
		},
		{
			name:     "different arrays",
			a:        []any{float64(1), float64(2), float64(3)},
			b:        []any{float64(1), float64(2), float64(4)},
			expected: false,
		},
		{
			name:     "identical strings",
			a:        "test",
			b:        "test",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "test1",
			b:        "test2",
			expected: false,
		},
		{
			name:     "identical numbers",
			a:        float64(42),
			b:        float64(42),
			expected: true,
		},
		{
			name:     "different numbers",
			a:        float64(42),
			b:        float64(43),
			expected: false,
		},
		{
			name:     "empty objects",
			a:        map[string]any{},
			b:        map[string]any{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.CompareJSON(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("CompareJSON() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	util := NewJSONUtil()

	data := map[string]any{
		"name":  "test",
		"count": float64(42),
		"user": map[string]any{
			"name": "Alice",
		},
	}

	tests := []struct {
		name      string
		path      string
		expectOk  bool
		expectVal any
	}{
		{
			name:      "existing top-level string",
			path:      "name",
			expectOk:  true,
			expectVal: "test",
		},
		{
			name:      "existing top-level number",
			path:      "count",
			expectOk:  true,
			expectVal: float64(42),
		},
		{
			name:     "non-existent key",
			path:     "nonexistent",
			expectOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := util.GetValue(data, tt.path)
			if ok != tt.expectOk {
				t.Errorf("GetValue() ok = %v, expected %v", ok, tt.expectOk)
			}
			if tt.expectOk && val != tt.expectVal {
				t.Errorf("GetValue() value = %v, expected %v", val, tt.expectVal)
			}
		})
	}
}

func TestSetValue(t *testing.T) {
	util := NewJSONUtil()

	tests := []struct {
		name  string
		path  string
		value any
	}{
		{
			name:  "set string value",
			path:  "name",
			value: "test",
		},
		{
			name:  "set number value",
			path:  "count",
			value: 42,
		},
		{
			name: "set object value",
			path: "user",
			value: map[string]any{
				"name": "Alice",
			},
		},
		{
			name:  "set array value",
			path:  "items",
			value: []any{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]any{}
			util.SetValue(data, tt.path, tt.value)

			val, ok := data[tt.path]
			if !ok {
				t.Error("SetValue() did not set the value")
				return
			}

			if !util.CompareJSON(val, tt.value) {
				t.Errorf("SetValue() set wrong value: got %v, expected %v", val, tt.value)
			}
		})
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	util := NewJSONUtil()

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "jta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // Cleanup, error not critical

	original := map[string]any{
		"app": map[string]any{
			"name":        "Jta",
			"description": "AI-powered JSON translation tool",
			"version":     "1.0.0",
		},
		"settings": map[string]any{
			"language": "en",
			"theme":    "dark",
		},
		"items": []any{
			map[string]any{"id": float64(1), "name": "Item 1"},
			map[string]any{"id": float64(2), "name": "Item 2"},
		},
	}

	testFile := filepath.Join(tmpDir, "roundtrip.json")

	// Save
	err = util.SaveJSON(testFile, original)
	if err != nil {
		t.Fatalf("Failed to save JSON: %v", err)
	}

	// Load
	loaded, err := util.LoadJSON(testFile)
	if err != nil {
		t.Fatalf("Failed to load JSON: %v", err)
	}

	// Compare
	if !deepEqual(original, loaded) {
		t.Error("Round-trip data does not match original")
	}
}

// deepEqual performs deep comparison of two values
func deepEqual(a, b any) bool {
	switch va := a.(type) {
	case map[string]any:
		vb, ok := b.(map[string]any)
		if !ok {
			return false
		}
		if len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !deepEqual(v, vb[k]) {
				return false
			}
		}
		return true
	case []any:
		vb, ok := b.([]any)
		if !ok {
			return false
		}
		if len(va) != len(vb) {
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

func TestDeepCopy_ErrorCases(t *testing.T) {
	util := NewJSONUtil()
	
	// Test with nil input - should handle gracefully
	t.Run("nil input", func(t *testing.T) {
		result := util.DeepCopy(nil)
		if result != nil {
			t.Errorf("DeepCopy(nil) = %v, want nil", result)
		}
	})
	
	// Test with complex nested structure
	t.Run("deeply nested structure", func(t *testing.T) {
		data := map[string]any{
			"level1": map[string]any{
				"level2": map[string]any{
					"level3": map[string]any{
						"value": "deep",
						"array": []any{
							map[string]any{"nested": "item"},
						},
					},
				},
			},
		}
		
		copied := util.DeepCopy(data)
		if copied == nil {
			t.Fatal("DeepCopy returned nil for deeply nested structure")
		}
		
		if !util.CompareJSON(data, copied) {
			t.Error("Deeply nested copy does not match original")
		}
	})
}

func TestCompareJSON_EdgeCases(t *testing.T) {
	util := NewJSONUtil()
	
	tests := []struct {
		name     string
		a        any
		b        any
		expected bool
	}{
		{
			name:     "compare map with non-map",
			a:        map[string]any{"key": "value"},
			b:        "not a map",
			expected: false,
		},
		{
			name:     "compare array with non-array",
			a:        []any{1, 2, 3},
			b:        "not an array",
			expected: false,
		},
		{
			name:     "compare arrays with different lengths",
			a:        []any{1, 2, 3},
			b:        []any{1, 2},
			expected: false,
		},
		{
			name:     "compare maps with different lengths",
			a:        map[string]any{"a": 1, "b": 2},
			b:        map[string]any{"a": 1},
			expected: false,
		},
		{
			name:     "compare map with missing key in b",
			a:        map[string]any{"a": 1, "b": 2},
			b:        map[string]any{"a": 1, "c": 2},
			expected: false,
		},
		{
			name:     "compare booleans true",
			a:        true,
			b:        true,
			expected: true,
		},
		{
			name:     "compare booleans false",
			a:        false,
			b:        false,
			expected: true,
		},
		{
			name:     "compare different booleans",
			a:        true,
			b:        false,
			expected: false,
		},
		{
			name:     "compare nil values",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "compare nil with non-nil",
			a:        nil,
			b:        "something",
			expected: false,
		},
		{
			name:     "compare float with string",
			a:        float64(42),
			b:        "42",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.CompareJSON(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("CompareJSON(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
