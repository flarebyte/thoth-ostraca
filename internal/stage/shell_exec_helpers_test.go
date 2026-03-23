package stage

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

func requirePOSIXShell(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell tests require POSIX shell")
	}
}

func baseShellOpts() shellOptions {
	return shellOptions{
		enabled:          true,
		program:          "sh",
		argsT:            []string{"-c", "printf 'ok'"},
		workingDir:       ".",
		env:              map[string]string{},
		timeout:          1000,
		captureStdout:    true,
		captureStderr:    true,
		captureMaxBytes:  1024,
		strictTemplating: true,
		killProcessGroup: true,
		termGraceMs:      50,
	}
}

func TestRunCommand_SuccessCanonical(t *testing.T) {
	requirePOSIXShell(t)
	opts := baseShellOpts()
	r, err := runCommand(
		context.Background(),
		opts,
		Record{Mapped: map[string]any{"a": 1}},
	)
	if err != nil {
		t.Fatalf("runCommand err: %v", err)
	}
	if r.exitCode != 0 || r.timedOut || r.errorMsg != "" {
		t.Fatalf("unexpected result: %+v", r)
	}
	if r.stdout == nil || *r.stdout != "ok" {
		t.Fatalf("unexpected stdout: %#v", r.stdout)
	}
	if r.stderr == nil || *r.stderr != "" {
		t.Fatalf("unexpected stderr: %#v", r.stderr)
	}
	if r.stdoutTruncated || r.stderrTruncated {
		t.Fatalf("unexpected truncation flags: %+v", r)
	}
}

func TestRunCommand_NonZeroExitNoErrorField(t *testing.T) {
	requirePOSIXShell(t)
	opts := baseShellOpts()
	opts.argsT = []string{"-c", "printf 'bad' >&2; exit 7"}
	r, err := runCommand(context.Background(), opts, Record{})
	if err != nil {
		t.Fatalf("runCommand err: %v", err)
	}
	if r.exitCode != 7 || r.timedOut || r.errorMsg != "" {
		t.Fatalf("unexpected result: %+v", r)
	}
}

func TestRunCommand_TimeoutSetsDeterministicFields(t *testing.T) {
	requirePOSIXShell(t)
	opts := baseShellOpts()
	opts.timeout = 20
	opts.termGraceMs = 10
	opts.argsT = []string{"-c", "sleep 2"}
	r, err := runCommand(context.Background(), opts, Record{})
	if err != nil {
		t.Fatalf("runCommand err: %v", err)
	}
	if !r.timedOut || r.exitCode != -2 {
		t.Fatalf("unexpected timeout result: %+v", r)
	}
}

func TestRunCommand_TruncationExactMaxBytes(t *testing.T) {
	requirePOSIXShell(t)
	opts := baseShellOpts()
	opts.captureMaxBytes = 5
	opts.argsT = []string{"-c", "printf '0123456789'"}
	r, err := runCommand(context.Background(), opts, Record{})
	if err != nil {
		t.Fatalf("runCommand err: %v", err)
	}
	if r.stdout == nil || len(*r.stdout) != 5 || *r.stdout != "01234" {
		t.Fatalf("unexpected stdout: %#v", r.stdout)
	}
	if !r.stdoutTruncated {
		t.Fatalf("expected stdout truncated flag true")
	}
}

func TestRunCommand_CaptureDisabledOmitsStream(t *testing.T) {
	requirePOSIXShell(t)
	opts := baseShellOpts()
	opts.captureStdout = false
	opts.argsT = []string{"-c", "printf 'hello'"}
	r, err := runCommand(context.Background(), opts, Record{})
	if err != nil {
		t.Fatalf("runCommand err: %v", err)
	}
	if r.stdout != nil {
		t.Fatalf("expected stdout omitted when capture disabled")
	}
	if r.stdoutTruncated {
		t.Fatalf("expected stdoutTruncated false")
	}
}

func TestProcessShellRecord_KeepGoingStartFailurePopulatesShell(t *testing.T) {
	opts := baseShellOpts()
	opts.program = "this-program-does-not-exist-xyz"
	rec, envErr, fatal := processShellRecord(context.Background(), Record{Locator: "x"}, opts, "keep-going")
	if fatal != nil {
		t.Fatalf("fatal should be nil in keep-going: %v", fatal)
	}
	if envErr == nil || !strings.Contains(envErr.Message, "program this-program-does-not-exist-xyz") {
		t.Fatalf("unexpected envErr: %+v", envErr)
	}
	if rec.Shell == nil || rec.Shell.Error == nil || rec.Shell.ExitCode != -1 || rec.Shell.TimedOut {
		t.Fatalf("unexpected rec shell: %+v", rec.Shell)
	}
	if rec.Shell.Program != "this-program-does-not-exist-xyz" {
		t.Fatalf("expected shell program diagnostics, got %+v", rec.Shell)
	}
	if len(rec.Shell.Args) != 2 || rec.Shell.Args[0] != "-c" {
		t.Fatalf("expected rendered args diagnostics, got %+v", rec.Shell)
	}
	if rec.Error == nil {
		t.Fatalf("expected rec.Error")
	}
}

func TestRunCommand_StartFailurePreservesUnderlyingError(t *testing.T) {
	requirePOSIXShell(t)
	opts := baseShellOpts()
	opts.program = "/bin/sh"
	opts.workingDir = "does-not-exist"
	r, err := runCommand(context.Background(), opts, Record{})
	if err != nil {
		t.Fatalf("runCommand err: %v", err)
	}
	if r.exitCode != -1 {
		t.Fatalf("expected exitCode=-1, got %+v", r)
	}
	if !strings.Contains(r.errorMsg, "start failed:") {
		t.Fatalf("expected start failure details, got %q", r.errorMsg)
	}
	if !strings.Contains(r.errorMsg, "does-not-exist") {
		t.Fatalf("expected working dir in error, got %q", r.errorMsg)
	}
}

func TestProcessShellRecord_DecodeJSONStdout(t *testing.T) {
	requirePOSIXShell(t)
	opts := baseShellOpts()
	opts.decodeJSONStdout = true
	opts.argsT = []string{
		"-c",
		"printf '%s\\n' '{json}'",
	}
	rec, envErr, fatal := processShellRecord(
		context.Background(),
		Record{
			Locator: "x",
			Mapped: map[string]any{
				"locator": "x",
				"kind":    "go",
			},
		},
		opts,
		"keep-going",
	)
	if fatal != nil {
		t.Fatalf("fatal: %v", fatal)
	}
	if envErr != nil {
		t.Fatalf("unexpected envErr: %+v", envErr)
	}
	if rec.Shell == nil || rec.Shell.JSON == nil {
		t.Fatalf("expected decoded shell json")
	}
	decoded, ok := rec.Shell.JSON.(map[string]any)
	if !ok || decoded["locator"] != "x" {
		t.Fatalf("unexpected decoded json: %#v", rec.Shell.JSON)
	}
}

func TestRenderArgs_DirectPlaceholders(t *testing.T) {
	rec := Record{
		Locator: "sub/c.go",
		Mapped: map[string]any{
			"kind": "go",
			"nested": map[string]any{
				"name": "write",
			},
		},
	}
	args, err := renderArgs(
		[]string{
			"{locator}",
			"{file.base}",
			"{file.dir}",
			"{file.stem}",
			"{file.ext}",
			"{mapped.kind}",
			"{mapped.nested.name}",
			"{json}",
		},
		rec,
		true,
	)
	if err != nil {
		t.Fatalf("renderArgs err: %v", err)
	}
	want := []string{
		"sub/c.go",
		"c.go",
		"sub",
		"c",
		".go",
		"go",
		"write",
		`{"kind":"go","nested":{"name":"write"}}`,
	}
	if len(args) != len(want) {
		t.Fatalf("len(args)=%d want %d", len(args), len(want))
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("arg[%d]=%q want %q", i, args[i], want[i])
		}
	}
}

func TestRenderArgs_StrictTemplatingRejectsUnknownPlaceholder(t *testing.T) {
	_, err := renderArgs([]string{"{nope}"}, Record{}, true)
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := err.Error(); got != "strict templating: invalid placeholder {nope}" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestRenderArgs_MissingMappedValueFailsClearly(t *testing.T) {
	_, err := renderArgs(
		[]string{"{mapped.kind}"},
		Record{Mapped: map[string]any{"name": "x"}},
		true,
	)
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := err.Error(); got != "template placeholder {mapped.kind} missing value" {
		t.Fatalf("unexpected error: %q", got)
	}
}
