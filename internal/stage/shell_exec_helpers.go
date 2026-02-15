package stage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

const shellExecStage = "shell-exec"

type shellExecRes struct {
	idx   int
	rec   any
	envE  *Error
	fatal error
}

type shellOptions struct {
	enabled bool
	program string
	argsT   []string
	timeout int // milliseconds
}

// buildShellOptions derives execution options from envelope meta with defaults.
func buildShellOptions(in Envelope) shellOptions {
	// Defaults
	opts := shellOptions{enabled: false, program: "bash", argsT: nil, timeout: 60000}
	if in.Meta != nil && in.Meta.Shell != nil {
		opts.enabled = in.Meta.Shell.Enabled
		if in.Meta.Shell.Program != "" {
			opts.program = in.Meta.Shell.Program
		}
		if len(in.Meta.Shell.ArgsTemplate) > 0 {
			opts.argsT = in.Meta.Shell.ArgsTemplate
		}
		if in.Meta.Shell.TimeoutMs > 0 {
			opts.timeout = in.Meta.Shell.TimeoutMs
		}
	}
	return opts
}

// renderArgs applies the {json} replacement using the record's mapped value.
func renderArgs(argsT []string, mapped any) []string {
	mappedJSON, _ := json.Marshal(mapped)
	rendered := make([]string, len(argsT))
	for i := range argsT {
		rendered[i] = replaceJSON(argsT[i], string(mappedJSON))
	}
	return rendered
}

// runCommand executes the command with a timeout and returns result or error.
func runCommand(ctx context.Context, program string, args []string, timeoutMs int) (stdout, stderr string, exitCode int, timedOut bool, err error) {
	cctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(cctx, program, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	if cctx.Err() == context.DeadlineExceeded {
		return "", "", 0, true, nil
	}
	stdout = outBuf.String()
	stderr = errBuf.String()
	if runErr != nil {
		var ee *exec.ExitError
		if errors.As(runErr, &ee) {
			return stdout, stderr, ee.ExitCode(), false, nil
		}
		return stdout, stderr, 0, false, runErr
	}
	return stdout, stderr, 0, false, nil
}

// processShellRecord validates, renders, executes and returns updated record or errors.
func processShellRecord(ctx context.Context, r any, opts shellOptions, mode string) (any, *Error, error) {
	rec, ok := r.(Record)
	if !ok {
		if mode == "keep-going" {
			return r, &Error{Stage: shellExecStage, Message: "invalid record type"}, nil
		}
		return nil, nil, errors.New("shell-exec: invalid record type")
	}
	if rec.Error != nil {
		return rec, nil, nil
	}
	rendered := renderArgs(opts.argsT, rec.Mapped)
	stdout, stderr, exitCode, timedOut, err := runCommand(ctx, opts.program, rendered, opts.timeout)
	if timedOut {
		if mode == "keep-going" {
			rec.Error = &RecError{Stage: shellExecStage, Message: "timeout"}
			return rec, &Error{Stage: shellExecStage, Locator: rec.Locator, Message: "timeout"}, nil
		}
		return nil, nil, fmt.Errorf("shell-exec: timeout")
	}
	if err != nil {
		if mode == "keep-going" {
			rec.Error = &RecError{Stage: shellExecStage, Message: err.Error()}
			return rec, &Error{Stage: shellExecStage, Locator: rec.Locator, Message: err.Error()}, nil
		}
		return nil, nil, fmt.Errorf("shell-exec: %v", err)
	}
	sr := &ShellResult{Stdout: stdout, Stderr: stderr, ExitCode: exitCode}
	rec.Shell = sr
	return rec, nil, nil
}

// replaceJSON replaces exact token {json} with provided JSON string.
func replaceJSON(s, json string) string {
	// Replace exactly the token {json}
	// No templating beyond this.
	// Simple non-allocating approach
	out := []byte{}
	i := 0
	for i < len(s) {
		if i+6 <= len(s) && s[i:i+6] == "{json}" {
			out = append(out, []byte(json)...)
			i += 6
		} else {
			out = append(out, s[i])
			i++
		}
	}
	return string(out)
}
