package stage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestPerfSmoke_DiscoveryAndParse(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	const n = 200
	for i := 0; i < n; i++ {
		dir := filepath.Join(root, fmt.Sprintf("p%02d", i%10))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		name := fmt.Sprintf("f%03d.thoth.yaml", i)
		body := fmt.Sprintf("locator: item-%03d\nmeta:\n  v: %d\n", i, i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write yaml failed: %v", err)
		}
	}

	in := Envelope{
		Records: []Record{},
		Meta: &Meta{
			Discovery: &DiscoveryMeta{Root: root},
			Limits:    &LimitsMeta{MaxRecordsInMemory: defaultMaxRecordsInMemory},
			Workers:   1,
		},
	}

	start := time.Now()
	discovered, err := discoverRunner(context.Background(), in, Deps{})
	if err != nil {
		t.Fatalf("discover-meta-files failed: %v", err)
	}
	parsed, err := parseValidateYAMLRunner(context.Background(), discovered, Deps{})
	if err != nil {
		t.Fatalf("parse-validate-yaml failed: %v", err)
	}
	elapsed := time.Since(start)

	if got := len(parsed.Records); got != n {
		t.Fatalf("unexpected record count: got %d want %d", got, n)
	}

	const budget = 10 * time.Second
	if elapsed > budget {
		if runtime.GOMAXPROCS(0) <= 1 {
			t.Skipf("perf smoke skipped on constrained runtime: %s", elapsed)
		}
		t.Fatalf("perf smoke exceeded budget: %s > %s", elapsed, budget)
	}
}
