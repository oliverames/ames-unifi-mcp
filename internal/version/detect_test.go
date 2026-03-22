package version

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		raw   string
		major int
		minor int
		patch int
	}{
		{"9.1.120", 9, 1, 120},
		{"10.2.97", 10, 2, 97},
		{"5.8.0", 5, 8, 0},
	}

	for _, tt := range tests {
		v, err := Parse(tt.raw)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.raw, err)
			continue
		}
		if v.Major != tt.major || v.Minor != tt.minor || v.Patch != tt.patch {
			t.Errorf("Parse(%q) = %d.%d.%d, want %d.%d.%d", tt.raw, v.Major, v.Minor, v.Patch, tt.major, tt.minor, tt.patch)
		}
	}
}

func TestAtLeast(t *testing.T) {
	v, _ := Parse("9.1.120")

	if !v.AtLeast(9, 0, 0) {
		t.Error("9.1.120 should be >= 9.0.0")
	}
	if !v.AtLeast(9, 1, 120) {
		t.Error("9.1.120 should be >= 9.1.120")
	}
	if v.AtLeast(9, 2, 0) {
		t.Error("9.1.120 should not be >= 9.2.0")
	}
	if v.AtLeast(10, 0, 0) {
		t.Error("9.1.120 should not be >= 10.0.0")
	}
	if !v.HasZBF() {
		t.Error("9.1.120 should support ZBF")
	}
}

func TestHasZBF(t *testing.T) {
	old, _ := Parse("8.5.0")
	if old.HasZBF() {
		t.Error("8.5.0 should not support ZBF")
	}

	new, _ := Parse("9.0.0")
	if !new.HasZBF() {
		t.Error("9.0.0 should support ZBF")
	}
}
