package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLuaSandboxRepo(t *testing.T, total int) string {
	t.Helper()
	root := repoRoot()
	tmpRoot := filepath.Join(root, "temp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	repo, err := os.MkdirTemp(tmpRoot, "phase4-lua-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	for i := 0; i < total; i++ {
		dir := filepath.Join(repo, "data")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		p := filepath.Join(dir, fmt.Sprintf("f%04d.thoth.yaml", i))
		body := fmt.Sprintf("locator: \"item/%04d\"\nmeta:\n  i: %d\n  enabled: true\n", i, i)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	return repo
}

func writeLuaSandboxConfig(t *testing.T, repo string, workers int, mapInline string, timeoutMs, instructionLimit int) string {
	t.Helper()
	root := repoRoot()
	tmpRoot := filepath.Join(root, "temp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	cfg, err := os.CreateTemp(tmpRoot, "lua-sandbox-*.cue")
	if err != nil {
		t.Fatalf("mktemp cfg: %v", err)
	}
	_ = cfg.Close()
	content := fmt.Sprintf("{\n  configVersion: \"v0\"\n  action: \"pipeline\"\n  discovery: { root: %q }\n  workers: %d\n  errors: { mode: \"fail-fast\" }\n  lua: {\n    timeoutMs: %d\n    instructionLimit: %d\n    memoryLimitBytes: 8388608\n    deterministicRandom: true\n    libs: { base: true, table: true, string: true, math: true }\n  }\n  map: { inline: %q }\n}\n", filepath.ToSlash(repo), workers, timeoutMs, instructionLimit, mapInline)
	if err := os.WriteFile(cfg.Name(), []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	return cfg.Name()
}

func TestLuaSandboxE2E_FailFastInfiniteLoop(t *testing.T) {
	bin := buildThoth(t)
	repo := writeLuaSandboxRepo(t, 2)
	cfg := writeLuaSandboxConfig(t, repo, 1, "return (function() while true do end end)()", 20, 100000000)
	r := runCmd(t, bin, "run", "--config", cfg)
	if r.code == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if string(r.stdout) != "" {
		t.Fatalf("expected empty stdout")
	}
	if got := string(r.stderr); got == "" || !strings.Contains(got, "lua-map: sandbox timeout") {
		t.Fatalf("unexpected stderr: %s", got)
	}
}

func TestLuaSandboxE2E_DeterministicRandomAcrossWorkers(t *testing.T) {
	bin := buildThoth(t)
	repo := writeLuaSandboxRepo(t, 50)
	code := "return { locator = locator, r = math.random(1, 1000000) }"
	cfg1 := writeLuaSandboxConfig(t, repo, 1, code, 2000, 1000000)
	cfg8 := writeLuaSandboxConfig(t, repo, 8, code, 2000, 1000000)
	r1 := runCmd(t, bin, "run", "--config", cfg1)
	r8 := runCmd(t, bin, "run", "--config", cfg8)
	if r1.code != 0 || len(r1.stderr) != 0 {
		t.Fatalf("workers=1 failed: code=%d stderr=%s", r1.code, string(r1.stderr))
	}
	if r8.code != 0 || len(r8.stderr) != 0 {
		t.Fatalf("workers=8 failed: code=%d stderr=%s", r8.code, string(r8.stderr))
	}
	if normalizeWorkersOut(r1.stdout) != normalizeWorkersOut(r8.stdout) {
		t.Fatalf("stdout drift workers=1 vs workers=8")
	}
}

func normalizeWorkersOut(b []byte) string {
	var v map[string]any
	if err := json.Unmarshal(b, &v); err != nil {
		return string(b)
	}
	if meta, ok := v["meta"].(map[string]any); ok {
		delete(meta, "workers")
	}
	out, err := json.Marshal(v)
	if err != nil {
		return string(b)
	}
	return string(append(out, '\n'))
}
