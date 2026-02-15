package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/laravel-ward/internal/tui/theme"
)

// RenderFooter renders the bottom help bar.
func RenderFooter(helpModel help.Model, keys help.KeyMap, t *theme.Theme, width int) string {
	helpView := helpModel.View(keys)
	return t.FooterBar.Width(width).Render(helpView)
}

// RenderSeparator renders a horizontal line separator.
func RenderSeparator(t *theme.Theme, width int) string {
	line := lipgloss.NewStyle().
		Foreground(t.Colors.Border).
		Render(repeat("─", width))
	return line
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		result = append(result, s...)
	}
	return string(result)
}
