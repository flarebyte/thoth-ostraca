// File Guide for dev/ai agents:
// Purpose: Normalize human-facing error text into deterministic, compact strings for records and envelopes.
// Responsibilities:
// - Collapse whitespace in error messages.
// - Provide the shared sanitizedError helper for envelope errors.
// - Supply a stable fallback message for empty errors.
// Architecture notes:
// - Whitespace collapsing is deliberate because tests assert deterministic error bytes across platforms and runtimes.
// - Record-level embedded errors can preserve richer formatting before envelope sanitization; do not assume both paths are identical.
package stage

import "strings"

func sanitizeErrorMessage(msg string) string {
	s := strings.Join(strings.Fields(msg), " ")
	if s == "" {
		return "error"
	}
	return s
}

func sanitizedError(e Error) Error {
	e.Message = sanitizeErrorMessage(e.Message)
	return e
}
