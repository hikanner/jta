package rtl

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/hikanner/jta/internal/domain"
)

// Processor handles RTL (Right-to-Left) text processing
type Processor struct {
	// Unicode directional marks
	lrmChar string // Left-to-Right Mark (U+200E)
	rlmChar string // Right-to-Left Mark (U+200F)

	// Patterns for detecting LTR text in RTL context
	urlPattern    *regexp.Regexp
	emailPattern  *regexp.Regexp
	numberPattern *regexp.Regexp
	latinPattern  *regexp.Regexp
}

// NewProcessor creates a new RTL processor
func NewProcessor() *Processor {
	return &Processor{
		lrmChar:       "\u200E", // Left-to-Right Mark
		rlmChar:       "\u200F", // Right-to-Left Mark
		urlPattern:    regexp.MustCompile(`https?://[^\s]+`),
		emailPattern:  regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		numberPattern: regexp.MustCompile(`\d+`),
		latinPattern:  regexp.MustCompile(`[a-zA-Z]+`),
	}
}

// ProcessText processes text for RTL languages
func (p *Processor) ProcessText(text string, targetLang string) string {
	lang, exists := domain.GetLanguage(targetLang)
	if !exists || !lang.IsRTL {
		return text
	}

	// Apply directional marks to LTR content
	processed := p.addDirectionalMarks(text)

	// Convert punctuation if language has special punctuation
	if len(lang.Punctuation) > 0 {
		processed = p.convertPunctuation(processed, lang.Punctuation)
	}

	return processed
}

// addDirectionalMarks adds Unicode directional marks around LTR text in RTL context
func (p *Processor) addDirectionalMarks(text string) string {
	result := text

	// Create a map to track protected regions (URLs and emails)
	protectedRanges := make([]struct{ start, end int }, 0)

	// Find and protect URLs
	urlMatches := p.urlPattern.FindAllStringIndex(text, -1)
	for _, match := range urlMatches {
		protectedRanges = append(protectedRanges, struct{ start, end int }{match[0], match[1]})
	}

	// Find and protect emails
	emailMatches := p.emailPattern.FindAllStringIndex(text, -1)
	for _, match := range emailMatches {
		protectedRanges = append(protectedRanges, struct{ start, end int }{match[0], match[1]})
	}

	// Wrap URLs with LRM marks
	result = p.urlPattern.ReplaceAllStringFunc(result, func(match string) string {
		return p.lrmChar + match + p.lrmChar
	})

	// Wrap email addresses with LRM marks
	result = p.emailPattern.ReplaceAllStringFunc(result, func(match string) string {
		return p.lrmChar + match + p.lrmChar
	})

	return result
}

// isStandaloneLatinWord checks if a Latin word is standalone
func (p *Processor) isStandaloneLatinWord(text, word string) bool {
	// Simple heuristic: check if word is surrounded by spaces or punctuation
	idx := strings.Index(text, word)
	if idx == -1 {
		return false
	}

	// Check character before
	if idx > 0 {
		before := rune(text[idx-1])
		if !unicode.IsSpace(before) && !unicode.IsPunct(before) {
			return false
		}
	}

	// Check character after
	endIdx := idx + len(word)
	if endIdx < len(text) {
		after := rune(text[endIdx])
		if !unicode.IsSpace(after) && !unicode.IsPunct(after) {
			return false
		}
	}

	return true
}

// convertPunctuation converts LTR punctuation to RTL equivalents
func (p *Processor) convertPunctuation(text string, punctuationMap map[string]string) string {
	result := text
	for ltr, rtl := range punctuationMap {
		result = strings.ReplaceAll(result, ltr, rtl)
	}
	return result
}

// ProcessBatch processes multiple texts for RTL languages
func (p *Processor) ProcessBatch(texts map[string]string, targetLang string) map[string]string {
	lang, exists := domain.GetLanguage(targetLang)
	if !exists || !lang.IsRTL {
		return texts
	}

	result := make(map[string]string, len(texts))
	for key, text := range texts {
		result[key] = p.ProcessText(text, targetLang)
	}

	return result
}

// NeedProcessing checks if a language needs RTL processing
func (p *Processor) NeedProcessing(targetLang string) bool {
	return domain.IsRTLLanguage(targetLang)
}

// AddLRM adds Left-to-Right Mark around text
func (p *Processor) AddLRM(text string) string {
	return p.lrmChar + text + p.lrmChar
}

// AddRLM adds Right-to-Left Mark around text
func (p *Processor) AddRLM(text string) string {
	return p.rlmChar + text + p.rlmChar
}

// StripDirectionalMarks removes all Unicode directional marks from text
func (p *Processor) StripDirectionalMarks(text string) string {
	text = strings.ReplaceAll(text, p.lrmChar, "")
	text = strings.ReplaceAll(text, p.rlmChar, "")
	return text
}
