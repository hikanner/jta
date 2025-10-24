package translator

import (
	"context"
	"fmt"
	"strings"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/format"
	"github.com/hikanner/jta/internal/provider"
)

// ReflectionEngine implements lightweight reflection mechanism
// Inspired by Andrew Ng's Translation Agent but optimized for efficiency
// Target: 1.2-1.5x API calls vs basic translation
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
		maxTokens:       2000,
		temperature:     0.3, // Lower temperature for consistency checking
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
	IssuesFound      []QualityIssue
	ImprovedTexts    map[string]string // Only improved translations
	ReflectionNeeded bool
	APICallsUsed     int
}

// QualityIssue represents a translation quality issue
type QualityIssue struct {
	Key         string
	Type        IssueType
	Description string
	Severity    IssueSeverity
}

// IssueType represents the type of quality issue
type IssueType string

const (
	IssueTypeTerminology  IssueType = "terminology"  // Terminology inconsistency
	IssueTypeFormat       IssueType = "format"       // Format element missing/damaged
	IssueTypeContext      IssueType = "context"      // Context mismatch
	IssueTypeNaturalness  IssueType = "naturalness"  // Unnatural translation
	IssueTypeCompleteness IssueType = "completeness" // Incomplete translation
)

// IssueSeverity represents the severity of an issue
type IssueSeverity string

const (
	SeverityCritical IssueSeverity = "critical" // Must fix
	SeverityHigh     IssueSeverity = "high"     // Should fix
	SeverityMedium   IssueSeverity = "medium"   // Nice to fix
	SeverityLow      IssueSeverity = "low"      // Optional
)

// Reflect performs lightweight reflection on translations
func (r *ReflectionEngine) Reflect(ctx context.Context, input ReflectionInput) (*ReflectionResult, error) {
	result := &ReflectionResult{
		IssuesFound:   []QualityIssue{},
		ImprovedTexts: make(map[string]string),
	}

	// Step 1: Quick quality check (no API call)
	issues := r.quickQualityCheck(input)
	result.IssuesFound = issues

	// Step 2: Determine if reflection is needed
	criticalIssues := r.filterCriticalIssues(issues)
	if len(criticalIssues) == 0 {
		result.ReflectionNeeded = false
		return result, nil
	}

	result.ReflectionNeeded = true

	// Step 3: Batch reflection for critical issues only (1 API call for multiple issues)
	if len(criticalIssues) > 0 {
		improved, err := r.batchReflectAndImprove(ctx, input, criticalIssues)
		if err != nil {
			return nil, fmt.Errorf("batch reflection failed: %w", err)
		}
		result.ImprovedTexts = improved
		result.APICallsUsed = 1 // Only 1 additional API call for all improvements
	}

	return result, nil
}

// quickQualityCheck performs fast quality checks without API calls
func (r *ReflectionEngine) quickQualityCheck(input ReflectionInput) []QualityIssue {
	var issues []QualityIssue

	for key, translated := range input.TranslatedTexts {
		source := input.SourceTexts[key]

		// Check 1: Format element integrity
		if issue := r.checkFormatIntegrity(key, source, translated); issue != nil {
			issues = append(issues, *issue)
		}

		// Check 2: Terminology consistency (if terminology provided)
		if input.Terminology != nil {
			if issue := r.checkTerminologyConsistency(key, source, translated, input.Terminology); issue != nil {
				issues = append(issues, *issue)
			}
		}

		// Check 3: Basic completeness (not empty, not too short)
		if issue := r.checkCompleteness(key, source, translated); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

// checkFormatIntegrity checks if format elements are preserved
func (r *ReflectionEngine) checkFormatIntegrity(key, source, translated string) *QualityIssue {
	report := r.formatProtector.GetValidationReport(source, translated)
	if !report.IsValid {
		errorMsg := "Format validation failed"
		if len(report.Errors) > 0 {
			errorMsg = fmt.Sprintf("Format validation failed: %s", report.Errors[0])
		}
		return &QualityIssue{
			Key:         key,
			Type:        IssueTypeFormat,
			Description: errorMsg,
			Severity:    SeverityCritical,
		}
	}

	return nil
}

// checkTerminologyConsistency checks if terminology is applied correctly
func (r *ReflectionEngine) checkTerminologyConsistency(key, source, translated string, term *domain.Terminology) *QualityIssue {
	// Check preserve terms (should remain in original)
	for _, preserveTerm := range term.PreserveTerms {
		if strings.Contains(source, preserveTerm) && !strings.Contains(translated, preserveTerm) {
			return &QualityIssue{
				Key:         key,
				Type:        IssueTypeTerminology,
				Description: fmt.Sprintf("Preserve term '%s' missing in translation", preserveTerm),
				Severity:    SeverityCritical,
			}
		}
	}

	// Check consistent terms (should use standard translation)
	// Get source terms and check if they appear in the source text
	if sourceTerms, ok := term.ConsistentTerms[term.SourceLanguage]; ok {
		for _, sourceTerm := range sourceTerms {
			if strings.Contains(strings.ToLower(source), strings.ToLower(sourceTerm)) {
				// Get the expected translation
				translation, hasTranslation := term.GetTermTranslation(sourceTerm, term.SourceLanguage)
				if hasTranslation && translation != sourceTerm {
					// Check if translation is used
					if !strings.Contains(strings.ToLower(translated), strings.ToLower(translation)) {
						return &QualityIssue{
							Key:         key,
							Type:        IssueTypeTerminology,
							Description: fmt.Sprintf("Consistent term '%s' may not use standard translation", sourceTerm),
							Severity:    SeverityHigh,
						}
					}
				}
			}
		}
	}

	return nil
}

// checkCompleteness checks basic completeness
func (r *ReflectionEngine) checkCompleteness(key, source, translated string) *QualityIssue {
	if translated == "" {
		return &QualityIssue{
			Key:         key,
			Type:        IssueTypeCompleteness,
			Description: "Translation is empty",
			Severity:    SeverityCritical,
		}
	}

	// Check if translation is suspiciously short compared to source
	sourceLen := len([]rune(source))
	translatedLen := len([]rune(translated))

	// Allow some flexibility - translation can be 20% of original length (different languages have different lengths)
	if sourceLen > 10 && translatedLen < sourceLen/5 {
		return &QualityIssue{
			Key:         key,
			Type:        IssueTypeCompleteness,
			Description: "Translation appears too short",
			Severity:    SeverityHigh,
		}
	}

	return nil
}

// filterCriticalIssues filters issues that require immediate fixing
func (r *ReflectionEngine) filterCriticalIssues(issues []QualityIssue) []QualityIssue {
	var critical []QualityIssue
	for _, issue := range issues {
		if issue.Severity == SeverityCritical || issue.Severity == SeverityHigh {
			critical = append(critical, issue)
		}
	}
	return critical
}

// batchReflectAndImprove performs batch reflection and improvement
func (r *ReflectionEngine) batchReflectAndImprove(
	ctx context.Context,
	input ReflectionInput,
	issues []QualityIssue,
) (map[string]string, error) {
	// Group issues by key
	issuesByKey := make(map[string][]QualityIssue)
	for _, issue := range issues {
		issuesByKey[issue.Key] = append(issuesByKey[issue.Key], issue)
	}

	// Build reflection prompt for all issues
	prompt := r.buildReflectionPrompt(input, issuesByKey)

	// Call AI provider once for all improvements
	req := &provider.CompletionRequest{
		Prompt:      prompt,
		Model:       r.provider.GetModelName(),
		Temperature: r.temperature,
		MaxTokens:   r.maxTokens,
		SystemMsg:   "You are a professional translator tasked with improving translation quality.",
	}

	resp, err := r.provider.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("reflection API call failed: %w", err)
	}

	// Parse improved translations from response
	improved := r.parseImprovedTranslations(resp.Content, issuesByKey)

	return improved, nil
}

// buildReflectionPrompt builds a prompt for batch reflection
func (r *ReflectionEngine) buildReflectionPrompt(
	input ReflectionInput,
	issuesByKey map[string][]QualityIssue,
) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Translation Quality Review\n\n"))
	sb.WriteString(fmt.Sprintf("Source Language: %s\n", input.SourceLang))
	sb.WriteString(fmt.Sprintf("Target Language: %s\n\n", input.TargetLang))

	// Add terminology context if available
	if input.Terminology != nil && len(input.Terminology.PreserveTerms) > 0 {
		sb.WriteString("## Terminology Requirements\n")
		sb.WriteString("Preserve terms (keep original): " + strings.Join(input.Terminology.PreserveTerms, ", ") + "\n\n")
	}

	sb.WriteString("## Translations with Issues\n\n")
	sb.WriteString("Please review and improve the following translations:\n\n")

	for key, issues := range issuesByKey {
		source := input.SourceTexts[key]
		translated := input.TranslatedTexts[key]

		sb.WriteString(fmt.Sprintf("### Key: %s\n", key))
		sb.WriteString(fmt.Sprintf("Source: %s\n", source))
		sb.WriteString(fmt.Sprintf("Current Translation: %s\n", translated))
		sb.WriteString("Issues:\n")
		for _, issue := range issues {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", issue.Severity, issue.Type, issue.Description))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n## Instructions\n")
	sb.WriteString("For each translation with issues, provide an improved version.\n")
	sb.WriteString("Format: KEY: improved translation\n")
	sb.WriteString("Only output improved translations, one per line.\n")

	return sb.String()
}

// parseImprovedTranslations parses improved translations from AI response
func (r *ReflectionEngine) parseImprovedTranslations(
	response string,
	issuesByKey map[string][]QualityIssue,
) map[string]string {
	improved := make(map[string]string)

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Expected format: "KEY: improved translation"
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			translation := strings.TrimSpace(parts[1])

			// Only accept improvements for keys with issues
			if _, hasIssues := issuesByKey[key]; hasIssues && translation != "" {
				improved[key] = translation
			}
		}
	}

	return improved
}

// ShouldReflect determines if reflection is needed for a batch
func (r *ReflectionEngine) ShouldReflect(translations map[string]string, terminology *domain.Terminology) bool {
	// Always reflect if we have terminology to enforce
	if terminology != nil && (len(terminology.PreserveTerms) > 0 || len(terminology.ConsistentTerms) > 0) {
		return true
	}

	// Reflection is always beneficial but we can skip for very small batches without terminology
	if len(translations) < 3 {
		return false
	}

	return true
}
