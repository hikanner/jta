package terminology

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hikanner/jta/internal/domain"
)

// TermRepository handles terminology definition storage
type TermRepository struct{}

// NewTermRepository creates a new term repository
func NewTermRepository() *TermRepository {
	return &TermRepository{}
}

// Load loads terminology from directory
func (r *TermRepository) Load(terminologyDir string) (*domain.Terminology, error) {
	path := filepath.Join(terminologyDir, "terminology.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read terminology file: %w", err)
	}

	var terminology domain.Terminology
	if err := json.Unmarshal(data, &terminology); err != nil {
		return nil, fmt.Errorf("failed to parse terminology: %w", err)
	}

	return &terminology, nil
}

// Save saves terminology to directory
func (r *TermRepository) Save(terminologyDir string, terminology *domain.Terminology) error {
	// Ensure directory exists
	if err := os.MkdirAll(terminologyDir, 0755); err != nil {
		return fmt.Errorf("failed to create terminology directory: %w", err)
	}

	path := filepath.Join(terminologyDir, "terminology.json")

	data, err := json.MarshalIndent(terminology, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal terminology: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write terminology file: %w", err)
	}

	return nil
}

// Exists checks if terminology file exists
func (r *TermRepository) Exists(terminologyDir string) bool {
	path := filepath.Join(terminologyDir, "terminology.json")
	_, err := os.Stat(path)
	return err == nil
}
