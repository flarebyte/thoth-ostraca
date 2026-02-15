package run

import (
	"context"
	"io"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

// executePipeline runs the fixed Phase 1 pipeline for `thoth run`.
func executePipeline(ctx context.Context, cfgPath string) (stage.Envelope, error) {
	in := stage.Envelope{Records: []stage.Record{}, Meta: &stage.Meta{ConfigPath: cfgPath}}
	stages := []string{
		"validate-config",
		"discover-meta-files",
		"parse-validate-yaml",
		"lua-filter",
		"lua-map",
		"shell-exec",
		"lua-postmap",
		"lua-reduce",
	}
	return runStages(ctx, in, stages)
}

// renderRunOutput prints final output to the provided writer while preserving
// existing behavior and exit conditions.
func renderRunOutput(out stage.Envelope, w io.Writer) error {
	// Attach contract version to final outputs
	if out.Meta == nil {
		out.Meta = &stage.Meta{}
	}
	out.Meta.ContractVersion = "1"
	if isLinesMode(out) {
		return renderLines(out, w)
	}
	return renderEnvelope(out, w)
}
