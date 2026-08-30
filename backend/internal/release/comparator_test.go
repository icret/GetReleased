package release

import "testing"

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name      string
		current   string
		candidate string
		want      bool
	}{
		{"newer", "1.0.0", "1.1.0", true},
		{"older", "1.1.0", "1.0.0", false},
		{"equal", "1.0.0", "1.0.0", false},
		{"invalid current", "x", "1.0.0", false},
		{"invalid candidate", "1.0.0", "x", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNewer(tt.current, tt.candidate); got != tt.want {
				t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.candidate, got, tt.want)
			}
		})
	}
}
