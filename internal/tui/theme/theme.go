package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/models"
)

// Colors defines the full color palette with adaptive light/dark support.
type Colors struct {
	Critical lipgloss.AdaptiveColor
	High     lipgloss.AdaptiveColor
	Medium   lipgloss.AdaptiveColor
	Low      lipgloss.AdaptiveColor
	Info     lipgloss.AdaptiveColor

	Primary    lipgloss.AdaptiveColor
	Secondary  lipgloss.AdaptiveColor
	Accent     lipgloss.AdaptiveColor
	Subtle     lipgloss.AdaptiveColor
	Text       lipgloss.AdaptiveColor
	TextDim    lipgloss.AdaptiveColor
	Border     lipgloss.AdaptiveColor
	Success    lipgloss.AdaptiveColor
	Warning    lipgloss.AdaptiveColor
	Error      lipgloss.AdaptiveColor
	Background lipgloss.AdaptiveColor

	StageActive   lipgloss.AdaptiveColor
	StageComplete lipgloss.AdaptiveColor
	StagePending  lipgloss.AdaptiveColor
}

// DefaultColors returns the default color scheme.
func DefaultColors() Colors {
	return Colors{
		Critical:   lipgloss.AdaptiveColor{Light: "#D32F2F", Dark: "#FF5252"},
		High:       lipgloss.AdaptiveColor{Light: "#F57C00", Dark: "#FFB74D"},
		Medium:     lipgloss.AdaptiveColor{Light: "#FFA000", Dark: "#FFD54F"},
		Low:        lipgloss.AdaptiveColor{Light: "#388E3C", Dark: "#81C784"},
		Info:       lipgloss.AdaptiveColor{Light: "#1976D2", Dark: "#64B5F6"},
		Primary:    lipgloss.AdaptiveColor{Light: "#5E35B1", Dark: "#B388FF"},
		Secondary:  lipgloss.AdaptiveColor{Light: "#455A64", Dark: "#B0BEC5"},
		Accent:     lipgloss.AdaptiveColor{Light: "#00ACC1", Dark: "#4DD0E1"},
		Subtle:     lipgloss.AdaptiveColor{Light: "#BDBDBD", Dark: "#424242"},
		Text:       lipgloss.AdaptiveColor{Light: "#212121", Dark: "#FAFAFA"},
		TextDim:    lipgloss.AdaptiveColor{Light: "#757575", Dark: "#9E9E9E"},
		Border:     lipgloss.AdaptiveColor{Light: "#E0E0E0", Dark: "#616161"},
		Success:    lipgloss.AdaptiveColor{Light: "#2E7D32", Dark: "#69F0AE"},
		Warning:    lipgloss.AdaptiveColor{Light: "#F9A825", Dark: "#FFD740"},
		Error:      lipgloss.AdaptiveColor{Light: "#C62828", Dark: "#FF5252"},
		Background: lipgloss.AdaptiveColor{Light: "#FAFAFA", Dark: "#1A1A2E"},

		StageActive:   lipgloss.AdaptiveColor{Light: "#1565C0", Dark: "#42A5F5"},
		StageComplete: lipgloss.AdaptiveColor{Light: "#2E7D32", Dark: "#69F0AE"},
		StagePending:  lipgloss.AdaptiveColor{Light: "#9E9E9E", Dark: "#616161"},
	}
}

// Theme holds all pre-built lipgloss styles.
type Theme struct {
	Colors Colors

	// Layout
	HeaderBar    lipgloss.Style
	FooterBar    lipgloss.Style
	ContentPanel lipgloss.Style
	SidePanel    lipgloss.Style

	// Stages
	ActiveStage    lipgloss.Style
	CompletedStage lipgloss.Style
	PendingStage   lipgloss.Style

	// Severity badges
	SeverityStyles map[models.Severity]lipgloss.Style

	// Table
	TableHeader   lipgloss.Style
	TableRow      lipgloss.Style
	TableRowAlt   lipgloss.Style
	TableSelected lipgloss.Style

	// Text
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Muted    lipgloss.Style
	Bold     lipgloss.Style
	Code     lipgloss.Style

	// Accents
	AccentStyle  lipgloss.Style
	SuccessStyle lipgloss.Style
	ErrorStyle   lipgloss.Style
	WarningStyle lipgloss.Style
}

// DefaultTheme creates a Theme from DefaultColors.
func DefaultTheme() *Theme {
	c := DefaultColors()

	t := &Theme{
		Colors: c,

		HeaderBar: lipgloss.NewStyle().
			Background(c.Primary).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
			Padding(0, 1).
			Bold(true),

		FooterBar: lipgloss.NewStyle().
			Foreground(c.TextDim).
			Padding(0, 1),

		ContentPanel: lipgloss.NewStyle().
			Padding(0, 1),

		SidePanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.Border).
			Padding(0, 1),

		ActiveStage: lipgloss.NewStyle().
			Foreground(c.StageActive).
			Bold(true),

		CompletedStage: lipgloss.NewStyle().
			Foreground(c.StageComplete),

		PendingStage: lipgloss.NewStyle().
			Foreground(c.StagePending),

		SeverityStyles: map[models.Severity]lipgloss.Style{
			models.SeverityCritical: lipgloss.NewStyle().
				Background(c.Critical).
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
				Bold(true).
				Padding(0, 1),
			models.SeverityHigh: lipgloss.NewStyle().
				Background(c.High).
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1A1A1A"}).
				Bold(true).
				Padding(0, 1),
			models.SeverityMedium: lipgloss.NewStyle().
				Background(c.Medium).
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1A1A1A"}).
				Padding(0, 1),
			models.SeverityLow: lipgloss.NewStyle().
				Background(c.Low).
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1A1A1A"}).
				Padding(0, 1),
			models.SeverityInfo: lipgloss.NewStyle().
				Background(c.Info).
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
				Padding(0, 1),
		},

		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(c.Primary).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(c.Border),

		TableRow: lipgloss.NewStyle().
			Foreground(c.Text),

		TableRowAlt: lipgloss.NewStyle().
			Foreground(c.Text),

		TableSelected: lipgloss.NewStyle().
			Foreground(c.Primary).
			Bold(true).
			Background(lipgloss.AdaptiveColor{Light: "#EDE7F6", Dark: "#2A2040"}),

		Title: lipgloss.NewStyle().
			Foreground(c.Text).
			Bold(true),

		Subtitle: lipgloss.NewStyle().
			Foreground(c.Secondary).
			Bold(true),

		Muted: lipgloss.NewStyle().
			Foreground(c.TextDim),

		Bold: lipgloss.NewStyle().
			Bold(true),

		Code: lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#F5F5F5", Dark: "#2D2D2D"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#D32F2F", Dark: "#FF8A80"}).
			Padding(0, 1),

		AccentStyle: lipgloss.NewStyle().
			Foreground(c.Accent),

		SuccessStyle: lipgloss.NewStyle().
			Foreground(c.Success),

		ErrorStyle: lipgloss.NewStyle().
			Foreground(c.Error),

		WarningStyle: lipgloss.NewStyle().
			Foreground(c.Warning),
	}

	return t
}

// SeverityColor returns the color for a severity level.
func (t *Theme) SeverityColor(s models.Severity) lipgloss.AdaptiveColor {
	switch s {
	case models.SeverityCritical:
		return t.Colors.Critical
	case models.SeverityHigh:
		return t.Colors.High
	case models.SeverityMedium:
		return t.Colors.Medium
	case models.SeverityLow:
		return t.Colors.Low
	case models.SeverityInfo:
		return t.Colors.Info
	default:
		return t.Colors.TextDim
	}
}
