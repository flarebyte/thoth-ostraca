package run

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

// executePipeline runs the fixed Phase 1 pipeline for `thoth run`.
func executePipeline(ctx context.Context, cfgPath string) (stage.Envelope, error) {
	in := stage.Envelope{Records: []any{}, Meta: &stage.Meta{ConfigPath: cfgPath}}
	e1, err := stage.Run(ctx, "validate-config", in, stage.Deps{})
	if err != nil {
		return stage.Envelope{}, err
	}
	e2, err := stage.Run(ctx, "discover-meta-files", e1, stage.Deps{})
	if err != nil {
		return stage.Envelope{}, err
	}
	e3, err := stage.Run(ctx, "parse-validate-yaml", e2, stage.Deps{})
	if err != nil {
		return stage.Envelope{}, err
	}
	e4, err := stage.Run(ctx, "lua-filter", e3, stage.Deps{})
	if err != nil {
		return stage.Envelope{}, err
	}
	e5, err := stage.Run(ctx, "lua-map", e4, stage.Deps{})
	if err != nil {
		return stage.Envelope{}, err
	}
	e6, err := stage.Run(ctx, "shell-exec", e5, stage.Deps{})
	if err != nil {
		return stage.Envelope{}, err
	}
	e7, err := stage.Run(ctx, "lua-postmap", e6, stage.Deps{})
	if err != nil {
		return stage.Envelope{}, err
	}
	e8, err := stage.Run(ctx, "lua-reduce", e7, stage.Deps{})
	if err != nil {
		return stage.Envelope{}, err
	}
	return e8, nil
}

// renderRunOutput prints final output to the provided writer while preserving
// existing behavior and exit conditions.
func renderRunOutput(out stage.Envelope, w io.Writer) error {
	// Lines mode
	if out.Meta != nil && out.Meta.Output != nil && out.Meta.Output.Lines {
		stripErrors := out.Meta != nil && out.Meta.Errors != nil && !out.Meta.Errors.EmbedErrors
		for _, r := range out.Records {
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
			if _, err := fmt.Fprint(w, buf.String()); err != nil {
				return err
			}
		}
		if out.Meta != nil && out.Meta.Errors != nil && out.Meta.Errors.Mode == "keep-going" {
			anyOK := false
			for _, r := range out.Records {
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

	// Full envelope mode
	if out.Meta != nil && out.Meta.Errors != nil && !out.Meta.Errors.EmbedErrors {
		for i, r := range out.Records {
			if rr, ok := r.(stage.Record); ok {
				rr.Error = nil
				out.Records[i] = rr
			}
		}
	}
	stage.SortEnvelopeErrors(&out)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, buf.String()); err != nil {
		return err
	}
	if out.Meta != nil && out.Meta.Errors != nil && out.Meta.Errors.Mode == "keep-going" {
		anyOK := false
		for _, r := range out.Records {
			if rec, ok := r.(stage.Record); ok {
				if rec.Error == nil {
					anyOK = true
					break
				}
			}
		}
		if !anyOK && len(out.Errors) > 0 {
			return fmt.Errorf("keep-going: no successful records")
		}
	}
	return nil
}
