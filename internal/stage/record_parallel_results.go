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
