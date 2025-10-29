package keyfilter

import (
	"testing"
)

func TestParsePatterns(t *testing.T) {
	filter := NewFilter()

	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{
			name:     "single exact pattern",
			input:    "settings.title",
			expected: 1,
			wantErr:  false,
		},
		{
			name:     "multiple patterns",
			input:    "settings.*,user.*,app.name",
			expected: 3,
			wantErr:  false,
		},
		{
			name:     "recursive pattern",
			input:    "settings.**",
			expected: 1,
			wantErr:  false,
		},
		{
			name:     "wildcard pattern",
			input:    "*.title",
			expected: 1,
			wantErr:  false,
		},
		{
			name:     "empty pattern",
			input:    "",
			expected: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns, err := filter.ParsePatterns(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePatterns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(patterns) != tt.expected {
				t.Errorf("ParsePatterns() got %d patterns, expected %d", len(patterns), tt.expected)
			}
		})
	}
}

func TestFilterKeys(t *testing.T) {
	filter := NewFilter()

	testData := map[string]any{
		"app": map[string]any{
			"name":    "Test App",
			"version": "1.0.0",
		},
		"settings": map[string]any{
			"title":       "Settings",
			"description": "App settings",
			"theme": map[string]any{
				"dark":  "Dark Mode",
				"light": "Light Mode",
			},
		},
		"buttons": map[string]any{
			"save":   "Save",
			"cancel": "Cancel",
		},
	}

	tests := []struct {
		name            string
		includePatterns string
		excludePatterns string
		expectedInclude int
	}{
		{
			name:            "filter settings only",
			includePatterns: "settings.*",
			excludePatterns: "",
			expectedInclude: 2, // settings.title, settings.description
		},
		{
			name:            "filter all with exclude",
			includePatterns: "",
			excludePatterns: "settings.**",
			expectedInclude: 4, // all except settings subtree (app.name, app.version, buttons.save, buttons.cancel)
		},
		{
			name:            "recursive pattern",
			includePatterns: "settings.**",
			excludePatterns: "",
			expectedInclude: 4, // settings.title, settings.description, settings.theme.dark, settings.theme.light
		},
		{
			name:            "exact match",
			includePatterns: "app.name",
			excludePatterns: "",
			expectedInclude: 1, // only app.name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			includePatterns, _ := filter.ParsePatterns(tt.includePatterns)
			excludePatterns, _ := filter.ParsePatterns(tt.excludePatterns)

			result, err := filter.FilterKeys(testData, includePatterns, excludePatterns)
			if err != nil {
				t.Errorf("FilterKeys() error = %v", err)
				return
			}

			if len(result.Included) != tt.expectedInclude {
				t.Errorf("FilterKeys() got %d included keys, expected %d. Keys: %v",
					len(result.Included), tt.expectedInclude, getKeys(result.Included))
			}
		})
	}
}

func TestMatcherExactMatch(t *testing.T) {
	filter := NewFilter()
	pattern, _ := filter.parsePattern("settings.title")

	tests := []struct {
		keyPath string
		want    bool
	}{
		{"settings.title", true},
		{"settings.description", false},
		{"settings", false},
		{"app.title", false},
	}

	for _, tt := range tests {
		t.Run(tt.keyPath, func(t *testing.T) {
			if got := filter.MatchKey(tt.keyPath, pattern); got != tt.want {
				t.Errorf("MatchKey(%q) = %v, want %v", tt.keyPath, got, tt.want)
			}
		})
	}
}

func TestMatcherSingleLevelWildcard(t *testing.T) {
	filter := NewFilter()
	pattern, _ := filter.parsePattern("settings.*")

	tests := []struct {
		keyPath string
		want    bool
	}{
		{"settings.title", true},
		{"settings.description", true},
		{"settings.theme.dark", false}, // too deep
		{"settings", false},
		{"app.title", false},
	}

	for _, tt := range tests {
		t.Run(tt.keyPath, func(t *testing.T) {
			if got := filter.MatchKey(tt.keyPath, pattern); got != tt.want {
				t.Errorf("MatchKey(%q) = %v, want %v", tt.keyPath, got, tt.want)
			}
		})
	}
}

func TestMatcherRecursiveWildcard(t *testing.T) {
	filter := NewFilter()
	pattern, _ := filter.parsePattern("settings.**")

	tests := []struct {
		keyPath string
		want    bool
	}{
		{"settings.title", true},
		{"settings.description", true},
		{"settings.theme.dark", true},
		{"settings.theme.light", true},
		{"settings", false},
		{"app.title", false},
	}

	for _, tt := range tests {
		t.Run(tt.keyPath, func(t *testing.T) {
			if got := filter.MatchKey(tt.keyPath, pattern); got != tt.want {
				t.Errorf("MatchKey(%q) = %v, want %v", tt.keyPath, got, tt.want)
			}
		})
	}
}

func TestRebuildJSON(t *testing.T) {
	filter := NewFilter()

	flattened := map[string]any{
		"app.name":            "Test App",
		"app.version":         "1.0.0",
		"settings.title":      "Settings",
		"settings.theme.dark": "Dark Mode",
	}

	rebuilt := filter.RebuildJSON(flattened)

	// Check structure
	if rebuilt["app"] == nil {
		t.Error("Expected 'app' key in rebuilt JSON")
	}

	app, ok := rebuilt["app"].(map[string]any)
	if !ok {
		t.Error("Expected 'app' to be a map")
	}

	if app["name"] != "Test App" {
		t.Errorf("Expected app.name = 'Test App', got %v", app["name"])
	}

	if app["version"] != "1.0.0" {
		t.Errorf("Expected app.version = '1.0.0', got %v", app["version"])
	}

	settings, ok := rebuilt["settings"].(map[string]any)
	if !ok {
		t.Error("Expected 'settings' to be a map")
	}

	if settings["title"] != "Settings" {
		t.Errorf("Expected settings.title = 'Settings', got %v", settings["title"])
	}

	theme, ok := settings["theme"].(map[string]any)
	if !ok {
		t.Error("Expected 'settings.theme' to be a map")
	}

	if theme["dark"] != "Dark Mode" {
		t.Errorf("Expected settings.theme.dark = 'Dark Mode', got %v", theme["dark"])
	}
}

// Helper function to get keys from a map
func getKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Test BuildKeyPath
func TestFilter_BuildKeyPath(t *testing.T) {
	filter := NewFilter()
	
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "single part",
			parts:    []string{"settings"},
			expected: "settings",
		},
		{
			name:     "two parts",
			parts:    []string{"settings", "title"},
			expected: "settings.title",
		},
		{
			name:     "three parts",
			parts:    []string{"settings", "theme", "dark"},
			expected: "settings.theme.dark",
		},
		{
			name:     "empty parts",
			parts:    []string{},
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.BuildKeyPath(tt.parts)
			if result != tt.expected {
				t.Errorf("BuildKeyPath() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// Test MatchMultiple
func TestMatcher_MatchMultiple(t *testing.T) {
	matcher := NewMatcher()
	filter := NewFilter()
	
	patterns, err := filter.ParsePatterns("settings.title,user.name,app.*")
	if err != nil {
		t.Fatalf("ParsePatterns() error = %v", err)
	}
	
	tests := []struct {
		name     string
		keyPath  string
		expected bool
	}{
		{
			name:     "matches first pattern",
			keyPath:  "settings.title",
			expected: true,
		},
		{
			name:     "matches second pattern",
			keyPath:  "user.name",
			expected: true,
		},
		{
			name:     "matches wildcard pattern",
			keyPath:  "app.version",
			expected: true,
		},
		{
			name:     "no match",
			keyPath:  "other.key",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.MatchMultiple(tt.keyPath, patterns)
			if result != tt.expected {
				t.Errorf("MatchMultiple() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test matchSingleLevel
func TestMatcher_MatchSingleLevel(t *testing.T) {
	matcher := NewMatcher()
	
	pattern := &KeyPattern{
		Pattern: "settings.*",
		Type:    PatternTypeSingleLevel,
		Parts:   []string{"settings", "*"},
	}
	
	tests := []struct {
		name     string
		keyPath  string
		expected bool
	}{
		{
			name:     "matches single level",
			keyPath:  "settings.title",
			expected: true,
		},
		{
			name:     "does not match nested",
			keyPath:  "settings.theme.dark",
			expected: false,
		},
		{
			name:     "does not match different prefix",
			keyPath:  "user.title",
			expected: false,
		},
		{
			name:     "does not match shorter path",
			keyPath:  "settings",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.matchSingleLevel(tt.keyPath, pattern)
			if result != tt.expected {
				t.Errorf("matchSingleLevel() = %v, want %v for key %s", result, tt.expected, tt.keyPath)
			}
		})
	}
}

// Test matchRecursive
func TestMatcher_MatchRecursive(t *testing.T) {
	matcher := NewMatcher()
	
	pattern := &KeyPattern{
		Pattern: "settings.**",
		Type:    PatternTypeRecursive,
		Parts:   []string{"settings", "**"},
	}
	
	tests := []struct {
		name     string
		keyPath  string
		expected bool
	}{
		{
			name:     "matches single level",
			keyPath:  "settings.title",
			expected: true,
		},
		{
			name:     "matches nested levels",
			keyPath:  "settings.theme.dark",
			expected: true,
		},
		{
			name:     "matches deep nested",
			keyPath:  "settings.theme.colors.primary.main",
			expected: true,
		},
		{
			name:     "does not match different prefix",
			keyPath:  "user.title",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.matchRecursive(tt.keyPath, pattern)
			if result != tt.expected {
				t.Errorf("matchRecursive() = %v, want %v for key %s", result, tt.expected, tt.keyPath)
			}
		})
	}
}

// Test matchWildcard
func TestMatcher_MatchWildcard(t *testing.T) {
	matcher := NewMatcher()
	
	pattern := &KeyPattern{
		Pattern: "*.title",
		Type:    PatternTypeWildcard,
		Parts:   []string{"*", "title"},
	}
	
	tests := []struct {
		name     string
		keyPath  string
		expected bool
	}{
		{
			name:     "matches with settings prefix",
			keyPath:  "settings.title",
			expected: true,
		},
		{
			name:     "matches with user prefix",
			keyPath:  "user.title",
			expected: true,
		},
		{
			name:     "matches with app prefix",
			keyPath:  "app.title",
			expected: true,
		},
		{
			name:     "does not match different suffix",
			keyPath:  "settings.name",
			expected: false,
		},
		{
			name:     "does not match nested",
			keyPath:  "settings.theme.title",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.matchWildcard(tt.keyPath, pattern)
			if result != tt.expected {
				t.Errorf("matchWildcard() = %v, want %v for key %s", result, tt.expected, tt.keyPath)
			}
		})
	}
}
