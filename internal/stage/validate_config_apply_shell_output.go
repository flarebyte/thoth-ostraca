package stage

import "github.com/flarebyte/thoth-ostraca/internal/config"

func applyShellMeta(out *Envelope, min config.Minimal) {
	if !min.Shell.HasSection {
		return
	}
	if out.Meta.Shell == nil {
		out.Meta.Shell = &ShellMeta{
			Enabled:          false,
			Program:          defaultShellProgram,
			WorkingDir:       defaultShellWorkingDir,
			Env:              map[string]string{},
			TimeoutMs:        defaultShellTimeoutMs,
			Capture:          ShellCaptureMeta{Stdout: true, Stderr: true, MaxBytes: defaultShellCaptureMaxBytes},
			StrictTemplating: true,
			KillProcessGroup: true,
			TermGraceMs:      defaultShellTermGraceMs,
		}
	}
	if min.Shell.HasEnabled {
		out.Meta.Shell.Enabled = min.Shell.Enabled
	}
	if min.Shell.HasProgram {
		out.Meta.Shell.Program = min.Shell.Program
	}
	if min.Shell.HasArgs {
		out.Meta.Shell.ArgsTemplate = append([]string(nil), min.Shell.ArgsTemplate...)
	}
	if min.Shell.HasWorkingDir {
		out.Meta.Shell.WorkingDir = min.Shell.WorkingDir
	}
	if min.Shell.HasEnv {
		out.Meta.Shell.Env = make(map[string]string, len(min.Shell.Env))
		for k, v := range min.Shell.Env {
			out.Meta.Shell.Env[k] = v
		}
	}
	if min.Shell.HasTimeout {
		out.Meta.Shell.TimeoutMs = min.Shell.TimeoutMs
	}
	if min.Shell.HasCaptureStdout {
		out.Meta.Shell.Capture.Stdout = min.Shell.CaptureStdout
	}
	if min.Shell.HasCaptureStderr {
		out.Meta.Shell.Capture.Stderr = min.Shell.CaptureStderr
	}
	if min.Shell.HasCaptureMax {
		out.Meta.Shell.Capture.MaxBytes = min.Shell.CaptureMaxBytes
	}
	if min.Shell.HasStrictTpl {
		out.Meta.Shell.StrictTemplating = min.Shell.StrictTemplating
	}
	if min.Shell.HasKillPG {
		out.Meta.Shell.KillProcessGroup = min.Shell.KillProcessGroup
	}
	if min.Shell.HasTermGrace {
		out.Meta.Shell.TermGraceMs = min.Shell.TermGraceMs
	}
	if out.Meta.Shell.Program == "" {
		out.Meta.Shell.Program = defaultShellProgram
	}
	if out.Meta.Shell.WorkingDir == "" {
		out.Meta.Shell.WorkingDir = defaultShellWorkingDir
	}
	if out.Meta.Shell.Env == nil {
		out.Meta.Shell.Env = map[string]string{}
	}
	if out.Meta.Shell.TimeoutMs < 0 {
		out.Meta.Shell.TimeoutMs = defaultShellTimeoutMs
	}
	if out.Meta.Shell.Capture.MaxBytes < 0 {
		out.Meta.Shell.Capture.MaxBytes = defaultShellCaptureMaxBytes
	}
	if out.Meta.Shell.TermGraceMs < 0 {
		out.Meta.Shell.TermGraceMs = defaultShellTermGraceMs
	}
}

func applyOutputMeta(out *Envelope, min config.Minimal) {
	if min.Output.HasOut || min.Output.HasPretty || min.Output.HasLines {
		if out.Meta.Output == nil {
			out.Meta.Output = &OutputMeta{}
		}
		if min.Output.HasOut {
			if min.Output.Out == "" {
				out.Meta.Output.Out = "-"
			} else {
				out.Meta.Output.Out = min.Output.Out
			}
		}
		if min.Output.HasPretty {
			out.Meta.Output.Pretty = min.Output.Pretty
		}
		if min.Output.HasLines {
			out.Meta.Output.Lines = min.Output.Lines
		}
	}
}
