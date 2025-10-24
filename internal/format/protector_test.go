package format

import (
	"strings"
	"testing"
)

func TestNewProtector(t *testing.T) {
	p := NewProtector()
	if p == nil {
		t.Fatal("NewProtector returned nil")
	}
	if p.placeholderPattern == nil {
		t.Error("placeholderPattern is nil")
	}
	if p.htmlPattern == nil {
		t.Error("htmlPattern is nil")
	}
	if p.urlPattern == nil {
		t.Error("urlPattern is nil")
	}
	if p.markdownPattern == nil {
		t.Error("markdownPattern is nil")
	}
}

func TestExtractPlaceholders(t *testing.T) {
	p := NewProtector()

	tests := []struct {
		name     string
		text     string
		expected int
		elemType ElementType
	}{
		{
			name:     "single curly brace placeholder",
			text:     "Hello {name}!",
			expected: 1,
			elemType: ElementTypePlaceholder,
		},
		{
			name:     "double curly brace placeholder",
			text:     "Hello {{userName}}!",
			expected: 1,
			elemType: ElementTypePlaceholder,
		},
		{
			name:     "printf style placeholders",
			text:     "Count: %d, Name: %s",
			expected: 2,
			elemType: ElementTypePlaceholder,
		},
		{
			name:     "python style placeholders",
			text:     "Hello %(name)s, you have %(count)d messages",
			expected: 2,
			elemType: ElementTypePlaceholder,
		},
		{
			name:     "mixed placeholders",
			text:     "Hello {name}, you have %d messages in {{folder}}",
			expected: 3,
			elemType: ElementTypePlaceholder,
		},
		{
			name:     "no placeholders",
			text:     "Just plain text",
			expected: 0,
			elemType: ElementTypePlaceholder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements := p.Extract(tt.text)
			placeholders := filterByType(elements, ElementTypePlaceholder)
			if len(placeholders) != tt.expected {
				t.Errorf("Expected %d placeholders, got %d", tt.expected, len(placeholders))
				for _, elem := range placeholders {
					t.Logf("Found: %s", elem.Value)
				}
			}
		})
	}
}

func TestExtractHTML(t *testing.T) {
	p := NewProtector()

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "simple HTML tag",
			text:     "This is <b>bold</b> text",
			expected: 2,
		},
		{
			name:     "self-closing tag",
			text:     "Line break<br/>here",
			expected: 1,
		},
		{
			name:     "tag with attributes",
			text:     `Click <a href="https://example.com">here</a>`,
			expected: 2,
		},
		{
			name:     "multiple tags",
			text:     "<div><p>Hello</p><span>World</span></div>",
			expected: 6,
		},
		{
			name:     "no HTML",
			text:     "Plain text without HTML",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements := p.Extract(tt.text)
			html := filterByType(elements, ElementTypeHTML)
			if len(html) != tt.expected {
				t.Errorf("Expected %d HTML elements, got %d", tt.expected, len(html))
				for _, elem := range html {
					t.Logf("Found: %s", elem.Value)
				}
			}
		})
	}
}

func TestExtractURLs(t *testing.T) {
	p := NewProtector()

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "http URL",
			text:     "Visit http://example.com for more info",
			expected: 1,
		},
		{
			name:     "https URL",
			text:     "Visit https://example.com for more info",
			expected: 1,
		},
		{
			name:     "multiple URLs",
			text:     "Check https://github.com and https://google.com",
			expected: 2,
		},
		{
			name:     "URL with path",
			text:     "See https://example.com/path/to/page",
			expected: 1,
		},
		{
			name:     "no URLs",
			text:     "Just plain text",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements := p.Extract(tt.text)
			urls := filterByType(elements, ElementTypeURL)
			if len(urls) != tt.expected {
				t.Errorf("Expected %d URLs, got %d", tt.expected, len(urls))
				for _, elem := range urls {
					t.Logf("Found: %s", elem.Value)
				}
			}
		})
	}
}

func TestExtractMarkdown(t *testing.T) {
	p := NewProtector()

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "bold with asterisks",
			text:     "This is **bold** text",
			expected: 1,
		},
		{
			name:     "italic with single asterisk",
			text:     "This is *italic* text",
			expected: 1,
		},
		{
			name:     "bold with underscores",
			text:     "This is __bold__ text",
			expected: 1,
		},
		{
			name:     "italic with single underscore",
			text:     "This is _italic_ text",
			expected: 1,
		},
		{
			name:     "markdown link",
			text:     "Click [here](https://example.com) for more",
			expected: 1,
		},
		{
			name:     "mixed markdown",
			text:     "**Bold**, *italic*, and [link](url)",
			expected: 3,
		},
		{
			name:     "no markdown",
			text:     "Plain text",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements := p.Extract(tt.text)
			markdown := filterByType(elements, ElementTypeMarkdown)
			if len(markdown) != tt.expected {
				t.Errorf("Expected %d markdown elements, got %d", tt.expected, len(markdown))
				for _, elem := range markdown {
					t.Logf("Found: %s", elem.Value)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	p := NewProtector()

	tests := []struct {
		name       string
		original   string
		translated string
		expectErr  bool
	}{
		{
			name:       "all placeholders preserved",
			original:   "Hello {name}, you have {count} messages",
			translated: "你好 {name}，你有 {count} 条消息",
			expectErr:  false,
		},
		{
			name:       "missing placeholder",
			original:   "Hello {name}",
			translated: "你好",
			expectErr:  true,
		},
		{
			name:       "HTML preserved",
			original:   "This is <b>bold</b> text",
			translated: "这是<b>粗体</b>文本",
			expectErr:  false,
		},
		{
			name:       "missing HTML tag",
			original:   "This is <b>bold</b> text",
			translated: "这是粗体文本",
			expectErr:  true,
		},
		{
			name:       "URL preserved",
			original:   "Visit https://example.com",
			translated: "访问 https://example.com",
			expectErr:  false,
		},
		{
			name:       "markdown preserved",
			original:   "This is **bold** text",
			translated: "这是 **bold** 文本",
			expectErr:  false,
		},
		{
			name:       "no format elements",
			original:   "Plain text",
			translated: "纯文本",
			expectErr:  false,
		},
		{
			name:       "complex format preserved",
			original:   "Hello {name}, visit <a href=\"url\">link</a> for %d updates",
			translated: "你好 {name}，访问<a href=\"url\">链接</a>获取 %d 更新",
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.Validate(tt.original, tt.translated)
			if (err != nil) != tt.expectErr {
				t.Errorf("Validate() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestGetValidationReport(t *testing.T) {
	p := NewProtector()

	tests := []struct {
		name              string
		original          string
		translated        string
		expectValid       bool
		expectMissing     int
		expectExtra       int
		expectErrorsCount int
	}{
		{
			name:              "valid translation",
			original:          "Hello {name}",
			translated:        "你好 {name}",
			expectValid:       true,
			expectMissing:     0,
			expectExtra:       0,
			expectErrorsCount: 0,
		},
		{
			name:              "missing placeholder",
			original:          "Hello {name} and {user}",
			translated:        "你好 {name}",
			expectValid:       false,
			expectMissing:     1,
			expectExtra:       0,
			expectErrorsCount: 1,
		},
		{
			name:              "extra placeholder",
			original:          "Hello",
			translated:        "你好 {extra}",
			expectValid:       true,
			expectMissing:     0,
			expectExtra:       1,
			expectErrorsCount: 0,
		},
		{
			name:              "missing multiple elements",
			original:          "Visit <a href=\"url\">site</a> at {time}",
			translated:        "访问网站",
			expectValid:       false,
			expectMissing:     3, // <a>, </a>, {time}
			expectExtra:       0,
			expectErrorsCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := p.GetValidationReport(tt.original, tt.translated)
			if report.IsValid != tt.expectValid {
				t.Errorf("Expected IsValid=%v, got %v", tt.expectValid, report.IsValid)
			}
			if len(report.MissingElements) != tt.expectMissing {
				t.Errorf("Expected %d missing elements, got %d", tt.expectMissing, len(report.MissingElements))
			}
			if len(report.ExtraElements) != tt.expectExtra {
				t.Errorf("Expected %d extra elements, got %d", tt.expectExtra, len(report.ExtraElements))
			}
			if len(report.Errors) != tt.expectErrorsCount {
				t.Errorf("Expected %d errors, got %d", tt.expectErrorsCount, len(report.Errors))
				for _, err := range report.Errors {
					t.Logf("Error: %s", err)
				}
			}
		})
	}
}

func TestHasFormatElements(t *testing.T) {
	p := NewProtector()

	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "has placeholder",
			text:     "Hello {name}",
			expected: true,
		},
		{
			name:     "has HTML",
			text:     "This is <b>bold</b>",
			expected: true,
		},
		{
			name:     "has URL",
			text:     "Visit https://example.com",
			expected: true,
		},
		{
			name:     "has markdown",
			text:     "This is **bold**",
			expected: true,
		},
		{
			name:     "plain text",
			text:     "Just plain text",
			expected: false,
		},
		{
			name:     "empty string",
			text:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.HasFormatElements(tt.text)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBuildFormatInstructions(t *testing.T) {
	p := NewProtector()

	tests := []struct {
		name     string
		text     string
		contains []string
	}{
		{
			name:     "with placeholders",
			text:     "Hello {name}",
			contains: []string{"FORMAT PRESERVATION", "Placeholders", "{name}"},
		},
		{
			name:     "with HTML",
			text:     "This is <b>bold</b>",
			contains: []string{"FORMAT PRESERVATION", "HTML tags", "<b>", "</b>"},
		},
		{
			name:     "with URL",
			text:     "Visit https://example.com",
			contains: []string{"FORMAT PRESERVATION", "URLs", "https://example.com"},
		},
		{
			name:     "with markdown",
			text:     "This is **bold**",
			contains: []string{"FORMAT PRESERVATION", "Markdown", "**bold**"},
		},
		{
			name:     "multiple types",
			text:     "Visit <a href=\"https://example.com\">{site}</a>",
			contains: []string{"FORMAT PRESERVATION", "Placeholders", "HTML", "URLs"},
		},
		{
			name:     "no format elements",
			text:     "Plain text",
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instructions := p.BuildFormatInstructions(tt.text)

			if len(tt.contains) == 0 {
				if instructions != "" {
					t.Errorf("Expected empty instructions, got: %s", instructions)
				}
				return
			}

			for _, substr := range tt.contains {
				if !strings.Contains(instructions, substr) {
					t.Errorf("Expected instructions to contain %q, but it doesn't.\nInstructions: %s", substr, instructions)
				}
			}
		})
	}
}

func TestGroupElements(t *testing.T) {
	elements := []FormatElement{
		{Type: ElementTypePlaceholder, Value: "{name}", Position: 0},
		{Type: ElementTypePlaceholder, Value: "{name}", Position: 10},
		{Type: ElementTypePlaceholder, Value: "{count}", Position: 20},
		{Type: ElementTypeHTML, Value: "<b>", Position: 30},
		{Type: ElementTypeHTML, Value: "</b>", Position: 40},
	}

	result := groupElements(elements)

	expected := map[string]int{
		"placeholder:{name}":  2,
		"placeholder:{count}": 1,
		"html:<b>":            1,
		"html:</b>":           1,
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d groups, got %d", len(expected), len(result))
	}

	for key, expectedCount := range expected {
		if result[key] != expectedCount {
			t.Errorf("Expected count for %q to be %d, got %d", key, expectedCount, result[key])
		}
	}
}

// Helper function to filter elements by type
func filterByType(elements []FormatElement, elemType ElementType) []FormatElement {
	var result []FormatElement
	for _, elem := range elements {
		if elem.Type == elemType {
			result = append(result, elem)
		}
	}
	return result
}
