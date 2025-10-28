package ui

import (
	"fmt"
	"strings"
	"time"
)

// Printer provides styled console output
type Printer struct {
	verbose bool
}

// NewPrinter creates a new printer
func NewPrinter(verbose bool) *Printer {
	return &Printer{
		verbose: verbose,
	}
}

// PrintHeader prints a section header
func (p *Printer) PrintHeader(text string) {
	fmt.Println(HeaderStyle.Render(text))
}

// PrintSuccess prints a success message
func (p *Printer) PrintSuccess(text string) {
	fmt.Println(SuccessStyle.Render(IconSuccess + " " + text))
}

// PrintError prints an error message
func (p *Printer) PrintError(text string) {
	fmt.Println(ErrorStyle.Render(IconError + " " + text))
}

// PrintWarning prints a warning message
func (p *Printer) PrintWarning(text string) {
	fmt.Println(WarningStyle.Render(IconWarning + " " + text))
}

// PrintInfo prints an info message
func (p *Printer) PrintInfo(text string) {
	fmt.Println(InfoStyle.Render(IconInfo + " " + text))
}

// PrintSubtle prints subtle text
func (p *Printer) PrintSubtle(text string) {
	fmt.Println(SubtleStyle.Render("   " + text))
}

// PrintStep prints a step with icon
func (p *Printer) PrintStep(icon, text string) {
	fmt.Printf("%s %s\n", icon, text)
}

// PrintProgress prints a simple progress indicator
func (p *Printer) PrintProgress(current, total int, text string) {
	percentage := float64(current) / float64(total) * 100
	barWidth := 40
	filled := int(float64(barWidth) * float64(current) / float64(total))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Printf("\r%s [%s] %.0f%% (%d/%d)",
		text,
		HighlightStyle.Render(bar),
		percentage,
		current,
		total,
	)

	if current == total {
		fmt.Println() // New line when complete
	}
}

// PrintStats prints statistics in a formatted box
func (p *Printer) PrintStats(stats map[string]any) {
	var lines []string

	lines = append(lines, HeaderStyle.Render(IconChart+" Statistics:"))

	for key, value := range stats {
		line := fmt.Sprintf("%s %s",
			LabelStyle.Render(key+":"),
			ValueStyle.Render(fmt.Sprintf("%v", value)),
		)
		lines = append(lines, "   "+line)
	}

	fmt.Println(StatsStyle.Render(strings.Join(lines, "\n")))
}

// PrintSeparator prints a visual separator
func (p *Printer) PrintSeparator() {
	fmt.Println(SubtleStyle.Render(strings.Repeat("─", 60)))
}

// PrintBox prints text in a bordered box
func (p *Printer) PrintBox(text string) {
	fmt.Println(BoxStyle.Render(text))
}

// FormatDuration formats a duration in a human-readable way
func (p *Printer) FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// FormatNumber formats a number with commas
func (p *Printer) FormatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result []string
	for i := len(s); i > 0; i -= 3 {
		start := max(i-3, 0)
		result = append([]string{s[start:i]}, result...)
	}

	return strings.Join(result, ",")
}

// PrintVerbose prints only in verbose mode
func (p *Printer) PrintVerbose(text string) {
	if p.verbose {
		fmt.Println(SubtleStyle.Render("   [verbose] " + text))
	}
}

// Spinner represents a loading spinner
type Spinner struct {
	frames []string
	index  int
	text   string
}

// NewSpinner creates a new spinner
func NewSpinner(text string) *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		index:  0,
		text:   text,
	}
}

// Next advances the spinner and prints it
func (s *Spinner) Next() {
	frame := s.frames[s.index]
	fmt.Printf("\r%s %s", HighlightStyle.Render(frame), s.text)
	s.index = (s.index + 1) % len(s.frames)
}

// Finish completes the spinner
func (s *Spinner) Finish(success bool) {
	if success {
		fmt.Printf("\r%s %s\n", SuccessStyle.Render(IconSuccess), s.text)
	} else {
		fmt.Printf("\r%s %s\n", ErrorStyle.Render(IconError), s.text)
	}
}
