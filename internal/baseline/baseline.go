package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/eljakani/ward/internal/models"
)

// Entry represents a single baselined finding.
type Entry struct {
	Fingerprint string `json:"fingerprint"`
	ID          string `json:"id"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Title       string `json:"title"`
	Severity    string `json:"severity"`
}

// Baseline is the on-disk format for suppressed findings.
type Baseline struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Entries   []Entry   `json:"entries"`

	// In-memory lookup
	fingerprints map[string]bool
}

// Load reads a baseline file from disk.
func Load(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading baseline %s: %w", path, err)
	}

	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("parsing baseline %s: %w", path, err)
	}

	b.fingerprints = make(map[string]bool, len(b.Entries))
	for _, e := range b.Entries {
		b.fingerprints[e.Fingerprint] = true
	}

	return &b, nil
}

// Save writes a baseline file to disk from the given findings.
func Save(path string, findings []models.Finding) error {
	entries := make([]Entry, 0, len(findings))
	for _, f := range findings {
		entries = append(entries, Entry{
			Fingerprint: f.Fingerprint(),
			ID:          f.ID,
			File:        f.File,
			Line:        f.Line,
			Title:       f.Title,
			Severity:    f.Severity.String(),
		})
	}

	b := Baseline{
		Version:   "1.0",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Entries:   entries,
	}

	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding baseline: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing baseline to %s: %w", path, err)
	}

	return nil
}

// IsBaselined returns true if the finding is suppressed by this baseline.
func (b *Baseline) IsBaselined(f models.Finding) bool {
	if b == nil || b.fingerprints == nil {
		return false
	}
	return b.fingerprints[f.Fingerprint()]
}

// Filter removes baselined findings from the list and returns:
// - filtered: findings NOT in the baseline (new/active issues)
// - suppressed: count of findings that were suppressed
func (b *Baseline) Filter(findings []models.Finding) (filtered []models.Finding, suppressed int) {
	if b == nil {
		return findings, 0
	}

	filtered = make([]models.Finding, 0, len(findings))
	for _, f := range findings {
		if b.IsBaselined(f) {
			suppressed++
		} else {
			filtered = append(filtered, f)
		}
	}
	return filtered, suppressed
}
