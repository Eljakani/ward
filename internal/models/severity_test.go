package models

import "testing"

func TestSeverityString(t *testing.T) {
	tests := []struct {
		sev  Severity
		want string
	}{
		{SeverityInfo, "Info"},
		{SeverityLow, "Low"},
		{SeverityMedium, "Medium"},
		{SeverityHigh, "High"},
		{SeverityCritical, "Critical"},
		{Severity(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.sev.String(); got != tt.want {
			t.Errorf("Severity(%d).String() = %q, want %q", tt.sev, got, tt.want)
		}
	}
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input string
		want  Severity
	}{
		{"critical", SeverityCritical},
		{"CRITICAL", SeverityCritical},
		{"high", SeverityHigh},
		{"High", SeverityHigh},
		{"medium", SeverityMedium},
		{"low", SeverityLow},
		{"info", SeverityInfo},
		{"unknown", SeverityInfo},
		{"", SeverityInfo},
	}

	for _, tt := range tests {
		if got := ParseSeverity(tt.input); got != tt.want {
			t.Errorf("ParseSeverity(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestSeverityWeight(t *testing.T) {
	if SeverityInfo.Weight() >= SeverityLow.Weight() {
		t.Error("Info should be less than Low")
	}
	if SeverityCritical.Weight() <= SeverityHigh.Weight() {
		t.Error("Critical should be greater than High")
	}
}

func TestAllSeverities(t *testing.T) {
	all := AllSeverities()
	if len(all) != 5 {
		t.Errorf("AllSeverities() returned %d items, want 5", len(all))
	}
	// Should be ascending
	for i := 1; i < len(all); i++ {
		if all[i] <= all[i-1] {
			t.Errorf("AllSeverities() not ascending at index %d", i)
		}
	}
}
