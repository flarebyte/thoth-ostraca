// File Guide for dev/ai agents:
// Purpose: Provide the small helpers that run one or many stage names under the `thoth run` command.
// Responsibilities:
// - Execute an ordered list of stage names in sequence.
// - Route individual stage execution through the progress reporter when enabled.
// - Supply the shared stage dependencies used by CLI runs.
// Architecture notes:
// - These helpers keep pipeline.go focused on action decisions instead of repeating stage loop boilerplate.
// - Progress wrapping is injected here so stage implementations remain unaware of CLI-specific reporting behavior.
package run

import (
	"context"
	"os"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

// runStages executes the provided list of stage names in order.
func runStages(ctx context.Context, in stage.Envelope, stages []string) (stage.Envelope, error) {
	out := in
	var err error
	deps := stage.Deps{Stderr: os.Stderr}
	for _, name := range stages {
		out, err = runStage(ctx, name, out, deps)
		if err != nil {
			return stage.Envelope{}, err
		}
	}
	return out, nil
}

func runStage(ctx context.Context, name string, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
	if reporter, _ := stage.ProgressReporterFromContext(ctx).(*progressReporter); reporter != nil {
		return reporter.runStage(ctx, name, in, deps)
	}
	return stage.Run(ctx, name, in, deps)
}
