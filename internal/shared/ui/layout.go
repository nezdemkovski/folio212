package ui

import "github.com/charmbracelet/lipgloss"

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
	logoStyle := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		Width(l.ContentWidth()).
		Align(lipgloss.Center)

	return logoStyle.Render(AppLogo)
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
