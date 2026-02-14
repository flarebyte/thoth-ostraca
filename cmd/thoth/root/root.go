package root

import (
	"github.com/flarebyte/thoth-ostraca/cmd/thoth/run"
	"github.com/flarebyte/thoth-ostraca/cmd/thoth/version"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for thoth.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thoth",
		Short: "CLI: A repository of interconnected data fragments archived for eternity by the scribe of the gods",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show help when no subcommand is provided.
			return cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Subcommands
	cmd.AddCommand(version.VersionCmd)
	cmd.AddCommand(run.Cmd)

	return cmd
}

// Execute runs the root command with provided args.
func Execute(args []string) error {
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	return cmd.Execute()
}
