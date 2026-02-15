package run

import (
	"context"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

// runStages executes the provided list of stage names in order.
func runStages(ctx context.Context, in stage.Envelope, stages []string) (stage.Envelope, error) {
	out := in
	var err error
	for _, name := range stages {
		out, err = stage.Run(ctx, name, out, stage.Deps{})
		if err != nil {
			return stage.Envelope{}, err
		}
	}
	return out, nil
}
