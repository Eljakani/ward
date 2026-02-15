package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/models"
	"github.com/eljakani/ward/internal/tui/theme"
)

// RenderLiveStats renders horizontal severity count badges.
//
// Example output:
//
//	Critical: 2  |  High: 5  |  Medium: 12  |  Low: 3  |  Info: 8
func RenderLiveStats(counts map[models.Severity]int, t *theme.Theme, width int) string {
	// Display in descending severity order (Critical first)
	severities := []models.Severity{
		models.SeverityCritical,
		models.SeverityHigh,
		models.SeverityMedium,
		models.SeverityLow,
		models.SeverityInfo,
	}

	var badges []string
	for _, sev := range severities {
		count := counts[sev]
		style := t.SeverityStyles[sev]
		badge := style.Render(fmt.Sprintf(" %s: %d ", sev.String(), count))
		badges = append(badges, badge)
	}

	separator := t.Muted.Render("  ")
	row := strings.Join(badges, separator)

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, row)
}

// RenderTotalFindings renders a compact total count.
func RenderTotalFindings(total int, t *theme.Theme) string {
	return t.Bold.Render(fmt.Sprintf("%d findings", total))
}
