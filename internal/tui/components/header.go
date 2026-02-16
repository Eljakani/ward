package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/tui/banner"
	"github.com/eljakani/ward/internal/tui/theme"
)

// HeaderData holds the information displayed in the header bar.
type HeaderData struct {
	ProjectName    string
	LaravelVersion string
	PHPVersion     string
	PackageCount   int
	ToolVersion    string
	ScanRunning    bool
	ScanComplete   bool
	ScanError      bool
}

// RenderHeader renders the top header bar across the full width.
func RenderHeader(data HeaderData, t *theme.Theme, width int) string {
	logo := " " + banner.RenderCompact() + " "

	sep := t.HeaderBar.Render(
		lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#BDBDBD", Dark: "#616161"}).
			Render(" | "),
	)

	project := ""
	if data.ProjectName != "" {
		project = t.HeaderBar.Render(fmt.Sprintf("Project: %s", data.ProjectName))
	}

	laravel := ""
	if data.LaravelVersion != "" {
		laravel = t.HeaderBar.Render(fmt.Sprintf("Laravel %s", data.LaravelVersion))
	}

	php := ""
	if data.PHPVersion != "" {
		php = t.HeaderBar.Render(fmt.Sprintf("PHP %s", data.PHPVersion))
	}

	packages := ""
	if data.PackageCount > 0 {
		packages = t.HeaderBar.Render(fmt.Sprintf("%d packages", data.PackageCount))
	}

	version := t.HeaderBar.Render(fmt.Sprintf("v%s", data.ToolVersion))

	var status string
	statusBg := lipgloss.NewStyle().Padding(0, 1).Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"})

	switch {
	case data.ScanError:
		status = statusBg.Background(t.Colors.Error).Render(" ERROR ")
	case data.ScanComplete:
		status = statusBg.Background(t.Colors.Success).Render(" COMPLETE ")
	case data.ScanRunning:
		status = statusBg.Background(t.Colors.StageActive).Render(" SCANNING ")
	default:
		status = t.Muted.Render(" READY ")
	}

	// Build left side
	leftParts := []string{logo}
	if project != "" {
		leftParts = append(leftParts, sep, project)
	}
	if laravel != "" {
		leftParts = append(leftParts, sep, laravel)
	}
	if php != "" {
		leftParts = append(leftParts, sep, php)
	}
	if packages != "" {
		leftParts = append(leftParts, sep, packages)
	}
	leftParts = append(leftParts, sep, version)
	left := lipgloss.JoinHorizontal(lipgloss.Center, leftParts...)

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(status)

	gap := width - leftWidth - rightWidth
	if gap < 0 {
		gap = 1
	}

	padding := fmt.Sprintf("%*s", gap, "")

	bar := t.HeaderBar.Width(width).Render(
		lipgloss.JoinHorizontal(lipgloss.Center, left, padding, status),
	)

	return bar
}
