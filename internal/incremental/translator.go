package incremental

import (
	"fmt"
	"maps"
)

// DiffResult represents the result of comparing source and target files
type DiffResult struct {
	New       map[string]any // New keys in source
	Modified  map[string]any // Modified keys (source text changed)
	Deleted   []string       // Keys in target but not in source
	Unchanged map[string]any // Keys unchanged
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
func (t *Translator) AnalyzeDiff(source, target map[string]any) (*DiffResult, error) {
	result := &DiffResult{
		New:       make(map[string]any),
		Modified:  make(map[string]any),
		Deleted:   []string{},
		Unchanged: make(map[string]any),
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
	translated, unchanged map[string]any,
	deleted []string,
) map[string]any {
	result := make(map[string]any)

	// Add unchanged
	maps.Copy(result, unchanged)

	// Add translated (overwrites unchanged if conflict)
	maps.Copy(result, translated)

	// Note: deleted keys are automatically excluded

	return result
}

// flattenJSON flattens nested JSON to dot notation
func (t *Translator) flattenJSON(data any, prefix string) map[string]any {
	result := make(map[string]any)

	switch v := data.(type) {
	case map[string]any:
		for key, value := range v {
			keyPath := key
			if prefix != "" {
				keyPath = prefix + "." + key
			}
			// Recursively flatten
			subResult := t.flattenJSON(value, keyPath)
			maps.Copy(result, subResult)
		}

	case []any:
		for i, value := range v {
			keyPath := fmt.Sprintf("%s[%d]", prefix, i)
			subResult := t.flattenJSON(value, keyPath)
			maps.Copy(result, subResult)
		}

	default:
		// Leaf value
		result[prefix] = v
	}

	return result
}

// compareValues compares two values for equality
func (t *Translator) compareValues(a, b any) bool {
	// Simple string comparison for now
	// In production, should handle different types properly
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func countKeys(data map[string]any) int {
	count := 0
	for _, value := range data {
		switch v := value.(type) {
		case map[string]any:
			count += countKeys(v)
		case []any:
			for _, item := range v {
				if m, ok := item.(map[string]any); ok {
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
