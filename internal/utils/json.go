package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bytedance/sonic"
)

// JSONUtil provides utilities for JSON operations
type JSONUtil struct{}

// NewJSONUtil creates a new JSON utility
func NewJSONUtil() *JSONUtil {
	return &JSONUtil{}
}

// LoadJSON loads JSON from a file
func (j *JSONUtil) LoadJSON(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var result map[string]interface{}
	err = sonic.Unmarshal(data, &result)
	if err != nil {
		// Fallback to standard JSON if sonic fails
		err = json.Unmarshal(data, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	return result, nil
}

// SaveJSON saves JSON to a file with pretty formatting
func (j *JSONUtil) SaveJSON(path string, data map[string]interface{}) error {
	// Use standard json for pretty printing
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	err = os.WriteFile(path, bytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DeepCopy creates a deep copy of a map
func (j *JSONUtil) DeepCopy(src map[string]interface{}) map[string]interface{} {
	// Simple implementation using JSON marshal/unmarshal
	bytes, err := sonic.Marshal(src)
	if err != nil {
		return nil
	}

	var dst map[string]interface{}
	err = sonic.Unmarshal(bytes, &dst)
	if err != nil {
		return nil
	}

	return dst
}

// CompareJSON compares two JSON objects and returns if they are equal
func (j *JSONUtil) CompareJSON(a, b interface{}) bool {
	return deepCompare(a, b)
}

// deepCompare performs deep comparison of two values
func deepCompare(a, b interface{}) bool {
	switch va := a.(type) {
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		if len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			bVal, exists := vb[k]
			if !exists {
				return false
			}
			if !deepCompare(v, bVal) {
				return false
			}
		}
		return true
	case []interface{}:
		vb, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !deepCompare(va[i], vb[i]) {
				return false
			}
		}
		return true
	case float64:
		vb, ok := b.(float64)
		if !ok {
			return false
		}
		return va == vb
	case string:
		vb, ok := b.(string)
		if !ok {
			return false
		}
		return va == vb
	case bool:
		vb, ok := b.(bool)
		if !ok {
			return false
		}
		return va == vb
	case nil:
		return b == nil
	default:
		return a == b
	}
}

// GetValue gets a value from nested JSON using dot notation path
func (j *JSONUtil) GetValue(data map[string]interface{}, path string) (interface{}, bool) {
	// Simple implementation - should handle dots in keys properly in production
	// For now, just return if path is in top level
	if val, ok := data[path]; ok {
		return val, true
	}

	return nil, false
}

// SetValue sets a value in nested JSON using dot notation path
func (j *JSONUtil) SetValue(data map[string]interface{}, path string, value interface{}) {
	// Simple implementation - just set at top level for now
	data[path] = value
}
