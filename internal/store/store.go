package store

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/eljakani/ward/internal/config"
	"github.com/eljakani/ward/internal/models"
)

// ScanRecord is the stored representation of a scan result.
type ScanRecord struct {
	ID          string         `json:"id"`
	ProjectName string         `json:"project_name"`
	ProjectPath string         `json:"project_path"`
	Timestamp   time.Time      `json:"timestamp"`
	Duration    string         `json:"duration"`
	FindingCount int           `json:"finding_count"`
	BySeverity  map[string]int `json:"by_severity"`
	ScannersRun []string       `json:"scanners_run"`
	FindingIDs  []string       `json:"finding_ids"` // unique finding IDs for diffing
}

// Diff represents the difference between two scans.
type Diff struct {
	NewFindings     []string `json:"new_findings"`
	ResolvedFindings []string `json:"resolved_findings"`
	TotalBefore     int      `json:"total_before"`
	TotalAfter      int      `json:"total_after"`
}

// Save stores a scan report in ~/.ward/store/.
func Save(report *models.ScanReport) (*ScanRecord, error) {
	storeDir, err := config.StoreDir()
	if err != nil {
		return nil, err
	}

	record := &ScanRecord{
		ID:           generateID(report),
		ProjectName:  report.ProjectContext.ProjectName,
		ProjectPath:  report.ProjectContext.RootPath,
		Timestamp:    report.CompletedAt,
		Duration:     report.Duration.String(),
		FindingCount: len(report.Findings),
		BySeverity:   make(map[string]int),
		ScannersRun:  report.ScannersRun,
		FindingIDs:   extractFindingKeys(report.Findings),
	}

	counts := report.CountBySeverity()
	for sev, count := range counts {
		record.BySeverity[sev.String()] = count
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshalling scan record: %w", err)
	}

	filename := fmt.Sprintf("%s_%s.json",
		record.Timestamp.Format("2006-01-02T15-04-05"),
		sanitizeName(record.ProjectName))

	path := filepath.Join(storeDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, fmt.Errorf("writing scan record: %w", err)
	}

	return record, nil
}

// ListRecords returns all stored scan records, most recent first.
func ListRecords() ([]ScanRecord, error) {
	storeDir, err := config.StoreDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(storeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var records []ScanRecord
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(storeDir, entry.Name()))
		if err != nil {
			continue
		}

		var record ScanRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}

		records = append(records, record)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	return records, nil
}

// LastRecord returns the most recent scan for a given project path.
func LastRecord(projectPath string) (*ScanRecord, error) {
	records, err := ListRecords()
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if r.ProjectPath == projectPath {
			return &r, nil
		}
	}
	return nil, nil
}

// CompareLast diffs the current scan against the most recent stored scan for the same project.
func CompareLast(report *models.ScanReport) (*Diff, error) {
	last, err := LastRecord(report.ProjectContext.RootPath)
	if err != nil || last == nil {
		return nil, err
	}

	currentKeys := extractFindingKeys(report.Findings)
	currentSet := toSet(currentKeys)
	previousSet := toSet(last.FindingIDs)

	diff := &Diff{
		TotalBefore: last.FindingCount,
		TotalAfter:  len(report.Findings),
	}

	for _, k := range currentKeys {
		if !previousSet[k] {
			diff.NewFindings = append(diff.NewFindings, k)
		}
	}

	for _, k := range last.FindingIDs {
		if !currentSet[k] {
			diff.ResolvedFindings = append(diff.ResolvedFindings, k)
		}
	}

	return diff, nil
}

func generateID(report *models.ScanReport) string {
	h := sha256.New()
	h.Write([]byte(report.ProjectContext.RootPath))
	h.Write([]byte(report.CompletedAt.String()))
	return fmt.Sprintf("%x", h.Sum(nil))[:12]
}

func extractFindingKeys(findings []models.Finding) []string {
	var keys []string
	for _, f := range findings {
		key := fmt.Sprintf("%s|%s|%d", f.ID, f.File, f.Line)
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, " ", "_")
	if len(name) > 40 {
		name = name[:40]
	}
	return name
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item] = true
	}
	return m
}
