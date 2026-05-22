package components

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/models"
	"github.com/eljakani/ward/internal/tui/theme"
)

// FindingDetail renders a detailed view of a single finding.
type FindingDetail struct {
	viewport viewport.Model
	finding  *models.Finding
	theme    *theme.Theme
	width    int
	height   int
}

// NewFindingDetail creates a new finding detail panel.
func NewFindingDetail(t *theme.Theme) *FindingDetail {
	vp := viewport.New(60, 20)
	return &FindingDetail{viewport: vp, theme: t}
}

// SetFinding sets the finding to display.
func (d *FindingDetail) SetFinding(f *models.Finding) {
	d.finding = f
	d.rebuildContent()
}

// SetSize updates the panel dimensions.
func (d *FindingDetail) SetSize(w, h int) {
	d.width = w
	d.height = h
	d.viewport.Width = w - 4
	d.viewport.Height = h - 4
	d.rebuildContent()
}

// HandleKey processes scrolling keys.
func (d *FindingDetail) HandleKey(msg tea.KeyMsg) {
	d.viewport, _ = d.viewport.Update(msg)
}

// CopyableText returns the current finding as plain text for clipboard copy.
func (d *FindingDetail) CopyableText() string {
	if d.finding == nil {
		return ""
	}
	f := d.finding
	var sb strings.Builder
	fmt.Fprintf(&sb, "Title:    %s\n", f.Title)
	fmt.Fprintf(&sb, "Severity: %s\n", f.Severity.String())
	fmt.Fprintf(&sb, "Category: %s\n", f.Category)
	fmt.Fprintf(&sb, "File:     %s:%d\n", f.File, f.Line)
	if f.Description != "" {
		fmt.Fprintf(&sb, "\nDescription:\n%s\n", f.Description)
	}
	if f.CodeSnippet != "" {
		fmt.Fprintf(&sb, "\nCode:\n%s\n", f.CodeSnippet)
	}
	if f.Remediation != "" {
		fmt.Fprintf(&sb, "\nRemediation:\n%s\n", f.Remediation)
	}
	if len(f.References) > 0 {
		fmt.Fprintf(&sb, "\nReferences:\n")
		for _, ref := range f.References {
			fmt.Fprintf(&sb, "  %s\n", ref)
		}
	}
	return sb.String()
}

// CopiedMsg is returned by CopyCmd when the copy succeeds.
type CopiedMsg struct{}

// CopyCmd returns a tea.Cmd that copies text to the clipboard via OSC52.
func CopyCmd(text string) tea.Cmd {
	return func() tea.Msg {
		encoded := base64.StdEncoding.EncodeToString([]byte(text))
		fmt.Fprintf(os.Stderr, "\033]52;c;%s\a", encoded)
		return CopiedMsg{}
	}
}

func (d *FindingDetail) rebuildContent() {
	if d.finding == nil {
		d.viewport.SetContent(d.theme.Muted.Render("  Select a finding to view details."))
		return
	}
	f := d.finding

	contentWidth := d.viewport.Width - 2
	if contentWidth < 20 {
		contentWidth = 20
	}

	sections := []string{
		// Title + Severity badge
		lipgloss.JoinHorizontal(lipgloss.Top,
			RenderSeverityBadge(f.Severity, d.theme),
			"  ",
			d.theme.Title.Render(f.Title),
		),
		"",
		// Category + Scanner
		d.theme.Muted.Render(fmt.Sprintf("  Category: %s  |  Scanner: %s", f.Category, f.Scanner)),
		"",
	}

	// Description
	if f.Description != "" {
		sections = append(sections,
			d.theme.Subtitle.Render("  Description"),
			wordWrap(f.Description, contentWidth),
			"",
		)
	}

	// Location
	if f.File != "" {
		location := fmt.Sprintf("  %s:%d", f.File, f.Line)
		sections = append(sections,
			d.theme.Subtitle.Render("  Location"),
			d.theme.AccentStyle.Render(location),
			"",
		)
	}

	// Code snippet
	if f.CodeSnippet != "" {
		codeBlock := d.theme.Code.Width(contentWidth - 2).Render(f.CodeSnippet)
		sections = append(sections,
			d.theme.Subtitle.Render("  Code"),
			"  "+codeBlock,
			"",
		)
	}

	// Remediation
	if f.Remediation != "" {
		sections = append(sections,
			d.theme.Subtitle.Render("  Remediation"),
			wordWrap(f.Remediation, contentWidth),
			"",
		)
	}

	// References
	if len(f.References) > 0 {
		sections = append(sections, d.theme.Subtitle.Render("  References"))
		for _, ref := range f.References {
			sections = append(sections, "  "+d.theme.AccentStyle.Render(ref))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	d.viewport.SetContent(content)
	d.viewport.GotoTop()
}

// View renders the finding detail panel.
func (d *FindingDetail) View() string {
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(d.theme.Colors.Border).
		Width(d.width).
		Height(d.height)

	title := d.theme.Subtitle.Render("  Finding Detail")
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", d.viewport.View())
	return border.Render(content)
}

func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	currentLine := "  " + words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) > width {
			lines = append(lines, currentLine)
			currentLine = "  " + word
		} else {
			currentLine += " " + word
		}
	}
	lines = append(lines, currentLine)
	return strings.Join(lines, "\n")
}
