package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout provides a consistent centered layout for all screens.
type Layout struct {
	width  int
	height int
}

func NewLayout(width, height int) Layout {
	return Layout{width: width, height: height}
}

func (l *Layout) UpdateDimensions(width, height int) {
	l.width = width
	l.height = height
}

// ContentWidth returns the recommended content width.
func (l Layout) ContentWidth() int {
	width := l.width - 8
	if width > 80 {
		return 80
	}
	if width < 60 {
		return 60
	}
	return width
}

func (l Layout) RenderLogo() string {
	lines := lipgloss.NewStyle().
		Width(l.ContentWidth()).
		Align(lipgloss.Center)

	// Two-tone logo for a bit more depth.
	raw := strings.Split(AppLogo, "\n")
	if len(raw) == 0 {
		return ""
	}

	cut := len(raw) / 2
	var rendered []string
	for i, line := range raw {
		color := ColorPrimary
		if i >= cut {
			color = ColorHighlight
		}
		style := lines.Foreground(color).Bold(true)
		rendered = append(rendered, style.Render(line))
	}

	return strings.Join(rendered, "\n")
}

func (l Layout) RenderSubtitle(text string) string {
	subtitleStyle := lipgloss.NewStyle().
		Foreground(ColorNeutral).
		Width(l.ContentWidth()).
		Align(lipgloss.Center).
		MarginTop(1).
		MarginBottom(2)

	return subtitleStyle.Render(text)
}

func (l Layout) RenderBody(content string) string {
	bodyStyle := lipgloss.NewStyle().
		Width(l.ContentWidth()).
		Align(lipgloss.Left)

	return bodyStyle.Render(content)
}

// RenderCentered centers content in the terminal using lipgloss.Place.
func (l Layout) RenderCentered(sections ...string) string {
	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(
		l.width,
		l.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
