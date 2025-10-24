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
		Short: "Jta - AI-powered JSON translation agent",
		Long: `Jta is an intelligent JSON translation tool powered by AI.

It provides:
  • Automatic terminology detection and management
  • Format preservation (placeholders, HTML, etc.)
  • Incremental translation (only translate changes)
  • Key filtering for selective translation
  • RTL language support
  • Multiple AI provider support (OpenAI, Anthropic, Google)`,
		Example: `  # Basic usage
  jta en.json --to zh

  # Multiple languages
  jta en.json --to zh,ja,ko

  # Specify provider and model
  jta en.json --to zh --provider anthropic --model claude-3-5-sonnet-20250116

  # Skip terminology detection
  jta en.json --to zh --skip-terms

  # Translate specific keys only
  jta en.json --to zh --keys "settings.*,user.*"

  # Force complete re-translation
  jta en.json --to zh --force`,
		Args: cobra.ExactArgs(1),
		RunE: runTranslate,
	}

	// Add flags
	cmd.Flags().StringVar(&targetLangs, "to", "", "Target language(s), comma-separated (required)")
	cmd.MarkFlagRequired("to")

	cmd.Flags().StringVar(&providerFlag, "provider", "openai", "AI provider (openai, anthropic, google)")
	cmd.Flags().StringVar(&modelFlag, "model", "", "Model name (uses default if not specified)")
	cmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "API key (or use environment variable)")

	cmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file or directory")
	cmd.Flags().StringVar(&terminologyFlag, "terminology", ".jta-terminology.json", "Terminology file path")

	cmd.Flags().BoolVar(&skipTerms, "skip-terms", false, "Skip term detection (still translates missing terms)")
	cmd.Flags().BoolVar(&noTerminology, "no-terminology", false, "Disable terminology management completely")

	cmd.Flags().StringVar(&keysFlag, "keys", "", "Only translate specified keys (glob patterns)")
	cmd.Flags().StringVar(&excludeKeysFlag, "exclude-keys", "", "Exclude specified keys (glob patterns)")

	cmd.Flags().BoolVar(&forceFlag, "force", false, "Force complete re-translation (ignore existing)")

	cmd.Flags().IntVar(&batchSizeFlag, "batch-size", 20, "Batch size for translation")
	cmd.Flags().IntVar(&concurrencyFlag, "concurrency", 3, "Concurrency for batch processing")

	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Non-interactive mode (auto-confirm)")
	cmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose output")

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
