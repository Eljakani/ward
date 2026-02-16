package rules

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eljakani/ward/internal/config"
	"github.com/eljakani/ward/internal/models"
)

// Scanner executes YAML-defined custom rules against the project.
type Scanner struct {
	rules []config.RuleDefinition
}

// New creates a rules scanner with the given rule definitions.
func New(rules []config.RuleDefinition) *Scanner {
	return &Scanner{rules: rules}
}

func (s *Scanner) Name() string        { return "rules-scanner" }
func (s *Scanner) Description() string { return "Custom YAML rule checks" }

func (s *Scanner) Scan(_ context.Context, project models.ProjectContext, emit func(models.Finding)) ([]models.Finding, error) {
	var findings []models.Finding

	for _, rule := range s.rules {
		if !rule.Enabled {
			continue
		}

		rf := s.evaluateRule(rule, project.RootPath)
		for _, f := range rf {
			findings = append(findings, f)
			emit(f)
		}
	}

	return findings, nil
}

func (s *Scanner) evaluateRule(rule config.RuleDefinition, root string) []models.Finding {
	var findings []models.Finding

	for _, pat := range rule.Patterns {
		pf := s.evaluatePattern(rule, pat, root)
		findings = append(findings, pf...)
	}

	return findings
}

func (s *Scanner) evaluatePattern(rule config.RuleDefinition, pat config.PatternDef, root string) []models.Finding {
	switch pat.Type {
	case "file-exists":
		return s.checkFileExists(rule, pat, root)
	case "regex", "contains":
		return s.checkFileContent(rule, pat, root)
	default:
		return nil
	}
}

// checkFileExists looks for files matching the pattern glob.
func (s *Scanner) checkFileExists(rule config.RuleDefinition, pat config.PatternDef, root string) []models.Finding {
	matches, _ := filepath.Glob(filepath.Join(root, pat.Pattern))
	found := len(matches) > 0

	// Negative = finding if pattern is ABSENT
	if pat.Negative {
		if found {
			return nil
		}
		return []models.Finding{s.buildFinding(rule, pat.Pattern, 0, "")}
	}

	// Normal = finding if file exists
	if !found {
		return nil
	}

	var findings []models.Finding
	for _, m := range matches {
		rel, _ := filepath.Rel(root, m)
		findings = append(findings, s.buildFinding(rule, rel, 0, ""))
	}
	return findings
}

// checkFileContent scans target files line-by-line for pattern matches.
func (s *Scanner) checkFileContent(rule config.RuleDefinition, pat config.PatternDef, root string) []models.Finding {
	files := resolveTarget(pat.Target, root)
	if len(files) == 0 {
		return nil
	}

	var re *regexp.Regexp
	if pat.Type == "regex" {
		var err error
		re, err = regexp.Compile(pat.Pattern)
		if err != nil {
			return nil // skip invalid regex
		}
	}

	var findings []models.Finding

	for _, fpath := range files {
		matches := scanFile(fpath, pat, re)

		rel, _ := filepath.Rel(root, fpath)

		if pat.Negative {
			// Finding if pattern was NOT found in this file
			if len(matches) == 0 {
				findings = append(findings, s.buildFinding(rule, rel, 0, ""))
			}
		} else {
			for _, m := range matches {
				findings = append(findings, s.buildFinding(rule, rel, m.line, m.text))
			}
		}
	}

	return findings
}

type match struct {
	line int
	text string
}

func scanFile(path string, pat config.PatternDef, re *regexp.Regexp) []match {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var matches []match
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		var matched bool
		switch pat.Type {
		case "regex":
			matched = re.MatchString(line)
		case "contains":
			matched = strings.Contains(line, pat.Pattern)
		}

		if matched {
			matches = append(matches, match{line: lineNum, text: strings.TrimSpace(line)})
		}
	}

	return matches
}

// resolveTarget converts a target name to a list of file paths.
func resolveTarget(target, root string) []string {
	patterns := targetGlobs(target, root)

	var files []string
	seen := make(map[string]bool)

	for _, pat := range patterns {
		matches, _ := filepath.Glob(pat)
		for _, m := range matches {
			if seen[m] {
				continue
			}
			info, err := os.Stat(m)
			if err != nil || info.IsDir() {
				continue
			}
			seen[m] = true
			files = append(files, m)
		}
	}

	// For recursive targets, also walk subdirectories
	if needsWalk(target) {
		ext := targetExt(target)
		if ext != "" {
			_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					if skipDir(info.Name()) {
						return filepath.SkipDir
					}
					return nil
				}
				if matchesExt(path, ext) && !seen[path] {
					seen[path] = true
					files = append(files, path)
				}
				return nil
			})
		}
	}

	return files
}

func targetGlobs(target, root string) []string {
	switch target {
	case "php-files":
		return []string{filepath.Join(root, "*.php"), filepath.Join(root, "app", "*.php")}
	case "blade-files":
		return []string{filepath.Join(root, "resources", "views", "*.blade.php")}
	case "config-files":
		return []string{filepath.Join(root, "config", "*.php")}
	case "env-files":
		return []string{filepath.Join(root, ".env"), filepath.Join(root, ".env.*")}
	case "routes-files":
		return []string{filepath.Join(root, "routes", "*.php")}
	case "migration-files":
		return []string{filepath.Join(root, "database", "migrations", "*.php")}
	case "js-files":
		return []string{filepath.Join(root, "resources", "js", "*.js"), filepath.Join(root, "resources", "js", "*.ts")}
	default:
		// If target looks like a glob pattern, use it directly
		if strings.ContainsAny(target, "*?[") {
			return []string{filepath.Join(root, target)}
		}
		return nil
	}
}

func needsWalk(target string) bool {
	switch target {
	case "php-files", "blade-files", "js-files":
		return true
	default:
		return false
	}
}

func targetExt(target string) string {
	switch target {
	case "php-files":
		return ".php"
	case "blade-files":
		return ".blade.php"
	case "js-files":
		return ".js" // also matches .jsx below
	default:
		return ""
	}
}

func matchesExt(path, ext string) bool {
	if ext == ".php" {
		return strings.HasSuffix(path, ".php")
	}
	if ext == ".blade.php" {
		return strings.HasSuffix(path, ".blade.php")
	}
	if ext == ".js" {
		return strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".ts") ||
			strings.HasSuffix(path, ".jsx") || strings.HasSuffix(path, ".tsx")
	}
	return strings.HasSuffix(path, ext)
}

func skipDir(name string) bool {
	switch name {
	case "vendor", "node_modules", ".git", "storage", ".idea", ".vscode":
		return true
	}
	return false
}

func (s *Scanner) buildFinding(rule config.RuleDefinition, file string, line int, snippet string) models.Finding {
	return models.Finding{
		ID:          rule.ID,
		Title:       rule.Title,
		Description: rule.Description,
		Severity:    parseSeverity(rule.Severity),
		Category:    rule.Category,
		Scanner:     s.Name(),
		File:        file,
		Line:        line,
		CodeSnippet: truncate(snippet, 200),
		Remediation: rule.Remediation,
		References:  rule.References,
	}
}

func parseSeverity(s string) models.Severity {
	switch strings.ToLower(s) {
	case "critical":
		return models.SeverityCritical
	case "high":
		return models.SeverityHigh
	case "medium":
		return models.SeverityMedium
	case "low":
		return models.SeverityLow
	default:
		return models.SeverityInfo
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + fmt.Sprintf("... (%d chars)", len(s))
}
