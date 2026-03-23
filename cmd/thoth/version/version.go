// File Guide for dev/ai agents:
// Purpose: Define the `thoth version` command and its human-readable versus JSON output modes.
// Responsibilities:
// - Register the version command and its flags.
// - Print the stable one-line version summary for normal CLI use.
// - Emit the richer JSON diagnostic object when explicitly requested.
// Architecture notes:
// - The human-readable path stays the default and is kept to exactly one stdout line because E2E tests rely on that stable contract.
// - JSON mode writes a human hint to stderr and structured data to stdout so machine-readable output stays clean.
package version

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/flarebyte/thoth-ostraca/internal/buildinfo"
	"github.com/spf13/cobra"
)

var (
	flagShort bool
	flagJSON  bool
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagShort || !flagJSON {
			// Keep Phase 1 stable output for E2E: exactly one line.
			if _, err := fmt.Fprintf(os.Stdout, "thoth %s\n", buildinfo.Summary()); err != nil {
				return err
			}
			return nil
		}

		// If JSON is requested explicitly, print a diagnostic object to stdout
		// and a human friendly line to stderr.
		_, _ = fmt.Fprintf(os.Stderr, "thoth version: %s\n", buildinfo.Summary())
		out := map[string]any{
			"version":   buildinfo.Version,
			"commit":    buildinfo.Commit,
			"date":      buildinfo.Date,
			"built_by":  buildinfo.BuiltBy,
			"go":        runtime.Version(),
			"go_os":     runtime.GOOS,
			"go_arch":   runtime.GOARCH,
			"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		}
		return encodeJSON(os.Stdout, out)
	},
}

func init() {
	VersionCmd.Flags().BoolVar(&flagShort, "short", false, "Print only the version string")
	VersionCmd.Flags().BoolVar(&flagJSON, "json", false, "Print detailed JSON version info")
}
