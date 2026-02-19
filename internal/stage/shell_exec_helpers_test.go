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
	r, err := runCommand(context.Background(), opts, map[string]any{"a": 1})
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
	r, err := runCommand(context.Background(), opts, nil)
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
	r, err := runCommand(context.Background(), opts, nil)
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
	r, err := runCommand(context.Background(), opts, nil)
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
	r, err := runCommand(context.Background(), opts, nil)
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
	if rec.Error == nil {
		t.Fatalf("expected rec.Error")
	}
}
