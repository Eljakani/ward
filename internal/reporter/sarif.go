package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eljakani/ward/internal/models"
)

// SARIFReporter generates a SARIF 2.1.0 report for GitHub Code Scanning integration.
type SARIFReporter struct {
	OutputDir string
}

func NewSARIFReporter(outputDir string) *SARIFReporter {
	if outputDir == "" {
		outputDir = "."
	}
	return &SARIFReporter{OutputDir: outputDir}
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
				ID:   f.ID,
				Name: f.Title,
				ShortDescription: sarifMessage{Text: f.Title},
				FullDescription:  sarifMessage{Text: f.Description},
				DefaultConfiguration: sarifRuleConfig{
					Level: severityToSARIFLevel(f.Severity),
				},
				Help: sarifMessage{Text: f.Remediation},
				Properties: sarifRuleProperties{
					Tags:     []string{f.Category},
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
			Message:   sarifMessage{Text: f.Title},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI:       f.File,
							URIBaseID: "%SRCROOT%",
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
		if len(f.References) > 0 {
			result.PartialFingerprints = map[string]string{"primaryLocationLineHash": f.ID + f.File}
		}
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
						Version:         "0.2.0",
						SemanticVersion: "0.2.0",
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
		return "critical"
	case models.SeverityHigh:
		return "high"
	case models.SeverityMedium:
		return "medium"
	case models.SeverityLow:
		return "low"
	default:
		return "informational"
	}
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
	URI       string `json:"uri"`
	URIBaseID string `json:"uriBaseId"`
}

type sarifRegion struct {
	StartLine int            `json:"startLine"`
	Snippet   *sarifSnippet  `json:"snippet,omitempty"`
}

type sarifSnippet struct {
	Text string `json:"text"`
}
