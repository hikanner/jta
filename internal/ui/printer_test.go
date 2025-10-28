package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestNewPrinter(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{"verbose printer", true},
		{"non-verbose printer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrinter(tt.verbose)
			if p == nil {
				t.Fatal("NewPrinter() returned nil")
			}
			if p.verbose != tt.verbose {
				t.Errorf("verbose = %v, want %v", p.verbose, tt.verbose)
			}
		})
	}
}

func TestPrinter_PrintHeader(t *testing.T) {
	p := NewPrinter(false)
	output := captureOutput(func() {
		p.PrintHeader("Test Header")
	})

	if !strings.Contains(output, "Test Header") {
		t.Errorf("PrintHeader output = %q, should contain 'Test Header'", output)
	}
}

func TestPrinter_PrintSuccess(t *testing.T) {
	p := NewPrinter(false)
	output := captureOutput(func() {
		p.PrintSuccess("Operation successful")
	})

	if !strings.Contains(output, "Operation successful") {
		t.Errorf("PrintSuccess output should contain message")
	}
	if !strings.Contains(output, IconSuccess) {
		t.Errorf("PrintSuccess output should contain success icon")
	}
}

func TestPrinter_PrintError(t *testing.T) {
	p := NewPrinter(false)
	output := captureOutput(func() {
		p.PrintError("Error occurred")
	})

	if !strings.Contains(output, "Error occurred") {
		t.Errorf("PrintError output should contain message")
	}
	if !strings.Contains(output, IconError) {
		t.Errorf("PrintError output should contain error icon")
	}
}

func TestPrinter_PrintWarning(t *testing.T) {
	p := NewPrinter(false)
	output := captureOutput(func() {
		p.PrintWarning("Warning message")
	})

	if !strings.Contains(output, "Warning message") {
		t.Errorf("PrintWarning output should contain message")
	}
	if !strings.Contains(output, IconWarning) {
		t.Errorf("PrintWarning output should contain warning icon")
	}
}

func TestPrinter_PrintInfo(t *testing.T) {
	p := NewPrinter(false)
	output := captureOutput(func() {
		p.PrintInfo("Info message")
	})

	if !strings.Contains(output, "Info message") {
		t.Errorf("PrintInfo output should contain message")
	}
	if !strings.Contains(output, IconInfo) {
		t.Errorf("PrintInfo output should contain info icon")
	}
}

func TestPrinter_PrintSubtle(t *testing.T) {
	p := NewPrinter(false)
	output := captureOutput(func() {
		p.PrintSubtle("Subtle text")
	})

	if !strings.Contains(output, "Subtle text") {
		t.Errorf("PrintSubtle output should contain message")
	}
}

func TestPrinter_PrintStep(t *testing.T) {
	p := NewPrinter(false)
	output := captureOutput(func() {
		p.PrintStep(IconRocket, "Step message")
	})

	if !strings.Contains(output, "Step message") {
		t.Errorf("PrintStep output should contain message")
	}
	if !strings.Contains(output, IconRocket) {
		t.Errorf("PrintStep output should contain icon")
	}
}

func TestPrinter_PrintProgress(t *testing.T) {
	p := NewPrinter(false)

	tests := []struct {
		name    string
		current int
		total   int
		text    string
	}{
		{"start progress", 0, 100, "Processing"},
		{"middle progress", 50, 100, "Processing"},
		{"complete progress", 100, 100, "Processing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				p.PrintProgress(tt.current, tt.total, tt.text)
			})

			if !strings.Contains(output, tt.text) {
				t.Errorf("PrintProgress output should contain text")
			}
		})
	}
}

func TestPrinter_PrintStats(t *testing.T) {
	p := NewPrinter(false)
	stats := map[string]any{
		"Files":    10,
		"Lines":    1000,
		"Duration": "5s",
	}

	output := captureOutput(func() {
		p.PrintStats(stats)
	})

	if !strings.Contains(output, "Statistics") {
		t.Errorf("PrintStats output should contain 'Statistics'")
	}
}

func TestPrinter_PrintSeparator(t *testing.T) {
	p := NewPrinter(false)
	output := captureOutput(func() {
		p.PrintSeparator()
	})

	if !strings.Contains(output, "─") {
		t.Errorf("PrintSeparator should output line character")
	}
}

func TestPrinter_PrintBox(t *testing.T) {
	p := NewPrinter(false)
	output := captureOutput(func() {
		p.PrintBox("Boxed text")
	})

	if !strings.Contains(output, "Boxed text") {
		t.Errorf("PrintBox output should contain message")
	}
}

func TestPrinter_FormatDuration(t *testing.T) {
	p := NewPrinter(false)

	tests := []struct {
		name     string
		duration time.Duration
		contains string
	}{
		{"milliseconds", 500 * time.Millisecond, "ms"},
		{"seconds", 5 * time.Second, "s"},
		{"minutes", 2 * time.Minute, "m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.FormatDuration(tt.duration)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("FormatDuration(%v) = %q, should contain %q", tt.duration, result, tt.contains)
			}
		})
	}
}

func TestPrinter_FormatNumber(t *testing.T) {
	p := NewPrinter(false)

	tests := []struct {
		name     string
		number   int
		expected string
	}{
		{"small number", 123, "123"},
		{"thousands", 1234, "1,234"},
		{"millions", 1234567, "1,234,567"},
		{"zero", 0, "0"},
		{"single digit", 5, "5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.FormatNumber(tt.number)
			if result != tt.expected {
				t.Errorf("FormatNumber(%d) = %q, want %q", tt.number, result, tt.expected)
			}
		})
	}
}

func TestPrinterVerbose(t *testing.T) {
	// Test that verbose flag is stored correctly
	verbosePrinter := NewPrinter(true)
	if !verbosePrinter.verbose {
		t.Error("verbose printer should have verbose=true")
	}

	quietPrinter := NewPrinter(false)
	if quietPrinter.verbose {
		t.Error("quiet printer should have verbose=false")
	}
}

func TestPrinter_AllPrintMethods(t *testing.T) {
	// Ensure all print methods don't panic
	p := NewPrinter(true)

	tests := []struct {
		name string
		fn   func()
	}{
		{"PrintHeader", func() { p.PrintHeader("test") }},
		{"PrintSuccess", func() { p.PrintSuccess("test") }},
		{"PrintError", func() { p.PrintError("test") }},
		{"PrintWarning", func() { p.PrintWarning("test") }},
		{"PrintInfo", func() { p.PrintInfo("test") }},
		{"PrintSubtle", func() { p.PrintSubtle("test") }},
		{"PrintStep", func() { p.PrintStep("🚀", "test") }},
		{"PrintProgress", func() { p.PrintProgress(50, 100, "test") }},
		{"PrintStats", func() { p.PrintStats(map[string]any{"key": "value"}) }},
		{"PrintSeparator", func() { p.PrintSeparator() }},
		{"PrintBox", func() { p.PrintBox("test") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()

			captureOutput(tt.fn)
		})
	}
}

func TestPrinter_FormatMethods(t *testing.T) {
	p := NewPrinter(false)

	// Test FormatDuration doesn't panic with edge cases
	durations := []time.Duration{
		0,
		1 * time.Nanosecond,
		1 * time.Millisecond,
		1 * time.Second,
		1 * time.Minute,
		1 * time.Hour,
	}

	for _, d := range durations {
		result := p.FormatDuration(d)
		if result == "" {
			t.Errorf("FormatDuration(%v) returned empty string", d)
		}
	}

	// Test FormatNumber doesn't panic with edge cases
	numbers := []int{0, 1, -1, 999, 1000, 999999, 1000000}

	for _, n := range numbers {
		result := p.FormatNumber(n)
		if result == "" {
			t.Errorf("FormatNumber(%d) returned empty string", n)
		}
	}
}
