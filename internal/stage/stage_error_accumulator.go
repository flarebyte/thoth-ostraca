// File Guide for dev/ai agents:
// Purpose: Provide the tiny shared helper that accumulates envelope and fatal errors during parallel stage processing.
// Responsibilities:
// - Append stage-level envelope errors when present.
// - Preserve the first fatal error seen across worker results.
// - Keep this repeated pattern out of stage-specific loops.
// Architecture notes:
// - This helper is deliberately narrow because it exists only to reduce repeated bookkeeping in parallel stage mergers.
// - First-error wins is intentional for fatal errors so fail-fast behavior stays deterministic.
package stage

func accumulateStageError(envErrs *[]Error, firstErr *error, envE *Error, fatal error) {
	if envE != nil {
		*envErrs = append(*envErrs, *envE)
	}
	if fatal != nil && *firstErr == nil {
		*firstErr = fatal
	}
}
