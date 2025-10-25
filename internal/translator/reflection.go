package translator

import (
	"context"
	"fmt"
	"strings"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/format"
	"github.com/hikanner/jta/internal/provider"
)

// ReflectionEngine implements Agentic reflection mechanism
// Following Andrew Ng's Translation Agent approach with two-step process:
// Step 1: Reflect - LLM evaluates translation quality and provides suggestions
// Step 2: Improve - LLM improves translation based on suggestions
// Target: 3x API calls (translate → reflect → improve)
type ReflectionEngine struct {
	provider        provider.AIProvider
	formatProtector *format.Protector
	maxTokens       int
	temperature     float32
}

// NewReflectionEngine creates a new reflection engine
func NewReflectionEngine(prov provider.AIProvider) *ReflectionEngine {
	return &ReflectionEngine{
		provider:        prov,
		formatProtector: format.NewProtector(),
		maxTokens:       3000,
		temperature:     0.3,
	}
}

// ReflectionInput contains input for reflection
type ReflectionInput struct {
	SourceTexts     map[string]string // key -> source text
	TranslatedTexts map[string]string // key -> translated text
	SourceLang      string
	TargetLang      string
	Terminology     *domain.Terminology
}

// ReflectionResult contains reflection results
type ReflectionResult struct {
	Suggestions      map[string]string // key -> expert suggestions from LLM
	ImprovedTexts    map[string]string // key -> improved translations
	ReflectionNeeded bool
	APICallsUsed     int // Should be 2 (reflect + improve)
}

// Reflect performs Agentic reflection on translations
// This is the main entry point following Andrew Ng's two-step approach
func (r *ReflectionEngine) Reflect(ctx context.Context, input ReflectionInput) (*ReflectionResult, error) {
	result := &ReflectionResult{
		Suggestions:   make(map[string]string),
		ImprovedTexts: make(map[string]string),
	}

	// For Agentic reflection, we always reflect on all translations
	// This allows LLM to discover quality issues that rule-based checks might miss
	if len(input.TranslatedTexts) == 0 {
		result.ReflectionNeeded = false
		return result, nil
	}

	result.ReflectionNeeded = true

	// Step 1: Reflection - LLM evaluates translations and provides suggestions
	suggestions, err := r.reflectStep(ctx, input)
	if err != nil {
		return nil, domain.NewTranslationError("reflection step failed", err).
			WithContext("source_lang", input.SourceLang).
			WithContext("target_lang", input.TargetLang).
			WithContext("translation_count", len(input.TranslatedTexts))
	}
	result.Suggestions = suggestions
	result.APICallsUsed++ // +1 API call for reflection

	// Step 2: Improvement - LLM improves translations based on suggestions
	improved, err := r.improveStep(ctx, input, suggestions)
	if err != nil {
		return nil, domain.NewTranslationError("improvement step failed", err).
			WithContext("source_lang", input.SourceLang).
			WithContext("target_lang", input.TargetLang).
			WithContext("suggestion_count", len(suggestions))
	}
	result.ImprovedTexts = improved
	result.APICallsUsed++ // +1 API call for improvement

	// Step 3: Validate format preservation after improvement
	for key, improvedText := range improved {
		sourceText := input.SourceTexts[key]
		if sourceText != "" {
			if err := r.formatProtector.Validate(sourceText, improvedText); err != nil {
				// Log warning but don't fail - format might be intentionally adjusted
				fmt.Printf("⚠️  Format validation warning for key '%s': %v\n", key, err)
			}
		}
	}

	return result, nil
}

// reflectStep performs the reflection step
// LLM evaluates translations across 4 dimensions: accuracy, fluency, style, terminology
func (r *ReflectionEngine) reflectStep(ctx context.Context, input ReflectionInput) (map[string]string, error) {
	// Build reflection prompt following Andrew Ng's approach
	prompt := r.buildReflectionPrompt(input)

	// Call AI provider
	req := &provider.CompletionRequest{
		Prompt:      prompt,
		Model:       r.provider.GetModelName(),
		Temperature: r.temperature,
		MaxTokens:   r.maxTokens,
		SystemMsg:   "You are an expert linguist and translator tasked with evaluating translation quality.",
	}

	resp, err := r.provider.Complete(ctx, req)
	if err != nil {
		return nil, domain.NewTranslationError("reflection API call failed", err).
			WithContext("source_lang", input.SourceLang).
			WithContext("target_lang", input.TargetLang)
	}

	// Parse suggestions from LLM response
	suggestions := r.parseReflectionSuggestions(resp.Content, input.TranslatedTexts)

	return suggestions, nil
}

// improveStep performs the improvement step
// LLM edits translations based on expert suggestions
func (r *ReflectionEngine) improveStep(
	ctx context.Context,
	input ReflectionInput,
	suggestions map[string]string,
) (map[string]string, error) {
	// Build improvement prompt following Andrew Ng's approach
	prompt := r.buildImprovementPrompt(input, suggestions)

	// Call AI provider
	req := &provider.CompletionRequest{
		Prompt:      prompt,
		Model:       r.provider.GetModelName(),
		Temperature: r.temperature,
		MaxTokens:   r.maxTokens,
		SystemMsg:   "You are an expert translator tasked with improving translations based on expert suggestions.",
	}

	resp, err := r.provider.Complete(ctx, req)
	if err != nil {
		return nil, domain.NewTranslationError("improvement API call failed", err).
			WithContext("source_lang", input.SourceLang).
			WithContext("target_lang", input.TargetLang)
	}

	// Parse improved translations from LLM response
	improved := r.parseImprovedTranslations(resp.Content, input.TranslatedTexts)

	return improved, nil
}

// buildReflectionPrompt builds the reflection prompt following Andrew Ng's approach
// Evaluates: (i) accuracy, (ii) fluency, (iii) style, (iv) terminology
func (r *ReflectionEngine) buildReflectionPrompt(input ReflectionInput) string {
	var sb strings.Builder

	// Task description
	sb.WriteString(fmt.Sprintf(
		"Your task is to carefully read source texts and translations from %s to %s, "+
			"and then give constructive criticism and helpful suggestions to improve the translations.\n\n",
		input.SourceLang, input.TargetLang,
	))

	// Add terminology requirements if provided
	if input.Terminology != nil && len(input.Terminology.PreserveTerms) > 0 {
		sb.WriteString("【Terminology Requirements】\n")
		sb.WriteString("The following terms must be preserved (kept in original form):\n")
		sb.WriteString(strings.Join(input.Terminology.PreserveTerms, ", "))
		sb.WriteString("\n\n")
	}

	if input.Terminology != nil && len(input.Terminology.ConsistentTerms) > 0 {
		sb.WriteString("Consistent terminology translations:\n")
		// Build terminology dictionary
		for lang, terms := range input.Terminology.ConsistentTerms {
			if lang == input.SourceLang {
				for _, term := range terms {
					if translation, ok := input.Terminology.GetTermTranslation(term, input.SourceLang); ok {
						sb.WriteString(fmt.Sprintf("- %s: %s\n", term, translation))
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	// Source texts section using XML-style tags
	sb.WriteString("<SOURCE_TEXTS>\n")
	for key, sourceText := range input.SourceTexts {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", key, sourceText))
	}
	sb.WriteString("</SOURCE_TEXTS>\n\n")

	// Translations section
	sb.WriteString("<TRANSLATIONS>\n")
	for key, translatedText := range input.TranslatedTexts {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", key, translatedText))
	}
	sb.WriteString("</TRANSLATIONS>\n\n")

	// Evaluation criteria (Andrew Ng's 4 dimensions)
	sb.WriteString("When writing suggestions, pay attention to whether there are ways to improve the translation's:\n")
	sb.WriteString(fmt.Sprintf(
		"(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),\n"+
			"(ii) fluency (by applying %s grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),\n"+
			"(iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),\n"+
			"(iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms in %s).\n\n",
		input.TargetLang, input.TargetLang,
	))

	// Output instructions
	sb.WriteString("Write a list of specific, helpful and constructive suggestions for improving the translation.\n")
	sb.WriteString("Each suggestion should address one specific part of the translation.\n\n")

	sb.WriteString("Output format:\n")
	sb.WriteString("[key] Suggestion text (or \"OK\" if no improvement needed)\n\n")

	sb.WriteString("Output only the suggestions and nothing else.")

	return sb.String()
}

// buildImprovementPrompt builds the improvement prompt following Andrew Ng's approach
func (r *ReflectionEngine) buildImprovementPrompt(input ReflectionInput, suggestions map[string]string) string {
	var sb strings.Builder

	// Task description
	sb.WriteString(fmt.Sprintf(
		"Your task is to carefully read, then edit, translations from %s to %s using expert suggestions and improve them.\n\n",
		input.SourceLang, input.TargetLang,
	))

	// Source texts section
	sb.WriteString("<SOURCE_TEXTS>\n")
	for key, sourceText := range input.SourceTexts {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", key, sourceText))
	}
	sb.WriteString("</SOURCE_TEXTS>\n\n")

	// Initial translations section
	sb.WriteString("<INITIAL_TRANSLATIONS>\n")
	for key, translatedText := range input.TranslatedTexts {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", key, translatedText))
	}
	sb.WriteString("</INITIAL_TRANSLATIONS>\n\n")

	// Expert suggestions section
	sb.WriteString("<EXPERT_SUGGESTIONS>\n")
	for key, suggestion := range suggestions {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", key, suggestion))
	}
	sb.WriteString("</EXPERT_SUGGESTIONS>\n\n")

	// Instructions
	sb.WriteString("Please take into account the expert suggestions when editing the translations. ")
	sb.WriteString("Edit the translations by ensuring:\n\n")
	sb.WriteString(fmt.Sprintf(
		"(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),\n"+
			"(ii) fluency (by applying %s grammar, spelling and punctuation rules and ensuring there are no unnecessary repetitions),\n"+
			"(iii) style (by ensuring the translations reflect the style of the source text),\n"+
			"(iv) terminology (inappropriate for context, inconsistent use), or\n"+
			"(v) other errors.\n\n",
		input.TargetLang,
	))

	// Format preservation reminder
	sb.WriteString("IMPORTANT: Preserve all format elements (placeholders like {variable}, HTML tags, special markers).\n\n")

	// Output format
	sb.WriteString("Output format:\n")
	sb.WriteString("[key] improved translation\n\n")

	sb.WriteString("Output only the improved translations and nothing else.")

	return sb.String()
}

// parseReflectionSuggestions parses suggestions from the reflection response
func (r *ReflectionEngine) parseReflectionSuggestions(response string, translations map[string]string) map[string]string {
	suggestions := make(map[string]string)

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Expected format: [key] suggestion text
		if strings.HasPrefix(line, "[") {
			closeBracket := strings.Index(line, "]")
			if closeBracket > 0 {
				key := line[1:closeBracket]
				suggestion := strings.TrimSpace(line[closeBracket+1:])

				// Only include if the key exists in translations
				if _, exists := translations[key]; exists && suggestion != "" {
					suggestions[key] = suggestion
				}
			}
		}
	}

	return suggestions
}

// parseImprovedTranslations parses improved translations from the improvement response
func (r *ReflectionEngine) parseImprovedTranslations(response string, originalTranslations map[string]string) map[string]string {
	improved := make(map[string]string)

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Expected format: [key] improved translation
		if strings.HasPrefix(line, "[") {
			closeBracket := strings.Index(line, "]")
			if closeBracket > 0 {
				key := line[1:closeBracket]
				translation := strings.TrimSpace(line[closeBracket+1:])

				// Only accept improvements for keys that exist in original
				if _, exists := originalTranslations[key]; exists && translation != "" {
					improved[key] = translation
				}
			}
		}
	}

	return improved
}

// ShouldReflect determines if reflection is needed for a batch
// In Agentic mode, we always reflect to allow LLM to discover quality issues
func (r *ReflectionEngine) ShouldReflect(translations map[string]string, terminology *domain.Terminology) bool {
	// Always reflect in Agentic mode (following Andrew Ng's approach)
	// LLM can discover subtle quality issues that rule-based checks miss
	return len(translations) > 0
}
