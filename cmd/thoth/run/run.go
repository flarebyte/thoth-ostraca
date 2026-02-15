package run

import (
	"bytes"
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
			// Optionally strip embedded errors
			stripErrors := e8.Meta != nil && e8.Meta.Errors != nil && !e8.Meta.Errors.EmbedErrors
			for _, r := range e8.Records {
				rec := r
				if stripErrors {
					if rr, ok := r.(stage.Record); ok {
						rr.Error = nil
						rec = rr
					}
				}
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				enc.SetEscapeHTML(false)
				if err := enc.Encode(rec); err != nil {
					return err
				}
				if _, err := fmt.Fprint(os.Stdout, buf.String()); err != nil {
					return err
				}
			}
			// In keep-going, determine exit code based on success presence
			if e8.Meta != nil && e8.Meta.Errors != nil && e8.Meta.Errors.Mode == "keep-going" {
				anyOK := false
				for _, r := range e8.Records {
					if rec, ok := r.(stage.Record); ok {
						if rec.Error == nil {
							anyOK = true
							break
						}
					}
				}
				if !anyOK {
					return fmt.Errorf("keep-going: no successful records")
				}
			}
			return nil
		}
		// Optionally strip embedded errors in full envelope mode
		if e8.Meta != nil && e8.Meta.Errors != nil && !e8.Meta.Errors.EmbedErrors {
			for i, r := range e8.Records {
				if rr, ok := r.(stage.Record); ok {
					rr.Error = nil
					e8.Records[i] = rr
				}
			}
		}
		// Ensure deterministic error ordering
		stage.SortEnvelopeErrors(&e8)
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(e8); err != nil {
			return err
		}
		if _, err := fmt.Fprint(os.Stdout, buf.String()); err != nil {
			return err
		}
		if e8.Meta != nil && e8.Meta.Errors != nil && e8.Meta.Errors.Mode == "keep-going" {
			anyOK := false
			for _, r := range e8.Records {
				if rec, ok := r.(stage.Record); ok {
					if rec.Error == nil {
						anyOK = true
						break
					}
				}
			}
			if !anyOK && len(e8.Errors) > 0 {
				return fmt.Errorf("keep-going: no successful records")
			}
		}
		return nil
	},
}

func init() {
	Cmd.Flags().StringVarP(&cfgPath, "config", "c", "", "Path to config file (.cue)")
}
