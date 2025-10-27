package terminology

import (
	"context"
	"fmt"
	"time"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/provider"
)

// Manager handles terminology detection and management
type Manager struct {
	provider              provider.AIProvider
	detector              *Detector
	termRepository        *TermRepository
	translationRepository *TranslationRepository
}

// NewManager creates a new terminology manager
func NewManager(provider provider.AIProvider) *Manager {
	detector := NewDetector(provider)
	return &Manager{
		provider:              provider,
		detector:              detector,
		termRepository:        NewTermRepository(),
		translationRepository: NewTranslationRepository(),
	}
}

// DetectTerms detects terms from a list of texts using LLM
func (m *Manager) DetectTerms(ctx context.Context, texts []string, sourceLang string) ([]domain.Term, error) {
	return m.detector.DetectTerms(ctx, texts, sourceLang)
}

// LoadTerminology loads terminology from directory
func (m *Manager) LoadTerminology(terminologyDir string) (*domain.Terminology, error) {
	return m.termRepository.Load(terminologyDir)
}

// SaveTerminology saves terminology to directory
func (m *Manager) SaveTerminology(terminologyDir string, terminology *domain.Terminology) error {
	return m.termRepository.Save(terminologyDir, terminology)
}

// TerminologyExists checks if terminology file exists
func (m *Manager) TerminologyExists(terminologyDir string) bool {
	return m.termRepository.Exists(terminologyDir)
}

// LoadTerminologyTranslation loads terminology translation from directory
func (m *Manager) LoadTerminologyTranslation(terminologyDir string, targetLang string) (*domain.TerminologyTranslation, error) {
	return m.translationRepository.Load(terminologyDir, targetLang)
}

// SaveTerminologyTranslation saves terminology translation to directory
func (m *Manager) SaveTerminologyTranslation(terminologyDir string, translation *domain.TerminologyTranslation) error {
	return m.translationRepository.Save(terminologyDir, translation)
}

// TranslationExists checks if translation file exists
func (m *Manager) TranslationExists(terminologyDir string, targetLang string) bool {
	return m.translationRepository.Exists(terminologyDir, targetLang)
}

// TranslateTerms translates terms to target language
func (m *Manager) TranslateTerms(ctx context.Context, terms []string, sourceLang, targetLang string) (map[string]string, error) {
	if len(terms) == 0 {
		return map[string]string{}, nil
	}

	// Build prompt for term translation
	prompt := m.buildTermTranslationPrompt(terms, sourceLang, targetLang)

	fmt.Printf("   📝 Calling LLM to translate %d terms...\n", len(terms))

	// Create independent 5-minute timeout for this LLM call
	callCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Call LLM
	resp, err := m.provider.Complete(callCtx, &provider.CompletionRequest{
		Prompt: prompt,
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
func (m *Manager) BuildPromptDictionary(terminology *domain.Terminology, translation *domain.TerminologyTranslation) string {
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
	if translation != nil && len(terminology.ConsistentTerms) > 0 && len(translation.Translations) > 0 {
		lines = append(lines, "📝 REQUIRED TRANSLATIONS:")
		for _, sourceTerm := range terminology.ConsistentTerms {
			if targetTerm, ok := translation.Translations[sourceTerm]; ok {
				lines = append(lines, fmt.Sprintf("   \"%s\" → \"%s\"", sourceTerm, targetTerm))
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
