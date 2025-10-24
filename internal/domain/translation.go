package domain

import "time"

// TranslationInput represents the input for translation
type TranslationInput struct {
	Source      map[string]interface{} // Source JSON data
	SourceLang  string
	TargetLang  string
	Terminology *Terminology
	Options     TranslationOptions
}

// TranslationOptions contains options for translation
type TranslationOptions struct {
	BatchSize     int
	Concurrency   int
	SkipTerms     bool // Skip term detection (but still translate missing terms)
	NoTerminology bool // Completely disable terminology management
	Force         bool // Force complete re-translation (ignore existing)
	Keys          []string
	ExcludeKeys   []string
}

// TranslationResult represents the result of translation
type TranslationResult struct {
	Target map[string]interface{} // Translated JSON data
	Stats  TranslationStats
	Errors []TranslationError
}

// TranslationStats contains statistics about the translation
type TranslationStats struct {
	TotalItems       int
	SuccessItems     int
	FailedItems      int
	SkippedItems     int
	Duration         time.Duration
	APICallsCount    int
	TotalTokens      int
	EstimatedCost    float64
	IncrementalStats *IncrementalStats // Only present for incremental translation
	FilterStats      *FilterStats      // Only present when key filtering is used
}

// IncrementalStats contains statistics for incremental translation
type IncrementalStats struct {
	NewKeys       int
	ModifiedKeys  int
	DeletedKeys   int
	UnchangedKeys int
}

// FilterStats contains statistics for key filtering
type FilterStats struct {
	TotalKeys    int
	IncludedKeys int
	ExcludedKeys int
}

// TranslationError represents an error during translation
type TranslationError struct {
	Key         string
	Message     string
	IsRetryable bool
}

// BatchItem represents a single item in a translation batch
type BatchItem struct {
	Key     string      // JSON key path (e.g., "settings.title")
	Text    string      // Text to translate
	Context string      // Context for the translation
	Value   interface{} // Original value (for non-string types)
}

// TranslatedItem represents a translated item
type TranslatedItem struct {
	Key            string
	OriginalText   string
	TranslatedText string
	Success        bool
	Error          string
}
