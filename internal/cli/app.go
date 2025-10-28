package cli

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/incremental"
	"github.com/hikanner/jta/internal/provider"
	"github.com/hikanner/jta/internal/terminology"
	"github.com/hikanner/jta/internal/translator"
	"github.com/hikanner/jta/internal/ui"
	"github.com/hikanner/jta/internal/utils"
)

// AppConfig contains application configuration
type AppConfig struct {
	Provider string
	Model    string
	APIKey   string
	Verbose  bool
}

// TranslateParams contains parameters for translation
type TranslateParams struct {
	SourcePath      string
	SourceLang      string
	TargetLang      string
	OutputPath      string
	TerminologyDir  string
	SkipTerminology bool
	NoTerminology   bool
	RedetectTerms   bool
	Incremental     bool
	Keys            string
	ExcludeKeys     string
	BatchSize       int
	Concurrency     int
	Yes             bool
}

// App is the main application
type App struct {
	provider    provider.AIProvider
	termManager *terminology.Manager
	engine      *translator.Engine
	incr        *incremental.Translator
	jsonUtil    *utils.JSONUtil
	config      AppConfig
	ui          *ui.Printer
}

// NewApp creates a new application instance
func NewApp(ctx context.Context, config AppConfig) (*App, error) {
	// Create AI provider
	providerType := provider.ProviderType(config.Provider)

	var prov provider.AIProvider
	var err error

	if config.APIKey != "" {
		// Use provided API key
		prov, err = provider.NewProvider(ctx, &provider.ProviderConfig{
			Type:   providerType,
			APIKey: config.APIKey,
			Model:  config.Model,
		})
	} else {
		// Use environment variable
		prov, err = provider.NewProviderFromEnv(ctx, providerType, config.Model)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Create terminology manager
	termManager := terminology.NewManager(prov)

	// Create translation engine
	engine := translator.NewEngine(prov, termManager)

	// Create incremental translator
	incrTranslator := incremental.NewTranslator()

	return &App{
		provider:    prov,
		termManager: termManager,
		engine:      engine,
		incr:        incrTranslator,
		jsonUtil:    utils.NewJSONUtil(),
		config:      config,
		ui:          ui.NewPrinter(config.Verbose),
	}, nil
}

// Translate performs the translation workflow
func (a *App) Translate(ctx context.Context, params TranslateParams) error {
	// Step 1: Load source JSON
	a.ui.PrintStep(ui.IconFile, "Loading source file...")
	source, err := a.jsonUtil.LoadJSON(params.SourcePath)
	if err != nil {
		a.ui.PrintError(fmt.Sprintf("Failed to load source: %v", err))
		return fmt.Errorf("failed to load source: %w", err)
	}
	a.ui.PrintSuccess("Source file loaded")

	// Step 2: Detect source language if not specified
	sourceLang := params.SourceLang
	if sourceLang == "" {
		// Auto-detect from filename (e.g., "en.json" -> "en")
		baseName := filepath.Base(params.SourcePath)
		sourceLang = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		a.ui.PrintSubtle(fmt.Sprintf("Detected source language: %s", sourceLang))
	}

	// Step 3: Determine output path
	outputPath := params.OutputPath
	if outputPath == "" {
		// Default: same directory as source, with target language suffix
		dir := filepath.Dir(params.SourcePath)
		ext := filepath.Ext(params.SourcePath)
		outputPath = filepath.Join(dir, params.TargetLang+ext)
	} else {
		// Check if outputPath is a directory
		if info, err := os.Stat(outputPath); err == nil && info.IsDir() {
			// If it's a directory, append the target language filename
			ext := filepath.Ext(params.SourcePath)
			outputPath = filepath.Join(outputPath, params.TargetLang+ext)
		}
	}

	// Step 4: Handle incremental translation mode
	var target map[string]any
	var diff *incremental.DiffResult

	if params.Incremental {
		// Incremental mode: check if target exists
		if _, err := os.Stat(outputPath); err == nil {
			target, err = a.jsonUtil.LoadJSON(outputPath)
			if err != nil {
				a.ui.PrintWarning(fmt.Sprintf("Failed to load existing target: %v", err))
			} else {
				// Analyze diff
				a.ui.PrintStep(ui.IconMagnify, "Analyzing changes (incremental mode)...")
				diff, err = a.incr.AnalyzeDiff(source, target)
				if err != nil {
					a.ui.PrintError(fmt.Sprintf("Failed to analyze diff: %v", err))
					return fmt.Errorf("failed to analyze diff: %w", err)
				}

				a.ui.PrintSubtle(fmt.Sprintf("New: %s keys", a.ui.FormatNumber(diff.Stats.NewCount)))
				a.ui.PrintSubtle(fmt.Sprintf("Modified: %s keys", a.ui.FormatNumber(diff.Stats.ModifiedCount)))
				a.ui.PrintSubtle(fmt.Sprintf("Unchanged: %s keys", a.ui.FormatNumber(diff.Stats.UnchangedCount)))

				if !a.incr.ShouldTranslate(diff, false) {
					a.ui.PrintSuccess("No changes detected, skipping translation")
					return nil
				}

				if !params.Yes {
					a.ui.PrintInfo(fmt.Sprintf("Will translate %d keys, keep %d unchanged",
						diff.Stats.NewCount+diff.Stats.ModifiedCount, diff.Stats.UnchangedCount))
					fmt.Print("Continue? [Y/n] ")

					var response string
					_, _ = fmt.Scanln(&response)
					if strings.ToLower(response) == "n" {
						a.ui.PrintWarning("Cancelled by user")
						return fmt.Errorf("cancelled by user")
					}
				}
			}
		}
	}

	// Step 5: Handle terminology
	var term *domain.Terminology
	var termTranslation *domain.TerminologyTranslation

	if !params.NoTerminology {
		// Load or detect terminology
		if a.termManager.TerminologyExists(params.TerminologyDir) {
			a.ui.PrintStep(ui.IconBook, "Loading terminology...")
			term, err = a.termManager.LoadTerminology(params.TerminologyDir)
			if err != nil {
				a.ui.PrintError(fmt.Sprintf("Failed to load terminology: %v", err))
				return fmt.Errorf("failed to load terminology: %w", err)
			}

			// Check source language match
			if term.SourceLanguage != sourceLang {
				if params.RedetectTerms {
					a.ui.PrintWarning(fmt.Sprintf("Source language changed (%s -> %s), re-detecting terminology...", term.SourceLanguage, sourceLang))
					term = nil // Force re-detection
				} else {
					a.ui.PrintWarning(fmt.Sprintf("Source language mismatch: terminology is %s, but source is %s", term.SourceLanguage, sourceLang))
					a.ui.PrintWarning("Use --redetect-terms to re-detect terminology for the new source language")
					return fmt.Errorf("source language mismatch")
				}
			} else {
				a.ui.PrintSuccess("Terminology loaded")
			}
		}

		// Detect terminology if not loaded
		if term == nil && !params.SkipTerminology {
			a.ui.PrintStep(ui.IconMagnify, "Detecting terminology...")

			// Extract texts for detection
			texts := extractTexts(source)
			a.ui.PrintSubtle(fmt.Sprintf("Extracted %d text strings from source file", len(texts)))

			terms, err := a.termManager.DetectTerms(ctx, texts, sourceLang)
			if err != nil {
				a.ui.PrintWarning(fmt.Sprintf("Failed to detect terms: %v", err))
			} else {
				// Build terminology
				term = &domain.Terminology{
					SourceLanguage:  sourceLang,
					PreserveTerms:   []string{},
					ConsistentTerms: []string{},
				}

				for _, t := range terms {
					if t.Type == domain.TermTypePreserve {
						term.PreserveTerms = append(term.PreserveTerms, t.Term)
					} else {
						term.ConsistentTerms = append(term.ConsistentTerms, t.Term)
					}
				}

				a.ui.PrintSuccess(fmt.Sprintf("Detected %s terms", a.ui.FormatNumber(len(terms))))
				a.ui.PrintSubtle(fmt.Sprintf("  Preserve: %d, Consistent: %d", len(term.PreserveTerms), len(term.ConsistentTerms)))

				// Save terminology
				shouldSave := params.Yes
				if !params.Yes {
					fmt.Printf("Save terminology to %s/terminology.json? [Y/n] ", params.TerminologyDir)
					var response string
					_, _ = fmt.Scanln(&response)
					shouldSave = strings.ToLower(response) != "n"
				}

				if shouldSave {
					err = a.termManager.SaveTerminology(params.TerminologyDir, term)
					if err != nil {
						a.ui.PrintWarning(fmt.Sprintf("Failed to save terminology: %v", err))
					} else {
						a.ui.PrintSuccess("Terminology saved")
					}
				}
			}
		}

		// Load or translate terminology
		if term != nil && len(term.ConsistentTerms) > 0 {
			if a.termManager.TranslationExists(params.TerminologyDir, params.TargetLang) {
				a.ui.PrintStep(ui.IconBook, "Loading terminology translation...")
				termTranslation, err = a.termManager.LoadTerminologyTranslation(params.TerminologyDir, params.TargetLang)
				if err != nil {
					a.ui.PrintError(fmt.Sprintf("Failed to load translation: %v", err))
					return fmt.Errorf("failed to load terminology translation: %w", err)
				}
				a.ui.PrintSuccess("Terminology translation loaded")
			}

			// Check for missing translations
			missingTerms := term.GetMissingTranslations(termTranslation)
			if len(missingTerms) > 0 {
				a.ui.PrintStep(ui.IconRobot, fmt.Sprintf("Translating %d missing terms...", len(missingTerms)))

				// Create timeout context for term translation (1 minute)
				translateCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
				defer cancel()

				translations, err := a.termManager.TranslateTerms(translateCtx, missingTerms, sourceLang, params.TargetLang)
				if err != nil {
					a.ui.PrintWarning(fmt.Sprintf("Failed to translate terms: %v", err))
				} else {
					// Create or update translation
					if termTranslation == nil {
						termTranslation = &domain.TerminologyTranslation{
							SourceLanguage: sourceLang,
							TargetLanguage: params.TargetLang,
							Translations:   translations,
						}
					} else {
						// Merge new translations
						maps.Copy(termTranslation.Translations, translations)
					}

					a.ui.PrintSuccess("Terms translated")

					// Save translation
					shouldSave := params.Yes
					if !params.Yes {
						fmt.Printf("Save terminology translation to %s/terminology.%s.json? [Y/n] ", params.TerminologyDir, params.TargetLang)
						var response string
						_, _ = fmt.Scanln(&response)
						shouldSave = strings.ToLower(response) != "n"
					}

					if shouldSave {
						err = a.termManager.SaveTerminologyTranslation(params.TerminologyDir, termTranslation)
						if err != nil {
							a.ui.PrintWarning(fmt.Sprintf("Failed to save translation: %v", err))
						} else {
							a.ui.PrintSuccess("Terminology translation saved")
						}
					}
				}
			}
		}
	}

	// Step 6: Parse key filter patterns
	var keyPatterns []string
	var excludeKeyPatterns []string

	if params.Keys != "" {
		keyPatterns = []string{params.Keys}
	}

	if params.ExcludeKeys != "" {
		excludeKeyPatterns = []string{params.ExcludeKeys}
	}

	// Step 7: Translate
	a.ui.PrintStep(ui.IconRobot, "Translating...")

	// Setup progress callbacks for detailed output
	a.setupProgressCallbacks()

	result, err := a.engine.Translate(ctx, domain.TranslationInput{
		Source:                 source,
		SourceLang:             sourceLang,
		TargetLang:             params.TargetLang,
		Terminology:            term,
		TerminologyTranslation: termTranslation,
		Options: domain.TranslationOptions{
			BatchSize:     params.BatchSize,
			Concurrency:   params.Concurrency,
			SkipTerms:     params.SkipTerminology,
			NoTerminology: params.NoTerminology,
			Incremental:   params.Incremental,
			Keys:          keyPatterns,
			ExcludeKeys:   excludeKeyPatterns,
		},
	})

	if err != nil {
		a.ui.PrintError(fmt.Sprintf("Translation failed: %v", err))
		return fmt.Errorf("translation failed: %w", err)
	}

	// Print completion summary
	fmt.Println()
	a.ui.PrintSuccess("Translation completed")

	// Step 8: Save result
	a.ui.PrintStep(ui.IconSave, "Saving translation...")
	err = a.jsonUtil.SaveJSON(outputPath, result.Target)
	if err != nil {
		a.ui.PrintError(fmt.Sprintf("Failed to save: %v", err))
		return fmt.Errorf("failed to save result: %w", err)
	}
	a.ui.PrintSuccess(fmt.Sprintf("Saved to %s", outputPath))

	// Step 9: Print stats
	fmt.Println() // Empty line for spacing
	a.ui.PrintHeader("Translation Statistics")

	// Build stats map
	stats := make(map[string]any)

	// Print filter stats if filtering was applied
	if result.Stats.FilterStats != nil {
		stats["Filtered"] = fmt.Sprintf("%d included, %d excluded (of %d total)",
			result.Stats.FilterStats.IncludedKeys,
			result.Stats.FilterStats.ExcludedKeys,
			result.Stats.FilterStats.TotalKeys)
	}

	stats["Total items"] = result.Stats.TotalItems
	stats["Success"] = result.Stats.SuccessItems
	stats["Failed"] = result.Stats.FailedItems
	stats["Duration"] = result.Stats.Duration.String()
	stats["API calls"] = result.Stats.APICallsCount

	a.ui.PrintStats(stats)

	return nil
}

func extractTexts(data any) []string {
	var texts []string

	switch v := data.(type) {
	case map[string]any:
		for _, value := range v {
			texts = append(texts, extractTexts(value)...)
		}
	case []any:
		for _, value := range v {
			texts = append(texts, extractTexts(value)...)
		}
	case string:
		if v != "" {
			texts = append(texts, v)
		}
	}

	return texts
}

// setupProgressCallbacks configures progress callbacks for detailed output
func (a *App) setupProgressCallbacks() {
	var batchCount int
	var successCount int
	var retryCount int
	var mu sync.Mutex

	// Setup batch processor callback
	a.engine.GetBatchProcessor().SetProgressCallback(func(event translator.BatchProgressEvent) {
		mu.Lock()
		defer mu.Unlock()

		switch event.Type {
		case "start":
			if batchCount == 0 {
				// First batch - print preparation info
				fmt.Printf("   📦 Split into %d batches (%d items/batch, concurrency: %d)\n",
					event.TotalBatches, event.BatchSize, event.Concurrency)
				fmt.Println()
			}
			batchCount++
			fmt.Printf("[Batch %d] 📝 Translating...\n", event.BatchIndex)

		case "complete":
			successCount++
			fmt.Printf("[Batch %d] ✓ Translated      (%.1fs, %d tokens)\n",
				event.BatchIndex, event.Duration.Seconds(), event.Tokens)

		case "retry":
			retryCount++
			fmt.Printf("[Batch %d] ⚠️  Translation failed (attempt %d/%d): %v\n",
				event.BatchIndex, event.Attempt, event.MaxAttempts, event.Error)
			fmt.Printf("[Batch %d] ⏰  Retry in %ds...\n", event.BatchIndex, 1<<uint(event.Attempt-1))

		case "error":
			fmt.Printf("[Batch %d] ❌ Translation failed after %d attempts: %v\n",
				event.BatchIndex, event.MaxAttempts, event.Error)
		}
	})

	// Setup reflection engine callback
	// Reflection progress is now logged per-batch in BatchProcessor
	// No need for global reflection callback
}
