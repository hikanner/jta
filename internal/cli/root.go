package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// Flags
	targetLangs     string
	providerFlag    string
	modelFlag       string
	apiKeyFlag      string
	outputFlag      string
	terminologyFlag string
	skipTerms       bool
	noTerminology   bool
	keysFlag        string
	excludeKeysFlag string
	forceFlag       bool
	batchSizeFlag   int
	concurrencyFlag int
	yesFlag         bool
	verboseFlag     bool
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jta <source> --to <languages>",
		Short: "Jta - Agentic JSON Translation Agent",
		Long: `Jta - AI-powered Agentic JSON Translation tool with intelligent quality optimization

A production-ready command-line tool that uses AI to translate JSON i18n files with
exceptional accuracy and consistency. Following Andrew Ng's Translation Agent methodology,
it features true Agentic reflection (translate → reflect → improve).

Key Features:
  ⭐ Agentic Translation - LLM self-evaluates and improves translations (3x API: translate, reflect, improve)
  📚 Smart Terminology - Auto-detection and consistent term management
  🔒 Format Protection - Preserves placeholders, HTML, URLs, and markdown
  ⚡ Incremental Mode - Only translates new/changed content
  🎯 Key Filtering - Selective translation with glob patterns
  🌍 RTL Support - Arabic, Hebrew, Persian, Urdu with bidirectional markers
  🚀 Multi-Provider - OpenAI (GPT-4o), Anthropic (Claude 3.5), Google (Gemini 2.0)
  �� Concurrent Processing - Fast batch translation with configurable concurrency`,
		Example: `  # Basic usage (Agentic reflection enabled by default)
  jta en.json --to zh

  # Multiple languages with shared terminology
  jta en.json --to zh,ja,ko

  # Use Claude for higher quality (recommended for production)
  jta en.json --to zh --provider anthropic --model claude-3-5-sonnet-20250116

  # With custom terminology file
  jta en.json --to zh --terminology ./config/tech-terms.json

  # Incremental translation (only new/changed content)
  jta en.json --to zh --output zh.json  # Run again after source changes

  # Selective translation with key filtering
  jta en.json --to zh --keys "settings.*,user.*" --exclude-keys "internal.*"

  # Fast mode: skip terminology detection
  jta en.json --to zh --skip-terms

  # Force complete re-translation (ignore existing)
  jta en.json --to zh --force`,
		Args: cobra.ExactArgs(1),
		RunE: runTranslate,
	}

	// Add flags
	cmd.Flags().StringVar(&targetLangs, "to", "", "Target language(s), comma-separated (e.g., zh,ja,ko) [REQUIRED]")
	cmd.MarkFlagRequired("to")

	// AI Provider settings
	cmd.Flags().StringVar(&providerFlag, "provider", "openai", "AI provider: openai, anthropic, or google")
	cmd.Flags().StringVar(&modelFlag, "model", "", "Model name (default: gpt-4o, claude-3-5-sonnet-20250116, gemini-2.0-flash-exp)")
	cmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "API key (or use OPENAI_API_KEY/ANTHROPIC_API_KEY/GEMINI_API_KEY env)")

	// Output settings
	cmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file path (default: <target-lang>.json in source directory)")
	cmd.Flags().StringVar(&terminologyFlag, "terminology", ".jta-terminology.json", "Terminology file path (auto-created if missing)")

	// Terminology management
	cmd.Flags().BoolVar(&skipTerms, "skip-terms", false, "Skip auto-detection (use existing terminology file only)")
	cmd.Flags().BoolVar(&noTerminology, "no-terminology", false, "Disable terminology management (faster but less consistent)")

	// Key filtering
	cmd.Flags().StringVar(&keysFlag, "keys", "", "Include only these keys (glob patterns, e.g., 'settings.*,user.*')")
	cmd.Flags().StringVar(&excludeKeysFlag, "exclude-keys", "", "Exclude these keys (glob patterns, e.g., 'internal.*,debug.*')")

	// Translation behavior
	cmd.Flags().BoolVar(&forceFlag, "force", false, "Force complete re-translation (ignore incremental mode)")

	// Performance tuning
	cmd.Flags().IntVar(&batchSizeFlag, "batch-size", 20, "Items per API call (10-50 recommended, larger = fewer calls but slower)")
	cmd.Flags().IntVar(&concurrencyFlag, "concurrency", 3, "Parallel API requests (1-5 recommended, higher = faster but may hit rate limits)")

	// UI behavior
	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Non-interactive mode (skip confirmations, useful for CI/CD)")
	cmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose output (show Agentic reflection steps and API details)")

	return cmd
}

func runTranslate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	sourcePath := args[0]

	// Validate source file exists
	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf("source file not found: %s", sourcePath)
	}

	// Parse target languages
	langs := strings.Split(targetLangs, ",")
	for i, lang := range langs {
		langs[i] = strings.TrimSpace(lang)
	}

	// Create application
	app, err := NewApp(ctx, AppConfig{
		Provider: providerFlag,
		Model:    modelFlag,
		APIKey:   apiKeyFlag,
		Verbose:  verboseFlag,
	})

	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	// Run translation for each target language
	for _, targetLang := range langs {
		fmt.Printf("\n🚀 Translating to %s...\n", targetLang)

		err := app.Translate(ctx, TranslateParams{
			SourcePath:    sourcePath,
			TargetLang:    targetLang,
			OutputPath:    outputFlag,
			TermPath:      terminologyFlag,
			SkipTerms:     skipTerms,
			NoTerminology: noTerminology,
			Keys:          keysFlag,
			ExcludeKeys:   excludeKeysFlag,
			Force:         forceFlag,
			BatchSize:     batchSizeFlag,
			Concurrency:   concurrencyFlag,
			Yes:           yesFlag,
		})

		if err != nil {
			return fmt.Errorf("translation failed for %s: %w", targetLang, err)
		}

		fmt.Printf("✅ Translation completed for %s\n", targetLang)
	}

	return nil
}

// Execute runs the root command
func Execute() error {
	return NewRootCmd().Execute()
}
