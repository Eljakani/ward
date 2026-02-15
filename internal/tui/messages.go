package tui

import "time"

// ViewID identifies which view is currently active.
type ViewID int

const (
	ViewScan    ViewID = iota
	ViewResults
)

// switchViewMsg requests a view change.
type switchViewMsg struct {
	view ViewID
}

// tickMsg drives spinner and animation updates.
type tickMsg time.Time
