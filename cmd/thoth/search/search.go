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
