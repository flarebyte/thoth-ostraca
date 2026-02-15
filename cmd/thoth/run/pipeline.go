package run

import (
	"context"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

// executePipeline runs the fixed Phase 1 pipeline for `thoth run`.
func executePipeline(ctx context.Context, cfgPath string) (stage.Envelope, error) {
	in := stage.Envelope{Records: []stage.Record{}, Meta: &stage.Meta{ConfigPath: cfgPath}}
	stages := []string{
		"validate-config",
		"discover-meta-files",
		"parse-validate-yaml",
		"validate-locators",
		"lua-filter",
		"lua-map",
		"shell-exec",
		"lua-postmap",
		"lua-reduce",
		"write-output",
	}
	return runStages(ctx, in, stages)
}

// output is handled by the write-output stage.
