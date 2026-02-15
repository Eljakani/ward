package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/laravel-ward/internal/eventbus"
	"github.com/eljakani/laravel-ward/internal/models"
	"github.com/eljakani/laravel-ward/internal/tui/components"
	"github.com/eljakani/laravel-ward/internal/tui/theme"
)

// ScanView renders the scanning-in-progress screen.
type ScanView struct {
	theme *theme.Theme

	// Sub-components
	currentStage models.PipelineStage
	scannerPanel *components.ScannerPanel
	eventLog     *components.EventLog
	scanComplete bool

	// Data
	severityCounts map[models.Severity]int

	// Layout
	width  int
	height int
}

// NewScanView creates a new scan view.
func NewScanView(t *theme.Theme) *ScanView {
	return &ScanView{
		theme:          t,
		scannerPanel:   components.NewScannerPanel(t),
		eventLog:       components.NewEventLog(t),
		severityCounts: make(map[models.Severity]int),
	}
}

// SetSize updates dimensions and propagates to sub-components.
func (v *ScanView) SetSize(w, h int) {
	v.width = w
	v.height = h

	// Layout allocation:
	//   stage progress:  1 line
	//   separator:       1 line
	//   live stats:      1 line
	//   separator:       1 line
	//   body:            remaining
	bodyH := h - 4
	if bodyH < 4 {
		bodyH = 4
	}

	scannerW := int(float64(w) * 0.30)
	logW := w - scannerW - 3 // account for borders

	v.scannerPanel.SetSize(scannerW, bodyH)
	v.eventLog.SetSize(logW, bodyH)
}

// UpdateScanners updates the scanner info list.
func (v *ScanView) UpdateScanners(s []models.ScannerInfo) {
	v.scannerPanel.SetScanners(s)
}

// UpdateStage sets the current pipeline stage.
func (v *ScanView) UpdateStage(stage models.PipelineStage) {
	v.currentStage = stage
}

// UpdateStats sets the severity counts.
func (v *ScanView) UpdateStats(counts map[models.Severity]int) {
	v.severityCounts = counts
}

// UpdateEventLog sets the event log.
func (v *ScanView) UpdateEventLog(events []eventbus.Event) {
	v.eventLog.SetEvents(events)
}

// SetScanComplete marks the scan as complete for stage rendering.
func (v *ScanView) SetScanComplete(complete bool) {
	v.scanComplete = complete
}

// Tick forwards tick to scanner panel for spinner animation.
func (v *ScanView) Tick(msg tea.Msg) tea.Cmd {
	return v.scannerPanel.Tick(msg)
}

// HandleKey delegates scrolling to event log.
func (v *ScanView) HandleKey(msg tea.KeyMsg) {
	v.eventLog.HandleKey(msg)
}

// View renders the complete scan view.
func (v *ScanView) View(width, height int) string {
	if width == 0 || height == 0 {
		return ""
	}

	// 1. Pipeline stage progress
	stageRow := components.RenderStageProgress(v.currentStage, v.scanComplete, v.theme, width)

	// 2. Separator
	sep := components.RenderSeparator(v.theme, width)

	// 3. Live stats
	statsRow := components.RenderLiveStats(v.severityCounts, v.theme, width)

	// 4. Body: scanner panel (left) + event log (right)
	scannerView := v.scannerPanel.View()
	logView := v.eventLog.View()
	body := lipgloss.JoinHorizontal(lipgloss.Top, scannerView, " ", logView)

	return lipgloss.JoinVertical(lipgloss.Left,
		stageRow,
		sep,
		statsRow,
		sep,
		body,
	)
}
