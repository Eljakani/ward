package updater

import "testing"

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input string
		want  []int
	}{
		{"v0.2.0", []int{0, 2, 0}},
		{"0.2.0", []int{0, 2, 0}},
		{"v1.10.3", []int{1, 10, 3}},
		{"v1.0.0-beta", []int{1, 0, 0}},
		{"dev", nil},
		{"", nil},
		{"v1.2", nil},
	}

	for _, tt := range tests {
		got := parseSemver(tt.input)
		if tt.want == nil {
			if got != nil {
				t.Errorf("parseSemver(%q) = %v, want nil", tt.input, got)
			}
			continue
		}
		if got == nil {
			t.Errorf("parseSemver(%q) = nil, want %v", tt.input, tt.want)
			continue
		}
		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Errorf("parseSemver(%q)[%d] = %d, want %d", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest  string
		current string
		want    bool
	}{
		{"v0.3.0", "v0.2.0", true},
		{"v0.2.0", "v0.2.0", false},
		{"v0.1.0", "v0.2.0", false},
		{"v1.0.0", "v0.99.99", true},
		{"v0.2.1", "v0.2.0", true},
		{"dev", "v0.2.0", false},
		{"v0.3.0", "dev", false},
	}

	for _, tt := range tests {
		got := isNewer(tt.latest, tt.current)
		if got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}

func TestFormatNotice(t *testing.T) {
	notice := formatNotice("v0.2.0", "v0.3.0", "https://github.com/Eljakani/ward/releases/tag/v0.3.0")
	if notice == "" {
		t.Error("formatNotice should return a non-empty string")
	}
}
