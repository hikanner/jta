package translator

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/format"
	"github.com/hikanner/jta/internal/provider"
	"golang.org/x/sync/errgroup"
)

// BatchStats contains statistics from batch processing
type BatchStats struct {
	APICallsCount int
	TotalTokens   int
}

// BatchProgressCallback is called for batch progress updates
type BatchProgressCallback func(event BatchProgressEvent)

// BatchProgressEvent represents a batch processing event
type BatchProgressEvent struct {
	Type         string // "start", "complete", "retry", "error"
	BatchIndex   int
	TotalBatches int
	BatchSize    int
	Concurrency  int
	Attempt      int
	MaxAttempts  int
	Duration     time.Duration
	Tokens       int
	Error        error
}

// BatchProcessor handles batch translation with concurrency
type BatchProcessor struct {
	provider         provider.AIProvider
	formatProtector  *format.Protector
	reflectionEngine *ReflectionEngine
	progressCallback BatchProgressCallback
}

// SetProgressCallback sets the progress callback function
func (bp *BatchProcessor) SetProgressCallback(callback BatchProgressCallback) {
	bp.progressCallback = callback
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(provider provider.AIProvider, reflectionEngine *ReflectionEngine) *BatchProcessor {
	return &BatchProcessor{
		provider:         provider,
		formatProtector:  format.NewProtector(),
		reflectionEngine: reflectionEngine,
	}
}

// ProcessBatches processes multiple batches with concurrency control
func (bp *BatchProcessor) ProcessBatches(
	ctx context.Context,
	batches [][]domain.BatchItem,
	sourceLang, targetLang string,
	termDict string,
	terminology *domain.Terminology,
	terminologyTranslation *domain.TerminologyTranslation,
	concurrency int,
) (map[string]string, BatchStats, error) {
	if concurrency <= 0 {
		concurrency = 3 // default concurrency
	}

	results := make(map[string]string)
	var resultsMu sync.Mutex

	stats := BatchStats{}
	var statsMu sync.Mutex

	// Track failed batches
	failedBatches := make(map[int]error)
	var failedMu sync.Mutex

	// Create error group with concurrency limit
	// Don't use errgroup's context to avoid canceling other batches on first failure
	g := new(errgroup.Group)
	g.SetLimit(concurrency)

	// Process each batch
	for i, batch := range batches {
		batchIdx := i
		batchItems := batch

		g.Go(func() error {
			// Record total batch time (including translation and reflection)
			batchTotalStart := time.Now()

			// Notify batch start
			if bp.progressCallback != nil {
				bp.progressCallback(BatchProgressEvent{
					Type:         "start",
					BatchIndex:   batchIdx + 1,
					TotalBatches: len(batches),
					BatchSize:    len(batchItems),
					Concurrency:  concurrency,
				})
			}

			// Process with retries
			maxRetries := 3
			var batchResults map[string]string
			var batchTokens int
			var err error

			for attempt := range maxRetries {
				startTime := time.Now()

				batchResults, batchTokens, err = bp.processSingleBatchOnce(
					ctx,
					batchItems,
					sourceLang,
					targetLang,
					termDict,
				)

				duration := time.Since(startTime)

				if err == nil {
					// Success
					if bp.progressCallback != nil {
						bp.progressCallback(BatchProgressEvent{
							Type:         "complete",
							BatchIndex:   batchIdx + 1,
							TotalBatches: len(batches),
							BatchSize:    len(batchItems),
							Concurrency:  concurrency,
							Duration:     duration,
							Tokens:       batchTokens,
						})
					}
					break
				}

				// Failed
				if attempt < maxRetries-1 {
					// Will retry
					if bp.progressCallback != nil {
						bp.progressCallback(BatchProgressEvent{
							Type:         "retry",
							BatchIndex:   batchIdx + 1,
							TotalBatches: len(batches),
							BatchSize:    len(batchItems),
							Concurrency:  concurrency,
							Attempt:      attempt + 1,
							MaxAttempts:  maxRetries,
							Error:        err,
						})
					}

					// Exponential backoff
					backoff := time.Duration(1<<uint(attempt)) * time.Second
					time.Sleep(backoff)
				} else {
					// Final failure
					if bp.progressCallback != nil {
						bp.progressCallback(BatchProgressEvent{
							Type:         "error",
							BatchIndex:   batchIdx + 1,
							TotalBatches: len(batches),
							BatchSize:    len(batchItems),
							Concurrency:  concurrency,
							Attempt:      attempt + 1,
							MaxAttempts:  maxRetries,
							Duration:     duration,
							Error:        err,
						})
					}
				}
			}

			if err != nil {
				// Record failure but don't return error (don't cancel other batches)
				failedMu.Lock()
				failedBatches[batchIdx] = err
				failedMu.Unlock()
				return nil // Don't propagate error to avoid canceling other batches
			}

			// Apply reflection to this batch if reflection engine is available
			if bp.reflectionEngine != nil && bp.reflectionEngine.ShouldReflect(batchResults, terminology) {
				// Build reflection input for this batch
				reflectionInput := ReflectionInput{
					SourceTexts:            make(map[string]string),
					TranslatedTexts:        batchResults,
					SourceLang:             sourceLang,
					TargetLang:             targetLang,
					Terminology:            terminology,
					TerminologyTranslation: terminologyTranslation,
				}

				// Extract source texts from batch items
				for _, item := range batchItems {
					reflectionInput.SourceTexts[item.Key] = item.Text
				}

				// Create progress callback for this batch (captures batchIdx for this specific batch)
				progressCallback := func(event ReflectionProgressEvent) {
					switch event.Type {
					case "reflecting_start":
						fmt.Printf("[Batch %d] 🔍 Reflecting...\n", batchIdx+1)
					case "reflected_complete":
						if event.Count > 0 {
							fmt.Printf("[Batch %d] ✓ Reflected       (%.1fs) Found %d suggestions\n",
								batchIdx+1, event.Duration.Seconds(), event.Count)
						} else {
							fmt.Printf("[Batch %d] ✓ Reflected       (%.1fs) All OK\n",
								batchIdx+1, event.Duration.Seconds())
						}
					case "improving_start":
						fmt.Printf("[Batch %d] ✨ Improving...\n", batchIdx+1)
					case "improved_complete":
						fmt.Printf("[Batch %d] ✓ Improved        (%.1fs) Updated %d translations\n",
							batchIdx+1, event.Duration.Seconds(), event.Count)
					}
				}

				// Perform reflection with the callback
				reflectionResult, reflectErr := bp.reflectionEngine.Reflect(ctx, reflectionInput, progressCallback)

				if reflectErr != nil {
					// Log error but don't fail the batch
					fmt.Printf("[Batch %d] ✗ Reflection failed: %v\n", batchIdx+1, reflectErr)
				} else if reflectionResult.ReflectionNeeded && len(reflectionResult.ImprovedTexts) > 0 {
					// Apply improvements
					maps.Copy(batchResults, reflectionResult.ImprovedTexts)

					// Update API call count
					statsMu.Lock()
					stats.APICallsCount += reflectionResult.APICallsUsed
					statsMu.Unlock()
				}
			}

			// Update results
			resultsMu.Lock()
			maps.Copy(results, batchResults)
			resultsMu.Unlock()

			// Update stats
			statsMu.Lock()
			stats.APICallsCount++
			stats.TotalTokens += batchTokens
			statsMu.Unlock()

			// Print final completion
			batchTotalElapsed := time.Since(batchTotalStart)
			if bp.reflectionEngine != nil && bp.reflectionEngine.ShouldReflect(batchResults, terminology) {
				fmt.Printf("[Batch %d] ✅ COMPLETE       (total: %.1fs)\n", batchIdx+1, batchTotalElapsed.Seconds())
			} else {
				fmt.Printf("[Batch %d] ✅ COMPLETE       (%.1fs, no reflection)\n", batchIdx+1, batchTotalElapsed.Seconds())
			}

			return nil
		})
	}

	// Wait for all batches to complete
	if err := g.Wait(); err != nil {
		return results, stats, err
	}

	// Check if any batches failed
	if len(failedBatches) > 0 {
		// If too many batches failed, return error
		failureRate := float64(len(failedBatches)) / float64(len(batches))
		if failureRate > 0.5 {
			// More than 50% failed - critical error
			firstFailedIdx := -1
			var firstError error
			for idx, err := range failedBatches {
				if firstFailedIdx == -1 || idx < firstFailedIdx {
					firstFailedIdx = idx
					firstError = err
				}
			}
			return results, stats, domain.NewTranslationError(
				fmt.Sprintf("batch %d failed", firstFailedIdx+1),
				firstError,
			).WithContext("failed_batches", len(failedBatches)).
				WithContext("total_batches", len(batches))
		}
		// If less than 50% failed, continue with partial results
		// The caller will get partial translations
	}

	return results, stats, nil
}

// processSingleBatchOnce processes a single batch of items (one attempt, no retries)
func (bp *BatchProcessor) processSingleBatchOnce(
	ctx context.Context,
	items []domain.BatchItem,
	sourceLang, targetLang string,
	termDict string,
) (map[string]string, int, error) {
	// Build batch translation prompt
	prompt := bp.buildBatchPrompt(items, sourceLang, targetLang, termDict)

	// Create independent 5-minute timeout for this LLM call
	callCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Call AI provider (single attempt)
	resp, err := bp.provider.Complete(callCtx, &provider.CompletionRequest{
		Prompt: prompt,
	})

	if err != nil {
		return nil, 0, err
	}

	// Parse response
	results, err := bp.parseBatchResponse(resp.Content, items)
	if err != nil {
		return nil, resp.Usage.TotalTokens, domain.NewFormatError("failed to parse response", err).
			WithContext("item_count", len(items))
	}

	// Validate format preservation
	for key, translated := range results {
		original := ""
		for _, item := range items {
			if item.Key == key {
				original = item.Text
				break
			}
		}

		if original != "" {
			// Validate format markers, but don't fail on validation errors
			// In production, might want to log warnings or fix automatically
			_ = bp.formatProtector.Validate(original, translated)
		}
	}

	return results, resp.Usage.TotalTokens, nil
}

// buildBatchPrompt builds the prompt for batch translation
func (bp *BatchProcessor) buildBatchPrompt(
	items []domain.BatchItem,
	sourceLang, targetLang string,
	termDict string,
) string {
	var builder strings.Builder

	// Get target language name
	targetLangName := targetLang // simplified, should use language map

	builder.WriteString(fmt.Sprintf(`You are a professional localization translator specialized in UI/UX content.

Task: Translate the following %s texts to %s for a JSON i18n file.

`, sourceLang, targetLangName))

	// Add terminology dictionary if provided
	if termDict != "" {
		builder.WriteString("【Terminology Dictionary】\n")
		builder.WriteString(termDict)
		builder.WriteString("\n\n")
	}

	// Add format instructions
	builder.WriteString(`【Core Requirements】
1. 🔒 Keep all placeholders unchanged (e.g., {variable}, {{count}})
2. 🏷️ Keep all HTML tags and special markers unchanged
3. 📝 Follow terminology translations EXACTLY
4. ⚡ Return format: [ID] translation (no additional explanation)
5. 🎯 Maintain context consistency across related texts

`)

	// Add items
	builder.WriteString("【Texts to Translate】\n")
	for i, item := range items {
		builder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, item.Text))
	}

	builder.WriteString("\n【Translation Results】\n")

	return builder.String()
}

// parseBatchResponse parses the batch translation response
func (bp *BatchProcessor) parseBatchResponse(content string, items []domain.BatchItem) (map[string]string, error) {
	results := make(map[string]string)

	// Parse line by line
	lines := strings.SplitSeq(content, "\n")

	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for pattern: [N] translation
		if strings.HasPrefix(line, "[") {
			closeBracket := strings.Index(line, "]")
			if closeBracket > 0 {
				idxStr := line[1:closeBracket]
				translation := strings.TrimSpace(line[closeBracket+1:])

				// Parse index
				var idx int
				_, err := fmt.Sscanf(idxStr, "%d", &idx)
				if err == nil && idx > 0 && idx <= len(items) {
					item := items[idx-1]
					results[item.Key] = translation
				}
			}
		}
	}

	// If parsing failed, try to extract any useful translations
	if len(results) == 0 {
		return nil, domain.NewFormatError("failed to parse translations from response", nil).
			WithContext("item_count", len(items))
	}

	return results, nil
}
