package translator

import (
	"context"
	"fmt"
	"time"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/format"
	"github.com/hikanner/jta/internal/keyfilter"
	"github.com/hikanner/jta/internal/provider"
	"github.com/hikanner/jta/internal/rtl"
	"github.com/hikanner/jta/internal/terminology"
)

// Engine is the main translation engine
type Engine struct {
	provider         provider.AIProvider
	termManager      *terminology.Manager
	formatProtector  *format.Protector
	batchProcessor   *BatchProcessor
	keyFilter        *keyfilter.Filter
	rtlProcessor     *rtl.Processor
	reflectionEngine *ReflectionEngine
}

// NewEngine creates a new translation engine
func NewEngine(
	provider provider.AIProvider,
	termManager *terminology.Manager,
) *Engine {
	reflectionEngine := NewReflectionEngine(provider)
	return &Engine{
		provider:         provider,
		termManager:      termManager,
		formatProtector:  format.NewProtector(),
		batchProcessor:   NewBatchProcessor(provider, reflectionEngine),
		keyFilter:        keyfilter.NewFilter(),
		rtlProcessor:     rtl.NewProcessor(),
		reflectionEngine: reflectionEngine,
	}
}

// GetBatchProcessor returns the batch processor for setting callbacks
func (e *Engine) GetBatchProcessor() *BatchProcessor {
	return e.batchProcessor
}

// GetReflectionEngine returns the reflection engine for setting callbacks
func (e *Engine) GetReflectionEngine() *ReflectionEngine {
	return e.reflectionEngine
}

// Translate performs the complete translation workflow
func (e *Engine) Translate(ctx context.Context, input domain.TranslationInput) (*domain.TranslationResult, error) {
	startTime := time.Now()

	result := &domain.TranslationResult{
		Target: make(map[string]interface{}),
		Stats: domain.TranslationStats{
			APICallsCount: 0,
		},
	}

	// Step 1: Apply key filtering if patterns are provided
	sourceData := input.Source
	if len(input.Options.Keys) > 0 || len(input.Options.ExcludeKeys) > 0 {
		includePatterns, err := e.parseKeyPatterns(input.Options.Keys)
		if err != nil {
			return nil, domain.NewValidationError("failed to parse include patterns", err).
				WithContext("patterns", input.Options.Keys)
		}

		excludePatterns, err := e.parseKeyPatterns(input.Options.ExcludeKeys)
		if err != nil {
			return nil, domain.NewValidationError("failed to parse exclude patterns", err).
				WithContext("patterns", input.Options.ExcludeKeys)
		}

		filterResult, err := e.keyFilter.FilterKeys(input.Source, includePatterns, excludePatterns)
		if err != nil {
			return nil, domain.NewValidationError("failed to filter keys", err)
		}

		// Rebuild filtered JSON structure
		sourceData = e.keyFilter.RebuildJSON(filterResult.Included)

		// Store filter stats
		result.Stats.FilterStats = &domain.FilterStats{
			TotalKeys:    filterResult.Stats.TotalKeys,
			IncludedKeys: filterResult.Stats.IncludedKeys,
			ExcludedKeys: filterResult.Stats.ExcludedKeys,
		}
	}

	// Step 2: Extract translatable items from source JSON
	items, err := e.extractTranslatableItems(sourceData, "")
	if err != nil {
		return nil, domain.NewFormatError("failed to extract translatable items", err)
	}

	result.Stats.TotalItems = len(items)

	// Step 2: Load terminology (if not disabled)
	var terminology *domain.Terminology
	var terminologyTranslation *domain.TerminologyTranslation
	if !input.Options.NoTerminology {
		terminology = input.Terminology
		terminologyTranslation = input.TerminologyTranslation
	}

	// Step 3: Build terminology dictionary for prompt
	var termDict string
	if terminology != nil {
		termDict = e.termManager.BuildPromptDictionary(terminology, terminologyTranslation)
	}

	// Step 4: Create batches for translation
	batches := e.createBatches(items, input.Options.BatchSize)

	// Step 5: Process batches with concurrency (includes per-batch reflection)
	translations, stats, err := e.batchProcessor.ProcessBatches(
		ctx,
		batches,
		input.SourceLang,
		input.TargetLang,
		termDict,
		terminology,
		terminologyTranslation,
		input.Options.Concurrency,
	)

	if err != nil {
		return nil, domain.NewTranslationError("batch processing failed", err).
			WithContext("source_lang", input.SourceLang).
			WithContext("target_lang", input.TargetLang).
			WithContext("batch_count", len(batches))
	}

	// Update stats
	result.Stats.APICallsCount = stats.APICallsCount
	result.Stats.TotalTokens = stats.TotalTokens
	result.Stats.SuccessItems = len(translations)
	result.Stats.FailedItems = result.Stats.TotalItems - result.Stats.SuccessItems

	// Note: Reflection is now done per-batch in ProcessBatches for better scalability
	// No need for global reflection here

	// Step 5.5: Apply RTL processing if target language is RTL
	if e.rtlProcessor.NeedProcessing(input.TargetLang) {
		translations = e.rtlProcessor.ProcessBatch(translations, input.TargetLang)
	}

	// Step 6: Rebuild JSON structure with translations
	rebuilt := e.rebuildJSONWithPath(sourceData, translations, "")
	if targetMap, ok := rebuilt.(map[string]interface{}); ok {
		result.Target = targetMap
	} else {
		result.Target = sourceData // fallback to source if rebuild fails
	}

	// Calculate duration
	result.Stats.Duration = time.Since(startTime)

	return result, nil
}

// extractTranslatableItems recursively extracts all translatable text from JSON
func (e *Engine) extractTranslatableItems(data interface{}, prefix string) ([]domain.BatchItem, error) {
	var items []domain.BatchItem

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			keyPath := key
			if prefix != "" {
				keyPath = prefix + "." + key
			}
			subItems, err := e.extractTranslatableItems(value, keyPath)
			if err != nil {
				return nil, err
			}
			items = append(items, subItems...)
		}

	case []interface{}:
		for i, value := range v {
			keyPath := fmt.Sprintf("%s[%d]", prefix, i)
			subItems, err := e.extractTranslatableItems(value, keyPath)
			if err != nil {
				return nil, err
			}
			items = append(items, subItems...)
		}

	case string:
		// Only add non-empty strings
		if v != "" {
			items = append(items, domain.BatchItem{
				Key:     prefix,
				Text:    v,
				Context: e.inferContext(prefix),
				Value:   v,
			})
		}

		// Ignore other types (numbers, booleans, null)
	}

	return items, nil
}

// inferContext infers the context from the key path
func (e *Engine) inferContext(keyPath string) string {
	// Simple context inference based on key names
	if containsAny(keyPath, []string{"title", "name", "label"}) {
		return "title"
	}
	if containsAny(keyPath, []string{"description", "desc", "detail"}) {
		return "description"
	}
	if containsAny(keyPath, []string{"button", "action", "cta"}) {
		return "action"
	}
	if containsAny(keyPath, []string{"error", "warning", "alert"}) {
		return "message"
	}
	return "general"
}

// createBatches creates translation batches from items
func (e *Engine) createBatches(items []domain.BatchItem, batchSize int) [][]domain.BatchItem {
	if batchSize <= 0 {
		batchSize = 20 // default batch size
	}

	var batches [][]domain.BatchItem
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}

	return batches
}

// rebuildJSONWithPath rebuilds the JSON structure with translations, tracking key paths
func (e *Engine) rebuildJSONWithPath(source interface{}, translations map[string]string, currentPath string) interface{} {
	switch v := source.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			keyPath := key
			if currentPath != "" {
				keyPath = currentPath + "." + key
			}
			result[key] = e.rebuildJSONWithPath(value, translations, keyPath)
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			keyPath := fmt.Sprintf("%s[%d]", currentPath, i)
			result[i] = e.rebuildJSONWithPath(value, translations, keyPath)
		}
		return result

	case string:
		// Check if we have a translation for this key path
		if translation, ok := translations[currentPath]; ok {
			return translation
		}
		// Return original value if no translation found
		return v

	default:
		// Non-string values (numbers, booleans, null) remain unchanged
		return v
	}
}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

// parseKeyPatterns parses key patterns from string slice
func (e *Engine) parseKeyPatterns(patterns []string) ([]*keyfilter.KeyPattern, error) {
	if len(patterns) == 0 {
		return nil, nil
	}

	var result []*keyfilter.KeyPattern
	for _, pattern := range patterns {
		parsed, err := e.keyFilter.ParsePatterns(pattern)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed...)
	}

	return result, nil
}
