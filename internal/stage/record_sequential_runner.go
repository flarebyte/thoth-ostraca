// File Guide for dev/ai agents:
// Purpose: Provide the common sequential record-stage loop used by create, load, and other ordered file operations.
// Responsibilities:
// - Run a per-record function over the envelope records in order.
// - Apply keep-going versus fail-fast error behavior consistently.
// - Append sanitized envelope errors and embedded record errors when configured.
// Architecture notes:
// - Sequential stages share this helper so ordered filesystem operations do not each reimplement the same error-mode logic.
// - The helper sanitizes both returned errors and embedded errors centrally to keep user-visible output stable.
package stage

import "fmt"

func runSequentialRecordStage(
	in Envelope,
	stage string,
	mode string,
	embed bool,
	fn func(Record) (Record, *Error, error),
) (Envelope, error) {
	out := in
	var envErrs []Error
	for i, r := range in.Records {
		if r.Error != nil {
			continue
		}
		rec, envE, err := fn(r)
		if envE != nil {
			se := sanitizedError(*envE)
			envErrs = append(envErrs, se)
			envE = &se
		}
		if err != nil {
			if mode == "keep-going" {
				if embed {
					rec = r
					msg := sanitizeErrorMessage(err.Error())
					if envE != nil {
						msg = envE.Message
					}
					rec.Error = &RecError{Stage: stage, Message: msg}
				}
				out.Records[i] = rec
				continue
			}
			return Envelope{}, fmt.Errorf("%s: %s", stage, sanitizeErrorMessage(err.Error()))
		}
		out.Records[i] = rec
	}
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
		SortEnvelopeErrors(&out)
	}
	return out, nil
}
