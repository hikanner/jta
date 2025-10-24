package translator

import (
	"testing"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/provider"
)

func TestCheckFormatIntegrity(t *testing.T) {
	mockProvider := &provider.OpenAIProvider{}
	engine := NewReflectionEngine(mockProvider)

	tests := []struct {
		name        string
		key         string
		source      string
		translated  string
		expectIssue bool
	}{
		{
			name:        "format preserved",
			key:         "message",
			source:      "You have {count} credits",
			translated:  "您有 {count} 个点数",
			expectIssue: false,
		},
		{
			name:        "format missing",
			key:         "message",
			source:      "You have {count} credits",
			translated:  "您有个点数",
			expectIssue: true,
		},
		{
			name:        "HTML preserved",
			key:         "desc",
			source:      "Use <b>bold</b> text",
			translated:  "使用<b>粗体</b>文本",
			expectIssue: false,
		},
		{
			name:        "HTML missing",
			key:         "desc",
			source:      "Use <b>bold</b> text",
			translated:  "使用粗体文本",
			expectIssue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := engine.checkFormatIntegrity(tt.key, tt.source, tt.translated)
			hasIssue := issue != nil

			if hasIssue != tt.expectIssue {
				t.Errorf("checkFormatIntegrity() hasIssue = %v, expectIssue %v", hasIssue, tt.expectIssue)
				if issue != nil {
					t.Logf("Issue: %+v", issue)
				}
			}

			if hasIssue && issue.Type != IssueTypeFormat {
				t.Errorf("Expected issue type Format, got %s", issue.Type)
			}

			if hasIssue && issue.Severity != SeverityCritical {
				t.Errorf("Expected severity Critical, got %s", issue.Severity)
			}
		})
	}
}

func TestCheckTerminologyConsistency(t *testing.T) {
	mockProvider := &provider.OpenAIProvider{}
	engine := NewReflectionEngine(mockProvider)

	terminology := &domain.Terminology{
		SourceLanguage: "en",
		PreserveTerms: []string{
			"MyApp",
			"API",
		},
		ConsistentTerms: map[string][]string{
			"en": {"credits"},
			"zh": {"点数"},
		},
	}

	tests := []struct {
		name        string
		key         string
		source      string
		translated  string
		expectIssue bool
		issueType   IssueType
	}{
		{
			name:        "preserve term kept",
			key:         "title",
			source:      "Welcome to MyApp",
			translated:  "欢迎使用 MyApp",
			expectIssue: false,
		},
		{
			name:        "preserve term missing",
			key:         "title",
			source:      "Welcome to MyApp",
			translated:  "欢迎使用我的应用",
			expectIssue: true,
			issueType:   IssueTypeTerminology,
		},
		{
			name:        "consistent term used",
			key:         "message",
			source:      "You have 10 credits",
			translated:  "您有 10 个点数",
			expectIssue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := engine.checkTerminologyConsistency(tt.key, tt.source, tt.translated, terminology)
			hasIssue := issue != nil

			if hasIssue != tt.expectIssue {
				t.Errorf("checkTerminologyConsistency() hasIssue = %v, expectIssue %v", hasIssue, tt.expectIssue)
				if issue != nil {
					t.Logf("Issue: %+v", issue)
				}
			}

			if hasIssue && issue.Type != tt.issueType {
				t.Errorf("Expected issue type %s, got %s", tt.issueType, issue.Type)
			}
		})
	}
}

func TestCheckCompleteness(t *testing.T) {
	mockProvider := &provider.OpenAIProvider{}
	engine := NewReflectionEngine(mockProvider)

	tests := []struct {
		name        string
		key         string
		source      string
		translated  string
		expectIssue bool
		severity    IssueSeverity
	}{
		{
			name:        "normal translation",
			key:         "message",
			source:      "Hello world",
			translated:  "你好世界",
			expectIssue: false,
		},
		{
			name:        "empty translation",
			key:         "message",
			source:      "Hello world",
			translated:  "",
			expectIssue: true,
			severity:    SeverityCritical,
		},
		{
			name:        "too short translation",
			key:         "message",
			source:      "This is a long message that should be translated properly",
			translated:  "好",
			expectIssue: true,
			severity:    SeverityHigh,
		},
		{
			name:        "short source allowed",
			key:         "ok",
			source:      "OK",
			translated:  "好",
			expectIssue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := engine.checkCompleteness(tt.key, tt.source, tt.translated)
			hasIssue := issue != nil

			if hasIssue != tt.expectIssue {
				t.Errorf("checkCompleteness() hasIssue = %v, expectIssue %v", hasIssue, tt.expectIssue)
				if issue != nil {
					t.Logf("Issue: %+v", issue)
				}
			}

			if hasIssue && issue.Type != IssueTypeCompleteness {
				t.Errorf("Expected issue type Completeness, got %s", issue.Type)
			}

			if hasIssue && issue.Severity != tt.severity {
				t.Errorf("Expected severity %s, got %s", tt.severity, issue.Severity)
			}
		})
	}
}

func TestQuickQualityCheck(t *testing.T) {
	mockProvider := &provider.OpenAIProvider{}
	engine := NewReflectionEngine(mockProvider)

	input := ReflectionInput{
		SourceTexts: map[string]string{
			"msg1": "You have {count} credits",
			"msg2": "Welcome to MyApp",
			"msg3": "This is a test message",
		},
		TranslatedTexts: map[string]string{
			"msg1": "您有个点数",    // Missing {count}
			"msg2": "欢迎使用我的应用", // Missing MyApp
			"msg3": "这是一个测试消息", // OK
		},
		SourceLang: "en",
		TargetLang: "zh",
		Terminology: &domain.Terminology{
			SourceLanguage: "en",
			PreserveTerms:  []string{"MyApp"},
		},
	}

	issues := engine.quickQualityCheck(input)

	// Should find at least 2 issues (msg1 format, msg2 terminology)
	if len(issues) < 2 {
		t.Errorf("quickQualityCheck() found %d issues, expected at least 2", len(issues))
		for _, issue := range issues {
			t.Logf("Found issue: %+v", issue)
		}
	}

	// Check that msg3 has no issues
	hasMsg3Issue := false
	for _, issue := range issues {
		if issue.Key == "msg3" {
			hasMsg3Issue = true
		}
	}
	if hasMsg3Issue {
		t.Error("msg3 should not have any issues")
	}
}

func TestFilterCriticalIssues(t *testing.T) {
	mockProvider := &provider.OpenAIProvider{}
	engine := NewReflectionEngine(mockProvider)

	issues := []QualityIssue{
		{Key: "k1", Type: IssueTypeFormat, Severity: SeverityCritical},
		{Key: "k2", Type: IssueTypeTerminology, Severity: SeverityHigh},
		{Key: "k3", Type: IssueTypeNaturalness, Severity: SeverityMedium},
		{Key: "k4", Type: IssueTypeContext, Severity: SeverityLow},
	}

	critical := engine.filterCriticalIssues(issues)

	// Should only include Critical and High severity
	if len(critical) != 2 {
		t.Errorf("filterCriticalIssues() got %d issues, want 2", len(critical))
	}

	for _, issue := range critical {
		if issue.Severity != SeverityCritical && issue.Severity != SeverityHigh {
			t.Errorf("filterCriticalIssues() included %s severity issue", issue.Severity)
		}
	}
}

func TestShouldReflect(t *testing.T) {
	mockProvider := &provider.OpenAIProvider{}
	engine := NewReflectionEngine(mockProvider)

	tests := []struct {
		name         string
		translations map[string]string
		terminology  *domain.Terminology
		expected     bool
	}{
		{
			name: "small batch without terminology",
			translations: map[string]string{
				"k1": "v1",
				"k2": "v2",
			},
			terminology: nil,
			expected:    false,
		},
		{
			name: "large batch without terminology",
			translations: map[string]string{
				"k1": "v1",
				"k2": "v2",
				"k3": "v3",
				"k4": "v4",
			},
			terminology: nil,
			expected:    true,
		},
		{
			name: "with terminology",
			translations: map[string]string{
				"k1": "v1",
			},
			terminology: &domain.Terminology{
				PreserveTerms: []string{"MyApp"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.ShouldReflect(tt.translations, tt.terminology)
			if result != tt.expected {
				t.Errorf("ShouldReflect() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseImprovedTranslations(t *testing.T) {
	mockProvider := &provider.OpenAIProvider{}
	engine := NewReflectionEngine(mockProvider)

	issuesByKey := map[string][]QualityIssue{
		"msg1": {{Key: "msg1", Type: IssueTypeFormat}},
		"msg2": {{Key: "msg2", Type: IssueTypeTerminology}},
	}

	tests := []struct {
		name     string
		response string
		expected map[string]string
	}{
		{
			name: "valid format",
			response: `
msg1: 您有 {count} 个点数
msg2: 欢迎使用 MyApp
`,
			expected: map[string]string{
				"msg1": "您有 {count} 个点数",
				"msg2": "欢迎使用 MyApp",
			},
		},
		{
			name: "with extra text",
			response: `
Here are the improvements:

msg1: 您有 {count} 个点数

Also:
msg2: 欢迎使用 MyApp
`,
			expected: map[string]string{
				"msg1": "您有 {count} 个点数",
				"msg2": "欢迎使用 MyApp",
			},
		},
		{
			name:     "empty response",
			response: "",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.parseImprovedTranslations(tt.response, issuesByKey)

			if len(result) != len(tt.expected) {
				t.Errorf("parseImprovedTranslations() got %d items, want %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("parseImprovedTranslations()[%s] = %q, want %q", key, result[key], expectedValue)
				}
			}
		})
	}
}
