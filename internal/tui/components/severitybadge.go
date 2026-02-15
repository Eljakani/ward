package components

import (
	"strings"

	"github.com/eljakani/ward/internal/models"
	"github.com/eljakani/ward/internal/tui/theme"
)

// RenderSeverityBadge renders a colored badge like " CRITICAL " or " HIGH ".
func RenderSeverityBadge(sev models.Severity, t *theme.Theme) string {
	style, ok := t.SeverityStyles[sev]
	if !ok {
		return sev.String()
	}
	return style.Render(strings.ToUpper(sev.String()))
}
