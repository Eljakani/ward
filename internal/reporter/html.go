package reporter

import (
	"context"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/eljakani/ward/internal/models"
)

// HTMLReporter generates a standalone HTML report.
type HTMLReporter struct {
	OutputDir string
}

func NewHTMLReporter(outputDir string) *HTMLReporter {
	if outputDir == "" {
		outputDir = "."
	}
	return &HTMLReporter{OutputDir: outputDir}
}

func (r *HTMLReporter) Name() string   { return "html" }
func (r *HTMLReporter) Format() string { return "html" }

func (r *HTMLReporter) Generate(_ context.Context, report *models.ScanReport) error {
	counts := report.CountBySeverity()

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Ward Security Report</title>
<style>
  :root {
    --bg: #0d1117; --surface: #161b22; --border: #30363d;
    --text: #e6edf3; --muted: #8b949e;
    --critical: #ff5252; --high: #ffb74d; --medium: #ffd54f;
    --low: #81c784; --info: #64b5f6; --accent: #b388ff;
  }
  * { margin:0; padding:0; box-sizing:border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
         background: var(--bg); color: var(--text); line-height: 1.6; padding: 2rem; }
  .container { max-width: 1100px; margin: 0 auto; }
  h1 { color: var(--accent); margin-bottom: 0.5rem; font-size: 1.8rem; }
  .subtitle { color: var(--muted); margin-bottom: 2rem; }
  .summary { display: flex; gap: 1rem; flex-wrap: wrap; margin-bottom: 2rem; }
  .stat { background: var(--surface); border: 1px solid var(--border);
          border-radius: 8px; padding: 1rem 1.5rem; min-width: 120px; text-align: center; }
  .stat .number { font-size: 2rem; font-weight: bold; }
  .stat .label { color: var(--muted); font-size: 0.85rem; text-transform: uppercase; }
  .stat.critical .number { color: var(--critical); }
  .stat.high .number { color: var(--high); }
  .stat.medium .number { color: var(--medium); }
  .stat.low .number { color: var(--low); }
  .stat.info .number { color: var(--info); }
  .stat.total .number { color: var(--accent); }
  .finding { background: var(--surface); border: 1px solid var(--border);
             border-radius: 8px; padding: 1.25rem; margin-bottom: 1rem; }
  .finding-header { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.5rem; }
  .badge { padding: 2px 10px; border-radius: 12px; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; }
  .badge.critical { background: var(--critical); color: #000; }
  .badge.high { background: var(--high); color: #000; }
  .badge.medium { background: var(--medium); color: #000; }
  .badge.low { background: var(--low); color: #000; }
  .badge.info { background: var(--info); color: #000; }
  .finding h3 { font-size: 1rem; }
  .finding .meta { color: var(--muted); font-size: 0.85rem; margin-bottom: 0.75rem; }
  .finding .description { margin-bottom: 0.75rem; }
  .finding pre { background: #0d1117; border: 1px solid var(--border);
                 border-radius: 6px; padding: 0.75rem; overflow-x: auto; font-size: 0.85rem;
                 margin-bottom: 0.75rem; color: var(--info); }
  .finding .remediation { background: rgba(179,136,255,0.08); border-left: 3px solid var(--accent);
                          padding: 0.75rem; border-radius: 0 6px 6px 0; font-size: 0.9rem; }
  .finding .references { margin-top: 0.5rem; }
  .finding .references a { color: var(--info); font-size: 0.85rem; }
  .footer { margin-top: 3rem; text-align: center; color: var(--muted); font-size: 0.85rem; }
  .footer a { color: var(--accent); }
  @media (prefers-color-scheme: light) {
    :root { --bg:#fff; --surface:#f6f8fa; --border:#d0d7de; --text:#1f2328; --muted:#656d76; }
    .finding pre { background: #f6f8fa; }
  }
</style>
</head>
<body>
<div class="container">
`)

	// Header
	sb.WriteString(fmt.Sprintf(`<h1>Ward Security Report</h1>
<p class="subtitle">%s &mdash; Laravel %s &mdash; %s &mdash; %d scanner(s)</p>
`, esc(report.ProjectContext.ProjectName), esc(report.ProjectContext.LaravelVersion),
		report.Duration.Round(1e6), len(report.ScannersRun)))

	// Summary stats
	sb.WriteString(`<div class="summary">`)
	sb.WriteString(fmt.Sprintf(`<div class="stat total"><div class="number">%d</div><div class="label">Total</div></div>`, len(report.Findings)))
	for _, sev := range []models.Severity{models.SeverityCritical, models.SeverityHigh, models.SeverityMedium, models.SeverityLow, models.SeverityInfo} {
		if c := counts[sev]; c > 0 {
			sb.WriteString(fmt.Sprintf(`<div class="stat %s"><div class="number">%d</div><div class="label">%s</div></div>`,
				strings.ToLower(sev.String()), c, sev.String()))
		}
	}
	sb.WriteString(`</div>`)

	// Findings
	for _, f := range report.Findings {
		sevClass := strings.ToLower(f.Severity.String())
		sb.WriteString(`<div class="finding">`)
		sb.WriteString(fmt.Sprintf(`<div class="finding-header"><span class="badge %s">%s</span><h3>%s</h3></div>`,
			sevClass, f.Severity.String(), esc(f.Title)))
		sb.WriteString(fmt.Sprintf(`<div class="meta">%s &bull; %s:%d &bull; %s</div>`,
			esc(f.ID), esc(f.File), f.Line, esc(f.Category)))
		sb.WriteString(fmt.Sprintf(`<div class="description">%s</div>`, esc(f.Description)))
		if f.CodeSnippet != "" {
			sb.WriteString(fmt.Sprintf(`<pre>%s</pre>`, esc(f.CodeSnippet)))
		}
		if f.Remediation != "" {
			sb.WriteString(fmt.Sprintf(`<div class="remediation">%s</div>`, esc(f.Remediation)))
		}
		if len(f.References) > 0 {
			sb.WriteString(`<div class="references">`)
			for _, ref := range f.References {
				sb.WriteString(fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener">%s</a> `, esc(ref), esc(ref)))
			}
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</div>`)
	}

	// Footer
	sb.WriteString(fmt.Sprintf(`<div class="footer">Generated by <a href="https://github.com/Eljakani/ward">Ward</a> v0.1.0</div>
</div>
</body>
</html>`))

	outPath := filepath.Join(r.OutputDir, "ward-report.html")
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing HTML report to %s: %w", outPath, err)
	}

	return nil
}

func esc(s string) string {
	return html.EscapeString(s)
}
