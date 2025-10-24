package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/incremental"
	"github.com/hikanner/jta/internal/provider"
	"github.com/hikanner/jta/internal/terminology"
	"github.com/hikanner/jta/internal/translator"
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
	SourcePath    string
	TargetLang    string
	OutputPath    string
	TermPath      string
	SkipTerms     bool
	NoTerminology bool
	Keys          string
	ExcludeKeys   string
	Force         bool
	BatchSize     int
	Concurrency   int
	Yes           bool
}

// App is the main application
type App struct {
	provider    provider.AIProvider
	termManager *terminology.Manager
	engine      *translator.Engine
	incr        *incremental.Translator
	jsonUtil    *utils.JSONUtil
	config      AppConfig
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
	termRepo := terminology.NewJSONRepository()
	termManager := terminology.NewManager(prov, termRepo)

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
	}, nil
}

// Translate performs the translation workflow
func (a *App) Translate(ctx context.Context, params TranslateParams) error {
	// Step 1: Load source JSON
	fmt.Println("📖 Loading source file...")
	source, err := a.jsonUtil.LoadJSON(params.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to load source: %w", err)
	}

	// Step 2: Determine output path
	outputPath := params.OutputPath
	if outputPath == "" {
		// Default: same directory as source, with target language suffix
		dir := filepath.Dir(params.SourcePath)
		ext := filepath.Ext(params.SourcePath)
		outputPath = filepath.Join(dir, params.TargetLang+ext)
	}

	// Step 3: Check if target exists (for incremental translation)
	var target map[string]interface{}
	targetExists := false

	if !params.Force {
		if _, err := os.Stat(outputPath); err == nil {
			target, err = a.jsonUtil.LoadJSON(outputPath)
			if err == nil {
				targetExists = true
			}
		}
	}

	// Step 4: Analyze diff if target exists
	var diff *incremental.DiffResult
	if targetExists {
		fmt.Println("🔍 Analyzing changes...")
		diff, err = a.incr.AnalyzeDiff(source, target)
		if err != nil {
			return fmt.Errorf("failed to analyze diff: %w", err)
		}

		fmt.Printf("   ✨ New: %d keys\n", diff.Stats.NewCount)
		fmt.Printf("   🔄 Modified: %d keys\n", diff.Stats.ModifiedCount)
		fmt.Printf("   ✅ Unchanged: %d keys\n", diff.Stats.UnchangedCount)

		if !a.incr.ShouldTranslate(diff, params.Force) {
			fmt.Println("✅ No changes detected, skipping translation")
			return nil
		}

		if !params.Yes {
			fmt.Printf("\n💡 Will translate %d keys, keep %d unchanged. Continue? [Y/n] ",
				diff.Stats.NewCount+diff.Stats.ModifiedCount, diff.Stats.UnchangedCount)

			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) == "n" {
				return fmt.Errorf("cancelled by user")
			}
		}
	}

	// Step 5: Handle terminology
	var term *domain.Terminology

	if !params.NoTerminology {
		if a.termManager.TerminologyExists(params.TermPath) {
			fmt.Println("📚 Loading terminology...")
			term, err = a.termManager.LoadTerminology(params.TermPath)
			if err != nil {
				return fmt.Errorf("failed to load terminology: %w", err)
			}
		} else if !params.SkipTerms {
			fmt.Println("🔍 Detecting terminology...")

			// Extract texts for detection
			texts := extractTexts(source)

			terms, err := a.termManager.DetectTerms(ctx, texts, "en")
			if err != nil {
				fmt.Printf("⚠️  Failed to detect terms: %v\n", err)
			} else {
				// Build terminology
				term = &domain.Terminology{
					SourceLanguage:  "en",
					PreserveTerms:   []string{},
					ConsistentTerms: make(map[string][]string),
				}

				for _, t := range terms {
					if t.Type == domain.TermTypePreserve {
						term.AddPreserveTerm(t.Term)
					} else {
						term.AddConsistentTerm("en", t.Term)
					}
				}

				fmt.Printf("✨ Detected %d terms\n", len(terms))

				// Save terminology
				if !params.Yes {
					fmt.Printf("Save terminology to %s? [Y/n] ", params.TermPath)
					var response string
					fmt.Scanln(&response)
					if strings.ToLower(response) != "n" {
						err = a.termManager.SaveTerminology(params.TermPath, term)
						if err != nil {
							fmt.Printf("⚠️  Failed to save terminology: %v\n", err)
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
	fmt.Println("🤖 Translating...")

	result, err := a.engine.Translate(ctx, domain.TranslationInput{
		Source:      source,
		SourceLang:  "en",
		TargetLang:  params.TargetLang,
		Terminology: term,
		Options: domain.TranslationOptions{
			BatchSize:     params.BatchSize,
			Concurrency:   params.Concurrency,
			SkipTerms:     params.SkipTerms,
			NoTerminology: params.NoTerminology,
			Force:         params.Force,
			Keys:          keyPatterns,
			ExcludeKeys:   excludeKeyPatterns,
		},
	})

	if err != nil {
		return fmt.Errorf("translation failed: %w", err)
	}

	// Step 7: Save result
	fmt.Println("💾 Saving translation...")
	err = a.jsonUtil.SaveJSON(outputPath, result.Target)
	if err != nil {
		return fmt.Errorf("failed to save result: %w", err)
	}

	// Step 8: Print stats
	fmt.Printf("\n📊 Statistics:\n")

	// Print filter stats if filtering was applied
	if result.Stats.FilterStats != nil {
		fmt.Printf("   Filtered keys: %d included, %d excluded (of %d total)\n",
			result.Stats.FilterStats.IncludedKeys,
			result.Stats.FilterStats.ExcludedKeys,
			result.Stats.FilterStats.TotalKeys)
	}

	fmt.Printf("   Total items: %d\n", result.Stats.TotalItems)
	fmt.Printf("   Success: %d\n", result.Stats.SuccessItems)
	fmt.Printf("   Failed: %d\n", result.Stats.FailedItems)
	fmt.Printf("   Duration: %v\n", result.Stats.Duration)
	fmt.Printf("   API calls: %d\n", result.Stats.APICallsCount)
	fmt.Printf("   Output: %s\n", outputPath)

	return nil
}

func extractTexts(data interface{}) []string {
	var texts []string

	switch v := data.(type) {
	case map[string]interface{}:
		for _, value := range v {
			texts = append(texts, extractTexts(value)...)
		}
	case []interface{}:
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
