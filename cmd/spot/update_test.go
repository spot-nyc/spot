package main

import "testing"

func TestDetectInstallSource(t *testing.T) {
	cases := []struct {
		name string
		path string
		want installSource
	}{
		{"homebrew apple silicon", "/opt/homebrew/Cellar/spot/0.1.0/bin/spot", sourceHomebrew},
		{"homebrew intel", "/usr/local/Cellar/spot/0.1.0/bin/spot", sourceHomebrew},
		{"scoop unix separators", "/Users/alice/scoop/apps/spot/current/spot.exe", sourceScoop},
		{"scoop windows separators", `C:\Users\alice\scoop\apps\spot\current\spot.exe`, sourceScoop},
		{"go install in default GOPATH", "/Users/alice/go/bin/spot", sourceGoInstall},
		{"install.sh user-local", "/Users/alice/.local/bin/spot", sourceInstallScript},
		{"install.sh usr-local not symlinked", "/usr/local/bin/spot", sourceInstallScript},
		{"unknown path", "/tmp/whatever/spot", sourceUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := detectInstallSource(tc.path, detectHelpers{
				evalSymlinks: func(p string) (string, error) { return p, nil },
				gopath:       "/Users/alice/go",
				homeDir:      "/Users/alice",
			})
			if got != tc.want {
				t.Errorf("detectInstallSource(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestDetectInstallSource_HomebrewSymlink(t *testing.T) {
	got := detectInstallSource("/usr/local/bin/spot", detectHelpers{
		evalSymlinks: func(p string) (string, error) {
			return "/usr/local/Cellar/spot/0.1.0/bin/spot", nil
		},
		gopath:  "/home/user/go",
		homeDir: "/home/user",
	})
	if got != sourceHomebrew {
		t.Errorf("expected sourceHomebrew via symlink resolution, got %v", got)
	}
}
