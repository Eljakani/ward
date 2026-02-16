package rules

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/eljakani/ward/internal/config"
	"github.com/eljakani/ward/internal/models"
)

func TestRulesScanner_RegexMatch(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "app"), 0755)
	os.WriteFile(filepath.Join(dir, "app", "Service.php"), []byte(`<?php
$password = "hardcoded123";
`), 0644)

	rules := []config.RuleDefinition{
		{
			ID:       "TEST-001",
			Title:    "Hardcoded password",
			Severity: "high",
			Category: "Secrets",
			Enabled:  true,
			Patterns: []config.PatternDef{
				{Type: "regex", Target: "php-files", Pattern: `\$password\s*=\s*"[a-zA-Z0-9]+"`},
			},
		},
	}

	s := New(rules)
	pc := models.ProjectContext{RootPath: dir}
	findings, err := s.Scan(context.Background(), pc, func(f models.Finding) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].ID != "TEST-001" {
		t.Errorf("finding ID = %q, want %q", findings[0].ID, "TEST-001")
	}
	if findings[0].Line != 2 {
		t.Errorf("finding line = %d, want 2", findings[0].Line)
	}
}

func TestRulesScanner_ContainsMatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("SECRET_KEY=abc123\n"), 0644)

	rules := []config.RuleDefinition{
		{
			ID:       "TEST-002",
			Title:    "Secret in env",
			Severity: "medium",
			Enabled:  true,
			Patterns: []config.PatternDef{
				{Type: "contains", Target: "env-files", Pattern: "SECRET_KEY"},
			},
		},
	}

	s := New(rules)
	pc := models.ProjectContext{RootPath: dir}
	findings, _ := s.Scan(context.Background(), pc, func(f models.Finding) {})

	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
}

func TestRulesScanner_NegativePattern(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "app"), 0755)
	// File WITHOUT the expected pattern
	os.WriteFile(filepath.Join(dir, "app", "Model.php"), []byte(`<?php
class User extends Model {
}
`), 0644)

	rules := []config.RuleDefinition{
		{
			ID:       "TEST-003",
			Title:    "Missing fillable",
			Severity: "medium",
			Enabled:  true,
			Patterns: []config.PatternDef{
				{Type: "contains", Target: "php-files", Pattern: "$fillable", Negative: true},
			},
		},
	}

	s := New(rules)
	pc := models.ProjectContext{RootPath: dir}
	findings, _ := s.Scan(context.Background(), pc, func(f models.Finding) {})

	if len(findings) != 1 {
		t.Errorf("expected 1 finding for negative pattern, got %d", len(findings))
	}
}

func TestRulesScanner_DisabledRule(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("SECRET=abc\n"), 0644)

	rules := []config.RuleDefinition{
		{
			ID:      "TEST-004",
			Enabled: false,
			Patterns: []config.PatternDef{
				{Type: "contains", Target: "env-files", Pattern: "SECRET"},
			},
		},
	}

	s := New(rules)
	pc := models.ProjectContext{RootPath: dir}
	findings, _ := s.Scan(context.Background(), pc, func(f models.Finding) {})

	if len(findings) != 0 {
		t.Errorf("disabled rule should produce no findings, got %d", len(findings))
	}
}

func TestRulesScanner_FileExists(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env.production"), []byte(""), 0644)

	rules := []config.RuleDefinition{
		{
			ID:       "TEST-005",
			Title:    "Production env file exists",
			Severity: "low",
			Enabled:  true,
			Patterns: []config.PatternDef{
				{Type: "file-exists", Pattern: ".env.production"},
			},
		},
	}

	s := New(rules)
	pc := models.ProjectContext{RootPath: dir}
	findings, _ := s.Scan(context.Background(), pc, func(f models.Finding) {})

	if len(findings) != 1 {
		t.Errorf("expected 1 finding for file-exists, got %d", len(findings))
	}
}

func TestSkipDir(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"vendor", true},
		{"node_modules", true},
		{".git", true},
		{"storage", true},
		{"app", false},
		{"src", false},
	}

	for _, tt := range tests {
		if got := skipDir(tt.name); got != tt.want {
			t.Errorf("skipDir(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
