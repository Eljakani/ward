package models

import (
	"crypto/sha256"
	"fmt"
)

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

// Fingerprint returns a stable hash identifying this finding across scans.
// Based on rule ID + file + line so it stays consistent even if descriptions change.
func (f Finding) Fingerprint() string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", f.ID, f.File, f.Line)))
	return fmt.Sprintf("%x", h[:12]) // 24-char hex
}
