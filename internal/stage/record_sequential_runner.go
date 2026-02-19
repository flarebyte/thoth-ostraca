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
			envErrs = append(envErrs, *envE)
		}
		if err != nil {
			if mode == "keep-going" {
				if embed {
					rec = r
					msg := err.Error()
					if envE != nil {
						msg = envE.Message
					}
					rec.Error = &RecError{Stage: stage, Message: msg}
				}
				out.Records[i] = rec
				continue
			}
			return Envelope{}, fmt.Errorf("%w", err)
		}
		out.Records[i] = rec
	}
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
	}
	return out, nil
}
