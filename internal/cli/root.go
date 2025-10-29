package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hikanner/jta/internal/domain"
	"github.com/spf13/cobra"
)

var (
	// Flags
	targetLangs        string
	providerFlag       string
	modelFlag          string
	apiKeyFlag         string
	sourceLangFlag     string
	outputFlag         string
	terminologyDirFlag string
	skipTerminology    bool
	noTerminology      bool
	redetectTerms      bool
	incrementalFlag    bool
	keysFlag           string
	excludeKeysFlag    string
	batchSizeFlag      int
	concurrencyFlag    int
	yesFlag            bool
	verboseFlag        bool
	listLanguagesFlag  bool
	versionFlag        bool
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
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
  🚀 Multi-Provider - OpenAI (GPT-5), Anthropic (Claude 4.5), Gemini (Gemini 2.5)
  �� Concurrent Processing - Fast batch translation with configurable concurrency`,
		Example: `  # Basic usage (Agentic reflection enabled by default)
  jta en.json --to zh

  # Multiple languages with shared terminology
  jta en.json --to zh,ja,ko

  # Use Claude for higher quality (recommended for production)
  jta en.json --to zh --provider anthropic --model claude-sonnet-4-5

  # With custom terminology directory
  jta en.json --to zh --terminology-dir ./config/.jta

  # Incremental translation (only new/changed content)
  jta en.json --to zh --incremental --output zh.json

  # Selective translation with key filtering
  jta en.json --to zh --keys "settings.*,user.*" --exclude-keys "internal.*"

  # Fast mode: skip terminology detection
  jta en.json --to zh --skip-terminology`,
		Args: cobra.MaximumNArgs(1),
		RunE: runTranslate,
	}

	// Add flags
	rootCmd.Flags().StringVar(&targetLangs, "to", "", "Target language(s), comma-separated (e.g., zh,ja,ko) [REQUIRED]")
	rootCmd.Flags().BoolVar(&listLanguagesFlag, "list-languages", false, "List all supported languages and exit")
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Print version information and exit")

	// AI Provider settings
	rootCmd.Flags().StringVar(&providerFlag, "provider", "openai", "AI provider: openai, anthropic, or gemini")
	rootCmd.Flags().StringVar(&modelFlag, "model", "", "Model name (default: gpt-5, claude-sonnet-4-5, gemini-2.5-flash)")
	rootCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "API key (or use OPENAI_API_KEY/ANTHROPIC_API_KEY/GEMINI_API_KEY env)")

	// Source settings
	rootCmd.Flags().StringVar(&sourceLangFlag, "source-lang", "", "Source language (auto-detected from filename if not specified)")

	// Output settings
	rootCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file path (default: <target-lang>.json in source directory)")

	// Terminology management
	rootCmd.Flags().StringVar(&terminologyDirFlag, "terminology-dir", ".jta", "Terminology directory (default: .jta/)")
	rootCmd.Flags().BoolVar(&skipTerminology, "skip-terminology", false, "Skip term detection (use existing terminology)")
	rootCmd.Flags().BoolVar(&noTerminology, "no-terminology", false, "Disable terminology management completely")
	rootCmd.Flags().BoolVar(&redetectTerms, "redetect-terms", false, "Re-detect terminology (use when source language changes)")

	// Translation behavior
	rootCmd.Flags().BoolVar(&incrementalFlag, "incremental", false, "Incremental translation (only translate new/modified content)")

	// Key filtering
	rootCmd.Flags().StringVar(&keysFlag, "keys", "", "Include only these keys (glob patterns, e.g., 'settings.*,user.*')")
	rootCmd.Flags().StringVar(&excludeKeysFlag, "exclude-keys", "", "Exclude these keys (glob patterns, e.g., 'internal.*,debug.*')")

	// Performance tuning
	rootCmd.Flags().IntVar(&batchSizeFlag, "batch-size", 20, "Items per API call (10-50 recommended, larger = fewer calls but slower)")
	rootCmd.Flags().IntVar(&concurrencyFlag, "concurrency", 3, "Parallel API requests (1-5 recommended, higher = faster but may hit rate limits)")

	// UI behavior
	rootCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Non-interactive mode (skip confirmations, useful for CI/CD)")
	rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose output (show Agentic reflection steps and API details)")

	return rootCmd
}

func runTranslate(cmd *cobra.Command, args []string) error {
	// Handle --version flag
	if versionFlag {
		PrintVersion()
		return nil
	}

	// Handle --list-languages flag
	if listLanguagesFlag {
		listLanguages()
		return nil
	}

	// Require source file when not listing languages
	if len(args) == 0 {
		return fmt.Errorf("source file is required")
	}

	// Require --to flag when translating
	if targetLangs == "" {
		return fmt.Errorf("--to flag is required")
	}

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
			SourcePath:      sourcePath,
			SourceLang:      sourceLangFlag,
			TargetLang:      targetLang,
			OutputPath:      outputFlag,
			TerminologyDir:  terminologyDirFlag,
			SkipTerminology: skipTerminology,
			NoTerminology:   noTerminology,
			RedetectTerms:   redetectTerms,
			Incremental:     incrementalFlag,
			Keys:            keysFlag,
			ExcludeKeys:     excludeKeysFlag,
			BatchSize:       batchSizeFlag,
			Concurrency:     concurrencyFlag,
			Yes:             yesFlag,
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

func listLanguages() {
	fmt.Println("🌍 Supported Languages")
	fmt.Println()

	// Separate LTR and RTL languages
	var ltrLangs []string
	var rtlLangs []string

	for code, lang := range domain.SupportedLanguages {
		entry := fmt.Sprintf("  %s  %-6s  %s (%s)", lang.Flag, code, lang.NativeName, lang.Name)
		if lang.IsRTL {
			rtlLangs = append(rtlLangs, entry)
		} else {
			ltrLangs = append(ltrLangs, entry)
		}
	}

	// Sort for consistent output
	sort.Strings(ltrLangs)
	sort.Strings(rtlLangs)

	// Print LTR languages
	fmt.Println("Left-to-Right (LTR):")
	for _, entry := range ltrLangs {
		fmt.Println(entry)
	}

	// Print RTL languages
	if len(rtlLangs) > 0 {
		fmt.Println()
		fmt.Println("Right-to-Left (RTL):")
		for _, entry := range rtlLangs {
			fmt.Println(entry)
		}
	}

	fmt.Println()
	fmt.Printf("Total: %d languages\n", len(domain.SupportedLanguages))
}
