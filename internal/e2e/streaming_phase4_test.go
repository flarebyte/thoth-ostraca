package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePhase4Repo(t *testing.T, total int) string {
	t.Helper()
	tmpRoot := filepath.Join(repoRoot(), "temp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	repo, err := os.MkdirTemp(tmpRoot, "phase4-stream-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	for i := 0; i < total; i++ {
		dir := filepath.Join(repo, "tree", fmt.Sprintf("%02d", i%17))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir repo tree: %v", err)
		}
		p := filepath.Join(dir, fmt.Sprintf("f%04d.thoth.yaml", i))
		body := fmt.Sprintf("locator: \"item/%04d\"\nmeta:\n  i: %d\n  enabled: true\n", i, i)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("write repo file: %v", err)
		}
	}
	return repo
}

func writePhase4Config(t *testing.T, repo string, workers int, lines bool, maxRecords int) string {
	t.Helper()
	tmpRoot := filepath.Join(repoRoot(), "temp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	cfg, err := os.CreateTemp(tmpRoot, "phase4-*.cue")
	if err != nil {
		t.Fatalf("mktemp cfg: %v", err)
	}
	_ = cfg.Close()
	content := fmt.Sprintf("{\n  configVersion: \"v0\"\n  action: \"pipeline\"\n  discovery: { root: %q }\n  output: { lines: %t }\n  limits: { maxRecordsInMemory: %d }\n  workers: %d\n  errors: { mode: \"fail-fast\" }\n}\n",
		filepath.ToSlash(repo), lines, maxRecords, workers)
	if err := os.WriteFile(cfg.Name(), []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	return cfg.Name()
}

func TestStreamingNDJSON_MaxRecordsLimitRespected(t *testing.T) {
	bin := buildThoth(t)
	repo := writePhase4Repo(t, 200)
	cfg1 := writePhase4Config(t, repo, 1, true, 50)
	cfg8 := writePhase4Config(t, repo, 8, true, 50)

	r1 := runCmd(t, bin, "run", "--config", cfg1)
	r8 := runCmd(t, bin, "run", "--config", cfg8)

	if r1.code != 0 || len(r1.stderr) != 0 {
		t.Fatalf("workers=1 failed: code=%d stderr=%s", r1.code, string(r1.stderr))
	}
	if r8.code != 0 || len(r8.stderr) != 0 {
		t.Fatalf("workers=8 failed: code=%d stderr=%s", r8.code, string(r8.stderr))
	}
	lines := strings.Split(strings.TrimSpace(string(r1.stdout)), "\n")
	if len(lines) != 200 {
		t.Fatalf("line count mismatch: got %d want 200", len(lines))
	}
	if string(r1.stdout) != string(r8.stdout) {
		t.Fatalf("ndjson output drift between workers")
	}
}

func TestBufferedMode_OverMaxRecordsFails(t *testing.T) {
	bin := buildThoth(t)
	repo := writePhase4Repo(t, 200)
	cfg := writePhase4Config(t, repo, 1, false, 50)

	r := runCmd(t, bin, "run", "--config", cfg)
	if r.code == 0 {
		t.Fatalf("expected non-zero exit")
	}
	errOut := string(r.stderr)
	if !strings.Contains(errOut, "maxRecordsInMemory") {
		t.Fatalf("stderr missing maxRecordsInMemory: %s", errOut)
	}
	if !strings.Contains(errOut, "output.lines=true") {
		t.Fatalf("stderr missing lines suggestion: %s", errOut)
	}
}
