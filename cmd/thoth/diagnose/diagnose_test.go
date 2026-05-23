package diagnose

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	runpkg "github.com/flarebyte/thoth-ostraca/cmd/thoth/run"
	"github.com/flarebyte/thoth-ostraca/internal/stage"
	"github.com/spf13/cobra"
)

func resetDiagFlags() {
	flagStage = ""
	flagStageIndex = -1
	flagUntilStage = ""
	flagUntilIndex = -1
	flagIn = ""
	flagDumpIn = ""
	flagDumpOut = ""
	flagDumpDir = ""
	flagPrepare = ""
	flagPreparePipeline = ""
	flagConfig = ""
	flagRoot = "."
	flagNoGit = false
	flagOut = "-"
	flagPretty = false
	flagLines = false
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	old := os.Stdout
	os.Stdout = w
	err = fn()
	_ = w.Close()
	os.Stdout = old
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	return buf.String(), err
}

func registerDiagPassStage(name string) {
	stage.Register(name, func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
		out := in
		if out.Meta == nil {
			out.Meta = &stage.Meta{}
		}
		out.Meta.Stage = name
		return out, nil
	})
}

func TestRelativizeRoot(t *testing.T) {
	if got := relativizeRoot(""); got != "" {
		t.Fatalf("got %q", got)
	}
	if got := relativizeRoot("a/b"); got != "a/b" {
		t.Fatalf("got %q", got)
	}
	cwd, _ := os.Getwd()
	inside := filepath.Join(cwd, "x", "y")
	if got := relativizeRoot(inside); got != "x/y" {
		t.Fatalf("got %q", got)
	}
}

func TestWriteJSONFileAndMaybeDump(t *testing.T) {
	resetDiagFlags()
	d := t.TempDir()
	p := filepath.Join(d, "a", "b.json")
	if err := writeJSONFile(p, map[string]any{"ok": true}); err != nil {
		t.Fatalf("writeJSONFile err: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("stat err: %v", err)
	}

	flagDumpIn = filepath.Join(d, "in.json")
	if err := maybeDumpIn(stage.Envelope{Records: []stage.Record{}}); err != nil {
		t.Fatalf("maybeDumpIn err: %v", err)
	}
	if _, err := os.Stat(flagDumpIn); err != nil {
		t.Fatalf("dump in missing: %v", err)
	}
	flagDumpOut = filepath.Join(d, "out.json")
	if err := maybeDumpOut(stage.Envelope{Records: []stage.Record{}}); err != nil {
		t.Fatalf("maybeDumpOut err: %v", err)
	}
	if _, err := os.Stat(flagDumpOut); err != nil {
		t.Fatalf("dump out missing: %v", err)
	}

	flagDumpDir = filepath.Join(d, "dir")
	if err := dumpStageBoundary(1, "s", "in", stage.Envelope{Records: []stage.Record{}}); err != nil {
		t.Fatalf("dumpStageBoundary err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(flagDumpDir, "001_s_in.json")); err != nil {
		t.Fatalf("dump boundary missing: %v", err)
	}
}

func TestWriteJSONFile_MkdirFailure(t *testing.T) {
	resetDiagFlags()
	d := t.TempDir()
	blocker := filepath.Join(d, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	// Parent path is a file, so MkdirAll should fail.
	err := writeJSONFile(filepath.Join(blocker, "x.json"), map[string]any{"ok": true})
	if err == nil {
		t.Fatalf("expected mkdir failure")
	}
}

func TestPrepareDiagnoseInputAndPrint(t *testing.T) {
	resetDiagFlags()
	d := t.TempDir()
	in := filepath.Join(d, "in.json")
	_ = os.WriteFile(in, []byte(`{"records":[{"locator":"a"}]}`), 0o644)
	env, err := prepareDiagnoseInput(in, "", "", false)
	if err != nil || len(env.Records) != 1 {
		t.Fatalf("prepare from file failed: %v %+v", err, env)
	}

	flagOut = "tmp.json"
	flagPretty = true
	env2, err := prepareDiagnoseInput("", "cfg.cue", "root", true)
	if err != nil {
		t.Fatalf("prepare no file err: %v", err)
	}
	if env2.Meta == nil || env2.Meta.ConfigPath != "cfg.cue" || env2.Meta.Discovery == nil || !env2.Meta.Discovery.NoGitignore {
		t.Fatalf("unexpected meta: %+v", env2.Meta)
	}
	if env2.Meta.Output == nil || !env2.Meta.Output.Pretty {
		t.Fatalf("expected output meta")
	}

	buf := &bytes.Buffer{}
	if err := printEnvelopeOneLine(buf, stage.Envelope{Records: []stage.Record{{Locator: "z"}}}); err != nil {
		t.Fatalf("print err: %v", err)
	}
	if !strings.Contains(buf.String(), `"contractVersion":"1"`) {
		t.Fatalf("missing contract version: %s", buf.String())
	}
}

func TestResolveHelpers(t *testing.T) {
	resetDiagFlags()
	stages := []string{"a", "b", "c"}
	if _, err := resolveStageByIndex(stages, 9, "--x"); err == nil {
		t.Fatalf("expected index error")
	}
	if idx, err := findStageIndexByName(stages, "b", "--s"); err != nil || idx != 1 {
		t.Fatalf("unexpected idx/err: %d %v", idx, err)
	}
	if _, err := findStageIndexByName(stages, "q", "--s"); err == nil {
		t.Fatalf("expected unknown stage error")
	}

	flagStage = "b"
	if got, err := resolveTargetStage(stages); err != nil || got != "b" {
		t.Fatalf("resolveTargetStage got=%q err=%v", got, err)
	}
	flagStage = ""
	flagStageIndex = 2
	if got, err := resolveTargetStage(stages); err != nil || got != "c" {
		t.Fatalf("resolveTargetStage index got=%q err=%v", got, err)
	}

	flagUntilIndex = 1
	if idx, ok, err := resolveUntilIndex(stages); err != nil || !ok || idx != 1 {
		t.Fatalf("resolveUntilIndex got=%d ok=%v err=%v", idx, ok, err)
	}
	flagUntilIndex = -1
	flagUntilStage = "c"
	if idx, ok, err := resolveUntilIndex(stages); err != nil || !ok || idx != 2 {
		t.Fatalf("resolveUntilIndex by name got=%d ok=%v err=%v", idx, ok, err)
	}

	flagPreparePipeline = "validate"
	if _, err := resolvePreparedAction(); err != nil {
		t.Fatalf("resolvePreparedAction err: %v", err)
	}
	flagPreparePipeline = "bad"
	if _, err := resolvePreparedAction(); err == nil {
		t.Fatalf("expected invalid action error")
	}
}

func TestRunStageSequenceAndRender(t *testing.T) {
	resetDiagFlags()
	registerDiagPassStage("diag-test-pass")
	env := stage.Envelope{Records: []stage.Record{{Locator: "x"}}}
	out, err := runStageSequence(env, []string{"diag-test-pass"})
	if err != nil {
		t.Fatalf("runStageSequence err: %v", err)
	}
	if out.Meta == nil || out.Meta.Stage != "diag-test-pass" {
		t.Fatalf("unexpected out meta: %+v", out.Meta)
	}

	stdout, err := captureStdout(t, func() error {
		return runStagesAndRender(stage.Envelope{Records: []stage.Record{}}, []string{"diag-test-pass"})
	})
	if err != nil {
		t.Fatalf("runStagesAndRender err: %v", err)
	}
	if !strings.Contains(stdout, `"contractVersion":"1"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestRunDiagnoseModes(t *testing.T) {
	registerDiagPassStage("diag-test-pass")
	d := t.TempDir()
	inPath := filepath.Join(d, "in.json")
	_ = os.WriteFile(inPath, []byte(`{"records":[{"locator":"a"}]}`), 0o644)

	// with --in
	resetDiagFlags()
	flagIn = inPath
	flagStage = "diag-test-pass"
	stdout, err := captureStdout(t, runDiagnoseWithIn)
	if err != nil {
		t.Fatalf("runDiagnoseWithIn err: %v", err)
	}
	if !strings.Contains(stdout, `"locator":"a"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}

	// with --prepare
	resetDiagFlags()
	flagPrepare = "meta-files"
	flagRoot = d
	flagStage = "diag-test-pass"
	stdout, err = captureStdout(t, runDiagnoseWithPrepare)
	if err != nil {
		t.Fatalf("runDiagnoseWithPrepare err: %v", err)
	}
	if !strings.Contains(stdout, `"contractVersion":"1"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}

	// default
	resetDiagFlags()
	flagStage = "diag-test-pass"
	stdout, err = captureStdout(t, func() error { return runDiagnoseDefault(&cobra.Command{}) })
	if err != nil {
		t.Fatalf("runDiagnoseDefault err: %v", err)
	}
	if !strings.Contains(stdout, `"contractVersion":"1"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestRunDiagnoseWithIn_MissingStageError(t *testing.T) {
	resetDiagFlags()
	d := t.TempDir()
	inPath := filepath.Join(d, "in.json")
	_ = os.WriteFile(inPath, []byte(`{"records":[]}`), 0o644)
	flagIn = inPath
	err := runDiagnoseWithIn()
	if err == nil || !strings.Contains(err.Error(), "missing required flag: --stage") {
		t.Fatalf("expected missing stage error, got %v", err)
	}
}

func TestRunDiagnoseWithPrepare_InvalidModeError(t *testing.T) {
	resetDiagFlags()
	flagPrepare = "bad-mode"
	flagStage = "diag-test-pass"
	if err := runDiagnoseWithPrepare(); err == nil {
		t.Fatalf("expected invalid prepare mode error")
	}
}

func TestBaseDiscoveryOverridesAndPreparedPipeline(t *testing.T) {
	resetDiagFlags()
	flagPreparePipeline = "validate"
	flagStage = "discover-meta-files"
	d := t.TempDir()
	flagRoot = d
	flagNoGit = true

	cmd := &cobra.Command{}
	cmd.Flags().String("root", ".", "")
	cmd.Flags().Bool("no-gitignore", false, "")
	_ = cmd.Flags().Set("root", d)
	_ = cmd.Flags().Set("no-gitignore", "true")

	root, noGit := baseDiscoveryOverrides(cmd)
	if root == "" || !noGit {
		t.Fatalf("unexpected overrides root=%q noGit=%v", root, noGit)
	}

	stdout, err := captureStdout(t, func() error { return runDiagnoseWithPreparedPipeline(cmd) })
	if err != nil {
		t.Fatalf("runDiagnoseWithPreparedPipeline err: %v", err)
	}
	if !strings.Contains(stdout, `"contractVersion":"1"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}

	// until-stage branch
	resetDiagFlags()
	flagPreparePipeline = "validate"
	flagUntilStage = "write-output"
	cmd2 := &cobra.Command{}
	cmd2.Flags().String("root", ".", "")
	cmd2.Flags().Bool("no-gitignore", false, "")
	stdout, err = captureStdout(t, func() error { return runDiagnoseWithPreparedPipeline(cmd2) })
	if err != nil {
		t.Fatalf("runDiagnoseWithPreparedPipeline until err: %v", err)
	}
	if !strings.Contains(stdout, `"records"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}

	// sanity: stage order still available for expected actions
	if _, err := runpkg.PreparedActionStages("validate", &stage.Meta{}); err != nil {
		t.Fatalf("PreparedActionStages err: %v", err)
	}
}

func TestRunDiagnoseWithPreparedPipeline_ErrorBranches(t *testing.T) {
	// invalid action
	resetDiagFlags()
	flagPreparePipeline = "nope"
	if err := runDiagnoseWithPreparedPipeline(&cobra.Command{}); err == nil {
		t.Fatalf("expected invalid prepare-pipeline action error")
	}

	// invalid until-stage
	resetDiagFlags()
	flagPreparePipeline = "validate"
	flagUntilStage = "does-not-exist"
	if err := runDiagnoseWithPreparedPipeline(&cobra.Command{}); err == nil {
		t.Fatalf("expected unknown until-stage error")
	}

	// missing target stage when no until specified
	resetDiagFlags()
	flagPreparePipeline = "validate"
	flagStage = ""
	if err := runDiagnoseWithPreparedPipeline(&cobra.Command{}); err == nil {
		t.Fatalf("expected missing stage error")
	}

	// validate-config stage error path
	resetDiagFlags()
	flagPreparePipeline = "validate"
	flagStage = "discover-meta-files"
	flagConfig = "does-not-exist.cue"
	if err := runDiagnoseWithPreparedPipeline(&cobra.Command{}); err == nil {
		t.Fatalf("expected validate-config error")
	}
}

func TestRunDiagnoseDefault_MissingStageError(t *testing.T) {
	resetDiagFlags()
	flagStage = "   "
	if err := runDiagnoseDefault(&cobra.Command{}); err == nil {
		t.Fatalf("expected missing stage error")
	}
}

func TestCmdRunEBranches(t *testing.T) {
	registerDiagPassStage("diag-test-pass")
	resetDiagFlags()
	if err := Cmd.RunE(Cmd, nil); err == nil {
		t.Fatalf("expected missing stage error")
	}

	d := t.TempDir()
	inPath := filepath.Join(d, "in.json")
	_ = os.WriteFile(inPath, []byte(`{"records":[]}`), 0o644)
	resetDiagFlags()
	flagIn = inPath
	flagStage = "diag-test-pass"
	_, err := captureStdout(t, func() error { return Cmd.RunE(Cmd, nil) })
	if err != nil {
		t.Fatalf("Cmd.RunE with in err: %v", err)
	}

	resetDiagFlags()
	flagPrepare = "meta-files"
	flagRoot = d
	flagStage = "diag-test-pass"
	_, err = captureStdout(t, func() error { return Cmd.RunE(Cmd, nil) })
	if err != nil {
		t.Fatalf("Cmd.RunE with prepare err: %v", err)
	}

	resetDiagFlags()
	flagPreparePipeline = "validate"
	flagStage = "discover-meta-files"
	cmd := &cobra.Command{}
	cmd.Flags().String("root", ".", "")
	cmd.Flags().Bool("no-gitignore", false, "")
	_, err = captureStdout(t, func() error { return Cmd.RunE(cmd, nil) })
	if err != nil {
		t.Fatalf("Cmd.RunE with prepared pipeline err: %v", err)
	}
}

func TestPrepareDiagnoseInputInvalidJSON(t *testing.T) {
	resetDiagFlags()
	d := t.TempDir()
	in := filepath.Join(d, "bad.json")
	_ = os.WriteFile(in, []byte("{"), 0o644)
	if _, err := prepareDiagnoseInput(in, "", "", false); err == nil {
		t.Fatalf("expected invalid input JSON error")
	}
}

func TestPrintEnvelopeOneLine_JSONRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	err := printEnvelopeOneLine(&buf, stage.Envelope{Records: []stage.Record{{Locator: "a"}}})
	if err != nil {
		t.Fatalf("printEnvelopeOneLine err: %v", err)
	}
	line := strings.TrimSpace(buf.String())
	var got stage.Envelope
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("unmarshal output err: %v; line=%s", err, line)
	}
	if got.Meta == nil || got.Meta.ContractVersion != "1" {
		t.Fatalf("unexpected contract version: %+v", got.Meta)
	}
}

type errWriter struct{}

func (errWriter) Write(p []byte) (int, error) { return 0, errors.New("write failed") }

func TestPrintEnvelopeOneLine_WriteError(t *testing.T) {
	err := printEnvelopeOneLine(errWriter{}, stage.Envelope{Records: []stage.Record{{Locator: "a"}}})
	if err == nil || !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("expected write failure, got %v", err)
	}
}

func TestRunStageSequence_DumpBoundaryError(t *testing.T) {
	resetDiagFlags()
	registerDiagPassStage("diag-test-pass")
	d := t.TempDir()
	blocker := filepath.Join(d, "not-a-dir")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	flagDumpDir = blocker
	_, err := runStageSequence(stage.Envelope{Records: []stage.Record{}}, []string{"diag-test-pass"})
	if err == nil {
		t.Fatalf("expected dump boundary error")
	}
}

func TestRunStagesAndRender_DumpOutError(t *testing.T) {
	resetDiagFlags()
	registerDiagPassStage("diag-test-pass")
	d := t.TempDir()
	blocker := filepath.Join(d, "not-a-dir")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	flagDumpOut = filepath.Join(blocker, "out.json")
	err := runStagesAndRender(stage.Envelope{Records: []stage.Record{}}, []string{"diag-test-pass"})
	if err == nil {
		t.Fatalf("expected dump out error")
	}
}
