package configscan

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/eljakani/ward/internal/models"
)

func setupConfigProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	os.MkdirAll(configDir, 0755)

	for name, content := range files {
		path := filepath.Join(configDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func scanConfig(t *testing.T, dir string) []models.Finding {
	t.Helper()
	s := New()
	pc := models.ProjectContext{RootPath: dir}
	findings, err := s.Scan(context.Background(), pc, func(f models.Finding) {})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	return findings
}

func findConfigByID(findings []models.Finding, id string) *models.Finding {
	for _, f := range findings {
		if f.ID == id {
			return &f
		}
	}
	return nil
}

func TestConfigScanner_NameAndDescription(t *testing.T) {
	s := New()
	if s.Name() != "config-scanner" {
		t.Errorf("Name() = %q, want %q", s.Name(), "config-scanner")
	}
	if s.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestConfigScanner_DebugHardcoded(t *testing.T) {
	dir := setupConfigProject(t, map[string]string{
		"app.php": `<?php
return [
    'debug' => true,
    'key' => env('APP_KEY'),
];`,
	})
	findings := scanConfig(t, dir)

	f := findConfigByID(findings, "CFG-001")
	if f == nil {
		t.Fatal("Expected CFG-001 (debug hardcoded true)")
	}
	if f.Severity != models.SeverityHigh {
		t.Errorf("CFG-001 severity = %v, want High", f.Severity)
	}
}

func TestConfigScanner_DebugViaEnv(t *testing.T) {
	dir := setupConfigProject(t, map[string]string{
		"app.php": `<?php
return [
    'debug' => env('APP_DEBUG', false),
    'key' => env('APP_KEY'),
];`,
	})
	findings := scanConfig(t, dir)

	// debug via env() should NOT trigger CFG-001
	if f := findConfigByID(findings, "CFG-001"); f != nil {
		t.Error("Unexpected CFG-001 when debug uses env()")
	}
}

func TestConfigScanner_SessionInsecure(t *testing.T) {
	dir := setupConfigProject(t, map[string]string{
		"session.php": `<?php
return [
    'secure' => false,
    'http_only' => false,
    'same_site' => 'none',
];`,
	})
	findings := scanConfig(t, dir)

	// Should flag insecure session settings
	if f := findConfigByID(findings, "CFG-005"); f == nil {
		t.Error("Expected CFG-005 for session secure=false")
	}
	if f := findConfigByID(findings, "CFG-006"); f == nil {
		t.Error("Expected CFG-006 for session http_only=false")
	}
}

func TestConfigScanner_CORSWildcard(t *testing.T) {
	dir := setupConfigProject(t, map[string]string{
		"cors.php": `<?php
return [
    'allowed_origins' => ['*'],
    'supports_credentials' => true,
];`,
	})
	findings := scanConfig(t, dir)

	if f := findConfigByID(findings, "CFG-009"); f == nil {
		t.Error("Expected CFG-009 for CORS wildcard origin")
	}
}

func TestConfigScanner_NoConfigDir(t *testing.T) {
	dir := t.TempDir()
	// No config/ directory at all — should not crash
	findings := scanConfig(t, dir)
	if len(findings) != 0 {
		t.Errorf("Expected 0 findings with no config dir, got %d", len(findings))
	}
}

func TestConfigScanner_EmitCallback(t *testing.T) {
	dir := setupConfigProject(t, map[string]string{
		"app.php": `<?php return ['debug' => true];`,
	})

	s := New()
	pc := models.ProjectContext{RootPath: dir}

	var emitted []models.Finding
	s.Scan(context.Background(), pc, func(f models.Finding) {
		emitted = append(emitted, f)
	})

	if len(emitted) == 0 {
		t.Error("Expected emit callback to be called")
	}
}
