package terminology

import (
	"context"
	"fmt"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/provider"
)

// Manager handles terminology detection and management
type Manager struct {
	provider   provider.AIProvider
	detector   *Detector
	repository Repository
}

// NewManager creates a new terminology manager
func NewManager(provider provider.AIProvider, repository Repository) *Manager {
	detector := NewDetector(provider)
	return &Manager{
		provider:   provider,
		detector:   detector,
		repository: repository,
	}
}

// DetectTerms detects terms from a list of texts using LLM
func (m *Manager) DetectTerms(ctx context.Context, texts []string, sourceLang string) ([]domain.Term, error) {
	return m.detector.DetectTerms(ctx, texts, sourceLang)
}

// LoadTerminology loads terminology from a file
func (m *Manager) LoadTerminology(path string) (*domain.Terminology, error) {
	return m.repository.Load(path)
}

// SaveTerminology saves terminology to a file
func (m *Manager) SaveTerminology(path string, terminology *domain.Terminology) error {
	return m.repository.Save(path, terminology)
}

// TerminologyExists checks if terminology file exists
func (m *Manager) TerminologyExists(path string) bool {
	return m.repository.Exists(path)
}

// TranslateTerms translates terms to target language
func (m *Manager) TranslateTerms(ctx context.Context, terms []string, sourceLang, targetLang string) (map[string]string, error) {
	if len(terms) == 0 {
		return map[string]string{}, nil
	}

	// Build prompt for term translation
	prompt := m.buildTermTranslationPrompt(terms, sourceLang, targetLang)

	// Call LLM
	resp, err := m.provider.Complete(ctx, &provider.CompletionRequest{
		Prompt:      prompt,
		Temperature: 0.3,
		MaxTokens:   0, // Let SDK use model-specific defaults
	})

	if err != nil {
		return nil, domain.NewTerminologyError("failed to translate terms", err).
			WithContext("source_lang", sourceLang).
			WithContext("target_lang", targetLang).
			WithContext("term_count", len(terms))
	}

	// Parse response
	translations, err := parseTermTranslations(resp.Content, terms)
	if err != nil {
		return nil, domain.NewFormatError("failed to parse term translations", err).
			WithContext("source_lang", sourceLang).
			WithContext("target_lang", targetLang)
	}

	return translations, nil
}

// BuildPromptDictionary builds a terminology dictionary for use in translation prompts
func (m *Manager) BuildPromptDictionary(terminology *domain.Terminology, targetLang string) string {
	if terminology == nil {
		return ""
	}

	var lines []string

	// Preserve terms (highest priority)
	if len(terminology.PreserveTerms) > 0 {
		lines = append(lines, "⚠️  CRITICAL - NEVER TRANSLATE THESE TERMS:")
		for _, term := range terminology.PreserveTerms {
			lines = append(lines, fmt.Sprintf("   \"%s\" → NEVER TRANSLATE, KEEP EXACTLY AS IS", term))
		}
		lines = append(lines, "")
	}

	// Consistent terms
	sourceTerms := terminology.ConsistentTerms[terminology.SourceLanguage]
	targetTerms := terminology.ConsistentTerms[targetLang]

	if len(sourceTerms) > 0 && len(targetTerms) > 0 {
		lines = append(lines, "📝 REQUIRED TRANSLATIONS:")
		for i, sourceTerm := range sourceTerms {
			if i < len(targetTerms) {
				lines = append(lines, fmt.Sprintf("   \"%s\" → \"%s\"", sourceTerm, targetTerms[i]))
			}
		}
		lines = append(lines, "")
	}

	if len(lines) > 0 {
		lines = append(lines, "🎯 INSTRUCTION: Follow these translations EXACTLY.")
	}

	return joinLines(lines)
}

func (m *Manager) buildTermTranslationPrompt(terms []string, sourceLang, targetLang string) string {
	termList := ""
	for i, term := range terms {
		termList += fmt.Sprintf("%d. \"%s\"\n", i+1, term)
	}

	return fmt.Sprintf(`You are a professional terminology translator.

Translate the following %s terms to %s. These are domain-specific terms that need accurate translation.

Terms to translate:
%s

Return ONLY a JSON object mapping each term to its translation:
{
  "term1": "translation1",
  "term2": "translation2"
}

Important:
- Keep brand names and technical acronyms unchanged if they are typically not translated
- Ensure consistency in terminology
- Use the most appropriate translation for the context
`, sourceLang, targetLang, termList)
}

func joinLines(lines []string) string {
	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}
