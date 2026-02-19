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
	results := runIndexedParallel(n, workers, func(idx int) recordParallelRes {
		r := in.Records[idx]
		rec, envE, fatal := processShellRecord(ctx, r, opts, mode)
		return recordParallelRes{idx: idx, rec: rec, envE: envE, fatal: fatal}
	})
	return mergeRecordParallelResults(out, results)
}

func init() { Register("shell-exec", shellExecRunner) }
