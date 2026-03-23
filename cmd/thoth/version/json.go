// File Guide for dev/ai agents:
// Purpose: Provide the small JSON encoding helper used by the version command.
// Responsibilities:
// - Encode a value to JSON on the provided writer.
// - Apply the stable pretty indent used by `thoth version --json`.
// - Keep JSON output concerns out of the version command body.
// Architecture notes:
// - This helper exists separately so the version command can stay focused on data selection rather than serialization details.
// - Indented JSON is part of the user-facing contract for version output, so the format is fixed here instead of left to callers.
package version

import (
	"encoding/json"
	"io"
)

func encodeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
