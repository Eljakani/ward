package eventbus

import (
	"time"

	"github.com/eljakani/laravel-ward/internal/models"
)

// EventType identifies the kind of event.
type EventType int

const (
	EventScanStarted EventType = iota
	EventScanCompleted
	EventScanFailed

	EventStageStarted
	EventStageCompleted

	EventScannerRegistered
	EventScannerStarted
	EventScannerCompleted
	EventScannerFailed
	EventScannerSkipped

	EventFindingDiscovered

	EventProgressUpdate
	EventLogMessage
)

func (t EventType) String() string {
	switch t {
	case EventScanStarted:
		return "scan.started"
	case EventScanCompleted:
		return "scan.completed"
	case EventScanFailed:
		return "scan.failed"
	case EventStageStarted:
		return "stage.started"
	case EventStageCompleted:
		return "stage.completed"
	case EventScannerRegistered:
		return "scanner.registered"
	case EventScannerStarted:
		return "scanner.started"
	case EventScannerCompleted:
		return "scanner.completed"
	case EventScannerFailed:
		return "scanner.failed"
	case EventScannerSkipped:
		return "scanner.skipped"
	case EventFindingDiscovered:
		return "finding.discovered"
	case EventProgressUpdate:
		return "progress.update"
	case EventLogMessage:
		return "log.message"
	default:
		return "unknown"
	}
}

// Event is the universal event envelope.
type Event struct {
	Type      EventType
	Timestamp time.Time
	Data      interface{}
}

// NewEvent creates a timestamped event.
func NewEvent(t EventType, data interface{}) Event {
	return Event{Type: t, Timestamp: time.Now(), Data: data}
}

// --- Payload structs ---

type ScanStartedData struct {
	ProjectPath  string
	ProjectName  string
	ScannerCount int
}

type ScanCompletedData struct {
	Report *models.ScanReport
}

type ScanFailedData struct {
	Error error
}

type StageStartedData struct {
	Stage models.PipelineStage
}

type StageCompletedData struct {
	Stage models.PipelineStage
}

type ScannerRegisteredData struct {
	Name        string
	Description string
}

type ScannerStartedData struct {
	Name string
}

type ScannerCompletedData struct {
	Name         string
	FindingCount int
}

type ScannerFailedData struct {
	Name  string
	Error error
}

type ScannerSkippedData struct {
	Name   string
	Reason string
}

type FindingDiscoveredData struct {
	Finding models.Finding
}

type ProgressUpdateData struct {
	ScannerName string
	Message     string
	Percent     float64
}

type LogMessageData struct {
	Level   string
	Message string
}
