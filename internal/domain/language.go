package domain

// Language represents a supported language
type Language struct {
	Code         string
	Name         string
	NativeName   string
	Flag         string            // Country flag emoji
	IsRTL        bool              // Right-to-left text direction
	Script       string            // Writing system (e.g., "latn", "hans", "arab")
	NumberSystem string            // Number system (e.g., "arabic-indic", "persian")
	Punctuation  map[string]string // Special punctuation for the language
}

// Common language codes
const (
	LangEnglish            = "en"
	LangChineseSimplified  = "zh"
	LangChineseTraditional = "zh-TW"
	LangJapanese           = "ja"
	LangKorean             = "ko"
	LangSpanish            = "es"
	LangFrench             = "fr"
	LangGerman             = "de"
	LangItalian            = "it"
	LangPortuguese         = "pt"
	LangRussian            = "ru"
	LangArabic             = "ar"
	LangHindi              = "hi"
	LangBengali            = "bn"
	LangThai               = "th"
	LangVietnamese         = "vi"
	LangIndonesian         = "id"
	LangMalay              = "ms"
	LangDutch              = "nl"
	LangPolish             = "pl"
	LangTurkish            = "tr"
	LangPersian            = "fa"
	LangHebrew             = "he"
	LangUrdu               = "ur"
	LangSinhala            = "si"
	LangNepali             = "ne"
	LangBurmese            = "my"
)

// SupportedLanguages defines all supported languages
var SupportedLanguages = map[string]Language{
	LangEnglish: {
		Code:       LangEnglish,
		Name:       "English",
		NativeName: "English",
		Flag:       "🇬🇧",
		IsRTL:      false,
		Script:     "latn",
	},
	LangChineseSimplified: {
		Code:       LangChineseSimplified,
		Name:       "Chinese (Simplified)",
		NativeName: "中文(简体)",
		Flag:       "🇨🇳",
		IsRTL:      false,
		Script:     "hans",
	},
	LangChineseTraditional: {
		Code:       LangChineseTraditional,
		Name:       "Chinese (Traditional)",
		NativeName: "中文(繁体)",
		Flag:       "🇨🇳",
		IsRTL:      false,
		Script:     "hant",
	},
	LangJapanese: {
		Code:       LangJapanese,
		Name:       "Japanese",
		NativeName: "日本語",
		Flag:       "🇯🇵",
		IsRTL:      false,
		Script:     "jpan",
	},
	LangKorean: {
		Code:       LangKorean,
		Name:       "Korean",
		NativeName: "한국어",
		Flag:       "🇰🇷",
		IsRTL:      false,
		Script:     "kore",
	},
	LangSpanish: {
		Code:       LangSpanish,
		Name:       "Spanish",
		NativeName: "Español",
		Flag:       "🇪🇸",
		IsRTL:      false,
		Script:     "latn",
	},
	LangFrench: {
		Code:       LangFrench,
		Name:       "French",
		NativeName: "Français",
		Flag:       "🇫🇷",
		IsRTL:      false,
		Script:     "latn",
	},
	LangGerman: {
		Code:       LangGerman,
		Name:       "German",
		NativeName: "Deutsch",
		Flag:       "🇩🇪",
		IsRTL:      false,
		Script:     "latn",
	},
	LangItalian: {
		Code:       LangItalian,
		Name:       "Italian",
		NativeName: "Italiano",
		Flag:       "🇮🇹",
		IsRTL:      false,
		Script:     "latn",
	},
	LangPortuguese: {
		Code:       LangPortuguese,
		Name:       "Portuguese",
		NativeName: "Português",
		Flag:       "🇵🇹",
		IsRTL:      false,
		Script:     "latn",
	},
	LangRussian: {
		Code:       LangRussian,
		Name:       "Russian",
		NativeName: "Русский",
		Flag:       "🇷🇺",
		IsRTL:      false,
		Script:     "cyrl",
	},
	LangArabic: {
		Code:         LangArabic,
		Name:         "Arabic",
		NativeName:   "العربية",
		Flag:         "🇸🇦",
		IsRTL:        true,
		Script:       "arab",
		NumberSystem: "arabic-indic",
		Punctuation: map[string]string{
			"?": "؟",
			",": "،",
			";": "؛",
		},
	},
	LangHindi: {
		Code:       LangHindi,
		Name:       "Hindi",
		NativeName: "हिन्दी",
		Flag:       "🇮🇳",
		IsRTL:      false,
		Script:     "deva",
	},
	LangBengali: {
		Code:       LangBengali,
		Name:       "Bengali",
		NativeName: "বাংলা",
		Flag:       "🇧🇩",
		IsRTL:      false,
		Script:     "beng",
	},
	LangThai: {
		Code:       LangThai,
		Name:       "Thai",
		NativeName: "ไทย",
		Flag:       "🇹🇭",
		IsRTL:      false,
		Script:     "thai",
	},
	LangVietnamese: {
		Code:       LangVietnamese,
		Name:       "Vietnamese",
		NativeName: "Tiếng Việt",
		Flag:       "🇻🇳",
		IsRTL:      false,
		Script:     "latn",
	},
	LangIndonesian: {
		Code:       LangIndonesian,
		Name:       "Indonesian",
		NativeName: "Bahasa Indonesia",
		Flag:       "🇮🇩",
		IsRTL:      false,
		Script:     "latn",
	},
	LangMalay: {
		Code:       LangMalay,
		Name:       "Malay",
		NativeName: "Bahasa Melayu",
		Flag:       "🇲🇾",
		IsRTL:      false,
		Script:     "latn",
	},
	LangDutch: {
		Code:       LangDutch,
		Name:       "Dutch",
		NativeName: "Nederlands",
		Flag:       "🇳🇱",
		IsRTL:      false,
		Script:     "latn",
	},
	LangPolish: {
		Code:       LangPolish,
		Name:       "Polish",
		NativeName: "Polski",
		Flag:       "🇵🇱",
		IsRTL:      false,
		Script:     "latn",
	},
	LangTurkish: {
		Code:       LangTurkish,
		Name:       "Turkish",
		NativeName: "Türkçe",
		Flag:       "🇹🇷",
		IsRTL:      false,
		Script:     "latn",
	},
	LangPersian: {
		Code:         LangPersian,
		Name:         "Persian",
		NativeName:   "فارسی",
		Flag:         "🇮🇷",
		IsRTL:        true,
		Script:       "arab",
		NumberSystem: "persian",
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
		Flag:       "🇮🇱",
		IsRTL:      true,
		Script:     "hebr",
	},
	LangUrdu: {
		Code:         LangUrdu,
		Name:         "Urdu",
		NativeName:   "اردو",
		Flag:         "🇵🇰",
		IsRTL:        true,
		Script:       "arab",
		NumberSystem: "arabic-indic",
		Punctuation: map[string]string{
			"?": "؟",
			",": "،",
		},
	},
	LangSinhala: {
		Code:       LangSinhala,
		Name:       "Sinhala",
		NativeName: "සිංහල",
		Flag:       "🇱🇰",
		IsRTL:      false,
		Script:     "sinh",
	},
	LangNepali: {
		Code:       LangNepali,
		Name:       "Nepali",
		NativeName: "नेपाली",
		Flag:       "🇳🇵",
		IsRTL:      false,
		Script:     "deva",
	},
	LangBurmese: {
		Code:       LangBurmese,
		Name:       "Burmese",
		NativeName: "မြန်မာ",
		Flag:       "🇲🇲",
		IsRTL:      false,
		Script:     "mymr",
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
