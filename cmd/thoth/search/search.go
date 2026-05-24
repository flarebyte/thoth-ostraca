// File Guide for dev/ai agents:
// Purpose: Expose the `thoth search` CLI subcommand so users can query `*.thoth.yaml` metadata without providing a CUE config.
// Responsibilities:
// - Define and parse search flags (`--root`, `--term`, `--fields`, `--out`).
// - Call the internal search engine and propagate execution errors.
// - Emit command output through Cobra's configured stdout stream.
// Architecture notes:
// - NewCmd returns a fresh command instance with locally scoped flag state to avoid shared global test/runtime mutation.
// - Search logic is delegated to internal/search so this file remains CLI wiring only.
package search

import (
	"context"

	internalsearch "github.com/flarebyte/thoth-ostraca/internal/search"
	"github.com/spf13/cobra"
)

// Cmd represents the `thoth search` command.
var Cmd = NewCmd()

// NewCmd creates a new `thoth search` command.
func NewCmd() *cobra.Command {
	var (
		localRoot   string
		localTerm   string
		localFields []string
		localOut    string
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search thoth metadata sidecar files",
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := internalsearch.Execute(context.Background(), internalsearch.Options{
				Root:   localRoot,
				Term:   localTerm,
				Fields: localFields,
				Out:    localOut,
			})
			if err != nil {
				return err
			}
			return internalsearch.WriteJSON(results, localOut, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&localRoot, "root", ".", "Root folder to search recursively")
	cmd.Flags().StringVar(&localTerm, "term", "", "Optional case-insensitive search term")
	cmd.Flags().StringSliceVar(&localFields, "fields", nil, "Optional comma-separated meta field keys to return")
	cmd.Flags().StringVar(&localOut, "out", "", "Optional output JSON file path; defaults to stdout")
	return cmd
}
