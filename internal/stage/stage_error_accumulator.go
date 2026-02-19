package stage

func accumulateStageError(envErrs *[]Error, firstErr *error, envE *Error, fatal error) {
	if envE != nil {
		*envErrs = append(*envErrs, *envE)
	}
	if fatal != nil && *firstErr == nil {
		*firstErr = fatal
	}
}
