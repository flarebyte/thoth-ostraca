// File Guide for dev/ai agents:
// Purpose: Define the minimal per-record error payload and the shared error-mode switch.
// Responsibilities:
// - Describe the embedded record error shape.
// - Resolve fail-fast vs keep-going behavior from envelope meta.
// - Expose whether record errors should be embedded.
// Architecture notes:
// - `errorMode` centralizes defaulting so stage runners do not drift on fail-fast semantics.
// - `RecError` intentionally stays smaller than envelope `Error`; the locator already lives on the record.
package stage

// RecError is a per-record error payload.
type RecError struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
}

func errorMode(meta *Meta) (mode string, embed bool) {
	mode = "fail-fast"
	if meta != nil && meta.Errors != nil {
		if meta.Errors.Mode != "" {
			mode = meta.Errors.Mode
		}
		embed = meta.Errors.EmbedErrors
	}
	return
}
