package domain

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

// Terminology represents the complete terminology configuration
type Terminology struct {
	SourceLanguage  string              `json:"sourceLanguage"`
	PreserveTerms   []string            `json:"preserveTerms"`
	ConsistentTerms map[string][]string `json:"consistentTerms"` // lang code -> terms
}

// GetTermTranslation returns the translation for a term in target language
func (t *Terminology) GetTermTranslation(term string, targetLang string) (string, bool) {
	// Check if it's a preserve term
	for _, preserveTerm := range t.PreserveTerms {
		if preserveTerm == term {
			return term, true // preserve terms stay the same
		}
	}

	// Check consistent terms
	sourceTerms := t.ConsistentTerms[t.SourceLanguage]
	targetTerms := t.ConsistentTerms[targetLang]

	// Find index in source language
	for i, sourceTerm := range sourceTerms {
		if sourceTerm == term {
			if i < len(targetTerms) {
				return targetTerms[i], true
			}
			break
		}
	}

	return "", false
}

// GetMissingTranslations returns terms that don't have translation in target language
func (t *Terminology) GetMissingTranslations(targetLang string) []string {
	sourceTerms := t.ConsistentTerms[t.SourceLanguage]
	targetTerms := t.ConsistentTerms[targetLang]

	var missing []string
	for i, sourceTerm := range sourceTerms {
		// If target doesn't have this index, or the translation is empty
		if i >= len(targetTerms) || targetTerms[i] == "" {
			missing = append(missing, sourceTerm)
		}
	}

	return missing
}

// AddPreserveTerm adds a term to preserve list
func (t *Terminology) AddPreserveTerm(term string) {
	// Check if already exists
	for _, existing := range t.PreserveTerms {
		if existing == term {
			return
		}
	}
	t.PreserveTerms = append(t.PreserveTerms, term)
}

// AddConsistentTerm adds a consistent term for a language
func (t *Terminology) AddConsistentTerm(lang string, term string) {
	if t.ConsistentTerms == nil {
		t.ConsistentTerms = make(map[string][]string)
	}

	// Check if already exists
	terms := t.ConsistentTerms[lang]
	for _, existing := range terms {
		if existing == term {
			return
		}
	}

	t.ConsistentTerms[lang] = append(terms, term)
}
