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
	if reporter := progressReporterFromContext(ctx); reporter != nil {
		return reporter.runStage(ctx, name, in, deps)
	}
	return stage.Run(ctx, name, in, deps)
}
