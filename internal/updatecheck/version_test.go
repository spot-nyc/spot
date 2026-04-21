package updatecheck

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"latest is newer patch", "v0.1.0", "v0.1.1", true},
		{"latest is newer minor", "v0.1.0", "v0.2.0", true},
		{"latest is newer major", "v0.1.0", "v1.0.0", true},
		{"latest is older", "v0.2.0", "v0.1.0", false},
		{"same version", "v0.1.0", "v0.1.0", false},
		{"dev current never announces", "dev", "v0.1.0", false},
		{"empty current never announces", "", "v0.1.0", false},
		{"missing v prefix on current", "0.1.0", "v0.1.1", true},
		{"missing v prefix on latest", "v0.1.0", "0.1.1", true},
		{"invalid latest", "v0.1.0", "not-a-version", false},
		{"invalid current", "not-a-version", "v0.1.0", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isNewer(tc.current, tc.latest)
			if got != tc.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
			}
		})
	}
}
