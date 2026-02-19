package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RuleDefinition is a single rule as written in a YAML file.
type RuleDefinition struct {
	ID          string       `yaml:"id"`
	Title       string       `yaml:"title"`
	Description string       `yaml:"description"`
	Severity    string       `yaml:"severity"` // critical, high, medium, low, info
	Category    string       `yaml:"category"`
	Enabled     bool         `yaml:"enabled"`
	Tags        []string     `yaml:"tags,omitempty"`
	Patterns    []PatternDef `yaml:"patterns,omitempty"`
	Remediation string       `yaml:"remediation,omitempty"`
	References  []string     `yaml:"references,omitempty"`
}

// PatternDef describes a single pattern check within a rule.
type PatternDef struct {
	Type           string `yaml:"type"`   // regex, contains, file-exists
	Target         string `yaml:"target"` // php-files, blade-files, config-files, env-files
	Pattern        string `yaml:"pattern"`
	Negative       bool   `yaml:"negative"`        // true = finding if pattern is ABSENT
	ExcludePattern string `yaml:"exclude_pattern"` // if line also matches this, skip it (reduce false positives)
}

// RuleFile is the top-level structure of a rules YAML file.
type RuleFile struct {
	Rules []RuleDefinition `yaml:"rules"`
}

// LoadRulesFromFile reads a single rules YAML file.
func LoadRulesFromFile(path string) ([]RuleDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading rules file %s: %w", path, err)
	}

	var rf RuleFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return nil, fmt.Errorf("parsing rules file %s: %w", path, err)
	}

	return rf.Rules, nil
}

// LoadRulesFromDir loads all .yaml and .yml files from a directory.
func LoadRulesFromDir(dir string) ([]RuleDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading rules dir %s: %w", dir, err)
	}

	var all []RuleDefinition
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		rules, err := LoadRulesFromFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		all = append(all, rules...)
	}

	return all, nil
}

// LoadAllRules loads rules from ~/.ward/rules plus any extra directories
// specified in the config.
func LoadAllRules(cfg *WardConfig) ([]RuleDefinition, error) {
	var all []RuleDefinition

	// Load from ~/.ward/rules
	rulesDir, err := RulesDir()
	if err != nil {
		return nil, err
	}
	rules, err := LoadRulesFromDir(rulesDir)
	if err != nil {
		return nil, err
	}
	all = append(all, rules...)

	// Load from extra directories in config
	for _, dir := range cfg.Rules.CustomDirs {
		rules, err := LoadRulesFromDir(dir)
		if err != nil {
			return nil, err
		}
		all = append(all, rules...)
	}

	// Apply overrides from config
	all = applyOverrides(all, cfg.Rules)

	return all, nil
}

func applyOverrides(rules []RuleDefinition, rc RulesConfig) []RuleDefinition {
	disabled := make(map[string]bool, len(rc.Disable))
	for _, id := range rc.Disable {
		disabled[id] = true
	}

	result := make([]RuleDefinition, 0, len(rules))
	for _, r := range rules {
		if disabled[r.ID] {
			continue
		}

		if ov, ok := rc.Override[r.ID]; ok {
			if ov.Severity != "" {
				r.Severity = ov.Severity
			}
			if ov.Enabled != nil && !*ov.Enabled {
				continue
			}
		}

		result = append(result, r)
	}

	return result
}
