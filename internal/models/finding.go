package models

// Finding represents a single security issue discovered by a scanner.
type Finding struct {
	ID          string
	Title       string
	Description string
	Severity    Severity
	Category    string
	Scanner     string
	File        string
	Line        int
	CodeSnippet string
	Remediation string
	References  []string
}
