package stage

import (
	"errors"
	"path/filepath"
)

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
