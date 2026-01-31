// Package ui provides centralized styling for the CLI template.
// All visual styling should be defined here - rendering code should never create inline styles.
//
// Design Philosophy:
// - Information density over decoration
// - Every visual element must serve a purpose
// - No emojis, no bright/neon colors, no heavy borders
// - Output must remain readable without color
package ui

import "github.com/charmbracelet/lipgloss"

// =============================================================================
// COLOR PALETTE (ANSI 256)
// =============================================================================
// Professional, muted colors that work well on both light and dark terminals.

var (
	ColorPrimary   = lipgloss.Color("111") // Soft cyan — titles, active focus
	ColorSecondary = lipgloss.Color("250") // Light gray — labels, headers
	ColorNeutral   = lipgloss.Color("245") // Muted gray — secondary text, meta
	ColorAccent    = lipgloss.Color("109") // Soft green — success states
	ColorWarning   = lipgloss.Color("180") // Soft amber — caution states
	ColorError     = lipgloss.Color("167") // Soft red — failure states
	ColorHighlight = lipgloss.Color("147") // Soft purple — selection, links

	// Extended palette for specific use cases
	ColorDim = lipgloss.Color("240") // Very muted — disabled, inactive
)

// =============================================================================
// STATUS INDICATORS (Symbols, not emojis)
// =============================================================================

const (
	SymbolDone    = "•" // Completed/checked item
	SymbolActive  = "→" // Currently active/selected
	SymbolRunning = "◉" // In-progress operation
	SymbolPending = "○" // Waiting/queued item
	SymbolWarning = "!" // Caution/attention needed
	SymbolError   = "✗" // Failed/error state
	SymbolInfo    = "·" // Informational bullet
)

// =============================================================================
// TYPOGRAPHY ROLE STYLES
// =============================================================================

var (
	// --- Structural Roles ---
	Title = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	Section = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Bold(true)

	Label = lipgloss.NewStyle().
		Foreground(ColorSecondary)

	Value = lipgloss.NewStyle().
		Foreground(ColorPrimary)

	Meta = lipgloss.NewStyle().
		Foreground(ColorNeutral)

	// --- State Roles ---
	ActiveStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorAccent)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight)

	DimStyle = lipgloss.NewStyle().
			Foreground(ColorDim)
)

// =============================================================================
// COMPONENT STYLES
// =============================================================================

var (
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	ProgressFilled = lipgloss.NewStyle().
			Foreground(ColorAccent)

	ProgressEmpty = lipgloss.NewStyle().
			Foreground(ColorDim)

	Container = lipgloss.NewStyle().
			Padding(1, 2)
)

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func StatusDone(message string) string {
	return SuccessStyle.Render(SymbolDone) + " " + Meta.Render(message)
}

func StatusRunning(message string) string {
	return ActiveStyle.Render(SymbolRunning) + " " + Value.Render(message)
}

func StatusWarning(message string) string {
	return WarningStyle.Render(SymbolWarning) + " " + WarningStyle.Render(message)
}

func StatusError(message string) string {
	return ErrorStyle.Render(SymbolError) + " " + ErrorStyle.Render(message)
}

func StatusInfo(message string) string {
	return Meta.Render(SymbolInfo) + " " + Meta.Render(message)
}

func SectionHeader(text string) string {
	return Section.Render(text)
}

func Bullet(text string) string {
	return "  " + Meta.Render(SymbolInfo) + " " + text
}
