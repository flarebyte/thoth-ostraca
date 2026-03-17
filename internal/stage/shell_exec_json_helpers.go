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
