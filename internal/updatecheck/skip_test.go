package updatecheck

import "testing"

func TestShouldSkip(t *testing.T) {
	cases := []struct {
		name           string
		currentVersion string
		env            map[string]string
		want           bool
	}{
		{"release build, no env vars", "v0.1.0", nil, false},
		{"dev build always skips", "dev", nil, true},
		{"empty version skips", "", nil, true},
		{"explicit opt-out", "v0.1.0", map[string]string{"SPOT_NO_UPDATE_CHECK": "1"}, true},
		{"opt-out=0 does not skip", "v0.1.0", map[string]string{"SPOT_NO_UPDATE_CHECK": "0"}, false},
		{"CI=true skips", "v0.1.0", map[string]string{"CI": "true"}, true},
		{"CI=1 skips", "v0.1.0", map[string]string{"CI": "1"}, true},
		{"GITHUB_ACTIONS skips", "v0.1.0", map[string]string{"GITHUB_ACTIONS": "true"}, true},
		{"BUILDKITE skips", "v0.1.0", map[string]string{"BUILDKITE": "true"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any CI-ish env vars the test harness might set.
			for _, k := range []string{"SPOT_NO_UPDATE_CHECK", "CI", "GITHUB_ACTIONS", "BUILDKITE"} {
				t.Setenv(k, "")
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			got := ShouldSkip(tc.currentVersion)
			if got != tc.want {
				t.Errorf("ShouldSkip(%q) = %v, want %v", tc.currentVersion, got, tc.want)
			}
		})
	}
}
