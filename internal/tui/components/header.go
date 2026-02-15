package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/laravel-ward/internal/tui/theme"
)

// HeaderData holds the information displayed in the header bar.
type HeaderData struct {
	ProjectName    string
	LaravelVersion string
	ToolVersion    string
	ScanRunning    bool
	ScanComplete   bool
	ScanError      bool
}

// RenderHeader renders the top header bar across the full width.
func RenderHeader(data HeaderData, t *theme.Theme, width int) string {
	logo := " WARD "

	project := ""
	if data.ProjectName != "" {
		project = fmt.Sprintf("  Project: %s", data.ProjectName)
	}

	laravel := ""
	if data.LaravelVersion != "" {
		laravel = fmt.Sprintf("  Laravel %s", data.LaravelVersion)
	}

	version := fmt.Sprintf("  v%s", data.ToolVersion)

	var status string
	switch {
	case data.ScanError:
		status = t.ErrorStyle.Render(" ERROR ")
	case data.ScanComplete:
		status = t.SuccessStyle.
			Background(t.Colors.Success).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Render(" COMPLETE ")
	case data.ScanRunning:
		status = t.WarningStyle.
			Background(t.Colors.StageActive).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Render(" SCANNING ")
	default:
		status = t.Muted.Render(" READY ")
	}

	left := logo + project + laravel + version
	right := status

	leftRendered := t.HeaderBar.Render(left)
	rightRendered := t.HeaderBar.Render(right)

	leftWidth := lipgloss.Width(leftRendered)
	rightWidth := lipgloss.Width(rightRendered)

	gap := width - leftWidth - rightWidth
	if gap < 0 {
		gap = 1
	}

	padding := t.HeaderBar.Render(fmt.Sprintf("%*s", gap, ""))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, padding, rightRendered)
}
