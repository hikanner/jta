package terminology

import (
	"encoding/json"
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
		return nil, domain.NewIOError("failed to read terminology file", err).
			WithContext("path", path)
	}

	var terminology domain.Terminology
	err = json.Unmarshal(data, &terminology)
	if err != nil {
		return nil, domain.NewFormatError("failed to parse terminology JSON", err).
			WithContext("path", path)
	}

	return &terminology, nil
}

// Save saves terminology to a JSON file
func (r *JSONRepository) Save(path string, terminology *domain.Terminology) error {
	// Marshal with indentation for readability
	data, err := json.MarshalIndent(terminology, "", "  ")
	if err != nil {
		return domain.NewFormatError("failed to marshal terminology", err).
			WithContext("path", path)
	}

	// Write to file
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return domain.NewIOError("failed to write terminology file", err).
			WithContext("path", path)
	}

	return nil
}

// Exists checks if terminology file exists
func (r *JSONRepository) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
