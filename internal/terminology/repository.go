package terminology

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hikanner/jta/internal/domain"
)

// Repository defines the interface for terminology storage
type Repository interface {
	Load(path string) (*domain.Terminology, error)
	Save(path string, terminology *domain.Terminology) error
	Exists(path string) bool
}

// JSONRepository implements Repository using JSON files
type JSONRepository struct{}

// NewJSONRepository creates a new JSON repository
func NewJSONRepository() *JSONRepository {
	return &JSONRepository{}
}

// Load loads terminology from a JSON file
func (r *JSONRepository) Load(path string) (*domain.Terminology, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read terminology file: %w", err)
	}

	var terminology domain.Terminology
	err = json.Unmarshal(data, &terminology)
	if err != nil {
		return nil, fmt.Errorf("failed to parse terminology JSON: %w", err)
	}

	return &terminology, nil
}

// Save saves terminology to a JSON file
func (r *JSONRepository) Save(path string, terminology *domain.Terminology) error {
	// Marshal with indentation for readability
	data, err := json.MarshalIndent(terminology, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal terminology: %w", err)
	}

	// Write to file
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write terminology file: %w", err)
	}

	return nil
}

// Exists checks if terminology file exists
func (r *JSONRepository) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
