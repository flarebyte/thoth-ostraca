package stage

import (
	"path/filepath"
	"testing"
)

// Policy guard: E2E tests belong in script/e2e (TypeScript).
// Keep a small allowlist for legacy Go E2E files until they are migrated.
func TestE2EPolicy_NoNewGoE2ETests(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "e2e", "*_test.go"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	allowed := map[string]bool{
		filepath.Clean(filepath.Join("..", "e2e", "determinism_test.go")): true,
	}
	for _, m := range matches {
		if !allowed[filepath.Clean(m)] {
			t.Fatalf("go e2e test not allowed: %s (use script/e2e/*.test.ts)", m)
		}
	}
}
