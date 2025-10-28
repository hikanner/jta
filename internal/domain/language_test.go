package domain

import "testing"

func TestIsRTLLanguage(t *testing.T) {
	tests := []struct {
		name     string
		langCode string
		expected bool
	}{
		{"Arabic is RTL", "ar", true},
		{"Hebrew is RTL", "he", true},
		{"Persian is RTL", "fa", true},
		{"Urdu is RTL", "ur", true},
		{"English is not RTL", "en", false},
		{"Spanish is not RTL", "es", false},
		{"Chinese is not RTL", "zh", false},
		{"Invalid code is not RTL", "invalid", false},
		{"Empty code is not RTL", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRTLLanguage(tt.langCode)
			if result != tt.expected {
				t.Errorf("IsRTLLanguage(%q) = %v, want %v", tt.langCode, result, tt.expected)
			}
		})
	}
}

func TestGetLanguage(t *testing.T) {
	tests := []struct {
		name       string
		langCode   string
		expectName string
		expectOk   bool
	}{
		{
			name:       "get English",
			langCode:   "en",
			expectName: "English",
			expectOk:   true,
		},
		{
			name:       "get Spanish",
			langCode:   "es",
			expectName: "Spanish",
			expectOk:   true,
		},
		{
			name:       "get Japanese",
			langCode:   "ja",
			expectName: "Japanese",
			expectOk:   true,
		},
		{
			name:       "get Arabic (RTL)",
			langCode:   "ar",
			expectName: "Arabic",
			expectOk:   true,
		},
		{
			name:     "invalid code",
			langCode: "invalid",
			expectOk: false,
		},
		{
			name:     "empty code",
			langCode: "",
			expectOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang, ok := GetLanguage(tt.langCode)

			if ok != tt.expectOk {
				t.Errorf("GetLanguage(%q) ok = %v, want %v", tt.langCode, ok, tt.expectOk)
			}

			if tt.expectOk {
				if lang.Name != tt.expectName {
					t.Errorf("GetLanguage(%q).Name = %q, want %q", tt.langCode, lang.Name, tt.expectName)
				}
				if lang.Code != tt.langCode {
					t.Errorf("GetLanguage(%q).Code = %q, want %q", tt.langCode, lang.Code, tt.langCode)
				}
			}
		})
	}
}

func TestValidateLanguageCode(t *testing.T) {
	tests := []struct {
		name     string
		langCode string
		expected bool
	}{
		{"valid English", "en", true},
		{"valid Spanish", "es", true},
		{"valid Chinese", "zh", true},
		{"valid Japanese", "ja", true},
		{"valid Arabic", "ar", true},
		{"valid French", "fr", true},
		{"valid German", "de", true},
		{"valid Italian", "it", true},
		{"valid Portuguese", "pt", true},
		{"valid Russian", "ru", true},
		{"valid Korean", "ko", true},
		{"valid Hindi", "hi", true},
		{"valid Turkish", "tr", true},
		{"valid Dutch", "nl", true},
		{"valid Polish", "pl", true},
		{"invalid code", "invalid", false},
		{"empty code", "", false},
		{"unknown code", "xx", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateLanguageCode(tt.langCode)
			if result != tt.expected {
				t.Errorf("ValidateLanguageCode(%q) = %v, want %v", tt.langCode, result, tt.expected)
			}
		})
	}
}

func TestLanguageStruct(t *testing.T) {
	// Test that Language struct has expected fields
	lang := Language{
		Code:       "en",
		Name:       "English",
		NativeName: "English",
		IsRTL:      false,
	}

	if lang.Code != "en" {
		t.Errorf("Language.Code = %q, want %q", lang.Code, "en")
	}
	if lang.Name != "English" {
		t.Errorf("Language.Name = %q, want %q", lang.Name, "English")
	}
	if lang.NativeName != "English" {
		t.Errorf("Language.NativeName = %q, want %q", lang.NativeName, "English")
	}
	if lang.IsRTL != false {
		t.Errorf("Language.IsRTL = %v, want %v", lang.IsRTL, false)
	}
}

func TestSupportedLanguages(t *testing.T) {
	// Verify that SupportedLanguages map is populated
	if len(SupportedLanguages) == 0 {
		t.Fatal("SupportedLanguages map should not be empty")
	}

	// Test a few known languages
	knownLanguages := map[string]string{
		"en": "English",
		"es": "Spanish",
		"fr": "French",
		"de": "German",
		"ja": "Japanese",
		"zh": "Chinese (Simplified)",
		"ar": "Arabic",
		"ru": "Russian",
		"ko": "Korean",
		"pt": "Portuguese",
	}

	for code, name := range knownLanguages {
		lang, exists := SupportedLanguages[code]
		if !exists {
			t.Errorf("SupportedLanguages should contain %q", code)
			continue
		}
		if lang.Name != name {
			t.Errorf("SupportedLanguages[%q].Name = %q, want %q", code, lang.Name, name)
		}
		if lang.Code != code {
			t.Errorf("SupportedLanguages[%q].Code = %q, want %q", code, lang.Code, code)
		}
	}

	// Verify RTL languages
	rtlLanguages := []string{"ar", "he", "fa", "ur"}
	for _, code := range rtlLanguages {
		lang, exists := SupportedLanguages[code]
		if !exists {
			t.Errorf("SupportedLanguages should contain RTL language %q", code)
			continue
		}
		if !lang.IsRTL {
			t.Errorf("SupportedLanguages[%q].IsRTL should be true", code)
		}
	}

	// Verify non-RTL languages
	ltrLanguages := []string{"en", "es", "fr", "de", "ja", "zh"}
	for _, code := range ltrLanguages {
		lang, exists := SupportedLanguages[code]
		if !exists {
			t.Errorf("SupportedLanguages should contain LTR language %q", code)
			continue
		}
		if lang.IsRTL {
			t.Errorf("SupportedLanguages[%q].IsRTL should be false", code)
		}
	}
}

func TestRTLLanguageCoverage(t *testing.T) {
	// Ensure all RTL languages are correctly marked
	expectedRTL := map[string]bool{
		"ar": true, // Arabic
		"he": true, // Hebrew
		"fa": true, // Persian
		"ur": true, // Urdu
	}

	for code, shouldBeRTL := range expectedRTL {
		if IsRTLLanguage(code) != shouldBeRTL {
			t.Errorf("IsRTLLanguage(%q) = %v, want %v", code, !shouldBeRTL, shouldBeRTL)
		}
	}
}
