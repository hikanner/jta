package format

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// ElementType represents the type of format element
type ElementType string

const (
	ElementTypePlaceholder ElementType = "placeholder"
	ElementTypeHTML        ElementType = "html"
	ElementTypeURL         ElementType = "url"
	ElementTypeMarkdown    ElementType = "markdown"
)

// FormatElement represents a format element in text
type FormatElement struct {
	Type     ElementType
	Value    string
	Position int
}

// ValidationReport contains the result of format validation
type ValidationReport struct {
	IsValid         bool
	MissingElements []FormatElement
	ExtraElements   []FormatElement
	Errors          []string
}

// Protector handles format protection and validation
type Protector struct {
	placeholderPattern *regexp.Regexp
	htmlPattern        *regexp.Regexp
	urlPattern         *regexp.Regexp
	markdownPattern    *regexp.Regexp
}

// NewProtector creates a new format protector
func NewProtector() *Protector {
	return &Protector{
		// Matches {variable}, {{variable}}, %s, %d, etc.
		placeholderPattern: regexp.MustCompile(`\{[^}]+\}|\{\{[^}]+\}\}|%[sd]|%\([^)]+\)[sd]`),
		// Matches HTML tags
		htmlPattern: regexp.MustCompile(`<[^>]+>`),
		// Matches URLs
		urlPattern: regexp.MustCompile(`https?://[^\s]+`),
		// Matches markdown syntax
		markdownPattern: regexp.MustCompile(`\*\*[^*]+\*\*|\*[^*]+\*|__[^_]+__|_[^_]+_|\[[^\]]+\]\([^)]+\)`),
	}
}

// Extract extracts all format elements from text
func (p *Protector) Extract(text string) []FormatElement {
	var elements []FormatElement

	// Extract placeholders
	matches := p.placeholderPattern.FindAllStringIndex(text, -1)
	for _, match := range matches {
		elements = append(elements, FormatElement{
			Type:     ElementTypePlaceholder,
			Value:    text[match[0]:match[1]],
			Position: match[0],
		})
	}

	// Extract HTML tags
	matches = p.htmlPattern.FindAllStringIndex(text, -1)
	for _, match := range matches {
		elements = append(elements, FormatElement{
			Type:     ElementTypeHTML,
			Value:    text[match[0]:match[1]],
			Position: match[0],
		})
	}

	// Extract URLs
	matches = p.urlPattern.FindAllStringIndex(text, -1)
	for _, match := range matches {
		elements = append(elements, FormatElement{
			Type:     ElementTypeURL,
			Value:    text[match[0]:match[1]],
			Position: match[0],
		})
	}

	// Extract Markdown
	matches = p.markdownPattern.FindAllStringIndex(text, -1)
	for _, match := range matches {
		elements = append(elements, FormatElement{
			Type:     ElementTypeMarkdown,
			Value:    text[match[0]:match[1]],
			Position: match[0],
		})
	}

	return elements
}

// Validate validates that all format elements are preserved in translation
func (p *Protector) Validate(original, translated string) error {
	report := p.GetValidationReport(original, translated)
	if !report.IsValid {
		return fmt.Errorf("format validation failed: %s", strings.Join(report.Errors, "; "))
	}
	return nil
}

// GetValidationReport generates a detailed validation report
func (p *Protector) GetValidationReport(original, translated string) ValidationReport {
	originalElements := p.Extract(original)
	translatedElements := p.Extract(translated)

	report := ValidationReport{
		IsValid: true,
		Errors:  []string{},
	}

	// Group elements by type and value for comparison
	originalMap := groupElements(originalElements)
	translatedMap := groupElements(translatedElements)

	// Check for missing elements
	for key, originalCount := range originalMap {
		translatedCount := translatedMap[key]
		if translatedCount < originalCount {
			report.IsValid = false
			report.Errors = append(report.Errors,
				fmt.Sprintf("Missing %d occurrence(s) of %s", originalCount-translatedCount, key))
			// Add to missing elements
			for _, elem := range originalElements {
				elemKey := fmt.Sprintf("%s:%s", elem.Type, elem.Value)
				if elemKey == key {
					report.MissingElements = append(report.MissingElements, elem)
				}
			}
		}
	}

	// Check for extra elements (usually not a problem, but report anyway)
	for key, translatedCount := range translatedMap {
		originalCount := originalMap[key]
		if translatedCount > originalCount {
			// Add to extra elements
			for _, elem := range translatedElements {
				elemKey := fmt.Sprintf("%s:%s", elem.Type, elem.Value)
				if elemKey == key {
					report.ExtraElements = append(report.ExtraElements, elem)
				}
			}
		}
	}

	return report
}

// HasFormatElements checks if text contains any format elements
func (p *Protector) HasFormatElements(text string) bool {
	return len(p.Extract(text)) > 0
}

// groupElements groups elements by type and value, counting occurrences
func groupElements(elements []FormatElement) map[string]int {
	result := make(map[string]int)
	for _, elem := range elements {
		key := fmt.Sprintf("%s:%s", elem.Type, elem.Value)
		result[key]++
	}
	return result
}

// BuildFormatInstructions builds instructions for the LLM about format preservation
func (p *Protector) BuildFormatInstructions(text string) string {
	elements := p.Extract(text)
	if len(elements) == 0 {
		return ""
	}

	var instructions []string
	instructions = append(instructions, "🔒 FORMAT PRESERVATION CRITICAL:")

	// Group by type
	placeholders := []string{}
	htmlTags := []string{}
	urls := []string{}
	markdown := []string{}

	for _, elem := range elements {
		switch elem.Type {
		case ElementTypePlaceholder:
			if !contains(placeholders, elem.Value) {
				placeholders = append(placeholders, elem.Value)
			}
		case ElementTypeHTML:
			if !contains(htmlTags, elem.Value) {
				htmlTags = append(htmlTags, elem.Value)
			}
		case ElementTypeURL:
			if !contains(urls, elem.Value) {
				urls = append(urls, elem.Value)
			}
		case ElementTypeMarkdown:
			if !contains(markdown, elem.Value) {
				markdown = append(markdown, elem.Value)
			}
		}
	}

	if len(placeholders) > 0 {
		instructions = append(instructions, fmt.Sprintf("   • Placeholders (NEVER translate): %s", strings.Join(placeholders, ", ")))
	}
	if len(htmlTags) > 0 {
		instructions = append(instructions, fmt.Sprintf("   • HTML tags (keep intact): %s", strings.Join(htmlTags, ", ")))
	}
	if len(urls) > 0 {
		instructions = append(instructions, fmt.Sprintf("   • URLs (keep intact): %s", strings.Join(urls, ", ")))
	}
	if len(markdown) > 0 {
		instructions = append(instructions, fmt.Sprintf("   • Markdown (preserve syntax): %s", strings.Join(markdown, ", ")))
	}

	return strings.Join(instructions, "\n")
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
