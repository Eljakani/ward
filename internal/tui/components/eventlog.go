package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eljakani/ward/internal/eventbus"
	"github.com/eljakani/ward/internal/tui/theme"
)

// EventLog is a scrollable event viewport.
type EventLog struct {
	viewport   viewport.Model
	events     []eventbus.Event
	theme      *theme.Theme
	autoScroll bool
	width      int
	height     int
}

// NewEventLog creates a new event log viewer.
func NewEventLog(t *theme.Theme) *EventLog {
	vp := viewport.New(80, 10)
	return &EventLog{
		viewport:   vp,
		theme:      t,
		autoScroll: true,
	}
}

// SetSize updates the viewport dimensions.
func (e *EventLog) SetSize(w, h int) {
	e.width = w
	e.height = h
	e.viewport.Width = w - 4
	e.viewport.Height = h - 2
}

// SetEvents replaces the event list and rebuilds the viewport content.
func (e *EventLog) SetEvents(events []eventbus.Event) {
	e.events = events
	e.rebuildContent()
}

func (e *EventLog) rebuildContent() {
	var lines []string
	for _, ev := range e.events {
		timestamp := e.theme.Muted.Render(ev.Timestamp.Format("15:04:05"))
		icon := eventTypeIcon(ev.Type, e.theme)
		msg := formatEventMessage(ev)
		line := fmt.Sprintf(" %s %s %s", timestamp, icon, msg)
		lines = append(lines, line)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	e.viewport.SetContent(content)
	if e.autoScroll {
		e.viewport.GotoBottom()
	}
}

// HandleKey processes scrolling keys.
func (e *EventLog) HandleKey(msg tea.KeyMsg) {
	e.viewport, _ = e.viewport.Update(msg)
}

// View renders the event log.
func (e *EventLog) View() string {
	title := e.theme.Subtitle.Render("  Event Log")

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(e.theme.Colors.Border).
		Width(e.width).
		Height(e.height)

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", e.viewport.View())
	return border.Render(content)
}

func eventTypeIcon(t eventbus.EventType, th *theme.Theme) string {
	switch t {
	case eventbus.EventScanStarted:
		return th.AccentStyle.Render("▶")
	case eventbus.EventScanCompleted:
		return th.SuccessStyle.Render("✓")
	case eventbus.EventScanFailed:
		return th.ErrorStyle.Render("✗")
	case eventbus.EventStageStarted:
		return th.ActiveStage.Render("●")
	case eventbus.EventStageCompleted:
		return th.CompletedStage.Render("✓")
	case eventbus.EventScannerRegistered:
		return th.Muted.Render("+")
	case eventbus.EventScannerStarted:
		return th.ActiveStage.Render("▸")
	case eventbus.EventScannerCompleted:
		return th.SuccessStyle.Render("✓")
	case eventbus.EventScannerFailed:
		return th.ErrorStyle.Render("✗")
	case eventbus.EventFindingDiscovered:
		return th.WarningStyle.Render("▲")
	case eventbus.EventProgressUpdate:
		return th.Muted.Render("…")
	case eventbus.EventLogMessage:
		return th.Muted.Render("·")
	default:
		return " "
	}
}

func formatEventMessage(ev eventbus.Event) string {
	switch data := ev.Data.(type) {
	case eventbus.ScanStartedData:
		return fmt.Sprintf("Scan started: %s (%d scanners)", data.ProjectName, data.ScannerCount)
	case eventbus.ScanCompletedData:
		return fmt.Sprintf("Scan completed: %d findings", len(data.Report.Findings))
	case eventbus.ScanFailedData:
		return fmt.Sprintf("Scan failed: %v", data.Error)
	case eventbus.StageStartedData:
		return fmt.Sprintf("Stage started: %s", data.Stage)
	case eventbus.StageCompletedData:
		return fmt.Sprintf("Stage completed: %s", data.Stage)
	case eventbus.ScannerRegisteredData:
		return fmt.Sprintf("Registered: %s", data.Name)
	case eventbus.ScannerStartedData:
		return fmt.Sprintf("Scanner started: %s", data.Name)
	case eventbus.ScannerCompletedData:
		return fmt.Sprintf("Scanner completed: %s (%d findings)", data.Name, data.FindingCount)
	case eventbus.ScannerFailedData:
		return fmt.Sprintf("Scanner failed: %s - %v", data.Name, data.Error)
	case eventbus.FindingDiscoveredData:
		return fmt.Sprintf("Finding: [%s] %s", data.Finding.Severity, data.Finding.Title)
	case eventbus.ProgressUpdateData:
		return fmt.Sprintf("%s: %s", data.ScannerName, data.Message)
	case eventbus.LogMessageData:
		return fmt.Sprintf("[%s] %s", data.Level, data.Message)
	default:
		return ev.Type.String()
	}
}
