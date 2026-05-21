package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eljakani/ward/internal/models"
)

// SARIFReporter generates a SARIF 2.1.0 report for GitHub Code Scanning integration.
type SARIFReporter struct {
	OutputDir string
	Version   string
}

func NewSARIFReporter(outputDir string, version string) *SARIFReporter {
	if outputDir == "" {
		outputDir = "."
	}
	if version == "" {
		version = "dev"
	}
	return &SARIFReporter{OutputDir: outputDir, Version: version}
}

func (r *SARIFReporter) Name() string   { return "sarif" }
func (r *SARIFReporter) Format() string { return "sarif" }

func (r *SARIFReporter) Generate(_ context.Context, report *models.ScanReport) error {
	// Build rules from unique finding IDs
	ruleIndex := make(map[string]int)
	var rules []sarifRule

	for _, f := range report.Findings {
		if _, exists := ruleIndex[f.ID]; !exists {
			ruleIndex[f.ID] = len(rules)
			rule := sarifRule{
				ID:               f.ID,
				Name:             sarifText(f.Title, f.ID),
				ShortDescription: sarifMessage{Text: sarifText(f.Title, f.ID)},
				FullDescription:  sarifMessage{Text: sarifText(f.Description, f.Title, f.ID)},
				DefaultConfiguration: sarifRuleConfig{
					Level: severityToSARIFLevel(f.Severity),
				},
				Help: sarifMessage{Text: sarifText(f.Remediation, f.Description, f.Title, f.ID)},
				Properties: sarifRuleProperties{
					Tags:     sarifTags(f.Category),
					Security: severityToSARIFSecurity(f.Severity),
				},
			}
			rules = append(rules, rule)
		}
	}

	// Build results
	var results []sarifResult
	for _, f := range report.Findings {
		result := sarifResult{
			RuleID:    f.ID,
			RuleIndex: ruleIndex[f.ID],
			Level:     severityToSARIFLevel(f.Severity),
			Message:   sarifMessage{Text: sarifText(f.Title, f.Description, f.ID)},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI: f.File,
						},
						Region: sarifRegion{
							StartLine: max(f.Line, 1),
						},
					},
				},
			},
		}
		if f.CodeSnippet != "" {
			result.Locations[0].PhysicalLocation.Region.Snippet = &sarifSnippet{Text: f.CodeSnippet}
		}
		result.PartialFingerprints = map[string]string{"primaryLocationLineHash/v1": f.Fingerprint()}
		results = append(results, result)
	}

	sarifDoc := sarifDocument{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:            "Ward",
						InformationURI:  "https://github.com/Eljakani/ward",
						Version:         r.Version,
						SemanticVersion: r.Version,
						Rules:           rules,
					},
				},
				Results: results,
			},
		},
	}

	data, err := json.MarshalIndent(sarifDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling SARIF report: %w", err)
	}

	outPath := filepath.Join(r.OutputDir, "ward-report.sarif")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("writing SARIF report to %s: %w", outPath, err)
	}

	return nil
}

func severityToSARIFLevel(s models.Severity) string {
	switch s {
	case models.SeverityCritical, models.SeverityHigh:
		return "error"
	case models.SeverityMedium:
		return "warning"
	default:
		return "note"
	}
}

func severityToSARIFSecurity(s models.Severity) string {
	switch s {
	case models.SeverityCritical:
		return "9.5"
	case models.SeverityHigh:
		return "7.5"
	case models.SeverityMedium:
		return "5.5"
	case models.SeverityLow:
		return "2.5"
	default:
		return "0.0"
	}
}

func sarifText(values ...string) string {
	for _, v := range values {
		if text := strings.TrimSpace(v); text != "" {
			return text
		}
	}
	return "Ward security finding"
}

func sarifTags(category string) []string {
	tags := []string{"security"}
	category = strings.TrimSpace(category)
	if category == "" || strings.EqualFold(category, "security") {
		return tags
	}
	return append(tags, category)
}

// SARIF 2.1.0 data structures

type sarifDocument struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name            string      `json:"name"`
	InformationURI  string      `json:"informationUri"`
	Version         string      `json:"version"`
	SemanticVersion string      `json:"semanticVersion"`
	Rules           []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID                   string              `json:"id"`
	Name                 string              `json:"name"`
	ShortDescription     sarifMessage        `json:"shortDescription"`
	FullDescription      sarifMessage        `json:"fullDescription"`
	DefaultConfiguration sarifRuleConfig     `json:"defaultConfiguration"`
	Help                 sarifMessage        `json:"help"`
	Properties           sarifRuleProperties `json:"properties"`
}

type sarifRuleConfig struct {
	Level string `json:"level"`
}

type sarifRuleProperties struct {
	Tags     []string `json:"tags"`
	Security string   `json:"security-severity"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifResult struct {
	RuleID              string            `json:"ruleId"`
	RuleIndex           int               `json:"ruleIndex"`
	Level               string            `json:"level"`
	Message             sarifMessage      `json:"message"`
	Locations           []sarifLocation   `json:"locations"`
	PartialFingerprints map[string]string `json:"partialFingerprints,omitempty"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int           `json:"startLine"`
	Snippet   *sarifSnippet `json:"snippet,omitempty"`
}

type sarifSnippet struct {
	Text string `json:"text"`
}
