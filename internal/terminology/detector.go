package terminology

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/provider"
)

const (
	// FULL_ANALYSIS_THRESHOLD is the token threshold for choosing detection strategy
	// Below this threshold: use full LLM analysis (faster, more accurate)
	// Above this threshold: use hybrid approach (statistical + LLM validation)
	FULL_ANALYSIS_THRESHOLD = 20000
)

// Detector handles terminology detection using LLM
type Detector struct {
	provider provider.AIProvider
}

// NewDetector creates a new detector
func NewDetector(provider provider.AIProvider) *Detector {
	return &Detector{
		provider: provider,
	}
}

// DetectTerms detects terminology from texts using LLM
func (d *Detector) DetectTerms(ctx context.Context, texts []string, sourceLang string) ([]domain.Term, error) {
	fmt.Printf("   🔍 DetectTerms called with %d texts, language: %s\n", len(texts), sourceLang)

	// Estimate token count
	estimatedTokens := d.estimateTokens(texts)
	fmt.Printf("   📊 Estimated tokens: %d (threshold: %d)\n", estimatedTokens, FULL_ANALYSIS_THRESHOLD)

	// Choose strategy based on file size
	if estimatedTokens <= FULL_ANALYSIS_THRESHOLD {
		// Strategy A: Small file - Full LLM analysis
		fmt.Printf("   ✅ Using full LLM analysis strategy\n")
		return d.analyzeWithLLM(ctx, texts, sourceLang)
	}

	// Strategy B: Large file - Hybrid approach (statistical + LLM validation)
	fmt.Printf("   ⚠️  File too large, using hybrid detection strategy\n")
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

	estimatedTokens := d.estimateTokens(texts)
	fmt.Printf("   📝 Analyzing %d texts with LLM (estimated ~%d tokens)...\n", len(texts), estimatedTokens)
	fmt.Printf("   ⏳ This may take 30-60 seconds for large files, please wait...\n")

	// Create independent 5-minute timeout for this LLM call
	callCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Call LLM (only once)
	resp, err := d.provider.Complete(callCtx, &provider.CompletionRequest{
		Prompt: prompt,
	})

	if err != nil {
		fmt.Printf("   ❌ LLM call failed: %v\n", err)
		return nil, domain.NewTerminologyError("LLM analysis failed", err).
			WithContext("language", lang).
			WithContext("text_count", len(texts))
	}

	fmt.Printf("   ✅ LLM response received (%d tokens)\n", resp.Usage.TotalTokens)

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
		return nil, domain.NewFormatError("failed to parse JSON", err).
			WithContext("json_content", jsonStr)
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

// CandidateWord represents a candidate term from statistical analysis
type CandidateWord struct {
	Word      string
	Frequency int
	Contexts  []string // Max 5 contexts
}

// hybridDetection performs hybrid detection for large files
// Step 1: Local statistical analysis (no LLM)
// Step 2: LLM batch validation
func (d *Detector) hybridDetection(ctx context.Context, texts []string, lang string) ([]domain.Term, error) {
	fmt.Printf("   🔄 Step 1: Extracting candidate terms using statistical analysis...\n")

	// Step 1: Extract candidate terms using local statistical analysis
	candidates := d.extractCandidatesSimplified(texts)

	fmt.Printf("   ✅ Found %d candidate terms\n", len(candidates))

	if len(candidates) == 0 {
		fmt.Printf("   ℹ️  No candidates found, returning empty list\n")
		return []domain.Term{}, nil
	}

	// Step 2: Validate candidates with LLM
	fmt.Printf("   🔄 Step 2: Validating candidates with LLM...\n")
	return d.validateWithLLM(ctx, candidates, lang)
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
		return nil, domain.NewFormatError("failed to parse translations", err).
			WithContext("json_content", jsonStr)
	}

	return translations, nil
}

// extractCandidatesSimplified extracts candidate terms using local statistical analysis
func (d *Detector) extractCandidatesSimplified(texts []string) map[string]*CandidateWord {
	candidates := make(map[string]*CandidateWord)

	for _, text := range texts {
		// Simple tokenization
		words := d.simpleTokenize(text)

		// Extract 1-3 word phrases
		for i := range words {
			// Single word
			d.addCandidate(candidates, words[i], text)

			// Bigram
			if i+1 < len(words) {
				phrase := words[i] + " " + words[i+1]
				d.addCandidate(candidates, phrase, text)
			}

			// Trigram
			if i+2 < len(words) {
				phrase := words[i] + " " + words[i+1] + " " + words[i+2]
				d.addCandidate(candidates, phrase, text)
			}
		}
	}

	return d.filterCandidates(candidates)
}

// addCandidate adds a word to candidates with frequency tracking
func (d *Detector) addCandidate(candidates map[string]*CandidateWord, word string, context string) {
	// Skip too short or too long
	if len(word) < 2 || len(word) > 50 {
		return
	}

	// Skip stop words
	if d.isStopWord(word) {
		return
	}

	word = strings.TrimSpace(word)

	if cand, exists := candidates[word]; exists {
		cand.Frequency++
		// Keep max 5 contexts
		if len(cand.Contexts) < 5 {
			cand.Contexts = append(cand.Contexts, context)
		}
	} else {
		candidates[word] = &CandidateWord{
			Word:      word,
			Frequency: 1,
			Contexts:  []string{context},
		}
	}
}

// filterCandidates filters candidates based on frequency and format
func (d *Detector) filterCandidates(candidates map[string]*CandidateWord) map[string]*CandidateWord {
	filtered := make(map[string]*CandidateWord)

	for word, info := range candidates {
		// Keep if: frequency >= 3 OR special format
		if info.Frequency >= 3 || d.isSpecialFormat(word) {
			filtered[word] = info
		}
	}

	return filtered
}

// simpleTokenize performs simple tokenization without external libraries
func (d *Detector) simpleTokenize(text string) []string {
	// Replace punctuation with spaces (keep hyphens and dots)
	text = strings.Map(func(r rune) rune {
		if r == '-' || r == '.' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			return ' '
		}
		return r
	}, text)

	words := strings.Fields(text)

	// Normalize case (preserve all-caps words like API)
	result := []string{}
	for _, word := range words {
		if word == strings.ToUpper(word) && len(word) >= 2 {
			result = append(result, word) // Keep all-caps
		} else {
			result = append(result, strings.ToLower(word))
		}
	}

	return result
}

// isStopWord checks if a word is a common stop word
func (d *Detector) isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"this": true, "that": true, "these": true, "those": true,
		"your": true, "you": true, "it": true, "its": true,
	}

	return stopWords[strings.ToLower(word)]
}

// isSpecialFormat checks if a word has special formatting (all-caps, version numbers, etc)
func (d *Detector) isSpecialFormat(word string) bool {
	// All-caps (e.g., API, JSON)
	if len(word) >= 2 && word == strings.ToUpper(word) && !strings.ContainsAny(word, " ") {
		return true
	}

	// Contains version numbers (e.g., FLUX.1, GPT-4)
	if strings.Contains(word, ".") || strings.ContainsAny(word, "0123456789") {
		return true
	}

	// CamelCase (e.g., MyApp, OpenAI)
	if len(word) > 1 && unicode.IsUpper(rune(word[0])) {
		for i := 1; i < len(word); i++ {
			if unicode.IsUpper(rune(word[i])) {
				return true
			}
		}
	}

	return false
}

// validateWithLLM validates candidates with LLM in batches
func (d *Detector) validateWithLLM(ctx context.Context, candidates map[string]*CandidateWord, lang string) ([]domain.Term, error) {
	batches := d.batchCandidates(candidates, 500)

	fmt.Printf("   📦 Split into %d batches (500 candidates per batch)\n", len(batches))

	allTerms := []domain.Term{}

	for i, batch := range batches {
		fmt.Printf("   🔄 Validating batch %d/%d...\n", i+1, len(batches))

		terms, err := d.validateBatchWithLLM(ctx, batch, lang)
		if err != nil {
			// Retry once for 504 errors
			if strings.Contains(err.Error(), "DEADLINE_EXCEEDED") || strings.Contains(err.Error(), "504") {
				fmt.Printf("   ⚠️  Batch %d failed with timeout, retrying once...\n", i+1)
				time.Sleep(10 * time.Second) // Brief pause before retry
				terms, err = d.validateBatchWithLLM(ctx, batch, lang)
			}

			if err != nil {
				fmt.Printf("   ❌ Batch %d failed: %v\n", i+1, err)
				return nil, domain.NewTerminologyError(fmt.Sprintf("batch %d validation failed", i+1), err)
			}
		}

		fmt.Printf("   ✅ Batch %d completed, found %d terms\n", i+1, len(terms))
		allTerms = append(allTerms, terms...)
	}

	fmt.Printf("   🎉 All batches completed, total %d terms found\n", len(allTerms))
	return allTerms, nil
}

// validateBatchWithLLM validates a single batch of candidates
func (d *Detector) validateBatchWithLLM(ctx context.Context, batch []*CandidateWord, lang string) ([]domain.Term, error) {
	prompt := d.buildValidationPrompt(batch, lang)

	// Calculate prompt size
	promptTokens := len(prompt) / 4 // rough estimate
	fmt.Printf("      📊 Batch prompt size: %d chars (~%d tokens)\n", len(prompt), promptTokens)
	fmt.Printf("      ⏳ Calling LLM API...\n")

	// Create independent 5-minute timeout for this LLM call
	batchCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	startTime := time.Now()

	resp, err := d.provider.Complete(batchCtx, &provider.CompletionRequest{
		Prompt: prompt,
	})

	elapsed := time.Since(startTime)

	if err != nil {
		fmt.Printf("      ❌ LLM API failed after %.1fs: %v\n", elapsed.Seconds(), err)
		return nil, err
	}

	fmt.Printf("      ✅ LLM API succeeded in %.1fs (%d tokens)\n", elapsed.Seconds(), resp.Usage.TotalTokens)

	return d.parseValidationResult(resp.Content)
}

// buildValidationPrompt builds the LLM prompt for candidate validation
func (d *Detector) buildValidationPrompt(candidates []*CandidateWord, lang string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`You are a terminology validation expert for JSON i18n files.

I have extracted candidate terms from a large %s JSON file using statistical analysis.
Your task: Verify which candidates are TRUE TERMS that need special handling for translation.

TRUE TERMS are:
1. PRESERVE (never translate): brand names, technical terms, product names, proper nouns
2. CONSISTENT (must translate uniformly): business domain terms, core concepts

NOT TERMS (ignore these):
- Common words that don't need special handling
- Generic phrases
- Complete sentences

Below are the candidates with their frequency and example contexts:

`, lang))

	for i, cand := range candidates {
		builder.WriteString(fmt.Sprintf("\n%d. Candidate: \"%s\"\n", i+1, cand.Word))
		builder.WriteString(fmt.Sprintf("   Frequency: %d times in file\n", cand.Frequency))
		builder.WriteString("   Example contexts:\n")
		for j, ctx := range cand.Contexts {
			if j >= 3 {
				break
			}
			builder.WriteString(fmt.Sprintf("   - \"%s\"\n", ctx))
		}
	}

	builder.WriteString(`

Return JSON array with your decisions (ONLY include terms where is_term is true):
[
  {
    "term": "API",
    "is_term": true,
    "type": "preserve",
    "reason": "Technical acronym, appears in multiple technical contexts"
  },
  {
    "term": "user profile",
    "is_term": true,
    "type": "consistent",
    "reason": "Core UI feature name, appears frequently across different contexts"
  }
]`)

	return builder.String()
}

// parseValidationResult parses LLM validation results
func (d *Detector) parseValidationResult(content string) ([]domain.Term, error) {
	jsonStr := extractJSON(content)

	var results []struct {
		Term   string `json:"term"`
		IsTerm bool   `json:"is_term"`
		Type   string `json:"type"`
		Reason string `json:"reason"`
	}

	err := json.Unmarshal([]byte(jsonStr), &results)
	if err != nil {
		return nil, domain.NewFormatError("failed to parse validation result", err).
			WithContext("json_content", jsonStr)
	}

	terms := []domain.Term{}
	for _, r := range results {
		if !r.IsTerm {
			continue
		}

		var termType domain.TermType
		if r.Type == "preserve" {
			termType = domain.TermTypePreserve
		} else {
			termType = domain.TermTypeConsistent
		}

		terms = append(terms, domain.Term{
			Term:   r.Term,
			Type:   termType,
			Reason: r.Reason,
		})
	}

	return terms, nil
}

// batchCandidates splits candidates into batches
func (d *Detector) batchCandidates(candidates map[string]*CandidateWord, batchSize int) [][]*CandidateWord {
	batches := [][]*CandidateWord{}
	currentBatch := []*CandidateWord{}

	for _, cand := range candidates {
		currentBatch = append(currentBatch, cand)

		if len(currentBatch) >= batchSize {
			batches = append(batches, currentBatch)
			currentBatch = []*CandidateWord{}
		}
	}

	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	return batches
}
