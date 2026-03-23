// File Guide for dev/ai agents:
// Purpose: Create consistent record-plus-envelope error results from one stage failure.
// Responsibilities:
// - Build the embedded record error when embedErrors is enabled.
// - Build the matching envelope error with locator context.
// - Keep stage failure shaping consistent across runners.
// Architecture notes:
// - The paired record/envelope error behavior is intentional; keep-going mode needs both views.
// - This helper should stay mechanical; message formatting belongs to upstream stage-specific logic.
package stage

func recordFailure(rec Record, stageName, msg string, embed bool) (Record, *Error) {
	rr := rec
	if embed {
		rr.Error = &RecError{Stage: stageName, Message: msg}
	}
	return rr, &Error{Stage: stageName, Locator: rec.Locator, Message: msg}
}
