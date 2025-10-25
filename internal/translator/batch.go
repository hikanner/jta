package translator

import (
	"context"
	"fmt"
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

// BatchProcessor handles batch translation with concurrency
type BatchProcessor struct {
	provider        provider.AIProvider
	formatProtector *format.Protector
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(provider provider.AIProvider) *BatchProcessor {
	return &BatchProcessor{
		provider:        provider,
		formatProtector: format.NewProtector(),
	}
}

// ProcessBatches processes multiple batches with concurrency control
func (bp *BatchProcessor) ProcessBatches(
	ctx context.Context,
	batches [][]domain.BatchItem,
	sourceLang, targetLang string,
	termDict string,
	concurrency int,
) (map[string]string, BatchStats, error) {
	if concurrency <= 0 {
		concurrency = 3 // default concurrency
	}

	results := make(map[string]string)
	var resultsMu sync.Mutex

	stats := BatchStats{}
	var statsMu sync.Mutex

	// Create error group with concurrency limit
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	// Process each batch
	for i, batch := range batches {
		batchIdx := i
		batchItems := batch

		g.Go(func() error {
			// Process single batch
			batchResults, batchTokens, err := bp.processSingleBatch(
				ctx,
				batchItems,
				sourceLang,
				targetLang,
				termDict,
			)

			if err != nil {
				return domain.NewTranslationError(fmt.Sprintf("batch %d failed", batchIdx), err).
					WithContext("batch_index", batchIdx).
					WithContext("batch_size", len(batchItems))
			}

			// Update results
			resultsMu.Lock()
			for key, value := range batchResults {
				results[key] = value
			}
			resultsMu.Unlock()

			// Update stats
			statsMu.Lock()
			stats.APICallsCount++
			stats.TotalTokens += batchTokens
			statsMu.Unlock()

			return nil
		})
	}

	// Wait for all batches to complete
	if err := g.Wait(); err != nil {
		return results, stats, err
	}

	return results, stats, nil
}

// processSingleBatch processes a single batch of items
func (bp *BatchProcessor) processSingleBatch(
	ctx context.Context,
	items []domain.BatchItem,
	sourceLang, targetLang string,
	termDict string,
) (map[string]string, int, error) {
	// Build batch translation prompt
	prompt := bp.buildBatchPrompt(items, sourceLang, targetLang, termDict)

	// Call AI provider with retries
	var resp *provider.CompletionResponse
	var err error
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = bp.provider.Complete(ctx, &provider.CompletionRequest{
			Prompt:      prompt,
			Temperature: 0.3,
			MaxTokens:   4000,
		})

		if err == nil {
			break
		}

		// Exponential backoff
		if attempt < maxRetries-1 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}
	}

	if err != nil {
		return nil, 0, domain.NewTranslationError(fmt.Sprintf("failed after %d retries", maxRetries), err).
			WithContext("retries", maxRetries).
			WithContext("item_count", len(items))
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
			if err := bp.formatProtector.Validate(original, translated); err != nil {
				// Log warning but don't fail
				// In production, might want to retry or fix automatically
			}
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
	lines := strings.Split(content, "\n")

	for _, line := range lines {
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
