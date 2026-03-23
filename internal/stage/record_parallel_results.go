// File Guide for dev/ai agents:
// Purpose: Merge per-record parallel stage results back into an envelope while preserving error semantics.
// Responsibilities:
// - Collect embedded record updates produced by parallel workers.
// - Accumulate sanitized envelope errors from parallel stages.
// - Return the first fatal error while preserving keep-going record mutations when appropriate.
// Architecture notes:
// - Parallel stages use this shared merger so shell, locator validation, and similar stages follow the same error-handling contract.
// - Envelope errors are sanitized here before appending so callers do not have to duplicate that consistency logic.
package stage

type recordParallelRes struct {
	idx   int
	rec   Record
	envE  *Error
	fatal error
}

func mergeRecordParallelResults(out Envelope, results []recordParallelRes) (Envelope, error) {
	var envErrs []Error
	var firstErr error
	for _, rr := range results {
		if rr.envE != nil {
			envErrs = append(envErrs, sanitizedError(*rr.envE))
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
		SortEnvelopeErrors(&out)
	}
	return out, nil
}
