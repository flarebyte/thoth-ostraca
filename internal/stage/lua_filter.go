// File Guide for dev/ai agents:
// Purpose: Run the Lua filter stage over records and keep only those whose predicate returns true.
// Responsibilities:
// - Build the configured Lua predicate for the current envelope.
// - Evaluate the predicate across records in parallel and collect keep/reject decisions.
// - Preserve filtered input metadata for diff-meta after record selection changes.
// Architecture notes:
// - The stage keeps index-position bookkeeping separate from output assembly so parallel evaluation can remain fast while final record order stays deterministic.
// - diff-meta updates meta.inputs here because later diff stages read the filtered locator set from envelope metadata, not directly from records.
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
	if out.Meta != nil && out.Meta.Config != nil && out.Meta.Config.Action == "diff-meta" {
		inputs := make([]string, 0, len(out.Records))
		for _, rec := range out.Records {
			inputs = append(inputs, rec.Locator)
		}
		out.Meta.Inputs = inputs
	}
	return out, nil
}

func init() { Register("lua-filter", luaFilterRunner) }
