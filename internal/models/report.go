package models

import "time"

// ScanReport is the final aggregate result of a scan.
type ScanReport struct {
	ProjectContext ProjectContext
	Findings       []Finding
	StartedAt      time.Time
	CompletedAt    time.Time
	Duration       time.Duration
	ScannersRun    []string
	ScannerErrors  map[string]string
}

// CountBySeverity returns a map of severity to finding count.
func (r *ScanReport) CountBySeverity() map[Severity]int {
	counts := make(map[Severity]int)
	for _, f := range r.Findings {
		counts[f.Severity]++
	}
	return counts
}

// FindingsByCategory groups findings by category.
func (r *ScanReport) FindingsByCategory() map[string][]Finding {
	grouped := make(map[string][]Finding)
	for _, f := range r.Findings {
		grouped[f.Category] = append(grouped[f.Category], f)
	}
	return grouped
}
