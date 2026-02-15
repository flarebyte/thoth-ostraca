package run

import (
	"context"
	"fmt"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

// executePipeline runs the fixed Phase 1 pipeline for `thoth run`.
func executePipeline(ctx context.Context, cfgPath string) (stage.Envelope, error) {
	// Always start by validating config to determine action
	in := stage.Envelope{Records: []stage.Record{}, Meta: &stage.Meta{ConfigPath: cfgPath}}
	out, err := stage.Run(ctx, "validate-config", in, stage.Deps{})
	if err != nil {
		return stage.Envelope{}, err
	}
	action := "pipeline"
	if out.Meta != nil && out.Meta.Config != nil && out.Meta.Config.Action != "" {
		action = out.Meta.Config.Action
	}
	switch action {
	case "pipeline", "nop":
		stages := []string{
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
		return runStages(ctx, out, stages)
	case "validate":
		stages := []string{
			"discover-meta-files",
			"parse-validate-yaml",
			"validate-locators",
			"write-output",
		}
		return runStages(ctx, out, stages)
	case "create-meta":
		stages := []string{
			"discover-input-files",
			"enrich-fileinfo",
			"write-meta-files",
			"write-output",
		}
		return runStages(ctx, out, stages)
	case "update-meta":
		stages := []string{
			"discover-input-files",
			"enrich-fileinfo",
			"load-existing-meta",
			"merge-meta",
			"write-updated-meta-files",
			"write-output",
		}
		return runStages(ctx, out, stages)
	case "diff-meta":
		stages := []string{
			"discover-input-files",
			"enrich-fileinfo",
			"discover-meta-files",
			"compute-meta-diff",
			"write-output",
		}
		return runStages(ctx, out, stages)
	default:
		// Should not happen; validate-config already enforced
		return stage.Envelope{}, fmt.Errorf("invalid action")
	}
}

// output is handled by the write-output stage.
