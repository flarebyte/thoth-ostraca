package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writeShellPhase4Repo(t *testing.T) string {
	t.Helper()
	tmpRoot := filepath.Join(repoRoot(), "temp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	repo, err := os.MkdirTemp(tmpRoot, "phase4-shell-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	meta := "locator: \"a\"\nmeta:\n  name: \"A\"\n"
	if err := os.WriteFile(filepath.Join(repo, "a.thoth.yaml"), []byte(meta), 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "sub", "known.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("write known: %v", err)
	}
	return repo
}

func cueStringList(items []string) string {
	parts := make([]string, 0, len(items))
	for _, it := range items {
		parts = append(parts, fmt.Sprintf("%q", it))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func writeShellPhase4Config(t *testing.T, repo string, workers int, shellBlock string) string {
	t.Helper()
	tmpRoot := filepath.Join(repoRoot(), "temp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	cfg, err := os.CreateTemp(tmpRoot, "phase4-shell-*.cue")
	if err != nil {
		t.Fatalf("mktemp cfg: %v", err)
	}
	_ = cfg.Close()
	content := fmt.Sprintf("{\n  configVersion: \"v0\"\n  action: \"pipeline\"\n  discovery: { root: %q }\n  workers: %d\n  shell: %s\n}\n", filepath.ToSlash(repo), workers, shellBlock)
	if err := os.WriteFile(cfg.Name(), []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	return cfg.Name()
}

func decodeRunEnvelope(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("json decode: %v\n%s", err, string(b))
	}
	return v
}

func shellStdoutFromRun(t *testing.T, b []byte) string {
	t.Helper()
	env := decodeRunEnvelope(t, b)
	recs, ok := env["records"].([]any)
	if !ok || len(recs) == 0 {
		t.Fatalf("missing records")
	}
	r, ok := recs[0].(map[string]any)
	if !ok {
		t.Fatalf("invalid record")
	}
	sm, ok := r["shell"].(map[string]any)
	if !ok {
		t.Fatalf("missing shell result")
	}
	out, _ := sm["stdout"].(string)
	return out
}

func TestShellPhase4_StrictTemplating(t *testing.T) {
	bin := buildThoth(t)
	repo := writeShellPhase4Repo(t)

	strictBlock := fmt.Sprintf("{ enabled: true, program: \"sh\", argsTemplate: %s, strictTemplating: true }", cueStringList([]string{"-c", "printf '%s' '{oops}'"}))
	cfgStrict := writeShellPhase4Config(t, repo, 1, strictBlock)
	rStrict := runCmd(t, bin, "run", "--config", cfgStrict)
	if rStrict.code == 0 {
		t.Fatalf("expected strict templating failure")
	}
	if !strings.Contains(string(rStrict.stderr), "strict templating: invalid placeholder {oops}") {
		t.Fatalf("unexpected stderr: %s", string(rStrict.stderr))
	}

	looseBlock := fmt.Sprintf("{ enabled: true, program: \"sh\", argsTemplate: %s, strictTemplating: false }", cueStringList([]string{"-c", "printf '%s' '{oops}'"}))
	cfgLoose := writeShellPhase4Config(t, repo, 1, looseBlock)
	rLoose := runCmd(t, bin, "run", "--config", cfgLoose)
	if rLoose.code != 0 || len(rLoose.stderr) != 0 {
		t.Fatalf("strict=false failed: code=%d stderr=%s", rLoose.code, string(rLoose.stderr))
	}
	if got := shellStdoutFromRun(t, rLoose.stdout); got != "{oops}" {
		t.Fatalf("unexpected stdout: %q", got)
	}
}

func TestShellPhase4_TimeoutTermination(t *testing.T) {
	bin := buildThoth(t)
	repo := writeShellPhase4Repo(t)
	block := fmt.Sprintf("{ enabled: true, program: \"sh\", argsTemplate: %s, timeoutMs: 30, termGraceMs: 10, killProcessGroup: true }", cueStringList([]string{"-c", "sleep 2"}))
	cfg := writeShellPhase4Config(t, repo, 1, block)
	r := runCmd(t, bin, "run", "--config", cfg)
	if r.code == 0 {
		t.Fatalf("expected timeout failure")
	}
	if !strings.Contains(string(r.stderr), "shell-exec: timeout") {
		t.Fatalf("unexpected stderr: %s", string(r.stderr))
	}
}

func TestShellPhase4_CaptureTruncation(t *testing.T) {
	bin := buildThoth(t)
	repo := writeShellPhase4Repo(t)
	block := fmt.Sprintf("{ enabled: true, program: \"sh\", argsTemplate: %s, capture: { stdout: true, stderr: true, maxBytes: 5 } }", cueStringList([]string{"-c", "printf '0123456789'"}))
	cfg := writeShellPhase4Config(t, repo, 1, block)
	r := runCmd(t, bin, "run", "--config", cfg)
	if r.code != 0 || len(r.stderr) != 0 {
		t.Fatalf("run failed: code=%d stderr=%s", r.code, string(r.stderr))
	}
	if got := shellStdoutFromRun(t, r.stdout); got != "01234" {
		t.Fatalf("unexpected stdout truncation: %q", got)
	}
}

func TestShellPhase4_WorkingDirAndEnvOverlay(t *testing.T) {
	bin := buildThoth(t)
	repo := writeShellPhase4Repo(t)
	blockDir := fmt.Sprintf("{ enabled: true, program: \"sh\", workingDir: \"sub\", argsTemplate: %s }", cueStringList([]string{"-c", "pwd"}))
	cfgDir := writeShellPhase4Config(t, repo, 1, blockDir)
	rDir := runCmd(t, bin, "run", "--config", cfgDir)
	if rDir.code != 0 || len(rDir.stderr) != 0 {
		t.Fatalf("workingDir run failed: code=%d stderr=%s", rDir.code, string(rDir.stderr))
	}
	if got := shellStdoutFromRun(t, rDir.stdout); !strings.Contains(strings.TrimSpace(got), "/sub") {
		t.Fatalf("unexpected pwd output: %q", got)
	}

	blockEnv := fmt.Sprintf("{ enabled: true, program: \"sh\", env: { THOTH_E2E_ENV: \"phase4\" }, argsTemplate: %s }", cueStringList([]string{"-c", "printf '%s' \"$THOTH_E2E_ENV\""}))
	cfgEnv := writeShellPhase4Config(t, repo, 1, blockEnv)
	rEnv := runCmd(t, bin, "run", "--config", cfgEnv)
	if rEnv.code != 0 || len(rEnv.stderr) != 0 {
		t.Fatalf("env run failed: code=%d stderr=%s", rEnv.code, string(rEnv.stderr))
	}
	if got := shellStdoutFromRun(t, rEnv.stdout); got != "phase4" {
		t.Fatalf("unexpected env stdout: %q", got)
	}
}

func TestShellPhase4_ProgramMatrix(t *testing.T) {
	bin := buildThoth(t)
	repo := writeShellPhase4Repo(t)
	programs := []string{"sh", "bash", "zsh"}
	for _, prog := range programs {
		prog := prog
		t.Run(prog, func(t *testing.T) {
			if _, err := exec.LookPath(prog); err != nil {
				if prog != "sh" {
					t.Skipf("%s not available", prog)
				}
				t.Fatalf("required shell %s missing", prog)
			}
			block := fmt.Sprintf("{ enabled: true, program: %q, argsTemplate: %s }", prog, cueStringList([]string{"-c", "printf 'ok'"}))
			cfg := writeShellPhase4Config(t, repo, 1, block)
			r := runCmd(t, bin, "run", "--config", cfg)
			if r.code != 0 || len(r.stderr) != 0 {
				t.Fatalf("program=%s failed: code=%d stderr=%s", prog, r.code, string(r.stderr))
			}
			if got := shellStdoutFromRun(t, r.stdout); got != "ok" {
				t.Fatalf("unexpected stdout: %q", got)
			}
		})
	}
}

func TestShellPhase4_Determinism_RerunAndWorkers(t *testing.T) {
	bin := buildThoth(t)
	repo := writeShellPhase4Repo(t)
	block := fmt.Sprintf("{ enabled: true, program: \"sh\", argsTemplate: %s }", cueStringList([]string{"-c", "printf '%s' '{json}'"}))
	cfg1 := writeShellPhase4Config(t, repo, 1, block)
	cfg8 := writeShellPhase4Config(t, repo, 8, block)

	r11 := runCmd(t, bin, "run", "--config", cfg1)
	r12 := runCmd(t, bin, "run", "--config", cfg1)
	r8 := runCmd(t, bin, "run", "--config", cfg8)

	if r11.code != 0 || len(r11.stderr) != 0 || r12.code != 0 || len(r12.stderr) != 0 || r8.code != 0 || len(r8.stderr) != 0 {
		t.Fatalf("determinism runs failed")
	}
	if string(r11.stdout) != string(r12.stdout) {
		t.Fatalf("stdout drift across reruns")
	}
	if normalizeWorkersOut(r11.stdout) != normalizeWorkersOut(r8.stdout) {
		t.Fatalf("stdout drift workers=1 vs workers=8")
	}
}
