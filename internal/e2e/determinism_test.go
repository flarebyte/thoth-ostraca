package e2e

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

type runResult struct {
	code   int
	stdout []byte
	stderr []byte
}

func buildThoth(t *testing.T) string {
	t.Helper()
	binDir := filepath.Join(".e2e-bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	bin := filepath.Join(binDir, "thoth")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-mod=vendor", "-o", bin, "./cmd/thoth")
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
	cmd.Dir = ""
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

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	_ = os.RemoveAll(dst)
	if err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		out := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	}); err != nil {
		t.Fatalf("copy: %v", err)
	}
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

func TestDeterminism_Pipeline_MultiRuns(t *testing.T) {
	bin := buildThoth(t)
	cfg := filepath.Join("testdata", "configs", "keep1_embed_true.cue")
	var runs []runResult
	for i := 0; i < 5; i++ {
		runs = append(runs, runCmd(t, bin, "run", "--config", cfg))
	}
	assertStable(t, runs)
}

func TestDeterminism_Pipeline_Workers(t *testing.T) {
	bin := buildThoth(t)
	cfg1 := filepath.Join("testdata", "configs", "workers1.cue")
	cfg2 := filepath.Join("testdata", "configs", "workers2.cue")
	// Generate workers=8 config
	cfg8 := filepath.Join("temp", "workers8_tmp.cue")
	_ = os.MkdirAll("temp", 0o755)
	content := "{\n  configVersion: \"v0\"\n  action: \"nop\"\n  discovery: { root: \"testdata/repos/keep1\" }\n  errors: { mode: \"keep-going\", embedErrors: true }\n  workers: 8\n  filter: { inline: \"return true\" }\n  map: { inline: \"return { locator = locator, name = meta and meta.name }\" }\n}"
	if err := os.WriteFile(cfg8, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg8: %v", err)
	}
	r1 := runCmd(t, bin, "run", "--config", cfg1)
	r2 := runCmd(t, bin, "run", "--config", cfg2)
	r8 := runCmd(t, bin, "run", "--config", cfg8)
	assertStable(t, []runResult{r1, r2, r8})
}

func TestDeterminism_Validate(t *testing.T) {
	bin := buildThoth(t)
	cfg := filepath.Join("testdata", "configs", "validate_only_ok.cue")
	var runs []runResult
	for i := 0; i < 5; i++ {
		runs = append(runs, runCmd(t, bin, "run", "--config", cfg))
	}
	assertStable(t, runs)
}

func TestDeterminism_DiffMeta(t *testing.T) {
	bin := buildThoth(t)
	cfg := filepath.Join("testdata", "configs", "diff1.cue")
	var runs []runResult
	for i := 0; i < 5; i++ {
		runs = append(runs, runCmd(t, bin, "run", "--config", cfg))
	}
	assertStable(t, runs)
}

func TestDeterminism_CreateMeta(t *testing.T) {
	bin := buildThoth(t)
	src := filepath.Join("testdata", "repos", "create1")
	cfgT := "{\n  configVersion: \"v0\"\n  action: \"create-meta\"\n  discovery: { root: \"%s\" }\n}"
	var baseOut []byte
	for i := 0; i < 5; i++ {
		repo := filepath.Join("temp", "create_det_", time.Now().UTC().Format("20060102150405"), fmtInt(i))
		copyTree(t, src, repo)
		cfg := filepath.Join(repo, "tmp.cue")
		data := []byte(fmtSprintf(cfgT, filepath.ToSlash(repo)))
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

func TestDeterminism_UpdateMeta(t *testing.T) {
	bin := buildThoth(t)
	src := filepath.Join("testdata", "repos", "update1")
	cfgT := "{\n  configVersion: \"v0\"\n  action: \"update-meta\"\n  discovery: { root: \"%s\" }\n}"
	var baseOut []byte
	for i := 0; i < 5; i++ {
		repo := filepath.Join("temp", "update_det_", time.Now().UTC().Format("20060102150405"), fmtInt(i))
		copyTree(t, src, repo)
		cfg := filepath.Join(repo, "tmp.cue")
		data := []byte(fmtSprintf(cfgT, filepath.ToSlash(repo)))
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

func fmtInt(i int) string                  { return strconv.Itoa(i) }
func fmtSprintf(f string, a ...any) string { return fmt.Sprintf(f, a...) }
