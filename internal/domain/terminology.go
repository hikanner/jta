package domain

import "slices"

// TermType represents the type of terminology
type TermType string

const (
	// TermTypePreserve indicates terms that should never be translated
	TermTypePreserve TermType = "preserve"
	// TermTypeConsistent indicates terms that must be translated consistently
	TermTypeConsistent TermType = "consistent"
)

// Term represents a single terminology entry
type Term struct {
	Term    string   `json:"term"`
	Type    TermType `json:"type"`
	Context string   `json:"context,omitempty"` // Context provided by LLM
	Reason  string   `json:"reason,omitempty"`  // Why detected as term
}

// Terminology represents the terminology definition (source language only)
type Terminology struct {
	SourceLanguage  string   `json:"sourceLanguage"`
	PreserveTerms   []string `json:"preserveTerms"`
	ConsistentTerms []string `json:"consistentTerms"`
}

// TerminologyTranslation represents terminology translations for a specific target language
type TerminologyTranslation struct {
	SourceLanguage string            `json:"sourceLanguage"`
	TargetLanguage string            `json:"targetLanguage"`
	Translations   map[string]string `json:"translations"` // term -> translation
}

// GetTermTranslation returns the translation for a term
func (tt *TerminologyTranslation) GetTermTranslation(term string) (string, bool) {
	translation, ok := tt.Translations[term]
	return translation, ok
}

// GetMissingTranslations returns terms that don't have translation
func (t *Terminology) GetMissingTranslations(translation *TerminologyTranslation) []string {
	var missing []string

	for _, term := range t.ConsistentTerms {
		if translation == nil {
			// No translation file exists, all terms are missing
			missing = append(missing, term)
		} else if _, ok := translation.Translations[term]; !ok {
			// Translation file exists but this term is missing
			missing = append(missing, term)
		}
	}

	return missing
}

// AddPreserveTerm adds a term to preserve list
func (t *Terminology) AddPreserveTerm(term string) {
	// Check if already exists
	if slices.Contains(t.PreserveTerms, term) {
		return
	}
	t.PreserveTerms = append(t.PreserveTerms, term)
}

// AddConsistentTerm adds a consistent term
func (t *Terminology) AddConsistentTerm(term string) {
	// Check if already exists
	if slices.Contains(t.ConsistentTerms, term) {
		return
	}
	t.ConsistentTerms = append(t.ConsistentTerms, term)
}

// NewTerminologyTranslation creates a new terminology translation
func NewTerminologyTranslation(sourceLang, targetLang string) *TerminologyTranslation {
	return &TerminologyTranslation{
		SourceLanguage: sourceLang,
		TargetLanguage: targetLang,
		Translations:   make(map[string]string),
	}
}

// AddTranslation adds a term translation
func (tt *TerminologyTranslation) AddTranslation(term, translation string) {
	tt.Translations[term] = translation
}
