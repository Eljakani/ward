package baseline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eljakani/ward/internal/models"
)

func sampleFindings() []models.Finding {
	return []models.Finding{
		{ID: "ENV-001", File: ".env", Line: 1, Title: "Debug enabled", Severity: models.SeverityHigh},
		{ID: "AUTH-001", File: "routes/web.php", Line: 10, Title: "No middleware", Severity: models.SeverityMedium},
		{ID: "SEC-001", File: "config/app.php", Line: 5, Title: "Hardcoded key", Severity: models.SeverityCritical},
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ward-baseline.json")

	findings := sampleFindings()

	// Save
	if err := Save(path, findings); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("Baseline file not created: %v", err)
	}

	// Load
	bl, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(bl.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(bl.Entries))
	}

	if bl.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", bl.Version)
	}
}

func TestFilter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ward-baseline.json")

	findings := sampleFindings()
	if err := Save(path, findings); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	bl, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// All original findings should be suppressed
	filtered, suppressed := bl.Filter(findings)
	if suppressed != 3 {
		t.Errorf("Expected 3 suppressed, got %d", suppressed)
	}
	if len(filtered) != 0 {
		t.Errorf("Expected 0 filtered findings, got %d", len(filtered))
	}

	// Add a new finding — it should pass through
	newFindings := append(findings, models.Finding{
		ID: "XSS-001", File: "resources/views/user.blade.php", Line: 42,
		Title: "Unescaped output", Severity: models.SeverityHigh,
	})
	filtered, suppressed = bl.Filter(newFindings)
	if suppressed != 3 {
		t.Errorf("Expected 3 suppressed, got %d", suppressed)
	}
	if len(filtered) != 1 {
		t.Errorf("Expected 1 new finding, got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].ID != "XSS-001" {
		t.Errorf("Expected XSS-001, got %s", filtered[0].ID)
	}
}

func TestFilterNilBaseline(t *testing.T) {
	var bl *Baseline
	findings := sampleFindings()
	filtered, suppressed := bl.Filter(findings)
	if suppressed != 0 {
		t.Errorf("Expected 0 suppressed for nil baseline, got %d", suppressed)
	}
	if len(filtered) != len(findings) {
		t.Errorf("Expected all findings to pass through nil baseline")
	}
}

func TestLoadNonExistent(t *testing.T) {
	_, err := Load("/nonexistent/path/baseline.json")
	if err == nil {
		t.Error("Expected error loading non-existent file")
	}
}
