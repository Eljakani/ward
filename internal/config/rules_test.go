package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRulesFromFile(t *testing.T) {
	dir := t.TempDir()
	content := `rules:
  - id: TEST-001
    title: "Test rule"
    severity: high
    category: test
    enabled: true
    patterns:
      - type: regex
        target: php-files
        pattern: 'test_pattern'
  - id: TEST-002
    title: "Test rule 2"
    severity: low
    category: test
    enabled: false
`
	path := filepath.Join(dir, "test.yaml")
	os.WriteFile(path, []byte(content), 0644)

	rules, err := LoadRulesFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}

	if rules[0].ID != "TEST-001" {
		t.Errorf("rule[0].ID = %q, want %q", rules[0].ID, "TEST-001")
	}
	if rules[0].Severity != "high" {
		t.Errorf("rule[0].Severity = %q, want %q", rules[0].Severity, "high")
	}
	if !rules[0].Enabled {
		t.Error("rule[0] should be enabled")
	}
	if rules[1].Enabled {
		t.Error("rule[1] should be disabled")
	}
}

func TestLoadRulesFromDir(t *testing.T) {
	dir := t.TempDir()

	file1 := `rules:
  - id: A-001
    title: "Rule A"
    enabled: true
`
	file2 := `rules:
  - id: B-001
    title: "Rule B"
    enabled: true
  - id: B-002
    title: "Rule B2"
    enabled: true
`
	os.WriteFile(filepath.Join(dir, "a.yaml"), []byte(file1), 0644)
	os.WriteFile(filepath.Join(dir, "b.yml"), []byte(file2), 0644)
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not a rule"), 0644)

	rules, err := LoadRulesFromDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rules) != 3 {
		t.Errorf("expected 3 rules, got %d", len(rules))
	}
}

func TestLoadRulesFromDir_NonExistent(t *testing.T) {
	rules, err := LoadRulesFromDir("/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for non-existent dir, got %d", len(rules))
	}
}

func TestApplyOverrides(t *testing.T) {
	rules := []RuleDefinition{
		{ID: "A-001", Severity: "high", Enabled: true},
		{ID: "A-002", Severity: "medium", Enabled: true},
		{ID: "A-003", Severity: "low", Enabled: true},
	}

	falseVal := false
	rc := RulesConfig{
		Disable: []string{"A-002"},
		Override: map[string]RuleOverride{
			"A-001": {Severity: "critical"},
			"A-003": {Enabled: &falseVal},
		},
	}

	result := applyOverrides(rules, rc)

	if len(result) != 1 {
		t.Fatalf("expected 1 rule after overrides, got %d", len(result))
	}
	if result[0].ID != "A-001" {
		t.Errorf("expected A-001, got %q", result[0].ID)
	}
	if result[0].Severity != "critical" {
		t.Errorf("severity should be overridden to critical, got %q", result[0].Severity)
	}
}
