package dependency

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eljakani/ward/internal/models"
)

const (
	osvQueryURL = "https://api.osv.dev/v1/query"
	osvBatchURL = "https://api.osv.dev/v1/querybatch"
	batchSize   = 100
	httpTimeout = 30 * time.Second
)

// Scanner checks installed packages against the OSV.dev vulnerability database.
type Scanner struct {
	client *http.Client
}

func New() *Scanner {
	return &Scanner{
		client: &http.Client{Timeout: httpTimeout},
	}
}

func (s *Scanner) Name() string        { return "dependency-scanner" }
func (s *Scanner) Description() string { return "Live CVE checks via OSV.dev (Packagist SBOM)" }

func (s *Scanner) Scan(ctx context.Context, project models.ProjectContext, emit func(models.Finding)) ([]models.Finding, error) {
	if len(project.InstalledPackages) == 0 {
		return nil, nil
	}

	// Step 1: Batch query to find which packages have known vulnerabilities
	vulnPackages, err := s.batchQuery(ctx, project.InstalledPackages)
	if err != nil {
		return nil, fmt.Errorf("querying OSV.dev: %w", err)
	}

	if len(vulnPackages) == 0 {
		return nil, nil
	}

	// Step 2: Fetch full vulnerability details for affected packages
	var findings []models.Finding

	for _, vp := range vulnPackages {
		vulns, err := s.queryPackage(ctx, vp.name, vp.version)
		if err != nil {
			continue // skip on error, don't fail the whole scan
		}

		for _, vuln := range vulns {
			f := vulnToFinding(s.Name(), vp.name, vp.version, vuln)
			findings = append(findings, f)
			emit(f)
		}
	}

	return findings, nil
}

type vulnPackage struct {
	name    string
	version string
}

// batchQuery sends all packages to OSV.dev batch endpoint and returns those with vulnerabilities.
func (s *Scanner) batchQuery(ctx context.Context, packages map[string]string) ([]vulnPackage, error) {
	// Build list of packages
	type query struct {
		Package struct {
			Name      string `json:"name"`
			Ecosystem string `json:"ecosystem"`
		} `json:"package"`
		Version string `json:"version"`
	}

	var allQueries []query
	var packageOrder []vulnPackage // preserve order for matching results

	for name, version := range packages {
		version = normalizeVersion(version)
		if version == "" {
			continue
		}
		q := query{Version: version}
		q.Package.Name = name
		q.Package.Ecosystem = "Packagist"
		allQueries = append(allQueries, q)
		packageOrder = append(packageOrder, vulnPackage{name: name, version: version})
	}

	// Send in batches
	var affected []vulnPackage

	for i := 0; i < len(allQueries); i += batchSize {
		end := i + batchSize
		if end > len(allQueries) {
			end = len(allQueries)
		}

		batch := allQueries[i:end]
		batchOrder := packageOrder[i:end]

		body, _ := json.Marshal(map[string]any{"queries": batch})
		req, err := http.NewRequestWithContext(ctx, "POST", osvBatchURL, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, err
		}

		var result struct {
			Results []struct {
				Vulns []struct {
					ID string `json:"id"`
				} `json:"vulns"`
			} `json:"results"`
		}

		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("OSV.dev returned status %d", resp.StatusCode)
		}

		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("parsing OSV.dev response: %w", err)
		}

		for j, r := range result.Results {
			if len(r.Vulns) > 0 && j < len(batchOrder) {
				affected = append(affected, batchOrder[j])
			}
		}
	}

	return affected, nil
}

// queryPackage fetches full vulnerability details for a single package+version.
func (s *Scanner) queryPackage(ctx context.Context, name, version string) ([]osvVuln, error) {
	body, _ := json.Marshal(map[string]any{
		"package": map[string]string{
			"name":      name,
			"ecosystem": "Packagist",
		},
		"version": version,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", osvQueryURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OSV.dev returned status %d", resp.StatusCode)
	}

	var result struct {
		Vulns []osvVuln `json:"vulns"`
	}

	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Vulns, nil
}

// OSV.dev response structures

type osvVuln struct {
	ID               string           `json:"id"`
	Summary          string           `json:"summary"`
	Details          string           `json:"details"`
	Aliases          []string         `json:"aliases"`
	References       []osvReference   `json:"references"`
	Affected         []osvAffected    `json:"affected"`
	DatabaseSpecific osvDBSpecific    `json:"database_specific"`
}

type osvReference struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type osvAffected struct {
	Package struct {
		Name      string `json:"name"`
		Ecosystem string `json:"ecosystem"`
	} `json:"package"`
	Ranges []osvRange `json:"ranges"`
}

type osvRange struct {
	Type   string     `json:"type"`
	Events []osvEvent `json:"events"`
}

type osvEvent struct {
	Introduced string `json:"introduced,omitempty"`
	Fixed      string `json:"fixed,omitempty"`
}

type osvDBSpecific struct {
	Severity string `json:"severity"`
	CWEIDs   []string `json:"cwe_ids"`
}

// vulnToFinding converts an OSV vulnerability to a Ward finding.
func vulnToFinding(scanner, pkgName, pkgVersion string, vuln osvVuln) models.Finding {
	// Extract CVE ID from aliases
	cveID := vuln.ID
	for _, alias := range vuln.Aliases {
		if strings.HasPrefix(alias, "CVE-") {
			cveID = alias
			break
		}
	}

	// Determine severity
	severity := parseSeverity(vuln.DatabaseSpecific.Severity)

	// Extract fixed version from affected ranges
	fixedVersion := extractFixedVersion(vuln.Affected, pkgName)

	// Build references
	var refs []string
	for _, ref := range vuln.References {
		if ref.Type == "ADVISORY" || ref.Type == "WEB" {
			refs = append(refs, ref.URL)
		}
	}
	// Limit to 3 refs
	if len(refs) > 3 {
		refs = refs[:3]
	}

	// Build description — use summary, fall back to truncated details
	description := vuln.Summary
	if description == "" && len(vuln.Details) > 0 {
		description = vuln.Details
		if len(description) > 300 {
			description = description[:300] + "..."
		}
	}

	// Build remediation
	remediation := fmt.Sprintf("Run: composer update %s", pkgName)
	if fixedVersion != "" {
		remediation = fmt.Sprintf("Upgrade %s to %s or later:\n  composer require %s:%s", pkgName, fixedVersion, pkgName, fixedVersion)
	}

	return models.Finding{
		ID:          cveID,
		Title:       fmt.Sprintf("[%s] %s@%s — %s", cveID, pkgName, pkgVersion, vuln.Summary),
		Description: description,
		Severity:    severity,
		Category:    "Dependencies",
		Scanner:     scanner,
		File:        "composer.lock",
		Remediation: remediation,
		References:  refs,
	}
}

func extractFixedVersion(affected []osvAffected, pkgName string) string {
	for _, a := range affected {
		if a.Package.Name != pkgName {
			continue
		}
		for _, r := range a.Ranges {
			for _, e := range r.Events {
				if e.Fixed != "" {
					return e.Fixed
				}
			}
		}
	}
	return ""
}

func parseSeverity(s string) models.Severity {
	switch strings.ToUpper(s) {
	case "CRITICAL":
		return models.SeverityCritical
	case "HIGH":
		return models.SeverityHigh
	case "MODERATE", "MEDIUM":
		return models.SeverityMedium
	case "LOW":
		return models.SeverityLow
	default:
		return models.SeverityMedium // default to medium for unknown
	}
}

func normalizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	// Skip dev/branch versions that OSV can't match
	if strings.HasPrefix(v, "dev-") || v == "" {
		return ""
	}
	return v
}
