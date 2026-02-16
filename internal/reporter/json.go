package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eljakani/ward/internal/models"
)

// jsonFinding is the JSON-serializable representation of a finding.
type jsonFinding struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Category    string   `json:"category"`
	Scanner     string   `json:"scanner"`
	File        string   `json:"file,omitempty"`
	Line        int      `json:"line,omitempty"`
	CodeSnippet string   `json:"code_snippet,omitempty"`
	Remediation string   `json:"remediation,omitempty"`
	References  []string `json:"references,omitempty"`
}

// jsonReport is the top-level JSON output structure.
type jsonReport struct {
	Project  jsonProject          `json:"project"`
	Summary  jsonSummary          `json:"summary"`
	Findings []jsonFinding        `json:"findings"`
}

type jsonProject struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	LaravelVersion string `json:"laravel_version,omitempty"`
	PHPVersion     string `json:"php_version,omitempty"`
}

type jsonSummary struct {
	TotalFindings int            `json:"total_findings"`
	BySeverity    map[string]int `json:"by_severity"`
	Duration      string         `json:"duration"`
	ScannersRun   []string       `json:"scanners_run"`
}

// JSONReporter writes a ward-report.json file.
type JSONReporter struct {
	OutputDir string
}

func NewJSONReporter(outputDir string) *JSONReporter {
	if outputDir == "" {
		outputDir = "."
	}
	return &JSONReporter{OutputDir: outputDir}
}

func (r *JSONReporter) Name() string   { return "json" }
func (r *JSONReporter) Format() string { return "json" }

func (r *JSONReporter) Generate(_ context.Context, report *models.ScanReport) error {
	jr := jsonReport{
		Project: jsonProject{
			Name:           report.ProjectContext.ProjectName,
			Path:           report.ProjectContext.RootPath,
			LaravelVersion: report.ProjectContext.LaravelVersion,
			PHPVersion:     report.ProjectContext.PHPVersion,
		},
		Summary: jsonSummary{
			TotalFindings: len(report.Findings),
			BySeverity:    make(map[string]int),
			Duration:      report.Duration.String(),
			ScannersRun:   report.ScannersRun,
		},
	}

	counts := report.CountBySeverity()
	for sev, count := range counts {
		jr.Summary.BySeverity[sev.String()] = count
	}

	jr.Findings = make([]jsonFinding, 0, len(report.Findings))
	for _, f := range report.Findings {
		jr.Findings = append(jr.Findings, jsonFinding{
			ID:          f.ID,
			Title:       f.Title,
			Description: f.Description,
			Severity:    f.Severity.String(),
			Category:    f.Category,
			Scanner:     f.Scanner,
			File:        f.File,
			Line:        f.Line,
			CodeSnippet: f.CodeSnippet,
			Remediation: f.Remediation,
			References:  f.References,
		})
	}

	data, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling report: %w", err)
	}

	outPath := filepath.Join(r.OutputDir, "ward-report.json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("writing report to %s: %w", outPath, err)
	}

	return nil
}
