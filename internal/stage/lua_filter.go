package stage

import (
	"context"
)

func luaFilterRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Determine predicate
	pred := buildLuaPredicate(in)

	out := in
	out.Records = out.Records[:0]

	mode, _ := errorMode(in.Meta)
	n := len(in.Records)
	keeps := make([]bool, n)
	outs := make([]Record, n)
	var envErrs []Error
	workers := getWorkers(in.Meta)
	results := runIndexedParallel(n, workers, func(idx int) luaFilterRes {
		r := in.Records[idx]
		keep, outRec, envE, fatal := processLuaFilterRecord(r, pred, mode, in.Meta)
		return luaFilterRes{idx: idx, keep: keep, out: outRec, envE: envE, fatal: fatal}
	})
	var firstErr error
	for _, rr := range results {
		accumulateStageError(&envErrs, &firstErr, rr.envE, rr.fatal)
		if rr.keep {
			keeps[rr.idx] = true
			outs[rr.idx] = rr.out
		}
	}
	if firstErr != nil {
		return Envelope{}, firstErr
	}
	appendSanitizedErrors(&out, envErrs)
	for i := 0; i < n; i++ {
		if keeps[i] {
			out.Records = append(out.Records, outs[i])
		}
	}
	return out, nil
}

func init() { Register("lua-filter", luaFilterRunner) }
