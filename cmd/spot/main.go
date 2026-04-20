// Command spot is the Spot CLI — manage reservations, searches, and restaurant lookup.
package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/internal/render"
)

// rootFlags holds persistent flag state shared by all subcommands.
type rootFlags struct {
	json bool
}

func newRootCmd() *cobra.Command {
	flags := &rootFlags{}

	cmd := &cobra.Command{
		Use:   "spot",
		Short: "Spot — manage reservations, searches, and restaurant lookup",
		Long: "The Spot CLI.\n\n" +
			"Monitor reservations, create searches, book tables, and query restaurants " +
			"from the command line.\n\n" +
			"Run `spot <command> --help` for detail on any command.",
		Version:       spot.Version,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}

	cmd.PersistentFlags().BoolVarP(&flags.json, "json", "j", false, "force JSON output (default: auto-detect based on TTY)")

	cmd.AddCommand(newAuthCmd(flags))
	cmd.AddCommand(newSearchesCmd(flags))
	cmd.AddCommand(newReservationsCmd(flags))

	return cmd
}

// resolveFormat determines the effective output format for a command writing
// to out. Honors the persistent --json flag; otherwise auto-detects based on
// whether out is a TTY. Pass cmd.OutOrStdout() from each command so tests
// that call SetOut(&buf) get deterministic JSON output.
func (f *rootFlags) resolveFormat(out io.Writer) render.Format {
	return render.ResolveWriter(f.json, out)
}

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		format := render.FormatTable
		if jsonFlag := cmd.PersistentFlags().Lookup("json"); jsonFlag != nil && jsonFlag.Value.String() == "true" {
			format = render.FormatJSON
		}

		if format == render.FormatJSON {
			RenderError(os.Stdout, format, err)
		} else {
			RenderError(os.Stderr, format, err)
		}

		os.Exit(ExitCodeFor(err))
	}
}
