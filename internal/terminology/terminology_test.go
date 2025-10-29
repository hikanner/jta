package terminology

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/provider"
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
				found := slices.Contains(got, wantTerm)
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

// ============================================================================
// Detector Tests
// ============================================================================

func TestNewDetector(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	detector := NewDetector(mockProvider)

	if detector == nil {
		t.Fatal("NewDetector() returned nil")
	}
	if detector.provider == nil {
		t.Error("Expected provider to be set")
	}
}

func TestEstimateTokens(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	detector := NewDetector(mockProvider)

	tests := []struct {
		name      string
		texts     []string
		minTokens int
		maxTokens int
	}{
		{
			name:      "Empty texts",
			texts:     []string{},
			minTokens: 0,
			maxTokens: 0,
		},
		{
			name:      "Single short text",
			texts:     []string{"Hello"},
			minTokens: 1,
			maxTokens: 2,
		},
		{
			name:      "Multiple texts",
			texts:     []string{"Hello world", "This is a test", "Another sentence"},
			minTokens: 8,
			maxTokens: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := detector.estimateTokens(tt.texts)
			if tokens < tt.minTokens || tokens > tt.maxTokens {
				t.Errorf("estimateTokens() = %d, want between %d and %d",
					tokens, tt.minTokens, tt.maxTokens)
			}
		})
	}
}

func TestBuildFullDocument(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	detector := NewDetector(mockProvider)

	texts := []string{"First text", "Second text", "Third text"}
	doc := detector.buildFullDocument(texts)

	// Verify document contains all texts
	for i, text := range texts {
		if !contains(doc, text) {
			t.Errorf("Document missing text[%d]: %s", i, text)
		}
	}

	// Verify document has structure (numbered entries)
	if !contains(doc, "[1]") {
		t.Error("Document should contain [1] marker")
	}
}

func TestBuildDetectionPrompt(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	detector := NewDetector(mockProvider)

	doc := "[1] Test document with API and OAuth"
	lang := "en"
	count := 1

	prompt := detector.buildDetectionPrompt(doc, lang, count)

	// Verify prompt contains required elements
	requiredElements := []string{
		"terminology",
		doc,
		lang,
		"JSON",
		"preserve",
		"consistent",
	}

	for _, elem := range requiredElements {
		if !contains(prompt, elem) {
			t.Errorf("Prompt missing required element: %s", elem)
		}
	}
}

func TestParseTermsFromJSON(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	detector := NewDetector(mockProvider)

	tests := []struct {
		name        string
		content     string
		expectError bool
		expectCount int
	}{
		{
			name: "Valid JSON with preserve and consistent terms",
			content: `{
				"preserveTerms": [
					{"term": "API", "reason": "Technical term", "frequency": 1, "examples": ["API usage"]},
					{"term": "OAuth", "reason": "Protocol", "frequency": 1, "examples": ["OAuth flow"]},
					{"term": "JSON", "reason": "Format", "frequency": 1, "examples": ["JSON data"]}
				],
				"consistentTerms": [
					{"term": "user", "reason": "Key concept", "frequency": 1, "examples": ["user data"]},
					{"term": "repository", "reason": "Core term", "frequency": 1, "examples": ["repository info"]}
				]
			}`,
			expectError: false,
			expectCount: 5,
		},
		{
			name: "Valid JSON with only preserve terms",
			content: `{
				"preserveTerms": [
					{"term": "GitHub", "reason": "Brand", "frequency": 1, "examples": ["GitHub usage"]},
					{"term": "API", "reason": "Technical", "frequency": 1, "examples": ["API call"]}
				],
				"consistentTerms": []
			}`,
			expectError: false,
			expectCount: 2,
		},
		{
			name: "Empty JSON",
			content: `{
				"preserveTerms": [],
				"consistentTerms": []
			}`,
			expectError: false,
			expectCount: 0,
		},
		{
			name:        "Invalid JSON",
			content:     `{ invalid json }`,
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terms, err := detector.parseTermsFromJSON(tt.content)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error for invalid JSON")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(terms) != tt.expectCount {
				t.Errorf("Got %d terms, want %d", len(terms), tt.expectCount)
			}
		})
	}
}

func TestDetector_SimpleTokenize(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	detector := NewDetector(mockProvider)

	tests := []struct {
		name     string
		text     string
		minWords int
	}{
		{
			name:     "Simple sentence",
			text:     "Hello world",
			minWords: 2,
		},
		{
			name:     "With punctuation",
			text:     "Hello, world! How are you?",
			minWords: 4,
		},
		{
			name:     "With special chars",
			text:     "GitHub API OAuth",
			minWords: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words := detector.simpleTokenize(tt.text)
			if len(words) < tt.minWords {
				t.Errorf("simpleTokenize() returned %d words, want at least %d",
					len(words), tt.minWords)
			}
		})
	}
}

func TestDetector_IsStopWord(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	detector := NewDetector(mockProvider)

	stopWords := []string{"the", "a", "an", "and", "or", "but", "is", "are", "was", "were"}
	nonStopWords := []string{"GitHub", "API", "user", "repository", "OAuth"}

	for _, word := range stopWords {
		if !detector.isStopWord(word) {
			t.Errorf("Expected '%s' to be a stop word", word)
		}
	}

	for _, word := range nonStopWords {
		if detector.isStopWord(word) {
			t.Errorf("Expected '%s' NOT to be a stop word", word)
		}
	}
}

func TestDetector_IsSpecialFormat(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	detector := NewDetector(mockProvider)

	tests := []struct {
		word     string
		expected bool
		reason   string
	}{
		// Should be special format (technical terms)
		{"API", true, "all caps"},
		{"JSON", true, "all caps"},
		{"OAuth", true, "CamelCase"},
		{"OpenAI", true, "CamelCase"},
		{"GPT-4", true, "contains number"},
		{"FLUX.1", true, "contains dot and number"},
		{"user123", true, "contains number"},

		// Should NOT be special format
		{"github", false, "lowercase"},
		{"user", false, "lowercase"},
		{"test", false, "lowercase"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result := detector.isSpecialFormat(tt.word)
			if result != tt.expected {
				t.Errorf("isSpecialFormat(%s) = %v, want %v (%s)",
					tt.word, result, tt.expected, tt.reason)
			}
		})
	}
}

func TestDetector_ExtractCandidatesSimplified(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	detector := NewDetector(mockProvider)

	texts := []string{
		"The GitHub API is very useful",
		"OAuth provides secure authentication",
		"The API documentation is comprehensive",
	}

	candidates := detector.extractCandidatesSimplified(texts)

	// Should contain some technical terms (at least one)
	if len(candidates) == 0 {
		t.Error("Expected some candidates to be extracted")
	}

	// API should be present (appears twice)
	if candidates["API"] != nil {
		if candidates["API"].Frequency != 2 {
			t.Errorf("API frequency = %d, want 2", candidates["API"].Frequency)
		}
	}

	// Should NOT contain stop words
	if candidates["the"] != nil {
		t.Error("Stop word 'the' should not be extracted")
	}
	if candidates["is"] != nil {
		t.Error("Stop word 'is' should not be extracted")
	}
}

func TestDetector_BuildValidationPrompt(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	detector := NewDetector(mockProvider)

	candidates := []*CandidateWord{
		{Word: "GitHub", Frequency: 5, Contexts: []string{"GitHub repository"}},
		{Word: "API", Frequency: 3, Contexts: []string{"REST API"}},
	}

	prompt := detector.buildValidationPrompt(candidates, "en")

	// Verify prompt contains candidates
	if !contains(prompt, "GitHub") {
		t.Error("Prompt should contain 'GitHub'")
	}
	if !contains(prompt, "API") {
		t.Error("Prompt should contain 'API'")
	}

	// Verify prompt has structure
	if !contains(prompt, "terminology") {
		t.Error("Prompt should mention 'terminology'")
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		contains string
	}{
		{
			name:     "JSON with markdown code block",
			content:  "Some text\n```json\n{\"key\": \"value\"}\n```\nMore text",
			contains: "key",
		},
		{
			name:     "Plain JSON",
			content:  `{"preserve": ["API"], "consistent": ["user"]}`,
			contains: "preserve",
		},
		{
			name:     "JSON with extra text",
			content:  `Here is the result: {"terms": ["GitHub"]} and that's it`,
			contains: "terms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.content)
			if !contains(result, tt.contains) {
				t.Errorf("extractJSON() result should contain '%s', got: %s",
					tt.contains, result)
			}
		})
	}
}

func TestDetector_AnalyzeWithLLM_Success(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	// Provide valid JSON response in the expected format
	mockProvider.AddResponse("```json\n" + `{
		"preserveTerms": [
			{"term": "GitHub", "reason": "Brand name", "frequency": 1, "examples": ["GitHub is a platform"]},
			{"term": "API", "reason": "Technical acronym", "frequency": 1, "examples": ["The API provides access"]},
			{"term": "OAuth", "reason": "Protocol name", "frequency": 1, "examples": ["OAuth enables authentication"]}
		],
		"consistentTerms": [
			{"term": "platform", "reason": "Key concept", "frequency": 1, "examples": ["GitHub is a platform"]},
			{"term": "access", "reason": "Important term", "frequency": 1, "examples": ["The API provides access"]}
		]
	}` + "\n```")

	detector := NewDetector(mockProvider)

	texts := []string{
		"GitHub is a platform",
		"The API provides access",
		"OAuth enables authentication",
	}

	terms, err := detector.analyzeWithLLM(context.Background(), texts, "en")
	if err != nil {
		t.Fatalf("analyzeWithLLM() error = %v, want nil", err)
	}

	// Should have parsed 5 terms (3 preserve + 2 consistent)
	if len(terms) < 3 {
		t.Errorf("Expected at least 3 terms, got %d", len(terms))
	}
}

func TestDetector_AnalyzeWithLLM_Error(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	mockProvider.SetError("API call failed")

	detector := NewDetector(mockProvider)

	texts := []string{"Test text"}

	_, err := detector.analyzeWithLLM(context.Background(), texts, "en")
	if err == nil {
		t.Error("Expected error from LLM failure")
	}
}

func TestDetector_DetectTerms_SmallFile(t *testing.T) {
	mockProvider := provider.NewMockProvider("test-model")
	// Provide valid JSON response in the expected format
	mockProvider.AddResponse("```json\n" + `{
		"preserveTerms": [
			{"term": "API", "reason": "Technical term", "frequency": 1, "examples": ["The API for user management"]}
		],
		"consistentTerms": [
			{"term": "user", "reason": "Key concept", "frequency": 1, "examples": ["user management"]}
		]
	}` + "\n```")

	detector := NewDetector(mockProvider)

	// Small file - should use full LLM analysis
	texts := []string{"The API for user management"}

	terms, err := detector.DetectTerms(context.Background(), texts, "en")
	if err != nil {
		t.Fatalf("DetectTerms() error = %v, want nil", err)
	}

	// Should have parsed 2 terms (1 preserve + 1 consistent)
	if len(terms) < 2 {
		t.Errorf("Expected at least 2 terms, got %d", len(terms))
	}
}

// Tests for Manager
func TestManager_DetectTerms_Simple(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	mockProvider.AddResponse("```json\n" + `{
		"preserveTerms": [
			{"term": "API", "reason": "Technical term"}
		],
		"consistentTerms": [
			{"term": "user", "reason": "Key concept"}
		]
	}` + "\n```")

	manager := NewManager(mockProvider)

	texts := []string{"The API for user management"}
	terms, err := manager.DetectTerms(context.Background(), texts, "en")

	if err != nil {
		t.Fatalf("DetectTerms() error = %v", err)
	}

	if len(terms) < 2 {
		t.Errorf("Expected at least 2 terms, got %d", len(terms))
	}
}

func TestManager_TranslateTerms_Simple(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	mockProvider.AddResponse("```json\n" + `{
		"user": "用户",
		"repository": "仓库"
	}` + "\n```")

	manager := NewManager(mockProvider)

	terms := []string{"user", "repository"}

	translations, err := manager.TranslateTerms(context.Background(), terms, "en", "zh")

	if err != nil {
		t.Fatalf("TranslateTerms() error = %v", err)
	}

	if len(translations) != 2 {
		t.Errorf("Expected 2 translations, got %d", len(translations))
	}

	if translations["user"] != "用户" {
		t.Errorf("Expected user -> 用户, got %s", translations["user"])
	}

	if translations["repository"] != "仓库" {
		t.Errorf("Expected repository -> 仓库, got %s", translations["repository"])
	}
}

func TestManager_LoadAndSaveTerminologyTranslation_Flow(t *testing.T) {
	tmpDir := t.TempDir()
	mockProvider := provider.NewMockProvider("gpt-4")
	manager := NewManager(mockProvider)

	translation := &domain.TerminologyTranslation{
		SourceLanguage: "en",
		TargetLanguage: "zh",
		Translations: map[string]string{
			"user": "用户",
			"API":  "API",
		},
	}

	// Test Save
	err := manager.SaveTerminologyTranslation(tmpDir, translation)
	if err != nil {
		t.Fatalf("SaveTerminologyTranslation() error = %v", err)
	}

	// Test Exists
	exists := manager.TranslationExists(tmpDir, "zh")
	if !exists {
		t.Error("TranslationExists() = false, want true")
	}

	// Test Load
	loaded, err := manager.LoadTerminologyTranslation(tmpDir, "zh")
	if err != nil {
		t.Fatalf("LoadTerminologyTranslation() error = %v", err)
	}

	if loaded.TargetLanguage != "zh" {
		t.Errorf("TargetLanguage = %s, want zh", loaded.TargetLanguage)
	}

	if len(loaded.Translations) != 2 {
		t.Errorf("len(Translations) = %d, want 2", len(loaded.Translations))
	}
}

// Tests for TranslationRepository
func TestTranslationRepository_SaveAndLoad_Flow(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewTranslationRepository()

	translation := &domain.TerminologyTranslation{
		SourceLanguage: "en",
		TargetLanguage: "zh",
		Translations: map[string]string{
			"hello": "你好",
			"world": "世界",
		},
	}

	// Test Save
	err := repo.Save(tmpDir, translation)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Test Exists
	if !repo.Exists(tmpDir, "zh") {
		t.Error("Exists() = false, want true after Save()")
	}

	// Test Load
	loaded, err := repo.Load(tmpDir, "zh")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.SourceLanguage != "en" {
		t.Errorf("SourceLanguage = %s, want en", loaded.SourceLanguage)
	}

	if loaded.TargetLanguage != "zh" {
		t.Errorf("TargetLanguage = %s, want zh", loaded.TargetLanguage)
	}

	if len(loaded.Translations) != 2 {
		t.Errorf("len(Translations) = %d, want 2", len(loaded.Translations))
	}

	if loaded.Translations["hello"] != "你好" {
		t.Errorf("Translations[hello] = %s, want 你好", loaded.Translations["hello"])
	}
}

func TestTranslationRepository_LoadNonExistent_Error(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewTranslationRepository()

	_, err := repo.Load(tmpDir, "fr")
	if err == nil {
		t.Error("Load() for non-existent file should return error")
	}
}

func TestTranslationRepository_ExistsNonExistent_False(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewTranslationRepository()

	if repo.Exists(tmpDir, "de") {
		t.Error("Exists() = true for non-existent file, want false")
	}
}

// Tests for TermRepository
func TestTermRepository_SaveAndLoad_Flow(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewTermRepository()

	terminology := &domain.Terminology{
		SourceLanguage:  "en",
		PreserveTerms:   []string{"GitHub", "API"},
		ConsistentTerms: []string{"user", "commit"},
	}

	// Test Save
	termFile := filepath.Join(tmpDir, "terminology.json")
	err := repo.Save(termFile, terminology)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Test Load
	loaded, err := repo.Load(termFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.SourceLanguage != "en" {
		t.Errorf("SourceLanguage = %s, want en", loaded.SourceLanguage)
	}

	if len(loaded.PreserveTerms) != 2 {
		t.Errorf("len(PreserveTerms) = %d, want 2", len(loaded.PreserveTerms))
	}

	if len(loaded.ConsistentTerms) != 2 {
		t.Errorf("len(ConsistentTerms) = %d, want 2", len(loaded.ConsistentTerms))
	}
}

func TestTermRepository_LoadNonExistent_Error(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewTermRepository()

	nonExistentFile := filepath.Join(tmpDir, "nonexistent.json")
	_, err := repo.Load(nonExistentFile)
	if err == nil {
		t.Error("Load() for non-existent file should return error")
	}
}

// Test batchCandidates
func TestDetector_BatchCandidates(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	detector := NewDetector(mockProvider)

	candidates := map[string]*CandidateWord{
		"term1": {Word: "term1", Frequency: 5, Contexts: []string{"context1"}},
		"term2": {Word: "term2", Frequency: 3, Contexts: []string{"context2"}},
		"term3": {Word: "term3", Frequency: 7, Contexts: []string{"context3"}},
		"term4": {Word: "term4", Frequency: 2, Contexts: []string{"context4"}},
		"term5": {Word: "term5", Frequency: 4, Contexts: []string{"context5"}},
	}

	tests := []struct {
		name          string
		batchSize     int
		expectedCount int
	}{
		{
			name:          "batch size 2",
			batchSize:     2,
			expectedCount: 3, // 5 items / 2 = 3 batches (2+2+1)
		},
		{
			name:          "batch size 3",
			batchSize:     3,
			expectedCount: 2, // 5 items / 3 = 2 batches (3+2)
		},
		{
			name:          "batch size larger than items",
			batchSize:     10,
			expectedCount: 1,
		},
		{
			name:          "batch size 1",
			batchSize:     1,
			expectedCount: 5, // each item in its own batch
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := detector.batchCandidates(candidates, tt.batchSize)

			if len(batches) != tt.expectedCount {
				t.Errorf("batchCandidates() created %d batches, want %d", len(batches), tt.expectedCount)
			}

			// Verify all candidates are included
			totalCandidates := 0
			for _, batch := range batches {
				totalCandidates += len(batch)
			}
			if totalCandidates != len(candidates) {
				t.Errorf("Total candidates in batches = %d, want %d", totalCandidates, len(candidates))
			}
		})
	}
}

// Test batchCandidates with empty input
func TestDetector_BatchCandidates_Empty(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	detector := NewDetector(mockProvider)

	candidates := map[string]*CandidateWord{}
	batches := detector.batchCandidates(candidates, 5)

	if len(batches) != 0 {
		t.Errorf("batchCandidates() with empty input created %d batches, want 0", len(batches))
	}
}

// Test parseValidationResult
func TestDetector_ParseValidationResult(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")
	detector := NewDetector(mockProvider)

	tests := []struct {
		name        string
		content     string
		expectError bool
		expectCount int
		expectTypes map[string]domain.TermType
	}{
		{
			name: "valid result with preserve terms",
			content: "```json\n" + `[
				{"term": "API", "is_term": true, "type": "preserve", "reason": "Technical term"},
				{"term": "user", "is_term": true, "type": "consistent", "reason": "Key concept"}
			]` + "\n```",
			expectError: false,
			expectCount: 2,
			expectTypes: map[string]domain.TermType{
				"API":  domain.TermTypePreserve,
				"user": domain.TermTypeConsistent,
			},
		},
		{
			name: "result with non-terms filtered",
			content: "```json\n" + `[
				{"term": "API", "is_term": true, "type": "preserve", "reason": "Technical term"},
				{"term": "the", "is_term": false, "type": "", "reason": "Common word"}
			]` + "\n```",
			expectError: false,
			expectCount: 1,
			expectTypes: map[string]domain.TermType{
				"API": domain.TermTypePreserve,
			},
		},
		{
			name:        "empty result",
			content:     "```json\n" + `[]` + "\n```",
			expectError: false,
			expectCount: 0,
		},
		{
			name:        "invalid JSON",
			content:     "not valid json",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terms, err := detector.parseValidationResult(tt.content)

			if tt.expectError && err == nil {
				t.Error("parseValidationResult() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("parseValidationResult() unexpected error = %v", err)
			}

			if !tt.expectError {
				if len(terms) != tt.expectCount {
					t.Errorf("parseValidationResult() returned %d terms, want %d", len(terms), tt.expectCount)
				}

				for _, term := range terms {
					if expectedType, ok := tt.expectTypes[term.Term]; ok {
						if term.Type != expectedType {
							t.Errorf("Term %s has type %v, want %v", term.Term, term.Type, expectedType)
						}
					}
				}
			}
		})
	}
}

// Test validateBatchWithLLM
func TestDetector_ValidateBatchWithLLM(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")

	// Mock LLM response
	mockProvider.AddResponse("```json\n" + `[
		{"term": "API", "is_term": true, "type": "preserve", "reason": "Technical term"},
		{"term": "repository", "is_term": true, "type": "consistent", "reason": "Core concept"}
	]` + "\n```")

	detector := NewDetector(mockProvider)

	batch := []*CandidateWord{
		{Word: "API", Frequency: 10, Contexts: []string{"Use the API"}},
		{Word: "repository", Frequency: 5, Contexts: []string{"GitHub repository"}},
	}

	terms, err := detector.validateBatchWithLLM(context.Background(), batch, "en")

	if err != nil {
		t.Fatalf("validateBatchWithLLM() error = %v", err)
	}

	if len(terms) != 2 {
		t.Errorf("validateBatchWithLLM() returned %d terms, want 2", len(terms))
	}

	// Verify term types
	foundPreserve := false
	foundConsistent := false
	for _, term := range terms {
		if term.Term == "API" && term.Type == domain.TermTypePreserve {
			foundPreserve = true
		}
		if term.Term == "repository" && term.Type == domain.TermTypeConsistent {
			foundConsistent = true
		}
	}

	if !foundPreserve {
		t.Error("Expected to find API as preserve term")
	}
	if !foundConsistent {
		t.Error("Expected to find repository as consistent term")
	}
}

// Test validateWithLLM with single batch
func TestDetector_ValidateWithLLM_SingleBatch(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")

	// Mock LLM response
	mockProvider.AddResponse("```json\n" + `[
		{"term": "API", "is_term": true, "type": "preserve", "reason": "Technical term"}
	]` + "\n```")

	detector := NewDetector(mockProvider)

	candidates := map[string]*CandidateWord{
		"API": {Word: "API", Frequency: 10, Contexts: []string{"Use the API"}},
	}

	terms, err := detector.validateWithLLM(context.Background(), candidates, "en")

	if err != nil {
		t.Fatalf("validateWithLLM() error = %v", err)
	}

	if len(terms) != 1 {
		t.Errorf("validateWithLLM() returned %d terms, want 1", len(terms))
	}
}

// Test validateWithLLM with error handling
func TestDetector_ValidateWithLLM_WithError(t *testing.T) {
	mockProvider := provider.NewMockProvider("gpt-4")

	// Set error to simulate API failure
	mockProvider.SetError("API error")

	detector := NewDetector(mockProvider)

	candidates := map[string]*CandidateWord{
		"API": {Word: "API", Frequency: 10, Contexts: []string{"Use the API"}},
	}

	_, err := detector.validateWithLLM(context.Background(), candidates, "en")

	if err == nil {
		t.Error("validateWithLLM() expected error but got none")
	}
}
