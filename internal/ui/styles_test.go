package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorDefinitions(t *testing.T) {
	colors := map[string]lipgloss.Color{
		"ColorPrimary":   ColorPrimary,
		"ColorSuccess":   ColorSuccess,
		"ColorWarning":   ColorWarning,
		"ColorError":     ColorError,
		"ColorInfo":      ColorInfo,
		"ColorSubtle":    ColorSubtle,
		"ColorHighlight": ColorHighlight,
	}

	for name, color := range colors {
		if string(color) == "" {
			t.Errorf("%s should not be empty", name)
		}
	}
}

func TestStyleDefinitions(t *testing.T) {
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"HeaderStyle", HeaderStyle},
		{"SuccessStyle", SuccessStyle},
		{"ErrorStyle", ErrorStyle},
		{"WarningStyle", WarningStyle},
		{"InfoStyle", InfoStyle},
		{"SubtleStyle", SubtleStyle},
		{"HighlightStyle", HighlightStyle},
		{"BoxStyle", BoxStyle},
		{"StatsStyle", StatsStyle},
		{"LabelStyle", LabelStyle},
		{"ValueStyle", ValueStyle},
	}

	for _, tt := range styles {
		t.Run(tt.name, func(t *testing.T) {
			// Test that style can render text
			result := tt.style.Render("test")
			if result == "" {
				t.Errorf("%s.Render() returned empty string", tt.name)
			}
		})
	}
}

func TestIconConstants(t *testing.T) {
	icons := map[string]string{
		"IconSuccess": IconSuccess,
		"IconError":   IconError,
		"IconWarning": IconWarning,
		"IconInfo":    IconInfo,
		"IconLoading": IconLoading,
		"IconRocket":  IconRocket,
		"IconBook":    IconBook,
		"IconRobot":   IconRobot,
		"IconMagnify": IconMagnify,
		"IconSparkle": IconSparkle,
		"IconFile":    IconFile,
		"IconSave":    IconSave,
		"IconChart":   IconChart,
		"IconTarget":  IconTarget,
		"IconCheck":   IconCheck,
		"IconCross":   IconCross,
	}

	for name, icon := range icons {
		if icon == "" {
			t.Errorf("%s should not be empty", name)
		}
	}
}

func TestHeaderStyle(t *testing.T) {
	result := HeaderStyle.Render("Test Header")
	if result == "" {
		t.Error("HeaderStyle should render text")
	}
	// HeaderStyle should be bold
	if !HeaderStyle.GetBold() {
		t.Error("HeaderStyle should be bold")
	}
}

func TestSuccessStyle(t *testing.T) {
	result := SuccessStyle.Render("Success")
	if result == "" {
		t.Error("SuccessStyle should render text")
	}
	if !SuccessStyle.GetBold() {
		t.Error("SuccessStyle should be bold")
	}
}

func TestErrorStyle(t *testing.T) {
	result := ErrorStyle.Render("Error")
	if result == "" {
		t.Error("ErrorStyle should render text")
	}
	if !ErrorStyle.GetBold() {
		t.Error("ErrorStyle should be bold")
	}
}

func TestWarningStyle(t *testing.T) {
	result := WarningStyle.Render("Warning")
	if result == "" {
		t.Error("WarningStyle should render text")
	}
	if !WarningStyle.GetBold() {
		t.Error("WarningStyle should be bold")
	}
}

func TestInfoStyle(t *testing.T) {
	result := InfoStyle.Render("Info")
	if result == "" {
		t.Error("InfoStyle should render text")
	}
}

func TestSubtleStyle(t *testing.T) {
	result := SubtleStyle.Render("Subtle")
	if result == "" {
		t.Error("SubtleStyle should render text")
	}
}

func TestHighlightStyle(t *testing.T) {
	result := HighlightStyle.Render("Highlight")
	if result == "" {
		t.Error("HighlightStyle should render text")
	}
	if !HighlightStyle.GetBold() {
		t.Error("HighlightStyle should be bold")
	}
}

func TestBoxStyle(t *testing.T) {
	result := BoxStyle.Render("Box")
	if result == "" {
		t.Error("BoxStyle should render text")
	}
	// BoxStyle should have border
	if BoxStyle.GetBorderStyle() == lipgloss.NormalBorder() {
		t.Error("BoxStyle should have a border set")
	}
}

func TestStatsStyle(t *testing.T) {
	result := StatsStyle.Render("Stats")
	if result == "" {
		t.Error("StatsStyle should render text")
	}
}

func TestLabelStyle(t *testing.T) {
	result := LabelStyle.Render("Label")
	if result == "" {
		t.Error("LabelStyle should render text")
	}
}

func TestValueStyle(t *testing.T) {
	result := ValueStyle.Render("Value")
	if result == "" {
		t.Error("ValueStyle should render text")
	}
	if !ValueStyle.GetBold() {
		t.Error("ValueStyle should be bold")
	}
}

func TestAllIconsAreUnique(t *testing.T) {
	icons := []string{
		IconSuccess, IconError, IconWarning, IconInfo,
		IconLoading, IconRocket, IconBook, IconRobot,
		IconMagnify, IconSparkle, IconFile, IconSave,
		IconChart, IconTarget, IconCheck, IconCross,
	}

	seen := make(map[string]bool)
	for _, icon := range icons {
		if seen[icon] {
			t.Errorf("Duplicate icon found: %q", icon)
		}
		seen[icon] = true
	}
}

func TestColorValues(t *testing.T) {
	tests := []struct {
		name     string
		color    lipgloss.Color
		expected string
	}{
		{"ColorPrimary", ColorPrimary, "#00D9FF"},
		{"ColorSuccess", ColorSuccess, "#00FF87"},
		{"ColorWarning", ColorWarning, "#FFD700"},
		{"ColorError", ColorError, "#FF5F87"},
		{"ColorInfo", ColorInfo, "#8787FF"},
		{"ColorSubtle", ColorSubtle, "#626262"},
		{"ColorHighlight", ColorHighlight, "#FF87D7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.color) != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.color, tt.expected)
			}
		})
	}
}

func TestStylesCanBeUsedConcurrently(t *testing.T) {
	// Test that styles are safe for concurrent use
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			HeaderStyle.Render("test")
			SuccessStyle.Render("test")
			ErrorStyle.Render("test")
			WarningStyle.Render("test")
			InfoStyle.Render("test")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestIconCategories(t *testing.T) {
	// Test that we have icons for different categories
	statusIcons := []string{IconSuccess, IconError, IconWarning, IconInfo}
	for _, icon := range statusIcons {
		if icon == "" {
			t.Error("Status icon should not be empty")
		}
	}

	progressIcons := []string{IconLoading, IconCheck, IconCross}
	for _, icon := range progressIcons {
		if icon == "" {
			t.Error("Progress icon should not be empty")
		}
	}

	functionalIcons := []string{IconFile, IconSave, IconChart}
	for _, icon := range functionalIcons {
		if icon == "" {
			t.Error("Functional icon should not be empty")
		}
	}

	decorativeIcons := []string{IconRocket, IconBook, IconRobot, IconMagnify, IconSparkle, IconTarget}
	for _, icon := range decorativeIcons {
		if icon == "" {
			t.Error("Decorative icon should not be empty")
		}
	}
}
