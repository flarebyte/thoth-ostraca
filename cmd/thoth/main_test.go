package main

import (
	"bytes"
	"os"
	"testing"
)

func TestRunSuccessAndErrorFormatting(t *testing.T) {
	buf := &bytes.Buffer{}
	if code := run([]string{"version"}, buf); code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", buf.String())
	}
}

func TestRunMissingCommandReturnsErrorCode(t *testing.T) {
	buf := &bytes.Buffer{}
	code := run([]string{"run"}, buf) // missing --config should error
	if code == 0 {
		t.Fatalf("expected non-zero exit code")
	}
	if buf.Len() == 0 {
		t.Fatalf("expected error output")
	}
}

func TestMainCallsOsExit(t *testing.T) {
	oldExit := osExit
	oldArgs := os.Args
	defer func() {
		osExit = oldExit
		os.Args = oldArgs
	}()

	called := false
	code := -1
	osExit = func(c int) {
		called = true
		code = c
	}
	os.Args = []string{"thoth", "version"}
	main()
	if !called {
		t.Fatalf("expected osExit to be called")
	}
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}
