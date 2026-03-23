// File Guide for dev/ai agents:
// Purpose: Decode shell stdout into structured JSON when the config explicitly asks for it.
// Responsibilities:
// - Parse captured stdout as JSON.
// - Fail clearly when stdout is missing or invalid.
// - Keep JSON decoding isolated from the rest of shell execution.
// Architecture notes:
// - JSON decoding is intentionally opt-in and isolated so raw stdout behavior remains backward compatible for non-JSON tools.
package stage

import (
	"encoding/json"
	"fmt"
)

func decodeShellStdoutJSON(stdout *string) (any, error) {
	if stdout == nil {
		return nil, fmt.Errorf("stdout missing")
	}
	var decoded any
	if err := json.Unmarshal([]byte(*stdout), &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}
