// File Guide for dev/ai agents:
// Purpose: Spawn and supervise the actual OS process for one shell-exec record run.
// Responsibilities:
// - Start the command with rendered args, env overlay, output capture, and timeout control.
// - Terminate timed-out processes and optionally their process group.
// - Return low-level shellRunResult details used by higher-level shell processing.
// Architecture notes:
// - Timeout handling is per record, not per stage, and the `-2` exit code convention for timeouts is intentional.
// - Process-group termination is deliberate to avoid leaving child processes behind for shell commands that spawn subcommands.
// - Diagnostic context is captured here because the low-level spawn path knows the final program, args, and working directory.
package stage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type limitedBuffer struct {
	max       int
	buf       bytes.Buffer
	truncated bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if b.max <= 0 {
		return n, nil
	}
	remain := b.max - b.buf.Len()
	if remain > 0 {
		if remain > len(p) {
			remain = len(p)
		}
		_, _ = b.buf.Write(p[:remain])
	}
	if len(p) > remain {
		b.truncated = true
	}
	return n, nil
}

func (b *limitedBuffer) String() string { return b.buf.String() }

type shellRunResult struct {
	exitCode        int
	stdout          *string
	stderr          *string
	stdoutTruncated bool
	stderrTruncated bool
	timedOut        bool
	errorMsg        string
	program         string
	workingDir      string
	args            []string
}

func strPtr(s string) *string { return &s }

// runCommand executes the command with timeout/termination and returns result or error.
func runCommand(ctx context.Context, opts shellOptions, rec Record) (shellRunResult, error) {
	args, err := renderArgs(opts.argsT, rec, opts.strictTemplating)
	if err != nil {
		return shellRunResult{}, err
	}
	baseRes := shellRunResult{
		program:    opts.program,
		workingDir: opts.workingDir,
		args:       append([]string(nil), args...),
	}
	if opts.workingDir != "" {
		info, statErr := os.Stat(opts.workingDir)
		if statErr != nil {
			baseRes.exitCode = -1
			baseRes.errorMsg = fmt.Sprintf(
				"program %s start failed: workingDir %s: %v",
				opts.program,
				opts.workingDir,
				statErr,
			)
			return baseRes, nil
		}
		if !info.IsDir() {
			baseRes.exitCode = -1
			baseRes.errorMsg = fmt.Sprintf(
				"program %s start failed: workingDir %s: not a directory",
				opts.program,
				opts.workingDir,
			)
			return baseRes, nil
		}
	}
	cmd := exec.Command(opts.program, args...)
	cmd.Dir = opts.workingDir
	cmd.Env = applyEnvOverlay(os.Environ(), opts.env)
	if opts.killProcessGroup {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	outBuf := &limitedBuffer{max: opts.captureMaxBytes}
	errBuf := &limitedBuffer{max: opts.captureMaxBytes}
	if opts.captureStdout {
		cmd.Stdout = outBuf
	} else {
		cmd.Stdout = io.Discard
	}
	if opts.captureStderr {
		cmd.Stderr = errBuf
	} else {
		cmd.Stderr = io.Discard
	}

	if err := cmd.Start(); err != nil {
		var ee *exec.Error
		if errors.As(err, &ee) {
			baseRes.exitCode = -1
			baseRes.errorMsg = fmt.Sprintf(
				"program %s not found: %v",
				opts.program,
				err,
			)
			return baseRes, nil
		}
		baseRes.exitCode = -1
		baseRes.errorMsg = fmt.Sprintf(
			"program %s start failed: %v",
			opts.program,
			err,
		)
		return baseRes, nil
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	timer := time.NewTimer(time.Duration(opts.timeout) * time.Millisecond)
	defer timer.Stop()

	var runErr error
	timedOut := false
	select {
	case runErr = <-done:
	case <-timer.C:
		timedOut = true
		signalProcess(cmd, opts.killProcessGroup, syscall.SIGTERM)
		grace := time.NewTimer(time.Duration(opts.termGraceMs) * time.Millisecond)
		select {
		case runErr = <-done:
			grace.Stop()
		case <-grace.C:
			signalProcess(cmd, opts.killProcessGroup, syscall.SIGKILL)
			runErr = <-done
		}
	}

	res := shellRunResult{
		exitCode:        0,
		stdoutTruncated: outBuf.truncated,
		stderrTruncated: errBuf.truncated,
		timedOut:        timedOut,
		program:         opts.program,
		workingDir:      opts.workingDir,
		args:            append([]string(nil), args...),
	}
	if opts.captureStdout {
		res.stdout = strPtr(outBuf.String())
	}
	if opts.captureStderr {
		res.stderr = strPtr(errBuf.String())
	}
	if timedOut {
		res.exitCode = -2
		return res, nil
	}
	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			res.exitCode = exitErr.ExitCode()
			return res, nil
		}
		res.exitCode = -1
		res.errorMsg = fmt.Sprintf(
			"program %s execution failed: %v",
			opts.program,
			runErr,
		)
		return res, nil
	}
	return res, nil
}

func signalProcess(cmd *exec.Cmd, killGroup bool, sig syscall.Signal) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	pid := cmd.Process.Pid
	if killGroup && pid > 0 {
		if err := syscall.Kill(-pid, sig); err == nil {
			return
		}
	}
	_ = cmd.Process.Signal(sig)
}

func applyEnvOverlay(base []string, overlay map[string]string) []string {
	if len(overlay) == 0 {
		return append([]string(nil), base...)
	}
	m := map[string]string{}
	for _, kv := range base {
		i := -1
		for j := 0; j < len(kv); j++ {
			if kv[j] == '=' {
				i = j
				break
			}
		}
		if i <= 0 {
			continue
		}
		m[kv[:i]] = kv[i+1:]
	}
	for k, v := range overlay {
		m[k] = v
	}
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}
