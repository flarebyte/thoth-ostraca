package version

import (
	"bytes"
	"testing"
)

func TestEncodeJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := encodeJSON(&buf, map[string]any{"a": 1}); err != nil {
		t.Fatalf("encodeJSON err: %v", err)
	}
	if got := buf.String(); got == "" || got[len(got)-1] != '\n' {
		t.Fatalf("expected trailing newline JSON, got %q", got)
	}
}
