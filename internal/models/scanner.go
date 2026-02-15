package models

import "context"

// Scanner is the interface that all security scanners implement.
type Scanner interface {
	Name() string
	Description() string
	Scan(ctx context.Context, project ProjectContext, emit func(Finding)) ([]Finding, error)
}

// ScannerStatus represents the current state of a scanner.
type ScannerStatus int

const (
	ScannerPending ScannerStatus = iota
	ScannerRunning
	ScannerDone
	ScannerError
	ScannerSkipped
)

func (s ScannerStatus) String() string {
	switch s {
	case ScannerPending:
		return "Pending"
	case ScannerRunning:
		return "Running"
	case ScannerDone:
		return "Done"
	case ScannerError:
		return "Error"
	case ScannerSkipped:
		return "Skipped"
	default:
		return "Unknown"
	}
}

// ScannerInfo holds runtime information about a scanner.
type ScannerInfo struct {
	Name         string
	Description  string
	Status       ScannerStatus
	FindingCount int
	Error        error
}
