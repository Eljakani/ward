package reporter

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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
	if location.PhysicalLocation.ArtifactLocation.URI == "" {
		t.Error("URI should not be empty")
	}

	if run.Tool.Driver.Rules[0].Properties.Security != "9.5" {
		t.Errorf("security-severity = %q, want 9.5", run.Tool.Driver.Rules[0].Properties.Security)
	}

	var raw map[string]any
	if err := json.Unmarshal(content, &raw); err != nil {
		t.Fatalf("failed to unmarshal raw SARIF: %v", err)
	}

	runs := raw["runs"].([]any)
	rawRun := runs[0].(map[string]any)
	results := rawRun["results"].([]any)
	rawResult := results[0].(map[string]any)
	locations := rawResult["locations"].([]any)
	rawLocation := locations[0].(map[string]any)
	physicalLocation := rawLocation["physicalLocation"].(map[string]any)
	artifactLocation := physicalLocation["artifactLocation"].(map[string]any)
	if _, ok := artifactLocation["uriBaseId"]; ok {
		t.Error("artifactLocation should not include uriBaseId")
	}
}
