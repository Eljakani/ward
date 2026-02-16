package env

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/eljakani/ward/internal/models"
)

func TestEnvScanner_DebugEnabled(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("APP_KEY=base64:abcdefghijklmnopqrstuvwxyz123456\nAPP_DEBUG=true\nAPP_ENV=production\n"), 0644)

	s := New()
	pc := models.ProjectContext{RootPath: dir}
	findings, err := s.Scan(context.Background(), pc, func(f models.Finding) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, f := range findings {
		if f.ID == "ENV-002" {
			found = true
			if f.Severity != models.SeverityHigh {
				t.Errorf("ENV-002 severity = %v, want High", f.Severity)
			}
		}
	}
	if !found {
		t.Error("expected ENV-002 finding for APP_DEBUG=true")
	}
}

func TestEnvScanner_EmptyAppKey(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("APP_KEY=\nAPP_ENV=production\n"), 0644)

	s := New()
	pc := models.ProjectContext{RootPath: dir}
	findings, _ := s.Scan(context.Background(), pc, func(f models.Finding) {})

	found := false
	for _, f := range findings {
		if f.ID == "ENV-003" && f.Severity == models.SeverityCritical {
			found = true
		}
	}
	if !found {
		t.Error("expected ENV-003 critical finding for empty APP_KEY")
	}
}

func TestEnvScanner_MissingEnvFile(t *testing.T) {
	dir := t.TempDir()

	s := New()
	pc := models.ProjectContext{RootPath: dir}
	findings, _ := s.Scan(context.Background(), pc, func(f models.Finding) {})

	if len(findings) != 1 || findings[0].ID != "ENV-001" {
		t.Errorf("expected single ENV-001 finding, got %d findings", len(findings))
	}
}

func TestEnvScanner_LocalEnv(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("APP_KEY=base64:abcdefghijklmnopqrstuvwxyz123456\nAPP_ENV=local\n"), 0644)

	s := New()
	pc := models.ProjectContext{RootPath: dir}
	findings, _ := s.Scan(context.Background(), pc, func(f models.Finding) {})

	found := false
	for _, f := range findings {
		if f.ID == "ENV-005" {
			found = true
		}
	}
	if !found {
		t.Error("expected ENV-005 finding for APP_ENV=local")
	}
}

func TestEnvScanner_ProductionEnv(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("APP_KEY=base64:abcdefghijklmnopqrstuvwxyz123456\nAPP_ENV=production\nAPP_DEBUG=false\n"), 0644)

	s := New()
	pc := models.ProjectContext{RootPath: dir}
	findings, _ := s.Scan(context.Background(), pc, func(f models.Finding) {})

	// Should not trigger ENV-002 or ENV-005
	for _, f := range findings {
		if f.ID == "ENV-002" {
			t.Error("unexpected ENV-002 for APP_DEBUG=false")
		}
		if f.ID == "ENV-005" {
			t.Error("unexpected ENV-005 for APP_ENV=production")
		}
	}
}

func TestEnvScanner_EnvExampleCredentials(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("APP_KEY=base64:abcdefghijklmnopqrstuvwxyz123456\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".env.example"), []byte("DB_PASSWORD=supersecretprod123\n"), 0644)

	s := New()
	pc := models.ProjectContext{RootPath: dir}
	findings, _ := s.Scan(context.Background(), pc, func(f models.Finding) {})

	found := false
	for _, f := range findings {
		if f.ID == "ENV-008" {
			found = true
		}
	}
	if !found {
		t.Error("expected ENV-008 for real credential in .env.example")
	}
}

func TestEnvScanner_EmitCallback(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("APP_KEY=\nAPP_DEBUG=true\n"), 0644)

	s := New()
	pc := models.ProjectContext{RootPath: dir}

	var emitted []models.Finding
	s.Scan(context.Background(), pc, func(f models.Finding) {
		emitted = append(emitted, f)
	})

	if len(emitted) == 0 {
		t.Error("expected emit callback to be called")
	}
}
