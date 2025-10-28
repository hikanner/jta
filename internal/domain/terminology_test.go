package domain

import "testing"

func TestTerminologyTranslation_GetTermTranslation(t *testing.T) {
	tests := []struct {
		name        string
		translation *TerminologyTranslation
		term        string
		wantTrans   string
		wantOk      bool
	}{
		{
			name: "existing term",
			translation: &TerminologyTranslation{
				Translations: map[string]string{
					"API":  "API",
					"user": "usuario",
				},
			},
			term:      "user",
			wantTrans: "usuario",
			wantOk:    true,
		},
		{
			name: "non-existing term",
			translation: &TerminologyTranslation{
				Translations: map[string]string{
					"API": "API",
				},
			},
			term:      "missing",
			wantTrans: "",
			wantOk:    false,
		},
		{
			name: "empty translations",
			translation: &TerminologyTranslation{
				Translations: map[string]string{},
			},
			term:      "any",
			wantTrans: "",
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTrans, gotOk := tt.translation.GetTermTranslation(tt.term)
			if gotOk != tt.wantOk {
				t.Errorf("GetTermTranslation() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotTrans != tt.wantTrans {
				t.Errorf("GetTermTranslation() = %q, want %q", gotTrans, tt.wantTrans)
			}
		})
	}
}

func TestTerminology_GetMissingTranslations(t *testing.T) {
	tests := []struct {
		name        string
		terminology *Terminology
		translation *TerminologyTranslation
		expected    []string
	}{
		{
			name: "no missing translations",
			terminology: &Terminology{
				ConsistentTerms: []string{"user", "account"},
			},
			translation: &TerminologyTranslation{
				Translations: map[string]string{
					"user":    "usuario",
					"account": "cuenta",
				},
			},
			expected: []string{},
		},
		{
			name: "some missing translations",
			terminology: &Terminology{
				ConsistentTerms: []string{"user", "account", "profile"},
			},
			translation: &TerminologyTranslation{
				Translations: map[string]string{
					"user": "usuario",
				},
			},
			expected: []string{"account", "profile"},
		},
		{
			name: "no translation file",
			terminology: &Terminology{
				ConsistentTerms: []string{"user", "account"},
			},
			translation: nil,
			expected:    []string{"user", "account"},
		},
		{
			name: "empty consistent terms",
			terminology: &Terminology{
				ConsistentTerms: []string{},
			},
			translation: &TerminologyTranslation{
				Translations: map[string]string{"user": "usuario"},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.terminology.GetMissingTranslations(tt.translation)

			if len(result) != len(tt.expected) {
				t.Errorf("GetMissingTranslations() returned %d items, want %d", len(result), len(tt.expected))
			}

			// Check all expected items are present
			for _, expected := range tt.expected {
				found := false
				for _, item := range result {
					if item == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetMissingTranslations() missing expected item %q", expected)
				}
			}
		})
	}
}

func TestTerminology_AddPreserveTerm(t *testing.T) {
	tests := []struct {
		name       string
		initial    []string
		termsToAdd []string
		expected   []string
	}{
		{
			name:       "add to empty list",
			initial:    []string{},
			termsToAdd: []string{"API"},
			expected:   []string{"API"},
		},
		{
			name:       "add new terms",
			initial:    []string{"API"},
			termsToAdd: []string{"OAuth", "SDK"},
			expected:   []string{"API", "OAuth", "SDK"},
		},
		{
			name:       "add duplicate term",
			initial:    []string{"API"},
			termsToAdd: []string{"API", "OAuth"},
			expected:   []string{"API", "OAuth"},
		},
		{
			name:       "add multiple duplicates",
			initial:    []string{"API", "OAuth"},
			termsToAdd: []string{"API", "OAuth", "SDK"},
			expected:   []string{"API", "OAuth", "SDK"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terminology := &Terminology{
				PreserveTerms: tt.initial,
			}

			for _, term := range tt.termsToAdd {
				terminology.AddPreserveTerm(term)
			}

			if len(terminology.PreserveTerms) != len(tt.expected) {
				t.Errorf("PreserveTerms length = %d, want %d", len(terminology.PreserveTerms), len(tt.expected))
			}

			for _, expected := range tt.expected {
				found := false
				for _, term := range terminology.PreserveTerms {
					if term == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("PreserveTerms missing expected term %q", expected)
				}
			}
		})
	}
}

func TestTerminology_AddConsistentTerm(t *testing.T) {
	tests := []struct {
		name       string
		initial    []string
		termsToAdd []string
		expected   []string
	}{
		{
			name:       "add to empty list",
			initial:    []string{},
			termsToAdd: []string{"user"},
			expected:   []string{"user"},
		},
		{
			name:       "add new terms",
			initial:    []string{"user"},
			termsToAdd: []string{"account", "profile"},
			expected:   []string{"user", "account", "profile"},
		},
		{
			name:       "add duplicate term",
			initial:    []string{"user"},
			termsToAdd: []string{"user", "account"},
			expected:   []string{"user", "account"},
		},
		{
			name:       "add multiple duplicates",
			initial:    []string{"user", "account"},
			termsToAdd: []string{"user", "account", "profile"},
			expected:   []string{"user", "account", "profile"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terminology := &Terminology{
				ConsistentTerms: tt.initial,
			}

			for _, term := range tt.termsToAdd {
				terminology.AddConsistentTerm(term)
			}

			if len(terminology.ConsistentTerms) != len(tt.expected) {
				t.Errorf("ConsistentTerms length = %d, want %d", len(terminology.ConsistentTerms), len(tt.expected))
			}

			for _, expected := range tt.expected {
				found := false
				for _, term := range terminology.ConsistentTerms {
					if term == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ConsistentTerms missing expected term %q", expected)
				}
			}
		})
	}
}

func TestNewTerminologyTranslation(t *testing.T) {
	tests := []struct {
		name       string
		sourceLang string
		targetLang string
	}{
		{"English to Spanish", "en", "es"},
		{"English to Japanese", "en", "ja"},
		{"Chinese to English", "zh", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewTerminologyTranslation(tt.sourceLang, tt.targetLang)

			if result == nil {
				t.Fatal("NewTerminologyTranslation() returned nil")
			}

			if result.SourceLanguage != tt.sourceLang {
				t.Errorf("SourceLanguage = %q, want %q", result.SourceLanguage, tt.sourceLang)
			}

			if result.TargetLanguage != tt.targetLang {
				t.Errorf("TargetLanguage = %q, want %q", result.TargetLanguage, tt.targetLang)
			}

			if result.Translations == nil {
				t.Error("Translations map should be initialized")
			}

			if len(result.Translations) != 0 {
				t.Errorf("Translations map should be empty, got %d items", len(result.Translations))
			}
		})
	}
}

func TestTerminologyStruct(t *testing.T) {
	// Test Terminology struct initialization
	terminology := &Terminology{
		SourceLanguage:  "en",
		PreserveTerms:   []string{"API", "SDK"},
		ConsistentTerms: []string{"user", "account"},
	}

	if terminology.SourceLanguage != "en" {
		t.Errorf("SourceLanguage = %q, want %q", terminology.SourceLanguage, "en")
	}

	if len(terminology.PreserveTerms) != 2 {
		t.Errorf("PreserveTerms length = %d, want 2", len(terminology.PreserveTerms))
	}

	if len(terminology.ConsistentTerms) != 2 {
		t.Errorf("ConsistentTerms length = %d, want 2", len(terminology.ConsistentTerms))
	}
}

func TestTerminologyTranslationStruct(t *testing.T) {
	// Test TerminologyTranslation struct initialization
	translation := &TerminologyTranslation{
		SourceLanguage: "en",
		TargetLanguage: "es",
		Translations: map[string]string{
			"user":    "usuario",
			"account": "cuenta",
		},
	}

	if translation.SourceLanguage != "en" {
		t.Errorf("SourceLanguage = %q, want %q", translation.SourceLanguage, "en")
	}

	if translation.TargetLanguage != "es" {
		t.Errorf("TargetLanguage = %q, want %q", translation.TargetLanguage, "es")
	}

	if len(translation.Translations) != 2 {
		t.Errorf("Translations length = %d, want 2", len(translation.Translations))
	}
}
