package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/models"
	"github.com/eljakani/ward/internal/tui/theme"
)

// RenderStageProgress renders a horizontal pipeline stage indicator.
//
// Example output:
//
//	✓ Provider  →  ✓ Resolvers  →  ● Scanners  →  ○ Post-Process  →  ○ Report
func RenderStageProgress(currentStage models.PipelineStage, scanComplete bool, t *theme.Theme, width int) string {
	stages := models.AllStages()
	var rendered []string

	for _, stage := range stages {
		var style lipgloss.Style
		var prefix string

		switch {
		case scanComplete:
			style = t.CompletedStage
			prefix = "✓ "
		case stage < currentStage:
			style = t.CompletedStage
			prefix = "✓ "
		case stage == currentStage:
			style = t.ActiveStage
			prefix = "● "
		default:
			style = t.PendingStage
			prefix = "○ "
		}

		label := style.Render(prefix + stage.String())
		rendered = append(rendered, label)
	}

	separator := t.Muted.Render(" → ")
	row := strings.Join(rendered, separator)

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, row)
}
