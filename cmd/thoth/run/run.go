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
		e6, err := stage.Run(context.Background(), "shell-exec", e5, stage.Deps{})
		if err != nil {
			return err
		}
		e7, err := stage.Run(context.Background(), "lua-postmap", e6, stage.Deps{})
		if err != nil {
			return err
		}
		e8, err := stage.Run(context.Background(), "lua-reduce", e7, stage.Deps{})
		if err != nil {
			return err
		}
		// Output modes
		if e8.Meta != nil && e8.Meta.Output != nil && e8.Meta.Output.Lines {
			for _, r := range e8.Records {
				b, err := json.Marshal(r)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(os.Stdout, string(b)); err != nil {
					return err
				}
			}
			return nil
		}
		b, err := json.Marshal(e8)
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
