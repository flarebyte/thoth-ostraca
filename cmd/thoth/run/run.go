// File Guide for dev/ai agents:
// Purpose: Define the `thoth run` Cobra command that executes config-driven actions from the CLI.
// Responsibilities:
// - Define the `run` command and its required `--config` flag.
// - Invoke the pipeline executor with a background context.
// - Apply final exit-rule evaluation after the pipeline completes.
// Architecture notes:
// - The command stays intentionally thin so business logic remains testable in package functions rather than embedded in Cobra callbacks.
// - Exit evaluation happens after pipeline execution instead of during stage runs so output writing and diagnostics complete before the process exits non-zero.
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
	Cmd.Flags().StringVarP(&cfgPath, "config", "c", "", "Config file path (.cue, required)")
}
