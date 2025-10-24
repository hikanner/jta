package incremental

import (
	"fmt"
)

// DiffResult represents the result of comparing source and target files
type DiffResult struct {
	New       map[string]interface{} // New keys in source
	Modified  map[string]interface{} // Modified keys (source text changed)
	Deleted   []string               // Keys in target but not in source
	Unchanged map[string]interface{} // Keys unchanged
	Stats     DiffStats
}

// DiffStats contains statistics about the diff
type DiffStats struct {
	NewCount       int
	ModifiedCount  int
	DeletedCount   int
	UnchangedCount int
	TotalKeys      int
}

// Translator handles incremental translation
type Translator struct{}

// NewTranslator creates a new incremental translator
func NewTranslator() *Translator {
	return &Translator{}
}

// AnalyzeDiff analyzes the difference between source and target
func (t *Translator) AnalyzeDiff(source, target map[string]interface{}) (*DiffResult, error) {
	result := &DiffResult{
		New:       make(map[string]interface{}),
		Modified:  make(map[string]interface{}),
		Deleted:   []string{},
		Unchanged: make(map[string]interface{}),
	}

	// If target is nil, everything is new
	if target == nil {
		result.New = source
		result.Stats.NewCount = countKeys(source)
		result.Stats.TotalKeys = result.Stats.NewCount
		return result, nil
	}

	// Flatten both JSONs for comparison
	sourceFlat := t.flattenJSON(source, "")
	targetFlat := t.flattenJSON(target, "")

	// Find new and modified keys
	for key, sourceValue := range sourceFlat {
		targetValue, exists := targetFlat[key]

		if !exists {
			// New key
			result.New[key] = sourceValue
			result.Stats.NewCount++
		} else if !t.compareValues(sourceValue, targetValue) {
			// Modified key (source text changed)
			result.Modified[key] = sourceValue
			result.Stats.ModifiedCount++
		} else {
			// Unchanged
			result.Unchanged[key] = targetValue
			result.Stats.UnchangedCount++
		}
	}

	// Find deleted keys
	for key := range targetFlat {
		if _, exists := sourceFlat[key]; !exists {
			result.Deleted = append(result.Deleted, key)
			result.Stats.DeletedCount++
		}
	}

	result.Stats.TotalKeys = len(sourceFlat)

	return result, nil
}

// ShouldTranslate determines if translation is needed based on diff
func (t *Translator) ShouldTranslate(result *DiffResult, force bool) bool {
	if force {
		return true
	}

	// Translate if there are new or modified keys
	return result.Stats.NewCount > 0 || result.Stats.ModifiedCount > 0
}

// MergeDiff merges translated content with unchanged content
func (t *Translator) MergeDiff(
	translated, unchanged map[string]interface{},
	deleted []string,
) map[string]interface{} {
	result := make(map[string]interface{})

	// Add unchanged
	for key, value := range unchanged {
		result[key] = value
	}

	// Add translated (overwrites unchanged if conflict)
	for key, value := range translated {
		result[key] = value
	}

	// Note: deleted keys are automatically excluded

	return result
}

// flattenJSON flattens nested JSON to dot notation
func (t *Translator) flattenJSON(data interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			keyPath := key
			if prefix != "" {
				keyPath = prefix + "." + key
			}
			// Recursively flatten
			subResult := t.flattenJSON(value, keyPath)
			for k, val := range subResult {
				result[k] = val
			}
		}

	case []interface{}:
		for i, value := range v {
			keyPath := fmt.Sprintf("%s[%d]", prefix, i)
			subResult := t.flattenJSON(value, keyPath)
			for k, val := range subResult {
				result[k] = val
			}
		}

	default:
		// Leaf value
		result[prefix] = v
	}

	return result
}

// compareValues compares two values for equality
func (t *Translator) compareValues(a, b interface{}) bool {
	// Simple string comparison for now
	// In production, should handle different types properly
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func countKeys(data map[string]interface{}) int {
	count := 0
	for _, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			count += countKeys(v)
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					count += countKeys(m)
				} else {
					count++
				}
			}
		default:
			count++
		}
	}
	return count
}
