// File Guide for dev/ai agents:
// Purpose: Run the shell-exec stage across records when shell analysis is enabled in the config.
// Responsibilities:
// - Build and validate stage-level shell options from envelope metadata.
// - Dispatch per-record shell execution in parallel.
// - Emit progress events and merge shell results back into the envelope.
// Architecture notes:
// - This file orchestrates the stage only; rendering, spawning, and result shaping live in helper files.
// - Progress reporting is kept here because it depends on stage-level worker fan-out, not individual shell mechanics.
package stage

import (
	"context"
	"fmt"
)

func shellExecRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// If not enabled, passthrough
	opts := buildShellOptions(in)
	if !opts.enabled {
		return in, nil
	}
	if err := validateShellOptions(opts); err != nil {
		return Envelope{}, fmt.Errorf("shell-exec: %v", err)
	}

	out := in
	mode, _ := errorMode(in.Meta)
	n := len(in.Records)
	workers := getWorkers(in.Meta)
	reporter := ProgressReporterFromContext(ctx)
	completed := 0
	results := runIndexedParallel(n, workers, func(idx int) recordParallelRes {
		r := in.Records[idx]
		rec, envE, fatal := processShellRecord(ctx, r, opts, mode)
		return recordParallelRes{idx: idx, rec: rec, envE: envE, fatal: fatal}
	})
	if reporter != nil {
		for range results {
			completed++
			reporter.ReportProgress(ProgressEvent{
				Stage:     "shell-exec",
				Event:     "progress",
				Completed: completed,
				Total:     n,
				Rejected:  0,
				Errors:    0,
			})
		}
	}
	return mergeRecordParallelResults(out, results)
}

func init() { Register("shell-exec", shellExecRunner) }
