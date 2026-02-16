package reporter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eljakani/ward/internal/models"
)

func testReport() *models.ScanReport {
	return &models.ScanReport{
		ProjectContext: models.ProjectContext{
			ProjectName:    "test/app",
			RootPath:       "/tmp/test",
			LaravelVersion: "^11.0",
			PHPVersion:     "^8.2",
		},
		Findings: []models.Finding{
			{
				ID:          "TEST-001",
				Title:       "Test Critical Finding",
				Description: "A critical test finding.",
				Severity:    models.SeverityCritical,
				Category:    "Test",
				Scanner:     "test-scanner",
				File:        "app/Test.php",
				Line:        42,
				CodeSnippet: "$password = 'hardcoded';",
				Remediation: "Fix it.",
				References:  []string{"https://example.com"},
			},
			{
				ID:       "TEST-002",
				Title:    "Test Low Finding",
				Severity: models.SeverityLow,
				Category: "Test",
				Scanner:  "test-scanner",
				File:     "config/app.php",
				Line:     10,
			},
		},
		StartedAt:   time.Now(),
		CompletedAt: time.Now(),
		Duration:    500 * time.Millisecond,
		ScannersRun: []string{"test-scanner"},
	}
}

func TestJSONReporter(t *testing.T) {
	dir := t.TempDir()
	r := NewJSONReporter(dir)

	if r.Name() != "json" {
		t.Errorf("Name() = %q, want %q", r.Name(), "json")
	}

	err := r.Generate(context.Background(), testReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "ward-report.json"))
	if err != nil {
		t.Fatalf("report file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "TEST-001") {
		t.Error("JSON report should contain finding ID")
	}
	if !strings.Contains(content, "test/app") {
		t.Error("JSON report should contain project name")
	}
}

func TestSARIFReporter(t *testing.T) {
	dir := t.TempDir()
	r := NewSARIFReporter(dir)

	if r.Name() != "sarif" {
		t.Errorf("Name() = %q, want %q", r.Name(), "sarif")
	}

	err := r.Generate(context.Background(), testReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "ward-report.sarif"))
	if err != nil {
		t.Fatalf("SARIF file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "sarif-schema-2.1.0") {
		t.Error("SARIF report should contain schema reference")
	}
	if !strings.Contains(content, "TEST-001") {
		t.Error("SARIF report should contain rule ID")
	}
	if !strings.Contains(content, "Ward") {
		t.Error("SARIF report should contain tool name")
	}
}

func TestHTMLReporter(t *testing.T) {
	dir := t.TempDir()
	r := NewHTMLReporter(dir)

	err := r.Generate(context.Background(), testReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "ward-report.html"))
	if err != nil {
		t.Fatalf("HTML file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "<!DOCTYPE html>") {
		t.Error("HTML report should contain doctype")
	}
	if !strings.Contains(content, "Test Critical Finding") {
		t.Error("HTML report should contain finding title")
	}
	if !strings.Contains(content, "Critical") {
		t.Error("HTML report should contain severity badge")
	}
}

func TestMarkdownReporter(t *testing.T) {
	dir := t.TempDir()
	r := NewMarkdownReporter(dir)

	err := r.Generate(context.Background(), testReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "ward-report.md"))
	if err != nil {
		t.Fatalf("Markdown file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# Ward Security Report") {
		t.Error("Markdown report should contain title")
	}
	if !strings.Contains(content, "TEST-001") {
		t.Error("Markdown report should contain finding ID")
	}
	if !strings.Contains(content, "Critical") {
		t.Error("Markdown report should contain severity section")
	}
}
