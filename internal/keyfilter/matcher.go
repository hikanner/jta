package keyfilter

import (
	"strings"
)

// Matcher handles pattern matching logic
type Matcher struct{}

// NewMatcher creates a new matcher
func NewMatcher() *Matcher {
	return &Matcher{}
}

// Match checks if a key path matches a pattern
func (m *Matcher) Match(keyPath string, pattern *KeyPattern) bool {
	switch pattern.Type {
	case PatternTypeExact:
		return m.matchExact(keyPath, pattern)
	case PatternTypeSingleLevel:
		return m.matchSingleLevel(keyPath, pattern)
	case PatternTypeRecursive:
		return m.matchRecursive(keyPath, pattern)
	case PatternTypeWildcard:
		return m.matchWildcard(keyPath, pattern)
	default:
		return false
	}
}

// matchExact performs exact string matching
func (m *Matcher) matchExact(keyPath string, pattern *KeyPattern) bool {
	return keyPath == pattern.Pattern
}

// matchSingleLevel matches single-level wildcard patterns
// Example: "settings.*" matches "settings.title" but not "settings.theme.dark"
func (m *Matcher) matchSingleLevel(keyPath string, pattern *KeyPattern) bool {
	if pattern.compiled != nil {
		return pattern.compiled.MatchString(keyPath)
	}

	// Fallback: manual matching
	keyParts := strings.Split(keyPath, ".")
	patternParts := pattern.Parts

	// Must have same number of parts
	if len(keyParts) != len(patternParts) {
		return false
	}

	// Match each part
	for i := range keyParts {
		if patternParts[i] == "*" {
			continue // Wildcard matches anything
		}
		if keyParts[i] != patternParts[i] {
			return false
		}
	}

	return true
}

// matchRecursive matches recursive wildcard patterns
// Example: "settings.**" matches "settings.title", "settings.theme.dark", etc.
func (m *Matcher) matchRecursive(keyPath string, pattern *KeyPattern) bool {
	if pattern.compiled != nil {
		return pattern.compiled.MatchString(keyPath)
	}

	// Fallback: check if keyPath starts with the prefix before **
	prefix := strings.TrimSuffix(pattern.Pattern, ".**")
	prefix = strings.TrimSuffix(prefix, "**")

	if prefix == "" {
		return true // Matches everything
	}

	return strings.HasPrefix(keyPath, prefix)
}

// matchWildcard matches wildcard at any position
// Example: "*.title" matches "settings.title", "user.title", etc.
func (m *Matcher) matchWildcard(keyPath string, pattern *KeyPattern) bool {
	if pattern.compiled != nil {
		return pattern.compiled.MatchString(keyPath)
	}

	// Fallback: manual matching
	keyParts := strings.Split(keyPath, ".")
	patternParts := pattern.Parts

	// Must have same number of parts
	if len(keyParts) != len(patternParts) {
		return false
	}

	// Match each part (from right to left is often more efficient for *.xxx patterns)
	for i := range keyParts {
		if patternParts[i] == "*" {
			continue
		}
		if keyParts[i] != patternParts[i] {
			return false
		}
	}

	return true
}

// MatchMultiple checks if key matches any pattern in a list
func (m *Matcher) MatchMultiple(keyPath string, patterns []*KeyPattern) bool {
	for _, pattern := range patterns {
		if m.Match(keyPath, pattern) {
			return true
		}
	}
	return false
}
