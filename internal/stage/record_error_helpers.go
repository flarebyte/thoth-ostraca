package stage

func recordFailure(rec Record, stageName, msg string, embed bool) (Record, *Error) {
	rr := rec
	if embed {
		rr.Error = &RecError{Stage: stageName, Message: msg}
	}
	return rr, &Error{Stage: stageName, Locator: rec.Locator, Message: msg}
}
