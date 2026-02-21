package run

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cfgPath string
)

// Cmd represents the `thoth run` command.
var Cmd = &cobra.Command{
	Use:           "run",
	Short:         "Run actions defined in a config",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgPath == "" {
			return fmt.Errorf("missing required flag: --config")
		}
		out, err := executePipeline(context.Background(), cfgPath)
		if err != nil {
			return err
		}
		return evaluateRunExit(out)
	},
}

func init() {
	Cmd.Flags().StringVarP(&cfgPath, "config", "c", "", "Path to config file (.cue)")
}
