package domain

// Language represents a supported language
type Language struct {
	Code        string
	Name        string
	NativeName  string
	IsRTL       bool
	Punctuation map[string]string // Special punctuation for the language
}

// Common language codes
const (
	LangEnglish           = "en"
	LangChineseSimplified = "zh"
	LangJapanese          = "ja"
	LangKorean            = "ko"
	LangSpanish           = "es"
	LangFrench            = "fr"
	LangGerman            = "de"
	LangArabic            = "ar"
	LangHebrew            = "he"
	LangPersian           = "fa"
	LangUrdu              = "ur"
)

// SupportedLanguages defines all supported languages
var SupportedLanguages = map[string]Language{
	LangEnglish: {
		Code:       LangEnglish,
		Name:       "English",
		NativeName: "English",
		IsRTL:      false,
	},
	LangChineseSimplified: {
		Code:       LangChineseSimplified,
		Name:       "Chinese (Simplified)",
		NativeName: "简体中文",
		IsRTL:      false,
	},
	LangJapanese: {
		Code:       LangJapanese,
		Name:       "Japanese",
		NativeName: "日本語",
		IsRTL:      false,
	},
	LangKorean: {
		Code:       LangKorean,
		Name:       "Korean",
		NativeName: "한국어",
		IsRTL:      false,
	},
	LangSpanish: {
		Code:       LangSpanish,
		Name:       "Spanish",
		NativeName: "Español",
		IsRTL:      false,
	},
	LangFrench: {
		Code:       LangFrench,
		Name:       "French",
		NativeName: "Français",
		IsRTL:      false,
	},
	LangGerman: {
		Code:       LangGerman,
		Name:       "German",
		NativeName: "Deutsch",
		IsRTL:      false,
	},
	LangArabic: {
		Code:       LangArabic,
		Name:       "Arabic",
		NativeName: "العربية",
		IsRTL:      true,
		Punctuation: map[string]string{
			"?": "؟",
			",": "،",
			";": "؛",
		},
	},
	LangHebrew: {
		Code:       LangHebrew,
		Name:       "Hebrew",
		NativeName: "עברית",
		IsRTL:      true,
	},
	LangPersian: {
		Code:       LangPersian,
		Name:       "Persian",
		NativeName: "فارسی",
		IsRTL:      true,
		Punctuation: map[string]string{
			"?": "؟",
			",": "،",
			";": "؛",
		},
	},
	LangUrdu: {
		Code:       LangUrdu,
		Name:       "Urdu",
		NativeName: "اردو",
		IsRTL:      true,
		Punctuation: map[string]string{
			"?": "؟",
			",": "،",
		},
	},
}

// IsRTLLanguage checks if a language code is RTL (right-to-left)
func IsRTLLanguage(langCode string) bool {
	lang, exists := SupportedLanguages[langCode]
	if !exists {
		return false
	}
	return lang.IsRTL
}

// GetLanguage returns the language definition for a code
func GetLanguage(langCode string) (Language, bool) {
	lang, exists := SupportedLanguages[langCode]
	return lang, exists
}

// ValidateLanguageCode checks if a language code is supported
func ValidateLanguageCode(langCode string) bool {
	_, exists := SupportedLanguages[langCode]
	return exists
}
