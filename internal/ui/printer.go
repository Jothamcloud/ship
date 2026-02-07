package ui

import (
	"fmt"
	"os"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// Printer handles formatted terminal output
type Printer struct {
	useColors bool
}

// NewPrinter creates a new Printer instance
func NewPrinter() *Printer {
	// Check if stdout is a terminal for color support
	useColors := true
	if os.Getenv("NO_COLOR") != "" {
		useColors = false
	}
	return &Printer{useColors: useColors}
}

// color wraps text in ANSI color codes if colors are enabled
func (p *Printer) color(c, text string) string {
	if !p.useColors {
		return text
	}
	return c + text + colorReset
}

// Info prints an informational message
func (p *Printer) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(p.color(colorBlue, "ℹ") + " " + msg)
}

// Success prints a success message
func (p *Printer) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(p.color(colorGreen, "✓") + " " + msg)
}

// Warning prints a warning message
func (p *Printer) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(p.color(colorYellow, "⚠") + " " + msg)
}

// Error prints an error message
func (p *Printer) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, p.color(colorRed, "✗")+" "+msg)
}

// Step prints a step message (action in progress)
func (p *Printer) Step(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(p.color(colorCyan, "→") + " " + msg)
}

// Hint prints a hint or suggestion
func (p *Printer) Hint(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(p.color(colorGray, "  "+msg))
}

// Divider prints a visual divider
func (p *Printer) Divider() {
	fmt.Println(p.color(colorGray, "─────────────────────────────────────────"))
}

// Plain prints a plain message without prefix
func (p *Printer) Plain(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// URL prints a clickable URL (in supported terminals like iTerm2, Hyper, Windows Terminal)
func (p *Printer) URL(label, url string) {
	if p.useColors {
		// OSC 8 hyperlink: ESC ] 8 ; ; URL ST TEXT ESC ] 8 ; ; ST
		// Using BEL (\a) as ST which has wider support
		link := fmt.Sprintf("\033]8;;%s\a%s%s%s\033]8;;\a", url, colorCyan, url, colorReset)
		fmt.Println(p.color(colorBlue, "ℹ") + " " + label + link)
	} else {
		fmt.Printf("ℹ %s%s\n", label, url)
	}
}
