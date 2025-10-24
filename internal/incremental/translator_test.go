package incremental

import (
	"testing"
)

func TestNewTranslator(t *testing.T) {
	tr := NewTranslator()
	if tr == nil {
		t.Fatal("NewTranslator returned nil")
	}
}

func TestAnalyzeDiff_NilTarget(t *testing.T) {
	tr := NewTranslator()

	source := map[string]interface{}{
		"name":  "test",
		"count": float64(42),
	}

	result, err := tr.AnalyzeDiff(source, nil)
	if err != nil {
		t.Fatalf("AnalyzeDiff failed: %v", err)
	}

	if result.Stats.NewCount != 2 {
		t.Errorf("Expected 2 new keys, got %d", result.Stats.NewCount)
	}

	if result.Stats.TotalKeys != 2 {
		t.Errorf("Expected 2 total keys, got %d", result.Stats.TotalKeys)
	}

	if len(result.New) != 2 {
		t.Errorf("Expected 2 items in New, got %d", len(result.New))
	}
}

func TestAnalyzeDiff_NoChanges(t *testing.T) {
	tr := NewTranslator()

	// Both are source files - same content
	source := map[string]interface{}{
		"name": "test",
		"age":  float64(30),
	}

	// Previous source file (not translation - this is source-to-source comparison)
	target := map[string]interface{}{
		"name": "test",
		"age":  float64(30),
	}

	result, err := tr.AnalyzeDiff(source, target)
	if err != nil {
		t.Fatalf("AnalyzeDiff failed: %v", err)
	}

	// Same source text = unchanged
	if result.Stats.UnchangedCount != 2 {
		t.Errorf("Expected 2 unchanged keys, got %d", result.Stats.UnchangedCount)
	}

	if result.Stats.NewCount != 0 {
		t.Errorf("Expected 0 new keys, got %d", result.Stats.NewCount)
	}

	if result.Stats.ModifiedCount != 0 {
		t.Errorf("Expected 0 modified keys, got %d", result.Stats.ModifiedCount)
	}
}

func TestAnalyzeDiff_NewKeys(t *testing.T) {
	tr := NewTranslator()

	source := map[string]interface{}{
		"name":        "test",
		"description": "new field",
		"age":         float64(30),
	}

	// Previous source (before adding description)
	target := map[string]interface{}{
		"name": "test",
		"age":  float64(30),
	}

	result, err := tr.AnalyzeDiff(source, target)
	if err != nil {
		t.Fatalf("AnalyzeDiff failed: %v", err)
	}

	if result.Stats.NewCount != 1 {
		t.Errorf("Expected 1 new key, got %d", result.Stats.NewCount)
	}

	if result.Stats.UnchangedCount != 2 {
		t.Errorf("Expected 2 unchanged keys, got %d", result.Stats.UnchangedCount)
	}

	if _, exists := result.New["description"]; !exists {
		t.Error("Expected 'description' to be in New")
	}
}

func TestAnalyzeDiff_ModifiedKeys(t *testing.T) {
	tr := NewTranslator()

	source := map[string]interface{}{
		"name": "test updated",
		"age":  float64(30),
	}

	target := map[string]interface{}{
		"name": "test",
		"age":  float64(30),
	}

	result, err := tr.AnalyzeDiff(source, target)
	if err != nil {
		t.Fatalf("AnalyzeDiff failed: %v", err)
	}

	if result.Stats.ModifiedCount != 1 {
		t.Errorf("Expected 1 modified key, got %d", result.Stats.ModifiedCount)
	}

	if result.Stats.UnchangedCount != 1 {
		t.Errorf("Expected 1 unchanged key, got %d", result.Stats.UnchangedCount)
	}

	if _, exists := result.Modified["name"]; !exists {
		t.Error("Expected 'name' to be in Modified")
	}
}

func TestAnalyzeDiff_DeletedKeys(t *testing.T) {
	tr := NewTranslator()

	source := map[string]interface{}{
		"name": "test",
	}

	// Previous source had more keys
	target := map[string]interface{}{
		"name": "test",
		"age":  float64(30),
	}

	result, err := tr.AnalyzeDiff(source, target)
	if err != nil {
		t.Fatalf("AnalyzeDiff failed: %v", err)
	}

	if result.Stats.DeletedCount != 1 {
		t.Errorf("Expected 1 deleted key, got %d", result.Stats.DeletedCount)
	}

	if len(result.Deleted) != 1 {
		t.Errorf("Expected 1 item in Deleted, got %d", len(result.Deleted))
	}

	if result.Deleted[0] != "age" {
		t.Errorf("Expected 'age' to be deleted, got %s", result.Deleted[0])
	}
}

func TestAnalyzeDiff_NestedObjects(t *testing.T) {
	tr := NewTranslator()

	source := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"age":  float64(30),
		},
		"settings": map[string]interface{}{
			"theme": "dark",
		},
	}

	target := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"age":  float64(25), // Changed
		},
		"settings": map[string]interface{}{
			"theme": "dark",
		},
	}

	result, err := tr.AnalyzeDiff(source, target)
	if err != nil {
		t.Fatalf("AnalyzeDiff failed: %v", err)
	}

	if result.Stats.ModifiedCount != 1 {
		t.Errorf("Expected 1 modified key, got %d", result.Stats.ModifiedCount)
	}

	if result.Stats.UnchangedCount != 2 {
		t.Errorf("Expected 2 unchanged keys, got %d", result.Stats.UnchangedCount)
	}
}

func TestAnalyzeDiff_Arrays(t *testing.T) {
	tr := NewTranslator()

	source := map[string]interface{}{
		"items": []interface{}{
			"item1",
			"item2",
			"item3",
		},
	}

	target := map[string]interface{}{
		"items": []interface{}{
			"item1",
			"item2",
		},
	}

	result, err := tr.AnalyzeDiff(source, target)
	if err != nil {
		t.Fatalf("AnalyzeDiff failed: %v", err)
	}

	// Third item is new
	if result.Stats.NewCount != 1 {
		t.Errorf("Expected 1 new key, got %d", result.Stats.NewCount)
	}

	if result.Stats.UnchangedCount != 2 {
		t.Errorf("Expected 2 unchanged keys, got %d", result.Stats.UnchangedCount)
	}
}

func TestShouldTranslate(t *testing.T) {
	tr := NewTranslator()

	tests := []struct {
		name     string
		result   *DiffResult
		force    bool
		expected bool
	}{
		{
			name: "has new keys",
			result: &DiffResult{
				Stats: DiffStats{NewCount: 1},
			},
			force:    false,
			expected: true,
		},
		{
			name: "has modified keys",
			result: &DiffResult{
				Stats: DiffStats{ModifiedCount: 1},
			},
			force:    false,
			expected: true,
		},
		{
			name: "no changes",
			result: &DiffResult{
				Stats: DiffStats{
					NewCount:      0,
					ModifiedCount: 0,
				},
			},
			force:    false,
			expected: false,
		},
		{
			name: "no changes but forced",
			result: &DiffResult{
				Stats: DiffStats{
					NewCount:      0,
					ModifiedCount: 0,
				},
			},
			force:    true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tr.ShouldTranslate(tt.result, tt.force)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMergeDiff(t *testing.T) {
	tr := NewTranslator()

	translated := map[string]interface{}{
		"name":        "测试",
		"description": "新字段",
	}

	unchanged := map[string]interface{}{
		"age":   float64(30),
		"theme": "dark",
	}

	deleted := []string{"oldField"}

	result := tr.MergeDiff(translated, unchanged, deleted)

	// Should have 4 keys (2 translated + 2 unchanged)
	if len(result) != 4 {
		t.Errorf("Expected 4 keys in result, got %d", len(result))
	}

	// Check translated keys
	if result["name"] != "测试" {
		t.Errorf("Expected name='测试', got %v", result["name"])
	}

	if result["description"] != "新字段" {
		t.Errorf("Expected description='新字段', got %v", result["description"])
	}

	// Check unchanged keys
	if result["age"] != float64(30) {
		t.Errorf("Expected age=30, got %v", result["age"])
	}

	if result["theme"] != "dark" {
		t.Errorf("Expected theme='dark', got %v", result["theme"])
	}

	// Check deleted key is not present
	if _, exists := result["oldField"]; exists {
		t.Error("Expected 'oldField' to be excluded")
	}
}

func TestMergeDiff_OverwriteUnchanged(t *testing.T) {
	tr := NewTranslator()

	translated := map[string]interface{}{
		"name": "新名称",
	}

	unchanged := map[string]interface{}{
		"name": "旧名称",
		"age":  float64(30),
	}

	deleted := []string{}

	result := tr.MergeDiff(translated, unchanged, deleted)

	// Translated should overwrite unchanged
	if result["name"] != "新名称" {
		t.Errorf("Expected name='新名称', got %v", result["name"])
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(result))
	}
}

func TestFlattenJSON(t *testing.T) {
	tr := NewTranslator()

	tests := []struct {
		name     string
		data     interface{}
		expected map[string]interface{}
	}{
		{
			name: "simple object",
			data: map[string]interface{}{
				"name": "test",
				"age":  float64(30),
			},
			expected: map[string]interface{}{
				"name": "test",
				"age":  float64(30),
			},
		},
		{
			name: "nested object",
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Alice",
					"age":  float64(30),
				},
			},
			expected: map[string]interface{}{
				"user.name": "Alice",
				"user.age":  float64(30),
			},
		},
		{
			name: "array",
			data: map[string]interface{}{
				"items": []interface{}{"a", "b", "c"},
			},
			expected: map[string]interface{}{
				"items[0]": "a",
				"items[1]": "b",
				"items[2]": "c",
			},
		},
		{
			name: "deeply nested",
			data: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"theme": "dark",
					},
				},
			},
			expected: map[string]interface{}{
				"app.settings.theme": "dark",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tr.flattenJSON(tt.data, "")

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(result))
			}

			for key, expectedVal := range tt.expected {
				actualVal, exists := result[key]
				if !exists {
					t.Errorf("Expected key %q to exist", key)
					continue
				}
				if actualVal != expectedVal {
					t.Errorf("Key %q: expected %v, got %v", key, expectedVal, actualVal)
				}
			}
		})
	}
}

func TestCompareValues(t *testing.T) {
	tr := NewTranslator()

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "equal strings",
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
			name:     "equal numbers",
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
			name:     "different types - treated as equal by fmt.Sprintf",
			a:        "42",
			b:        float64(42),
			expected: true, // Current implementation converts both to string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tr.compareValues(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCountKeys(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		expected int
	}{
		{
			name: "simple object",
			data: map[string]interface{}{
				"name": "test",
				"age":  float64(30),
			},
			expected: 2,
		},
		{
			name: "nested object",
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Alice",
					"age":  float64(30),
				},
			},
			expected: 2,
		},
		{
			name: "with array",
			data: map[string]interface{}{
				"items": []interface{}{"a", "b", "c"},
			},
			expected: 3,
		},
		{
			name: "deeply nested",
			data: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"theme": "dark",
						"lang":  "en",
					},
					"version": "1.0.0",
				},
			},
			expected: 3,
		},
		{
			name:     "empty object",
			data:     map[string]interface{}{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countKeys(tt.data)
			if result != tt.expected {
				t.Errorf("Expected %d keys, got %d", tt.expected, result)
			}
		})
	}
}
