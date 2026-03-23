// File Guide for dev/ai agents:
// Purpose: Provide test-only helpers that build temporary config files and execute validate-config in unit tests.
// Responsibilities:
// - Write inline config content to a temporary fixture path under temp/.
// - Build the minimal input envelope needed by the validate-config stage.
// - Invoke the stage runner and return its result to tests.
// Architecture notes:
// - This helper is production code only because Go test files in this package reuse it across multiple suites.
// - It writes to temp/ rather than t.TempDir so existing test fixtures and golden expectations can refer to stable relative paths when needed.
package stage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func runValidateConfigWithContent(t *testing.T, fileName, content string) (Envelope, error) {
	t.Helper()
	_ = os.MkdirAll("temp", 0o755)
	cfg := filepath.Join("temp", fileName)
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	in := Envelope{Records: []Record{}, Meta: &Meta{ConfigPath: cfg}}
	return Run(context.Background(), "validate-config", in, Deps{})
}
