package stage

import (
	"context"
	"fmt"
)

// processShellRecord validates, renders, executes and returns updated record or errors.
func processShellRecord(ctx context.Context, rec Record, opts shellOptions, mode string) (Record, *Error, error) {
	if rec.Error != nil {
		return rec, nil, nil
	}
	runRes, err := runCommand(ctx, opts, rec)
	if err != nil {
		msg := sanitizeErrorMessage(err.Error())
		if mode == "keep-going" {
			rec.Shell = &ShellResult{
				ExitCode:        -1,
				Stdout:          nil,
				Stderr:          nil,
				StdoutTruncated: false,
				StderrTruncated: false,
				TimedOut:        false,
				Error:           strPtr(msg),
			}
			rec.Error = &RecError{Stage: shellExecStage, Message: msg}
			return rec, &Error{Stage: shellExecStage, Locator: rec.Locator, Message: msg}, nil
		}
		return Record{}, nil, fmt.Errorf("shell-exec: %s", msg)
	}
	rec.Shell = &ShellResult{
		ExitCode:        runRes.exitCode,
		Stdout:          runRes.stdout,
		Stderr:          runRes.stderr,
		StdoutTruncated: runRes.stdoutTruncated,
		StderrTruncated: runRes.stderrTruncated,
		TimedOut:        runRes.timedOut,
	}
	if runRes.errorMsg != "" {
		msg := sanitizeErrorMessage(runRes.errorMsg)
		rec.Shell.Error = strPtr(msg)
		attachShellDiagnostics(rec.Shell, runRes)
		runRes.errorMsg = msg
	}
	if !runRes.timedOut && runRes.errorMsg == "" && opts.decodeJSONStdout {
		decoded, err := decodeShellStdoutJSON(runRes.stdout)
		if err != nil {
			msg := sanitizeErrorMessage("invalid JSON stdout: " + err.Error())
			attachShellDiagnostics(rec.Shell, runRes)
			rec.Error = &RecError{Stage: shellExecStage, Message: msg}
			if mode == "keep-going" {
				return rec, &Error{
					Stage:   shellExecStage,
					Locator: rec.Locator,
					Message: msg,
				}, nil
			}
			return Record{}, nil, fmt.Errorf("shell-exec: %s", msg)
		}
		rec.Shell.JSON = decoded
	}
	if runRes.timedOut {
		attachShellDiagnostics(rec.Shell, runRes)
		if mode == "keep-going" {
			rec.Error = &RecError{Stage: shellExecStage, Message: "timeout"}
			return rec, &Error{Stage: shellExecStage, Locator: rec.Locator, Message: "timeout"}, nil
		}
		return Record{}, nil, fmt.Errorf("shell-exec: timeout")
	}
	if runRes.errorMsg != "" {
		if mode == "keep-going" {
			rec.Error = &RecError{Stage: shellExecStage, Message: runRes.errorMsg}
			return rec, &Error{Stage: shellExecStage, Locator: rec.Locator, Message: runRes.errorMsg}, nil
		}
		return Record{}, nil, fmt.Errorf("shell-exec: %s", runRes.errorMsg)
	}
	return rec, nil, nil
}

func attachShellDiagnostics(shell *ShellResult, runRes shellRunResult) {
	if shell == nil {
		return
	}
	shell.Program = runRes.program
	shell.WorkingDir = runRes.workingDir
	if len(runRes.args) > 0 {
		shell.Args = append([]string(nil), runRes.args...)
	}
}
