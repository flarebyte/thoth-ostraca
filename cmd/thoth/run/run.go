package run

import (
	"fmt"
	"os"

	"github.com/flarebyte/thoth-ostraca/internal/config"
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
		if err := config.LoadAndValidate(cfgPath); err != nil {
			return err
		}
		// Success output must be a single JSON line.
		fmt.Fprintln(os.Stdout, `{"ok":true}`)
		return nil
	},
}

func init() {
	Cmd.Flags().StringVarP(&cfgPath, "config", "c", "", "Path to config file (.cue)")
}
