package keyfilter

import (
	"fmt"
	"maps"
	"regexp"
	"strings"
)

// PatternType represents the type of pattern
type PatternType string

const (
	PatternTypeExact       PatternType = "exact"     // Exact match: "settings.title"
	PatternTypeSingleLevel PatternType = "single"    // Single level wildcard: "settings.*"
	PatternTypeRecursive   PatternType = "recursive" // Recursive wildcard: "settings.**"
	PatternTypeWildcard    PatternType = "wildcard"  // Any position wildcard: "*.title"
)

// KeyPattern represents a parsed key pattern
type KeyPattern struct {
	Pattern  string         // Original pattern string
	Type     PatternType    // Pattern type
	Parts    []string       // Pattern parts split by "."
	IsGlob   bool           // Whether contains wildcard
	compiled *regexp.Regexp // Compiled regex (cached)
}

// FilterResult contains the result of key filtering
type FilterResult struct {
	Included map[string]any // Keys that match include patterns
	Excluded map[string]any // Keys that match exclude patterns
	Stats    FilterStats
}

// FilterStats contains statistics about filtering
type FilterStats struct {
	TotalKeys    int
	IncludedKeys int
	ExcludedKeys int
}

// Filter handles key filtering with glob patterns
type Filter struct {
	matcher *Matcher
}

// NewFilter creates a new key filter
func NewFilter() *Filter {
	return &Filter{
		matcher: NewMatcher(),
	}
}

// ParsePatterns parses comma-separated pattern strings
func (f *Filter) ParsePatterns(patterns string) ([]*KeyPattern, error) {
	if patterns == "" {
		return nil, nil
	}

	patternStrs := strings.Split(patterns, ",")
	var parsed []*KeyPattern

	for _, p := range patternStrs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		pattern, err := f.parsePattern(p)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", p, err)
		}

		parsed = append(parsed, pattern)
	}

	return parsed, nil
}

// parsePattern parses a single pattern string
func (f *Filter) parsePattern(pattern string) (*KeyPattern, error) {
	kp := &KeyPattern{
		Pattern: pattern,
		Parts:   strings.Split(pattern, "."),
	}

	// Determine pattern type
	if strings.Contains(pattern, "**") {
		kp.Type = PatternTypeRecursive
		kp.IsGlob = true
	} else if strings.Contains(pattern, "*") {
		// Check if * is at the beginning (*.title) or in the middle
		if strings.HasPrefix(pattern, "*.") {
			kp.Type = PatternTypeWildcard
		} else {
			kp.Type = PatternTypeSingleLevel
		}
		kp.IsGlob = true
	} else {
		kp.Type = PatternTypeExact
		kp.IsGlob = false
	}

	// Compile regex for complex patterns
	if kp.IsGlob {
		regex, err := f.patternToRegex(pattern)
		if err != nil {
			return nil, err
		}
		kp.compiled = regex
	}

	return kp, nil
}

// patternToRegex converts a glob pattern to regex
func (f *Filter) patternToRegex(pattern string) (*regexp.Regexp, error) {
	// Escape special regex characters except * and .
	escaped := regexp.QuoteMeta(pattern)

	// Replace escaped \* with regex patterns
	// ** matches any number of segments
	escaped = strings.ReplaceAll(escaped, `\*\*`, `.*`)
	// * matches single segment (no dots)
	escaped = strings.ReplaceAll(escaped, `\*`, `[^.]+`)

	// Escape dots in the original pattern represent literal dots
	escaped = strings.ReplaceAll(escaped, `\.`, `\.`)

	// Anchor the pattern
	escaped = "^" + escaped + "$"

	return regexp.Compile(escaped)
}

// FilterKeys filters JSON keys based on include and exclude patterns
func (f *Filter) FilterKeys(
	data map[string]any,
	includePatterns []*KeyPattern,
	excludePatterns []*KeyPattern,
) (*FilterResult, error) {
	result := &FilterResult{
		Included: make(map[string]any),
		Excluded: make(map[string]any),
		Stats:    FilterStats{},
	}

	// Flatten the JSON to get all key paths
	flattened := f.flattenJSON(data, "")
	result.Stats.TotalKeys = len(flattened)

	// If no include patterns, include everything by default
	if len(includePatterns) == 0 {
		maps.Copy(result.Included, flattened)
	} else {
		// Apply include patterns
		for key, value := range flattened {
			if f.matchesAnyPattern(key, includePatterns) {
				result.Included[key] = value
			}
		}
	}

	// Apply exclude patterns (takes priority)
	if len(excludePatterns) > 0 {
		for key := range result.Included {
			if f.matchesAnyPattern(key, excludePatterns) {
				delete(result.Included, key)
				result.Excluded[key] = flattened[key]
			}
		}
	}

	result.Stats.IncludedKeys = len(result.Included)
	result.Stats.ExcludedKeys = result.Stats.TotalKeys - result.Stats.IncludedKeys

	return result, nil
}

// matchesAnyPattern checks if key matches any of the patterns
func (f *Filter) matchesAnyPattern(keyPath string, patterns []*KeyPattern) bool {
	for _, pattern := range patterns {
		if f.matcher.Match(keyPath, pattern) {
			return true
		}
	}
	return false
}

// MatchKey checks if a key path matches a pattern
func (f *Filter) MatchKey(keyPath string, pattern *KeyPattern) bool {
	return f.matcher.Match(keyPath, pattern)
}

// BuildKeyPath constructs a key path from parts
func (f *Filter) BuildKeyPath(parts []string) string {
	return strings.Join(parts, ".")
}

// flattenJSON flattens nested JSON to dot notation
func (f *Filter) flattenJSON(data any, prefix string) map[string]any {
	result := make(map[string]any)

	switch v := data.(type) {
	case map[string]any:
		for key, value := range v {
			keyPath := key
			if prefix != "" {
				keyPath = prefix + "." + key
			}

			// Recursively flatten
			subResult := f.flattenJSON(value, keyPath)
			maps.Copy(result, subResult)
		}

	case []any:
		for i, value := range v {
			keyPath := fmt.Sprintf("%s[%d]", prefix, i)
			subResult := f.flattenJSON(value, keyPath)
			maps.Copy(result, subResult)
		}

	default:
		// Leaf value
		if prefix != "" {
			result[prefix] = v
		}
	}

	return result
}

// RebuildJSON rebuilds JSON structure from flattened keys
func (f *Filter) RebuildJSON(flattened map[string]any) map[string]any {
	result := make(map[string]any)

	for keyPath, value := range flattened {
		f.setNestedValue(result, keyPath, value)
	}

	return result
}

// setNestedValue sets a value in nested map using dot notation
func (f *Filter) setNestedValue(data map[string]any, keyPath string, value any) {
	parts := strings.Split(keyPath, ".")

	current := data
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]

		if _, exists := current[part]; !exists {
			current[part] = make(map[string]any)
		}

		if next, ok := current[part].(map[string]any); ok {
			current = next
		} else {
			// Type conflict, skip
			return
		}
	}

	// Set the final value
	current[parts[len(parts)-1]] = value
}
