package terminology

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hikanner/jta/internal/domain"
)

// TranslationRepository handles terminology translation storage
type TranslationRepository struct{}

// NewTranslationRepository creates a new translation repository
func NewTranslationRepository() *TranslationRepository {
	return &TranslationRepository{}
}

// Load loads terminology translation from directory
func (r *TranslationRepository) Load(terminologyDir string, targetLang string) (*domain.TerminologyTranslation, error) {
	path := filepath.Join(terminologyDir, fmt.Sprintf("terminology.%s.json", targetLang))

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read translation file: %w", err)
	}

	var translation domain.TerminologyTranslation
	if err := json.Unmarshal(data, &translation); err != nil {
		return nil, fmt.Errorf("failed to parse translation: %w", err)
	}

	return &translation, nil
}

// Save saves terminology translation to directory
func (r *TranslationRepository) Save(terminologyDir string, translation *domain.TerminologyTranslation) error {
	// Ensure directory exists
	if err := os.MkdirAll(terminologyDir, 0755); err != nil {
		return fmt.Errorf("failed to create terminology directory: %w", err)
	}

	path := filepath.Join(terminologyDir, fmt.Sprintf("terminology.%s.json", translation.TargetLanguage))

	data, err := json.MarshalIndent(translation, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal translation: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write translation file: %w", err)
	}

	return nil
}

// Exists checks if translation file exists
func (r *TranslationRepository) Exists(terminologyDir string, targetLang string) bool {
	path := filepath.Join(terminologyDir, fmt.Sprintf("terminology.%s.json", targetLang))
	_, err := os.Stat(path)
	return err == nil
}
