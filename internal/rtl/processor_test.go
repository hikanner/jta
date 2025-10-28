package rtl

import (
	"strings"
	"testing"
)

func TestProcessorCreation(t *testing.T) {
	processor := NewProcessor()
	if processor == nil {
		t.Fatal("NewProcessor() returned nil")
	}
	if processor.lrmChar != "\u200E" {
		t.Error("LRM character not set correctly")
	}
	if processor.rlmChar != "\u200F" {
		t.Error("RLM character not set correctly")
	}
}

func TestNeedProcessing(t *testing.T) {
	processor := NewProcessor()

	tests := []struct {
		lang     string
		expected bool
	}{
		{"ar", true},  // Arabic - RTL
		{"he", true},  // Hebrew - RTL
		{"fa", true},  // Persian - RTL
		{"ur", true},  // Urdu - RTL
		{"en", false}, // English - LTR
		{"zh", false}, // Chinese - LTR
		{"ja", false}, // Japanese - LTR
		{"xx", false}, // Invalid language
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := processor.NeedProcessing(tt.lang)
			if result != tt.expected {
				t.Errorf("NeedProcessing(%q) = %v, want %v", tt.lang, result, tt.expected)
			}
		})
	}
}

func TestProcessText_NonRTL(t *testing.T) {
	processor := NewProcessor()
	text := "Hello world"

	result := processor.ProcessText(text, "en")
	if result != text {
		t.Errorf("ProcessText() modified non-RTL text: got %q, want %q", result, text)
	}
}

func TestProcessText_Arabic(t *testing.T) {
	processor := NewProcessor()
	lrm := "\u200E"

	tests := []struct {
		name     string
		input    string
		checkFor string
		desc     string
	}{
		{
			name:     "URL protection",
			input:    "Visit https://example.com for more",
			checkFor: lrm + "https://example.com" + lrm,
			desc:     "URL should be wrapped with LRM marks",
		},
		{
			name:     "email protection",
			input:    "Contact user@example.com",
			checkFor: lrm + "user@example.com" + lrm,
			desc:     "Email should be wrapped with LRM marks",
		},
		{
			name:     "question mark conversion",
			input:    "What is this?",
			checkFor: "؟",
			desc:     "Question mark should be converted to Arabic ؟",
		},
		{
			name:     "comma conversion",
			input:    "First, second, third",
			checkFor: "،",
			desc:     "Commas should be converted to Arabic ،",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.ProcessText(tt.input, "ar")
			if !strings.Contains(result, tt.checkFor) {
				t.Errorf("%s\nInput: %q\nOutput: %q\nExpected to contain: %q",
					tt.desc, tt.input, result, tt.checkFor)
			}
		})
	}
}

func TestConvertPunctuation(t *testing.T) {
	processor := NewProcessor()

	punctuation := map[string]string{
		"?": "؟",
		",": "،",
		";": "؛",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"What?", "What؟"},
		{"Hello, world", "Hello، world"},
		{"First; second", "First؛ second"},
		{"Question? Yes, sure; okay", "Question؟ Yes، sure؛ okay"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := processor.convertPunctuation(tt.input, punctuation)
			if result != tt.expected {
				t.Errorf("convertPunctuation(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAddDirectionalMarks(t *testing.T) {
	processor := NewProcessor()
	lrm := "\u200E"

	tests := []struct {
		name  string
		input string
		check func(string) bool
	}{
		{
			name:  "URL gets LRM marks",
			input: "Visit https://example.com",
			check: func(result string) bool {
				return strings.Contains(result, lrm+"https://example.com"+lrm)
			},
		},
		{
			name:  "Email gets LRM marks",
			input: "Email: test@example.com",
			check: func(result string) bool {
				return strings.Contains(result, lrm+"test@example.com"+lrm)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.addDirectionalMarks(tt.input)
			if !tt.check(result) {
				t.Errorf("addDirectionalMarks(%q) = %q, check failed", tt.input, result)
			}
		})
	}
}

func TestProcessBatch(t *testing.T) {
	processor := NewProcessor()

	input := map[string]string{
		"title":   "What is this?",
		"message": "Hello, world",
		"url":     "Visit https://example.com",
	}

	// Test with Arabic (RTL)
	result := processor.ProcessBatch(input, "ar")

	if len(result) != len(input) {
		t.Errorf("ProcessBatch() returned %d items, want %d", len(result), len(input))
	}

	// Check that punctuation was converted
	if !strings.Contains(result["title"], "؟") {
		t.Error("ProcessBatch() did not convert question mark in title")
	}

	if !strings.Contains(result["message"], "،") {
		t.Error("ProcessBatch() did not convert comma in message")
	}

	// Test with English (LTR) - should not modify
	resultEn := processor.ProcessBatch(input, "en")
	if resultEn["title"] != input["title"] {
		t.Error("ProcessBatch() modified text for non-RTL language")
	}
}

func TestAddLRM(t *testing.T) {
	processor := NewProcessor()
	lrm := "\u200E"
	text := "test"

	result := processor.AddLRM(text)
	expected := lrm + text + lrm

	if result != expected {
		t.Errorf("AddLRM(%q) = %q, want %q", text, result, expected)
	}
}

func TestAddRLM(t *testing.T) {
	processor := NewProcessor()
	rlm := "\u200F"
	text := "test"

	result := processor.AddRLM(text)
	expected := rlm + text + rlm

	if result != expected {
		t.Errorf("AddRLM(%q) = %q, want %q", text, result, expected)
	}
}

func TestStripDirectionalMarks(t *testing.T) {
	processor := NewProcessor()
	lrm := "\u200E"
	rlm := "\u200F"

	tests := []struct {
		input    string
		expected string
	}{
		{lrm + "test" + lrm, "test"},
		{rlm + "test" + rlm, "test"},
		{lrm + "test" + rlm + "text", "testtext"},
		{"normal text", "normal text"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := processor.StripDirectionalMarks(tt.input)
			if result != tt.expected {
				t.Errorf("StripDirectionalMarks(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
