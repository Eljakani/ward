package configscan

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eljakani/ward/internal/models"
)

// Scanner checks Laravel config/*.php files for security misconfigurations.
type Scanner struct{}

func New() *Scanner { return &Scanner{} }

func (s *Scanner) Name() string        { return "config-scanner" }
func (s *Scanner) Description() string { return "Laravel configuration security checks" }

func (s *Scanner) Scan(_ context.Context, project models.ProjectContext, emit func(models.Finding)) ([]models.Finding, error) {
	configDir := filepath.Join(project.RootPath, "config")
	if _, err := os.Stat(configDir); err != nil {
		return nil, nil
	}

	var findings []models.Finding

	checks := []struct {
		file  string
		check func(string) []models.Finding
	}{
		{"app.php", s.checkApp},
		{"auth.php", s.checkAuth},
		{"session.php", s.checkSession},
		{"mail.php", s.checkMail},
		{"cors.php", s.checkCORS},
		{"database.php", s.checkDatabase},
		{"broadcasting.php", s.checkBroadcasting},
		{"logging.php", s.checkLogging},
	}

	for _, c := range checks {
		path := filepath.Join(configDir, c.file)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		ff := c.check(path)
		for _, f := range ff {
			findings = append(findings, f)
			emit(f)
		}
	}

	return findings, nil
}

func (s *Scanner) checkApp(path string) []models.Finding {
	var findings []models.Finding
	content := readFile(path)
	lines := toLines(content)

	// Debug mode hardcoded to true
	if line, n := findPattern(lines, `'debug'\s*=>\s*true`); n > 0 {
		findings = append(findings, models.Finding{
			ID:          "CFG-001",
			Title:       "Debug mode hardcoded to true in app.php",
			Description: "config/app.php has 'debug' => true instead of reading from env(). This means debug mode is always on, even in production.",
			Severity:    models.SeverityHigh,
			Category:    "Configuration",
			Scanner:     s.Name(),
			File:        "config/app.php",
			Line:        n,
			CodeSnippet: line,
			Remediation: "Use: 'debug' => env('APP_DEBUG', false),",
			References:  []string{"https://owasp.org/Top10/A05_2021-Security_Misconfiguration/"},
		})
	}

	// Cipher not AES-256-CBC
	if line, n := findPattern(lines, `'cipher'\s*=>\s*'(?i)((?!aes-256-cbc).+)'`); n > 0 {
		findings = append(findings, models.Finding{
			ID:          "CFG-002",
			Title:       "Non-standard encryption cipher configured",
			Description: "The application encryption cipher is not the recommended AES-256-CBC. Using a weaker cipher reduces the security of encrypted data.",
			Severity:    models.SeverityMedium,
			Category:    "Cryptography",
			Scanner:     s.Name(),
			File:        "config/app.php",
			Line:        n,
			CodeSnippet: line,
			Remediation: "Use: 'cipher' => 'AES-256-CBC',",
		})
	}

	return findings
}

func (s *Scanner) checkAuth(path string) []models.Finding {
	var findings []models.Finding
	content := readFile(path)
	lines := toLines(content)

	// Password reset expiry too long (> 120 minutes)
	if line, n := findPattern(lines, `'expire'\s*=>\s*(\d{3,})`); n > 0 {
		findings = append(findings, models.Finding{
			ID:          "CFG-003",
			Title:       "Password reset token expiry is very long",
			Description: "The password reset token expires after a very long period. Long-lived reset tokens increase the window for token theft and reuse.",
			Severity:    models.SeverityLow,
			Category:    "Authentication",
			Scanner:     s.Name(),
			File:        "config/auth.php",
			Line:        n,
			CodeSnippet: line,
			Remediation: "Set a reasonable expiry: 'expire' => 60, (60 minutes)",
		})
	}

	return findings
}

func (s *Scanner) checkSession(path string) []models.Finding {
	var findings []models.Finding
	content := readFile(path)
	lines := toLines(content)

	// Session cookie not httponly
	if line, n := findPattern(lines, `'http_only'\s*=>\s*false`); n > 0 {
		findings = append(findings, models.Finding{
			ID:          "CFG-004",
			Title:       "Session cookie missing HttpOnly flag",
			Description: "The session cookie HttpOnly flag is set to false. This allows JavaScript to access the session cookie, enabling theft through XSS attacks.",
			Severity:    models.SeverityHigh,
			Category:    "Configuration",
			Scanner:     s.Name(),
			File:        "config/session.php",
			Line:        n,
			CodeSnippet: line,
			Remediation: "Set: 'http_only' => true,",
			References:  []string{"https://cwe.mitre.org/data/definitions/1004.html"},
		})
	}

	// Session cookie not secure
	if line, n := findPattern(lines, `'secure'\s*=>\s*false`); n > 0 {
		findings = append(findings, models.Finding{
			ID:          "CFG-005",
			Title:       "Session cookie missing Secure flag",
			Description: "The session cookie Secure flag is false. The cookie will be sent over plain HTTP, allowing session hijacking via network sniffing.",
			Severity:    models.SeverityMedium,
			Category:    "Configuration",
			Scanner:     s.Name(),
			File:        "config/session.php",
			Line:        n,
			CodeSnippet: line,
			Remediation: "Set: 'secure' => env('SESSION_SECURE_COOKIE', true),",
			References:  []string{"https://cwe.mitre.org/data/definitions/614.html"},
		})
	}

	// SameSite not strict or lax
	if line, n := findPattern(lines, `'same_site'\s*=>\s*('none'|null)`); n > 0 {
		findings = append(findings, models.Finding{
			ID:          "CFG-006",
			Title:       "Session cookie SameSite set to none",
			Description: "The SameSite attribute is set to 'none', allowing the cookie to be sent with cross-site requests. This weakens CSRF protection.",
			Severity:    models.SeverityMedium,
			Category:    "Configuration",
			Scanner:     s.Name(),
			File:        "config/session.php",
			Line:        n,
			CodeSnippet: line,
			Remediation: "Set: 'same_site' => 'lax', (or 'strict' for maximum protection)",
		})
	}

	// Session lifetime very long (> 480 minutes = 8 hours)
	if line, n := findPattern(lines, `'lifetime'\s*=>\s*(\d{4,})`); n > 0 {
		findings = append(findings, models.Finding{
			ID:          "CFG-007",
			Title:       "Session lifetime is excessively long",
			Description: "Sessions persist for an unusually long time. Long session lifetimes increase the risk of session hijacking and unauthorized access from abandoned sessions.",
			Severity:    models.SeverityLow,
			Category:    "Configuration",
			Scanner:     s.Name(),
			File:        "config/session.php",
			Line:        n,
			CodeSnippet: line,
			Remediation: "Set a reasonable session lifetime: 'lifetime' => 120, (2 hours)",
		})
	}

	return findings
}

func (s *Scanner) checkMail(path string) []models.Finding {
	var findings []models.Finding
	content := readFile(path)
	lines := toLines(content)

	// Hardcoded mail credentials
	if line, n := findPattern(lines, `'password'\s*=>\s*'[^']{4,}'`); n > 0 {
		// Make sure it's not env()
		if !strings.Contains(line, "env(") {
			findings = append(findings, models.Finding{
				ID:          "CFG-008",
				Title:       "Mail password hardcoded in config",
				Description: "A mail password is hardcoded in config/mail.php instead of using env(). This credential is exposed to anyone with source access.",
				Severity:    models.SeverityHigh,
				Category:    "Secrets",
				Scanner:     s.Name(),
				File:        "config/mail.php",
				Line:        n,
				CodeSnippet: strings.Replace(line, line, maskConfigValue(line), 1),
				Remediation: "Use: 'password' => env('MAIL_PASSWORD'),",
			})
		}
	}

	return findings
}

func (s *Scanner) checkCORS(path string) []models.Finding {
	var findings []models.Finding
	content := readFile(path)
	lines := toLines(content)

	if line, n := findPattern(lines, `'allowed_origins'\s*=>\s*\[\s*'\*'\s*\]`); n > 0 {
		findings = append(findings, models.Finding{
			ID:          "CFG-009",
			Title:       "CORS allows all origins",
			Description: "config/cors.php allows requests from any origin ('*'). This permits cross-site data theft if authenticated endpoints return sensitive data.",
			Severity:    models.SeverityMedium,
			Category:    "Configuration",
			Scanner:     s.Name(),
			File:        "config/cors.php",
			Line:        n,
			CodeSnippet: line,
			Remediation: "Specify allowed origins: 'allowed_origins' => [env('FRONTEND_URL')],",
			References:  []string{"https://cwe.mitre.org/data/definitions/942.html"},
		})
	}

	if line, n := findPattern(lines, `'supports_credentials'\s*=>\s*true`); n > 0 {
		// Credentials + wildcard is especially dangerous
		if _, wn := findPattern(lines, `'allowed_origins'\s*=>\s*\[\s*'\*'\s*\]`); wn > 0 {
			findings = append(findings, models.Finding{
				ID:          "CFG-010",
				Title:       "CORS allows credentials with wildcard origin",
				Description: "CORS is configured with both 'supports_credentials' => true and wildcard allowed_origins. This combination allows any website to make authenticated requests to your API.",
				Severity:    models.SeverityHigh,
				Category:    "Configuration",
				Scanner:     s.Name(),
				File:        "config/cors.php",
				Line:        n,
				CodeSnippet: line,
				Remediation: "Never combine 'supports_credentials' => true with wildcard origins. Specify exact allowed origins.",
				References:  []string{"https://cwe.mitre.org/data/definitions/942.html"},
			})
		}
	}

	return findings
}

func (s *Scanner) checkDatabase(path string) []models.Finding {
	var findings []models.Finding
	content := readFile(path)
	lines := toLines(content)

	// Hardcoded database password
	if line, n := findPattern(lines, `'password'\s*=>\s*'[^']{4,}'`); n > 0 {
		if !strings.Contains(line, "env(") {
			findings = append(findings, models.Finding{
				ID:          "CFG-011",
				Title:       "Database password hardcoded in config",
				Description: "A database password is hardcoded in config/database.php. Use env() to keep credentials out of source.",
				Severity:    models.SeverityHigh,
				Category:    "Secrets",
				Scanner:     s.Name(),
				File:        "config/database.php",
				Line:        n,
				CodeSnippet: maskConfigValue(line),
				Remediation: "Use: 'password' => env('DB_PASSWORD', ''),",
			})
		}
	}

	return findings
}

func (s *Scanner) checkBroadcasting(path string) []models.Finding {
	var findings []models.Finding
	content := readFile(path)
	lines := toLines(content)

	// Hardcoded Pusher keys
	if line, n := findPattern(lines, `'(secret|key)'\s*=>\s*'[a-zA-Z0-9]{10,}'`); n > 0 {
		if !strings.Contains(line, "env(") {
			findings = append(findings, models.Finding{
				ID:          "CFG-012",
				Title:       "Broadcasting secret/key hardcoded in config",
				Description: "A Pusher or broadcasting service key is hardcoded instead of using env().",
				Severity:    models.SeverityMedium,
				Category:    "Secrets",
				Scanner:     s.Name(),
				File:        "config/broadcasting.php",
				Line:        n,
				CodeSnippet: maskConfigValue(line),
				Remediation: "Use: 'secret' => env('PUSHER_APP_SECRET'),",
			})
		}
	}

	return findings
}

func (s *Scanner) checkLogging(path string) []models.Finding {
	var findings []models.Finding
	content := readFile(path)
	lines := toLines(content)

	// Slack webhook URL in config
	if line, n := findPattern(lines, `hooks\.slack\.com/services`); n > 0 {
		if !strings.Contains(line, "env(") {
			findings = append(findings, models.Finding{
				ID:          "CFG-013",
				Title:       "Slack webhook URL hardcoded in logging config",
				Description: "A Slack webhook URL is hardcoded in config/logging.php. Webhook URLs are sensitive — anyone with the URL can post to your Slack channel.",
				Severity:    models.SeverityMedium,
				Category:    "Secrets",
				Scanner:     s.Name(),
				File:        "config/logging.php",
				Line:        n,
				CodeSnippet: maskConfigValue(line),
				Remediation: "Use: 'url' => env('LOG_SLACK_WEBHOOK_URL'),",
			})
		}
	}

	return findings
}

// Helpers

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func toLines(content string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func findPattern(lines []string, pattern string) (string, int) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", 0
	}
	for i, line := range lines {
		if re.MatchString(line) {
			return strings.TrimSpace(line), i + 1
		}
	}
	return "", 0
}

func maskConfigValue(line string) string {
	// Replace quoted values longer than 4 chars with masked version
	re := regexp.MustCompile(`=>\s*'([^']{4,})'`)
	return re.ReplaceAllStringFunc(line, func(match string) string {
		parts := strings.SplitN(match, "'", 3)
		if len(parts) >= 2 {
			val := parts[1]
			if len(val) > 4 {
				masked := fmt.Sprintf("=> '%s****%s'", val[:2], val[len(val)-2:])
				return masked
			}
		}
		return match
	})
}
