package stage

func appendSanitizedErrors(out *Envelope, envErrs []Error) {
	if len(envErrs) == 0 {
		return
	}
	for _, e := range envErrs {
		out.Errors = append(out.Errors, sanitizedError(e))
	}
	SortEnvelopeErrors(out)
}
