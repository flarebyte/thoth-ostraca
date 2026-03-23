// File Guide for dev/ai agents:
// Purpose: Append envelope-level errors in a single place while preserving deterministic output order.
// Responsibilities:
// - Add sanitized stage errors to the envelope.
// - Sort envelope errors after appending them.
// - Avoid repeated error-list boilerplate across stage runners.
// Architecture notes:
// - Error sanitization and ordering are intentional output-contract behavior, not incidental cleanup.
// - Keep this helper tiny; richer error semantics belong in the stage that produced the error.
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
