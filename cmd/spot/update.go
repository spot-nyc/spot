package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot/internal/render"
)

// installSource is where the current `spot` binary was installed from. We
// use this to pick the right upgrade command.
type installSource int

const (
	sourceUnknown installSource = iota
	sourceHomebrew
	sourceScoop
	sourceGoInstall
	sourceInstallScript
)

func (s installSource) name() string {
	switch s {
	case sourceHomebrew:
		return "homebrew"
	case sourceScoop:
		return "scoop"
	case sourceGoInstall:
		return "go-install"
	case sourceInstallScript:
		return "install-script"
	}
	return "unknown"
}

// detectHelpers isolates the OS-level lookups detectInstallSource needs, so
// the logic can be unit-tested with fake inputs.
type detectHelpers struct {
	evalSymlinks func(path string) (string, error)
	gopath       string
	homeDir      string
}

func realDetectHelpers() detectHelpers {
	home, _ := os.UserHomeDir()
	gopath := os.Getenv("GOPATH")
	if gopath == "" && home != "" {
		gopath = filepath.Join(home, "go")
	}
	return detectHelpers{
		evalSymlinks: filepath.EvalSymlinks,
		gopath:       gopath,
		homeDir:      home,
	}
}

// detectInstallSource inspects execPath (typically from os.Executable) and
// classifies how the binary was installed. Tests pass fake helpers to cover
// each branch without needing a real filesystem.
func detectInstallSource(execPath string, helpers detectHelpers) installSource {
	lower := strings.ToLower(strings.ReplaceAll(execPath, "\\", "/"))
	if strings.Contains(lower, "/scoop/apps/spot/") {
		return sourceScoop
	}

	resolved := execPath
	if helpers.evalSymlinks != nil {
		if r, err := helpers.evalSymlinks(execPath); err == nil {
			resolved = r
		}
	}

	if strings.Contains(resolved, "/Cellar/spot/") {
		return sourceHomebrew
	}

	if helpers.gopath != "" {
		gopathBin := filepath.Join(helpers.gopath, "bin", "spot")
		if execPath == gopathBin || resolved == gopathBin {
			return sourceGoInstall
		}
	}

	if helpers.homeDir != "" {
		localBin := filepath.Join(helpers.homeDir, ".local", "bin", "spot")
		if execPath == localBin || resolved == localBin {
			return sourceInstallScript
		}
	}

	if execPath == "/usr/local/bin/spot" || resolved == "/usr/local/bin/spot" {
		return sourceInstallScript
	}

	return sourceUnknown
}

// upgradeCommand returns the shell command a user should run to upgrade
// their installation.
func upgradeCommand(source installSource) string {
	switch source {
	case sourceHomebrew:
		return "brew upgrade spot"
	case sourceScoop:
		return "scoop update spot"
	case sourceGoInstall:
		return "go install github.com/spot-nyc/spot/cmd/spot@latest"
	case sourceInstallScript:
		return "curl -fsSL https://raw.githubusercontent.com/spot-nyc/spot/main/install.sh | sh"
	}
	return ""
}

func newUpdateCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Print the upgrade command for the detected install source",
		Long: "Inspects how `spot` was installed and prints the corresponding\n" +
			"upgrade command. Does not run it — copy-paste into your shell.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			execPath, err := os.Executable()
			if err != nil {
				return err
			}
			source := detectInstallSource(execPath, realDetectHelpers())

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				payload := map[string]any{
					"source":         source.name(),
					"upgradeCommand": upgradeCommand(source),
				}
				if source == sourceUnknown {
					payload["alternatives"] = map[string]string{
						"homebrew":     "brew upgrade spot",
						"scoop":        "scoop update spot",
						"goInstall":    "go install github.com/spot-nyc/spot/cmd/spot@latest",
						"installShell": "curl -fsSL https://raw.githubusercontent.com/spot-nyc/spot/main/install.sh | sh",
					}
				}
				return render.JSON(cmd.OutOrStdout(), payload)
			}

			if source == sourceUnknown {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Couldn't detect how spot was installed. Choose one:")
				_, _ = fmt.Fprintln(cmd.OutOrStdout())
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  Homebrew        brew upgrade spot")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  Scoop           scoop update spot")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  go install      go install github.com/spot-nyc/spot/cmd/spot@latest")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  curl installer  curl -fsSL https://raw.githubusercontent.com/spot-nyc/spot/main/install.sh | sh")
				return nil
			}

			labels := map[installSource]string{
				sourceHomebrew:      "Homebrew",
				sourceScoop:         "Scoop",
				sourceGoInstall:     "`go install`",
				sourceInstallScript: "the curl installer",
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "You installed spot via %s. Run:\n\n    %s\n\n", labels[source], upgradeCommand(source))
			return nil
		},
	}
}
