package stage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"syscall"
	"time"
)

const shellExecStage = "shell-exec"

type shellExecRes struct {
	idx   int
	rec   Record
	envE  *Error
	fatal error
}

type shellOptions struct {
	enabled          bool
	program          string
	argsT            []string
	workingDir       string
	env              map[string]string
	timeout          int
	captureStdout    bool
	captureStderr    bool
	captureMaxBytes  int
	strictTemplating bool
	killProcessGroup bool
	termGraceMs      int
}

var placeholderPattern = regexp.MustCompile(`\{[^{}]+\}`)

// buildShellOptions derives execution options from envelope meta with defaults.
func buildShellOptions(in Envelope) shellOptions {
	opts := shellOptions{
		enabled:          false,
		program:          defaultShellProgram,
		argsT:            nil,
		workingDir:       filepath.Join(".", defaultShellWorkingDir),
		env:              map[string]string{},
		timeout:          defaultShellTimeoutMs,
		captureStdout:    true,
		captureStderr:    true,
		captureMaxBytes:  defaultShellCaptureMaxBytes,
		strictTemplating: true,
		killProcessGroup: true,
		termGraceMs:      defaultShellTermGraceMs,
	}
	if in.Meta == nil || in.Meta.Shell == nil {
		return opts
	}
	cfg := in.Meta.Shell
	opts.enabled = cfg.Enabled
	if cfg.Program != "" {
		opts.program = cfg.Program
	}
	if len(cfg.ArgsTemplate) > 0 {
		opts.argsT = append([]string(nil), cfg.ArgsTemplate...)
	}
	root := "."
	if in.Meta.Discovery != nil && in.Meta.Discovery.Root != "" {
		root = in.Meta.Discovery.Root
	}
	wd := cfg.WorkingDir
	if wd == "" {
		wd = defaultShellWorkingDir
	}
	opts.workingDir = filepath.Join(root, wd)
	if cfg.Env != nil {
		opts.env = make(map[string]string, len(cfg.Env))
		for k, v := range cfg.Env {
			opts.env[k] = v
		}
	}
	if cfg.TimeoutMs >= 0 {
		opts.timeout = cfg.TimeoutMs
	}
	opts.captureStdout = cfg.Capture.Stdout
	opts.captureStderr = cfg.Capture.Stderr
	if cfg.Capture.MaxBytes >= 0 {
		opts.captureMaxBytes = cfg.Capture.MaxBytes
	}
	opts.strictTemplating = cfg.StrictTemplating
	opts.killProcessGroup = cfg.KillProcessGroup
	if cfg.TermGraceMs >= 0 {
		opts.termGraceMs = cfg.TermGraceMs
	}
	return opts
}

func validateShellOptions(opts shellOptions) error {
	if len(opts.argsT) == 0 {
		return errors.New("missing argsTemplate")
	}
	return nil
}

// renderArgs applies the {json} replacement using the record's mapped value.
func renderArgs(argsT []string, mapped any, strict bool) ([]string, error) {
	mappedJSON, _ := json.Marshal(mapped)
	rendered := make([]string, len(argsT))
	for i := range argsT {
		a := argsT[i]
		if strict {
			matches := placeholderPattern.FindAllString(a, -1)
			for _, m := range matches {
				if m != "{json}" {
					return nil, fmt.Errorf("strict templating: invalid placeholder %s", m)
				}
			}
		}
		rendered[i] = replaceJSON(a, string(mappedJSON))
	}
	return rendered, nil
}

type limitedBuffer struct {
	max int
	buf bytes.Buffer
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
	return n, nil
}

func (b *limitedBuffer) String() string { return b.buf.String() }

// runCommand executes the command with timeout/termination and returns result or error.
func runCommand(ctx context.Context, opts shellOptions, mapped any) (stdout, stderr string, exitCode int, timedOut bool, err error) {
	args, err := renderArgs(opts.argsT, mapped, opts.strictTemplating)
	if err != nil {
		return "", "", 0, false, err
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
			return "", "", 0, false, fmt.Errorf("program %s not found", opts.program)
		}
		return "", "", 0, false, fmt.Errorf("program %s start failed", opts.program)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	timer := time.NewTimer(time.Duration(opts.timeout) * time.Millisecond)
	defer timer.Stop()

	var runErr error
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

	stdout = outBuf.String()
	stderr = errBuf.String()
	if timedOut {
		return stdout, stderr, 0, true, nil
	}
	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			return stdout, stderr, exitErr.ExitCode(), false, nil
		}
		return stdout, stderr, 0, false, fmt.Errorf("program %s execution failed", opts.program)
	}
	return stdout, stderr, 0, false, nil
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

// processShellRecord validates, renders, executes and returns updated record or errors.
func processShellRecord(ctx context.Context, rec Record, opts shellOptions, mode string) (Record, *Error, error) {
	if rec.Error != nil {
		return rec, nil, nil
	}
	stdout, stderr, exitCode, timedOut, err := runCommand(ctx, opts, rec.Mapped)
	if timedOut {
		if mode == "keep-going" {
			rec.Error = &RecError{Stage: shellExecStage, Message: "timeout"}
			return rec, &Error{Stage: shellExecStage, Locator: rec.Locator, Message: "timeout"}, nil
		}
		return Record{}, nil, fmt.Errorf("shell-exec: timeout")
	}
	if err != nil {
		if mode == "keep-going" {
			rec.Error = &RecError{Stage: shellExecStage, Message: err.Error()}
			return rec, &Error{Stage: shellExecStage, Locator: rec.Locator, Message: err.Error()}, nil
		}
		return Record{}, nil, fmt.Errorf("shell-exec: %v", err)
	}
	sr := &ShellResult{Stdout: stdout, Stderr: stderr, ExitCode: exitCode}
	rec.Shell = sr
	return rec, nil, nil
}

// replaceJSON replaces exact token {json} with provided JSON string.
func replaceJSON(s, json string) string {
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
