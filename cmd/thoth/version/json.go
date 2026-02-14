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

