package terminology

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/provider"
)

const (
	// MAX_CONTEXT_TOKENS is the maximum token count for terminology detection
	// Conservative estimate suitable for all mainstream models (GPT-3.5+ supports 16K+)
	MAX_CONTEXT_TOKENS = 10000

	// CONTEXT_USAGE_RATIO is the actual usage ratio (reserve space for prompt and output)
	CONTEXT_USAGE_RATIO = 0.7
)

// Detector handles terminology detection using LLM
type Detector struct {
	provider  provider.AIProvider
	maxTokens int
}

// NewDetector creates a new detector
func NewDetector(provider provider.AIProvider) *Detector {
	return &Detector{
		provider:  provider,
		maxTokens: MAX_CONTEXT_TOKENS,
	}
}

// DetectTerms detects terminology from texts using LLM
func (d *Detector) DetectTerms(ctx context.Context, texts []string, sourceLang string) ([]domain.Term, error) {
	// Estimate token count
	estimatedTokens := d.estimateTokens(texts)

	// Calculate usable tokens (70% for text, 30% for prompt and output)
	maxUsableTokens := int(float64(d.maxTokens) * CONTEXT_USAGE_RATIO)

	// Choose strategy based on file size
	if estimatedTokens <= maxUsableTokens {
		// Strategy A: Small file - Full LLM analysis
		return d.analyzeWithLLM(ctx, texts, sourceLang)
	}

	// Strategy B: Large file - Hybrid approach (statistical + LLM validation)
	// Only used in rare scenarios
	return d.hybridDetection(ctx, texts, sourceLang)
}

// estimateTokens estimates the token count for texts
func (d *Detector) estimateTokens(texts []string) int {
	totalChars := 0
	for _, text := range texts {
		totalChars += len(text)
	}
	// Rough estimation: English averages 4 chars per token
	return totalChars / 4
}

// analyzeWithLLM performs full LLM analysis (for small files)
func (d *Detector) analyzeWithLLM(ctx context.Context, texts []string, lang string) ([]domain.Term, error) {
	// Build complete document
	doc := d.buildFullDocument(texts)

	// Build detection prompt
	prompt := d.buildDetectionPrompt(doc, lang, len(texts))

	// Call LLM (only once)
	resp, err := d.provider.Complete(ctx, &provider.CompletionRequest{
		Prompt:      prompt,
		Temperature: 0.3,
		MaxTokens:   2000,
	})

	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Parse result
	return d.parseTermsFromJSON(resp.Content)
}

// buildFullDocument builds a complete document from texts
func (d *Detector) buildFullDocument(texts []string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Total texts: %d\n\n", len(texts)))

	for i, text := range texts {
		builder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, text))
	}

	return builder.String()
}

// buildDetectionPrompt builds the prompt for term detection
func (d *Detector) buildDetectionPrompt(doc string, lang string, totalCount int) string {
	return fmt.Sprintf(`You are an expert terminology analyst for JSON internationalization files.

Your task: Analyze this COMPLETE %s JSON i18n file (containing %d texts) and identify terms that need special handling for translation consistency.

<DOCUMENT>
%s
</DOCUMENT>

Analysis Instructions:
1. Read through the ENTIRE document carefully
2. Notice which terms appear MULTIPLE TIMES in different contexts
3. Consider term importance based on:
   - Frequency of occurrence
   - Context (technical, business, branding)
   - Impact on translation consistency

Identify TWO types of terms:

A. PRESERVE (never translate):
   - Brand names (e.g., "MyApp", "OpenAI")
   - Technical terms (e.g., "API", "OAuth", "JSON")
   - Product names with versions (e.g., "FLUX.1", "GPT-4")
   - Proper nouns

B. CONSISTENT (must translate uniformly):
   - Business domain terms appearing multiple times
   - Core concepts specific to this application
   - Terms where inconsistent translation would confuse users

Response Format (JSON only, no explanation):
{
  "preserveTerms": [
    {
      "term": "API",
      "reason": "Technical acronym",
      "frequency": 15,
      "examples": ["API key", "API access", "API documentation"]
    }
  ],
  "consistentTerms": [
    {
      "term": "credits",
      "reason": "Core business concept",
      "frequency": 23,
      "examples": ["You have 10 credits", "Buy credits", "Unlimited credits"]
    }
  ]
}

Important:
- Only include terms that appear in the document
- Provide accurate frequency counts
- Include 2-3 example usages for each term
- Focus on quality over quantity (typically 5-15 terms total)`, lang, totalCount, doc)
}

// parseTermsFromJSON parses terms from LLM JSON response
func (d *Detector) parseTermsFromJSON(content string) ([]domain.Term, error) {
	// Extract JSON from potential markdown code blocks
	jsonStr := extractJSON(content)

	// Parse JSON
	var result struct {
		PreserveTerms []struct {
			Term      string   `json:"term"`
			Reason    string   `json:"reason"`
			Frequency int      `json:"frequency"`
			Examples  []string `json:"examples"`
		} `json:"preserveTerms"`
		ConsistentTerms []struct {
			Term      string   `json:"term"`
			Reason    string   `json:"reason"`
			Frequency int      `json:"frequency"`
			Examples  []string `json:"examples"`
		} `json:"consistentTerms"`
	}

	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var terms []domain.Term

	// Add preserve terms
	for _, t := range result.PreserveTerms {
		terms = append(terms, domain.Term{
			Term:    t.Term,
			Type:    domain.TermTypePreserve,
			Reason:  t.Reason,
			Context: fmt.Sprintf("Frequency: %d, Examples: %s", t.Frequency, strings.Join(t.Examples, "; ")),
		})
	}

	// Add consistent terms
	for _, t := range result.ConsistentTerms {
		terms = append(terms, domain.Term{
			Term:    t.Term,
			Type:    domain.TermTypeConsistent,
			Reason:  t.Reason,
			Context: fmt.Sprintf("Frequency: %d, Examples: %s", t.Frequency, strings.Join(t.Examples, "; ")),
		})
	}

	return terms, nil
}

// hybridDetection performs hybrid detection for large files
func (d *Detector) hybridDetection(ctx context.Context, texts []string, lang string) ([]domain.Term, error) {
	// For now, use simplified approach - just analyze first N texts
	// In production, this would use statistical pre-processing
	maxTexts := 100
	if len(texts) > maxTexts {
		texts = texts[:maxTexts]
	}

	return d.analyzeWithLLM(ctx, texts, lang)
}

// extractJSON extracts JSON from markdown code blocks or raw text
func extractJSON(content string) string {
	// Try to extract ```json ... ``` blocks
	start := strings.Index(content, "```json")
	if start != -1 {
		start += 7 // len("```json")
		end := strings.Index(content[start:], "```")
		if end != -1 {
			return strings.TrimSpace(content[start : start+end])
		}
	}

	// Try to extract ``` ... ``` blocks
	start = strings.Index(content, "```")
	if start != -1 {
		start += 3
		end := strings.Index(content[start:], "```")
		if end != -1 {
			return strings.TrimSpace(content[start : start+end])
		}
	}

	// Try to find JSON array or object start/end
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "[") || strings.HasPrefix(content, "{") {
		return content
	}

	// Find first { or [
	for i, c := range content {
		if c == '{' || c == '[' {
			return strings.TrimSpace(content[i:])
		}
	}

	return content
}

// parseTermTranslations parses term translations from LLM response
func parseTermTranslations(content string, terms []string) (map[string]string, error) {
	jsonStr := extractJSON(content)

	var translations map[string]string
	err := json.Unmarshal([]byte(jsonStr), &translations)
	if err != nil {
		return nil, fmt.Errorf("failed to parse translations: %w", err)
	}

	return translations, nil
}
