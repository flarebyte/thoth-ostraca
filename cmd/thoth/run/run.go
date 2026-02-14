package run

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
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
		in := stage.Envelope{Records: []any{}, Meta: &stage.Meta{ConfigPath: cfgPath}}
		e1, err := stage.Run(context.Background(), "validate-config", in, stage.Deps{})
		if err != nil {
			return err
		}
		e2, err := stage.Run(context.Background(), "discover-meta-files", e1, stage.Deps{})
		if err != nil {
			return err
		}
		e3, err := stage.Run(context.Background(), "parse-validate-yaml", e2, stage.Deps{})
		if err != nil {
			return err
		}
		e4, err := stage.Run(context.Background(), "lua-filter", e3, stage.Deps{})
		if err != nil {
			return err
		}
		e5, err := stage.Run(context.Background(), "lua-map", e4, stage.Deps{})
		if err != nil {
			return err
		}
		b, err := json.Marshal(e5)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(os.Stdout, string(b)); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	Cmd.Flags().StringVarP(&cfgPath, "config", "c", "", "Path to config file (.cue)")
}
