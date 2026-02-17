package stage

import (
	"context"
)

func shellExecRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// If not enabled, passthrough
	opts := buildShellOptions(in)
	if !opts.enabled {
		return in, nil
	}

	out := in
	mode, _ := errorMode(in.Meta)
	n := len(in.Records)
	type res = shellExecRes
	workers := getWorkers(in.Meta)
	var envErrs []Error
	results := runIndexedParallel(n, workers, func(idx int) res {
		r := in.Records[idx]
		rec, envE, fatal := processShellRecord(ctx, r, opts, mode)
		return res{idx: idx, rec: rec, envE: envE, fatal: fatal}
	})
	var firstErr error
	for _, rr := range results {
		if rr.envE != nil {
			envErrs = append(envErrs, *rr.envE)
		}
		if rr.fatal != nil && firstErr == nil {
			firstErr = rr.fatal
		}
		out.Records[rr.idx] = rr.rec
	}
	if firstErr != nil {
		return Envelope{}, firstErr
	}
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
	}
	return out, nil
}

func init() { Register("shell-exec", shellExecRunner) }
