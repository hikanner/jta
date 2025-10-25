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
		SourceLanguage: "en",
		PreserveTerms:  []string{"GitHub", "API", "OAuth"},
		ConsistentTerms: map[string][]string{
			"en": {"user", "repository", "commit"},
			"zh": {"用户", "仓库", "提交"},
		},
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
	os.WriteFile(testFile, []byte("{ invalid json }"), 0644)

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
	os.WriteFile(existingFile, []byte("{}"), 0644)

	repo := NewJSONRepository()

	if !repo.Exists(existingFile) {
		t.Error("Exists() = false, want true for existing file")
	}

	if repo.Exists(nonExistingFile) {
		t.Error("Exists() = true, want false for non-existing file")
	}
}

func TestManagerBuildPromptDictionary(t *testing.T) {
	repo := NewJSONRepository()
	manager := NewManager(nil, repo) // Provider not needed for BuildPromptDictionary

	terminology := &domain.Terminology{
		SourceLanguage: "en",
		PreserveTerms:  []string{"GitHub", "API"},
		ConsistentTerms: map[string][]string{
			"en": {"user", "repository"},
			"zh": {"用户", "仓库"},
		},
	}

	dict := manager.BuildPromptDictionary(terminology, "zh")

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
	repo := NewJSONRepository()
	manager := NewManager(nil, repo)

	terminology := &domain.Terminology{
		SourceLanguage:  "en",
		PreserveTerms:   []string{},
		ConsistentTerms: map[string][]string{},
	}

	dict := manager.BuildPromptDictionary(terminology, "zh")

	// Should return empty or minimal content for empty terminology
	// The exact format depends on implementation
	if dict != "" && len(dict) > 100 {
		t.Errorf("BuildPromptDictionary() for empty terminology = %d chars, expected short or empty", len(dict))
	}
}

func TestTerminologyGetTermTranslation(t *testing.T) {
	term := &domain.Terminology{
		SourceLanguage: "en",
		ConsistentTerms: map[string][]string{
			"en": {"user", "repository", "commit"},
			"zh": {"用户", "仓库", "提交"},
		},
	}

	tests := []struct {
		name       string
		sourceTerm string
		targetLang string
		want       string
		wantOk     bool
	}{
		{
			name:       "existing term",
			sourceTerm: "user",
			targetLang: "zh",
			want:       "用户",
			wantOk:     true,
		},
		{
			name:       "second term",
			sourceTerm: "repository",
			targetLang: "zh",
			want:       "仓库",
			wantOk:     true,
		},
		{
			name:       "non-existent term",
			sourceTerm: "unknown",
			targetLang: "zh",
			want:       "",
			wantOk:     false,
		},
		{
			name:       "target language without translation",
			sourceTerm: "user",
			targetLang: "fr",
			want:       "",
			wantOk:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := term.GetTermTranslation(tt.sourceTerm, tt.targetLang)

			if ok != tt.wantOk {
				t.Errorf("GetTermTranslation() ok = %v, want %v", ok, tt.wantOk)
			}

			if got != tt.want {
				t.Errorf("GetTermTranslation() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestManagerLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "terminology.json")

	repo := NewJSONRepository()
	manager := NewManager(nil, repo)

	// Create test terminology
	term := &domain.Terminology{
		SourceLanguage: "en",
		PreserveTerms:  []string{"Test"},
	}

	// Test save
	err := manager.SaveTerminology(testFile, term)
	if err != nil {
		t.Fatalf("SaveTerminology() error = %v, want nil", err)
	}

	// File should exist
	if !manager.TerminologyExists(testFile) {
		t.Error("File should exist after SaveTerminology()")
	}

	// Test load
	loaded, err := manager.LoadTerminology(testFile)
	if err != nil {
		t.Fatalf("LoadTerminology() error = %v, want nil", err)
	}

	if loaded.SourceLanguage != term.SourceLanguage {
		t.Errorf("Loaded SourceLanguage = %s, want %s", loaded.SourceLanguage, term.SourceLanguage)
	}
}

func TestManagerTerminologyExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing.json")
	nonExistingFile := filepath.Join(tmpDir, "non-existing.json")

	repo := NewJSONRepository()
	manager := NewManager(nil, repo)

	// Create existing file
	term := &domain.Terminology{
		SourceLanguage: "en",
	}
	manager.SaveTerminology(existingFile, term)

	if !manager.TerminologyExists(existingFile) {
		t.Error("TerminologyExists() = false, want true for existing file")
	}

	if manager.TerminologyExists(nonExistingFile) {
		t.Error("TerminologyExists() = true, want false for non-existing file")
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
