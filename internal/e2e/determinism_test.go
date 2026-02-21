package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/config"
	"github.com/flarebyte/thoth-ostraca/internal/testutil"
)

type runResult struct {
	code   int
	stdout []byte
	stderr []byte
}

func repoRoot() string {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		return filepath.Clean(filepath.Join("..", ".."))
	}
	return root
}

func buildThoth(t *testing.T) string {
	t.Helper()
	root := repoRoot()
	binDir := filepath.Join(root, ".e2e-bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	bin := filepath.Join(binDir, "thoth")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-mod=vendor", "-o", bin, "./cmd/thoth")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, string(out))
	}
	return bin
}

func runCmd(t *testing.T, bin string, args ...string) runResult {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = repoRoot()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = -1
		}
	}
	return runResult{code: code, stdout: stdout.Bytes(), stderr: stderr.Bytes()}
}

func assertStable(t *testing.T, runs []runResult) {
	t.Helper()
	if len(runs) < 2 {
		t.Fatalf("need >=2 runs")
	}
	a := runs[0]
	for i, r := range runs[1:] {
		if r.code != a.code {
			t.Fatalf("exit code drift at run %d: %d vs %d", i+1, r.code, a.code)
		}
		if !bytes.Equal(r.stdout, a.stdout) {
			t.Fatalf("stdout drift at run %d", i+1)
		}
		if !bytes.Equal(r.stderr, a.stderr) {
			t.Fatalf("stderr drift at run %d", i+1)
		}
	}
}

func normalizeWorkersJSON(b []byte) []byte {
	var v map[string]any
	if err := json.Unmarshal(b, &v); err != nil {
		return b
	}
	if meta, ok := v["meta"].(map[string]any); ok {
		delete(meta, "workers")
	}
	out, err := json.Marshal(v)
	if err != nil {
		return b
	}
	return append(out, '\n')
}

func TestDeterminism_Pipeline_MultiRuns(t *testing.T) {
	root := repoRoot()
	bin := buildThoth(t)
	cfg := filepath.Join(root, "testdata", "configs", "keep1_embed_true.cue")
	var runs []runResult
	for i := 0; i < 5; i++ {
		runs = append(runs, runCmd(t, bin, "run", "--config", cfg))
	}
	assertStable(t, runs)
}

func TestDeterminism_Pipeline_Workers(t *testing.T) {
	root := repoRoot()
	bin := buildThoth(t)
	cfg1 := filepath.Join(root, "testdata", "configs", "workers1.cue")
	cfg2 := filepath.Join(root, "testdata", "configs", "workers2.cue")
	// Generate workers=8 config
	cfg8 := filepath.Join(root, "temp", "workers8_tmp.cue")
	_ = os.MkdirAll(filepath.Join(root, "temp"), 0o755)
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"nop\"\n  discovery: { root: \"testdata/repos/keep1\" }\n  errors: { mode: \"keep-going\", embedErrors: true }\n  workers: 8\n  filter: { inline: \"return true\" }\n  map: { inline: \"if (meta and meta.name) == \\\"LuaErr\\\" then error(\\\"boom\\\") end; return { locator = locator, name = meta and meta.name }\" }\n}"
	if err := os.WriteFile(cfg8, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg8: %v", err)
	}
	r1 := runCmd(t, bin, "run", "--config", cfg1)
	r2 := runCmd(t, bin, "run", "--config", cfg2)
	r8 := runCmd(t, bin, "run", "--config", cfg8)
	if r1.code != 0 || r2.code != 0 || r8.code != 0 {
		t.Fatalf("expected zero exit codes")
	}
	if !bytes.Equal(normalizeWorkersJSON(r1.stdout), normalizeWorkersJSON(r2.stdout)) {
		t.Fatalf("stdout drift workers 1 vs 2")
	}
	if !bytes.Equal(normalizeWorkersJSON(r2.stdout), normalizeWorkersJSON(r8.stdout)) {
		t.Fatalf("stdout drift workers 2 vs 8")
	}
	if len(r1.stderr) != 0 || len(r2.stderr) != 0 || len(r8.stderr) != 0 {
		t.Fatalf("expected empty stderr")
	}
}

func TestDeterminism_Validate(t *testing.T) {
	root := repoRoot()
	bin := buildThoth(t)
	cfg := filepath.Join(root, "testdata", "configs", "validate_only_ok.cue")
	var runs []runResult
	for i := 0; i < 5; i++ {
		runs = append(runs, runCmd(t, bin, "run", "--config", cfg))
	}
	assertStable(t, runs)
}

func TestDeterminism_DiffMeta(t *testing.T) {
	root := repoRoot()
	bin := buildThoth(t)
	cfg := filepath.Join(root, "testdata", "configs", "diff1.cue")
	var runs []runResult
	for i := 0; i < 5; i++ {
		runs = append(runs, runCmd(t, bin, "run", "--config", cfg))
	}
	assertStable(t, runs)
}

func TestDeterminism_CreateMeta(t *testing.T) {
	root := repoRoot()
	bin := buildThoth(t)
	src := filepath.Join(root, "testdata", "repos", "create1")
	cfgT := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"create-meta\"\n  discovery: { root: \"%s\" }\n}"
	repo := filepath.Join(root, "temp", "create_det_repo")
	assertMetaActionDeterminism(t, bin, src, repo, cfgT)
}

func TestDeterminism_UpdateMeta(t *testing.T) {
	root := repoRoot()
	bin := buildThoth(t)
	src := filepath.Join(root, "testdata", "repos", "update1")
	cfgT := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"update-meta\"\n  discovery: { root: \"%s\" }\n}"
	repo := filepath.Join(root, "temp", "update_det_repo")
	assertMetaActionDeterminism(t, bin, src, repo, cfgT)
}

func assertMetaActionDeterminism(t *testing.T, bin, src, repo, cfgTemplate string) {
	t.Helper()
	var baseOut []byte
	for i := 0; i < 5; i++ {
		if err := testutil.CopyTree(src, repo); err != nil {
			t.Fatalf("copy: %v", err)
		}
		cfg := filepath.Join(repo, "tmp.cue")
		data := []byte(fmtSprintf(cfgTemplate, filepath.ToSlash(repo)))
		if err := os.WriteFile(cfg, data, 0o644); err != nil {
			t.Fatalf("write cfg: %v", err)
		}
		r := runCmd(t, bin, "run", "--config", cfg)
		if i == 0 {
			baseOut = r.stdout
		}
		if !bytes.Equal(r.stdout, baseOut) {
			t.Fatalf("stdout drift run %d", i)
		}
		if r.code != 0 || len(r.stderr) != 0 {
			t.Fatalf("unexpected status/stderr")
		}
	}
}

func fmtSprintf(f string, a ...any) string { return fmt.Sprintf(f, a...) }
