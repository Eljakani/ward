package reporter

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSARIFReporter_Generate_ValidatesOutput(t *testing.T) {
	dir := t.TempDir()
	r := NewSARIFReporter(dir, "1.0.0")

	report := testReport()
	err := r.Generate(context.Background(), report)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "ward-report.sarif"))
	if err != nil {
		t.Fatalf("failed to read generated report: %v", err)
	}

	var doc sarifDocument
	if err := json.Unmarshal(content, &doc); err != nil {
		t.Fatalf("generated invalid JSON: %v", err)
	}

	if doc.Version != "2.1.0" {
		t.Errorf("Version = %q, want 2.1.0", doc.Version)
	}

	if len(doc.Runs) == 0 {
		t.Fatal("Runs is empty")
	}

	run := doc.Runs[0]
	if run.Tool.Driver.Name != "Ward" {
		t.Errorf("Tool.Driver.Name = %q, want Ward", run.Tool.Driver.Name)
	}

	if len(run.Results) == 0 {
		t.Fatal("Results is empty")
	}

	result := run.Results[0]
	if len(result.Locations) == 0 {
		t.Fatal("Locations is empty")
	}

	location := result.Locations[0]
	// Verify URIBaseID is NOT present
	if location.PhysicalLocation.ArtifactLocation.URIBaseID != "" {
		t.Errorf("URIBaseID should be empty, got %q", location.PhysicalLocation.ArtifactLocation.URIBaseID)
	}

	// Verify URI is present and correct
	if location.PhysicalLocation.ArtifactLocation.URI == "" {
		t.Error("URI should not be empty")
	}
}
