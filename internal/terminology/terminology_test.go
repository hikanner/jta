package terminology

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hikanner/jta/internal/domain"
)

func TestNewJSONRepository(t *testing.T) {
	repo := NewJSONRepository()
	if repo == nil {
		t.Fatal("NewJSONRepository() returned nil")
	}
}

func TestJSONRepositorySaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-terminology.json")

	repo := NewJSONRepository()

	// Create test terminology
	term := &domain.Terminology{
		SourceLanguage:  "en",
		PreserveTerms:   []string{"GitHub", "API", "OAuth"},
		ConsistentTerms: []string{"user", "repository", "commit"},
	}

	// Test Save
	err := repo.Save(testFile, term)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	// Test Exists
	if !repo.Exists(testFile) {
		t.Error("Exists() = false, want true after Save()")
	}

	// Test Load
	loaded, err := repo.Load(testFile)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Verify loaded data
	if loaded.SourceLanguage != term.SourceLanguage {
		t.Errorf("SourceLanguage = %s, want %s", loaded.SourceLanguage, term.SourceLanguage)
	}

	if len(loaded.PreserveTerms) != len(term.PreserveTerms) {
		t.Errorf("PreserveTerms length = %d, want %d", len(loaded.PreserveTerms), len(term.PreserveTerms))
	}

	for i, pt := range loaded.PreserveTerms {
		if pt != term.PreserveTerms[i] {
			t.Errorf("PreserveTerms[%d] = %s, want %s", i, pt, term.PreserveTerms[i])
		}
	}

	if len(loaded.ConsistentTerms) != len(term.ConsistentTerms) {
		t.Errorf("ConsistentTerms length = %d, want %d", len(loaded.ConsistentTerms), len(term.ConsistentTerms))
	}
}

func TestJSONRepositoryLoadNonExistent(t *testing.T) {
	repo := NewJSONRepository()

	_, err := repo.Load("/non/existent/file.json")
	if err == nil {
		t.Error("Load() error = nil, want error for non-existent file")
	}
}

func TestJSONRepositoryLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	if err := os.WriteFile(testFile, []byte("{ invalid json }"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	repo := NewJSONRepository()
	_, err := repo.Load(testFile)
	if err == nil {
		t.Error("Load() error = nil, want error for invalid JSON")
	}
}

func TestJSONRepositoryExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing.json")
	nonExistingFile := filepath.Join(tmpDir, "non-existing.json")

	// Create existing file
	if err := os.WriteFile(existingFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	repo := NewJSONRepository()

	if !repo.Exists(existingFile) {
		t.Error("Exists() = false, want true for existing file")
	}

	if repo.Exists(nonExistingFile) {
		t.Error("Exists() = true, want false for non-existing file")
	}
}

func TestManagerBuildPromptDictionary(t *testing.T) {
	manager := NewManager(nil) // Provider not needed for BuildPromptDictionary

	terminology := &domain.Terminology{
		SourceLanguage:  "en",
		PreserveTerms:   []string{"GitHub", "API"},
		ConsistentTerms: []string{"user", "repository"},
	}

	translation := &domain.TerminologyTranslation{
		SourceLanguage: "en",
		TargetLanguage: "zh",
		Translations: map[string]string{
			"user":       "用户",
			"repository": "仓库",
		},
	}

	dict := manager.BuildPromptDictionary(terminology, translation)

	// Check that dictionary contains expected content
	if dict == "" {
		t.Error("BuildPromptDictionary() returned empty string")
	}

	// Should contain preserve terms
	if !contains(dict, "GitHub") {
		t.Error("Dictionary should contain 'GitHub'")
	}

	if !contains(dict, "API") {
		t.Error("Dictionary should contain 'API'")
	}

	// Should contain consistent terms mapping
	if !contains(dict, "user") {
		t.Error("Dictionary should contain 'user'")
	}

	if !contains(dict, "用户") {
		t.Error("Dictionary should contain '用户'")
	}
}

func TestManagerBuildPromptDictionaryEmpty(t *testing.T) {
	manager := NewManager(nil)

	terminology := &domain.Terminology{
		SourceLanguage:  "en",
		PreserveTerms:   []string{},
		ConsistentTerms: []string{},
	}

	dict := manager.BuildPromptDictionary(terminology, nil)

	// Should return empty or minimal content for empty terminology
	// The exact format depends on implementation
	if dict != "" && len(dict) > 100 {
		t.Errorf("BuildPromptDictionary() for empty terminology = %d chars, expected short or empty", len(dict))
	}
}

func TestTerminologyGetMissingTranslations(t *testing.T) {
	term := &domain.Terminology{
		SourceLanguage:  "en",
		ConsistentTerms: []string{"user", "repository", "commit"},
	}

	tests := []struct {
		name        string
		translation *domain.TerminologyTranslation
		want        []string
	}{
		{
			name:        "nil translation - all missing",
			translation: nil,
			want:        []string{"user", "repository", "commit"},
		},
		{
			name: "partial translation",
			translation: &domain.TerminologyTranslation{
				SourceLanguage: "en",
				TargetLanguage: "zh",
				Translations: map[string]string{
					"user": "用户",
				},
			},
			want: []string{"repository", "commit"},
		},
		{
			name: "complete translation",
			translation: &domain.TerminologyTranslation{
				SourceLanguage: "en",
				TargetLanguage: "zh",
				Translations: map[string]string{
					"user":       "用户",
					"repository": "仓库",
					"commit":     "提交",
				},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := term.GetMissingTranslations(tt.translation)

			if len(got) != len(tt.want) {
				t.Errorf("GetMissingTranslations() length = %d, want %d", len(got), len(tt.want))
				return
			}

			// Check that all expected missing terms are present
			for _, wantTerm := range tt.want {
				found := false
				for _, gotTerm := range got {
					if gotTerm == wantTerm {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetMissingTranslations() missing term %s", wantTerm)
				}
			}
		})
	}
}

func TestManagerLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()

	manager := NewManager(nil)

	// Create test terminology
	term := &domain.Terminology{
		SourceLanguage: "en",
		PreserveTerms:  []string{"Test"},
	}

	// Test save
	err := manager.SaveTerminology(tmpDir, term)
	if err != nil {
		t.Fatalf("SaveTerminology() error = %v, want nil", err)
	}

	// File should exist
	if !manager.TerminologyExists(tmpDir) {
		t.Error("File should exist after SaveTerminology()")
	}

	// Test load
	loaded, err := manager.LoadTerminology(tmpDir)
	if err != nil {
		t.Fatalf("LoadTerminology() error = %v, want nil", err)
	}

	if loaded.SourceLanguage != term.SourceLanguage {
		t.Errorf("Loaded SourceLanguage = %s, want %s", loaded.SourceLanguage, term.SourceLanguage)
	}
}

func TestManagerTerminologyExists(t *testing.T) {
	existingDir := t.TempDir()
	nonExistingDir := filepath.Join(t.TempDir(), "non-existing")

	manager := NewManager(nil)

	// Create existing terminology
	term := &domain.Terminology{
		SourceLanguage: "en",
	}
	if err := manager.SaveTerminology(existingDir, term); err != nil {
		t.Fatalf("Failed to save terminology: %v", err)
	}

	if !manager.TerminologyExists(existingDir) {
		t.Error("TerminologyExists() = false, want true for existing directory")
	}

	if manager.TerminologyExists(nonExistingDir) {
		t.Error("TerminologyExists() = true, want false for non-existing directory")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (findSubstring(s, substr) >= 0)
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
