package env

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eljakani/ward/internal/models"
)

// Scanner checks .env files for security issues.
type Scanner struct{}

func New() *Scanner       { return &Scanner{} }
func (s *Scanner) Name() string        { return "env-scanner" }
func (s *Scanner) Description() string { return "Environment file security checks" }

func (s *Scanner) Scan(_ context.Context, project models.ProjectContext, emit func(models.Finding)) ([]models.Finding, error) {
	var findings []models.Finding

	envPath := filepath.Join(project.RootPath, ".env")
	envVars, err := readEnvFile(envPath)
	if err != nil {
		f := models.Finding{
			ID:          "ENV-001",
			Title:       "No .env file found",
			Description: "The project has no .env file. While this may be intentional in containerized deployments, ensure environment configuration is provided through another mechanism.",
			Severity:    models.SeverityInfo,
			Category:    "Configuration",
			Scanner:     s.Name(),
			File:        ".env",
			Remediation: "Copy .env.example to .env and configure your environment variables.",
		}
		findings = append(findings, f)
		emit(f)
		return findings, nil
	}

	// APP_DEBUG=true
	if val, ok := envVars["APP_DEBUG"]; ok && strings.EqualFold(val, "true") {
		f := models.Finding{
			ID:          "ENV-002",
			Title:       "APP_DEBUG is enabled",
			Description: "APP_DEBUG is set to true. In production, this exposes detailed error messages including stack traces, database queries, and environment variables to end users.",
			Severity:    models.SeverityHigh,
			Category:    "Configuration",
			Scanner:     s.Name(),
			File:        ".env",
			Line:        findLine(envPath, "APP_DEBUG"),
			CodeSnippet: fmt.Sprintf("APP_DEBUG=%s", val),
			Remediation: "Set APP_DEBUG=false in your production .env file. Use Laravel's logging system for error tracking instead.",
			References:  []string{"https://owasp.org/Top10/A05_2021-Security_Misconfiguration/"},
		}
		findings = append(findings, f)
		emit(f)
	}

	// APP_KEY empty or default
	if val, ok := envVars["APP_KEY"]; ok {
		if val == "" {
			f := models.Finding{
				ID:          "ENV-003",
				Title:       "APP_KEY is empty",
				Description: "The application encryption key is not set. Laravel uses this key to encrypt cookies, sessions, and other sensitive data. Without it, encrypted data is insecure.",
				Severity:    models.SeverityCritical,
				Category:    "Cryptography",
				Scanner:     s.Name(),
				File:        ".env",
				Line:        findLine(envPath, "APP_KEY"),
				CodeSnippet: "APP_KEY=",
				Remediation: "Generate a new application key: php artisan key:generate",
				References:  []string{"https://cwe.mitre.org/data/definitions/321.html"},
			}
			findings = append(findings, f)
			emit(f)
		} else if isWeakKey(val) {
			f := models.Finding{
				ID:          "ENV-004",
				Title:       "APP_KEY appears to be a default or weak key",
				Description: "The application key looks like a default or placeholder value. This makes all encrypted data (sessions, cookies, passwords) predictable and breakable.",
				Severity:    models.SeverityCritical,
				Category:    "Cryptography",
				Scanner:     s.Name(),
				File:        ".env",
				Line:        findLine(envPath, "APP_KEY"),
				CodeSnippet: fmt.Sprintf("APP_KEY=%s", val),
				Remediation: "Generate a new application key: php artisan key:generate",
				References:  []string{"https://cwe.mitre.org/data/definitions/321.html"},
			}
			findings = append(findings, f)
			emit(f)
		}
	} else {
		f := models.Finding{
			ID:          "ENV-003",
			Title:       "APP_KEY is not defined",
			Description: "No APP_KEY variable found in .env. Laravel requires this key for all encryption operations.",
			Severity:    models.SeverityCritical,
			Category:    "Cryptography",
			Scanner:     s.Name(),
			File:        ".env",
			Remediation: "Add APP_KEY to .env and generate a key: php artisan key:generate",
			References:  []string{"https://cwe.mitre.org/data/definitions/321.html"},
		}
		findings = append(findings, f)
		emit(f)
	}

	// APP_ENV not production
	if val, ok := envVars["APP_ENV"]; ok {
		lower := strings.ToLower(val)
		if lower == "local" || lower == "development" || lower == "dev" {
			f := models.Finding{
				ID:          "ENV-005",
				Title:       fmt.Sprintf("APP_ENV is set to '%s'", val),
				Description: "The application environment suggests a non-production configuration. If this is a production server, this may cause debug features to be enabled and performance optimizations to be skipped.",
				Severity:    models.SeverityMedium,
				Category:    "Configuration",
				Scanner:     s.Name(),
				File:        ".env",
				Line:        findLine(envPath, "APP_ENV"),
				CodeSnippet: fmt.Sprintf("APP_ENV=%s", val),
				Remediation: "Set APP_ENV=production on production servers.",
			}
			findings = append(findings, f)
			emit(f)
		}
	}

	// Empty DB_PASSWORD
	if val, ok := envVars["DB_PASSWORD"]; ok && val == "" {
		f := models.Finding{
			ID:          "ENV-006",
			Title:       "Database password is empty",
			Description: "DB_PASSWORD is set to an empty string. While this may be valid for local development with trust authentication, it's a security risk if this configuration reaches production.",
			Severity:    models.SeverityLow,
			Category:    "Configuration",
			Scanner:     s.Name(),
			File:        ".env",
			Line:        findLine(envPath, "DB_PASSWORD"),
			CodeSnippet: "DB_PASSWORD=",
			Remediation: "Set a strong database password for non-local environments.",
		}
		findings = append(findings, f)
		emit(f)
	}

	// SESSION_DRIVER=file in production-looking env
	if val, ok := envVars["SESSION_DRIVER"]; ok && val == "file" {
		if env, ok := envVars["APP_ENV"]; ok && strings.EqualFold(env, "production") {
			f := models.Finding{
				ID:          "ENV-007",
				Title:       "File-based sessions in production",
				Description: "SESSION_DRIVER is set to 'file' in what appears to be a production environment. File sessions don't scale across multiple servers and are slower than alternatives.",
				Severity:    models.SeverityLow,
				Category:    "Configuration",
				Scanner:     s.Name(),
				File:        ".env",
				Line:        findLine(envPath, "SESSION_DRIVER"),
				CodeSnippet: fmt.Sprintf("SESSION_DRIVER=%s", val),
				Remediation: "Use redis, memcached, or database session drivers for production: SESSION_DRIVER=redis",
			}
			findings = append(findings, f)
			emit(f)
		}
	}

	// Check .env.example for real-looking credentials
	exampleFindings := s.checkEnvExample(project.RootPath)
	for _, f := range exampleFindings {
		findings = append(findings, f)
		emit(f)
	}

	return findings, nil
}

func (s *Scanner) checkEnvExample(root string) []models.Finding {
	exPath := filepath.Join(root, ".env.example")
	vars, err := readEnvFile(exPath)
	if err != nil {
		return nil
	}

	var findings []models.Finding
	sensitiveKeys := []string{"DB_PASSWORD", "MAIL_PASSWORD", "AWS_SECRET_ACCESS_KEY", "REDIS_PASSWORD", "PUSHER_APP_SECRET"}

	for _, key := range sensitiveKeys {
		val, ok := vars[key]
		if !ok || val == "" {
			continue
		}
		// Skip obvious placeholders
		lower := strings.ToLower(val)
		if lower == "null" || lower == "secret" || lower == "password" || lower == "your_password_here" || lower == "changeme" {
			continue
		}
		// If it's longer than 6 chars and not a placeholder, flag it
		if len(val) > 6 {
			f := models.Finding{
				ID:          "ENV-008",
				Title:       fmt.Sprintf("Potential real credential in .env.example: %s", key),
				Description: fmt.Sprintf("The .env.example file contains a value for %s that doesn't look like a placeholder. This file is typically committed to version control and should only contain example/placeholder values.", key),
				Severity:    models.SeverityMedium,
				Category:    "Secrets",
				Scanner:     s.Name(),
				File:        ".env.example",
				Line:        findLine(exPath, key),
				CodeSnippet: fmt.Sprintf("%s=%s", key, maskValue(val)),
				Remediation: fmt.Sprintf("Replace the value of %s in .env.example with a placeholder like 'your_%s_here'.", key, strings.ToLower(key)),
				References:  []string{"https://cwe.mitre.org/data/definitions/798.html"},
			}
			findings = append(findings, f)
		}
	}

	return findings
}

func readEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)
		vars[key] = val
	}
	return vars, scanner.Err()
}

func findLine(path, key string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	lineNum := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineNum++
		if strings.HasPrefix(strings.TrimSpace(scanner.Text()), key+"=") ||
			strings.HasPrefix(strings.TrimSpace(scanner.Text()), key+" =") {
			return lineNum
		}
	}
	return 0
}

func isWeakKey(val string) bool {
	lower := strings.ToLower(val)
	// All zeros/A's base64 key
	if strings.HasPrefix(lower, "base64:aaaaaaa") {
		return true
	}
	// Common test keys
	if lower == "somerandostrng" || lower == "somerandomstring" {
		return true
	}
	// Too short after base64: prefix
	if strings.HasPrefix(val, "base64:") && len(val) < 20 {
		return true
	}
	return false
}

func maskValue(val string) string {
	if len(val) <= 4 {
		return "****"
	}
	return val[:2] + strings.Repeat("*", len(val)-4) + val[len(val)-2:]
}
