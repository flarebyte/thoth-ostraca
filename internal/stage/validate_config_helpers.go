package stage

import "github.com/flarebyte/thoth-ostraca/internal/config"

const defaultMaxRecordsInMemory = 10000
const defaultLuaTimeoutMs = 2000
const defaultLuaInstructionLimit = 1000000
const defaultLuaMemoryLimitBytes = 8388608
const defaultShellProgram = "bash"
const defaultShellWorkingDir = "."
const defaultShellTimeoutMs = 60000
const defaultShellCaptureMaxBytes = 1048576
const defaultShellTermGraceMs = 2000

// sanitizeWorkers ensures a minimum of 1 worker when present.
func sanitizeWorkers(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

// applyMinimalToMeta mutates out.Meta to reflect values from the parsed minimal config.
// It mirrors the original field population and presence checks exactly.
func applyMinimalToMeta(out *Envelope, min config.Minimal) {
	if out.Meta == nil {
		out.Meta = &Meta{}
	}
	out.Meta.Config = &ConfigMeta{ConfigVersion: min.ConfigVersion, Action: min.Action}
	out.Meta.ConfigPath = "" // do not persist configPath in output

	// Discovery
	if min.Discovery.HasRoot || min.Discovery.HasNoGitignore || min.Discovery.HasFollowSymlink {
		if out.Meta.Discovery == nil {
			out.Meta.Discovery = &DiscoveryMeta{}
		}
		if min.Discovery.HasRoot {
			out.Meta.Discovery.Root = min.Discovery.Root
		}
		if min.Discovery.HasNoGitignore {
			out.Meta.Discovery.NoGitignore = min.Discovery.NoGitignore
		}
		if min.Discovery.HasFollowSymlink {
			out.Meta.Discovery.FollowSymlinks = min.Discovery.FollowSymlinks
		}
	}

	// Validation
	if min.Validation.HasAllowUnknownTop {
		if out.Meta.Validation == nil {
			out.Meta.Validation = &ValidationMeta{}
		}
		out.Meta.Validation.AllowUnknownTopLevel = min.Validation.AllowUnknownTopLevel
	}

	// Limits
	if out.Meta.Limits == nil {
		out.Meta.Limits = &LimitsMeta{MaxRecordsInMemory: defaultMaxRecordsInMemory}
	}
	if min.Limits.HasMaxYAMLBytes {
		out.Meta.Limits.MaxYAMLBytes = min.Limits.MaxYAMLBytes
	}
	if min.Limits.HasMaxRecordsInMemory {
		if min.Limits.MaxRecordsInMemory < 1 {
			out.Meta.Limits.MaxRecordsInMemory = 1
		} else {
			out.Meta.Limits.MaxRecordsInMemory = min.Limits.MaxRecordsInMemory
		}
	}
	if out.Meta.Limits.MaxRecordsInMemory < 1 {
		out.Meta.Limits.MaxRecordsInMemory = defaultMaxRecordsInMemory
	}

	// Lua: filter, map, postmap, reduce
	if min.Filter.HasInline {
		if out.Meta.Lua == nil {
			out.Meta.Lua = &LuaMeta{}
		}
		out.Meta.Lua.FilterInline = min.Filter.Inline
	}
	if min.Map.HasInline {
		if out.Meta.Lua == nil {
			out.Meta.Lua = &LuaMeta{}
		}
		out.Meta.Lua.MapInline = min.Map.Inline
	}
	if min.PostMap.HasInline {
		if out.Meta.Lua == nil {
			out.Meta.Lua = &LuaMeta{}
		}
		out.Meta.Lua.PostMapInline = min.PostMap.Inline
	}
	if min.Reduce.HasInline {
		if out.Meta.Lua == nil {
			out.Meta.Lua = &LuaMeta{}
		}
		out.Meta.Lua.ReduceInline = min.Reduce.Inline
	}

	// Lua sandbox (only when lua section is present in config)
	if min.LuaSandbox.HasSection {
		if out.Meta.LuaSandbox == nil {
			out.Meta.LuaSandbox = &LuaSandboxMeta{
				TimeoutMs:        defaultLuaTimeoutMs,
				InstructionLimit: defaultLuaInstructionLimit,
				MemoryLimitBytes: defaultLuaMemoryLimitBytes,
				Libs: LuaSandboxLibsMeta{
					Base:   true,
					Table:  true,
					String: true,
					Math:   true,
				},
				DeterministicRandom: true,
			}
		}
		if min.LuaSandbox.HasTimeoutMs {
			out.Meta.LuaSandbox.TimeoutMs = min.LuaSandbox.TimeoutMs
		}
		if min.LuaSandbox.HasInstructionLimit {
			out.Meta.LuaSandbox.InstructionLimit = min.LuaSandbox.InstructionLimit
		}
		if min.LuaSandbox.HasMemoryLimitBytes {
			out.Meta.LuaSandbox.MemoryLimitBytes = min.LuaSandbox.MemoryLimitBytes
		}
		if min.LuaSandbox.HasDeterministicRandom {
			out.Meta.LuaSandbox.DeterministicRandom = min.LuaSandbox.DeterministicRandom
		}
		if min.LuaSandbox.Libs.HasBase {
			out.Meta.LuaSandbox.Libs.Base = min.LuaSandbox.Libs.Base
		}
		if min.LuaSandbox.Libs.HasTable {
			out.Meta.LuaSandbox.Libs.Table = min.LuaSandbox.Libs.Table
		}
		if min.LuaSandbox.Libs.HasString {
			out.Meta.LuaSandbox.Libs.String = min.LuaSandbox.Libs.String
		}
		if min.LuaSandbox.Libs.HasMath {
			out.Meta.LuaSandbox.Libs.Math = min.LuaSandbox.Libs.Math
		}
	}

	// Shell
	if min.Shell.HasSection {
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

	// Output
	if min.Output.HasOut || min.Output.HasPretty || min.Output.HasLines {
		if out.Meta.Output == nil {
			out.Meta.Output = &OutputMeta{}
		}
		if min.Output.HasOut {
			out.Meta.Output.Out = min.Output.Out
		}
		if min.Output.HasPretty {
			out.Meta.Output.Pretty = min.Output.Pretty
		}
		if min.Output.HasLines {
			out.Meta.Output.Lines = min.Output.Lines
		}
	}

	// Errors
	if min.Errors.HasMode || min.Errors.HasEmbed {
		if out.Meta.Errors == nil {
			out.Meta.Errors = &ErrorsMeta{}
		}
		if min.Errors.HasMode {
			out.Meta.Errors.Mode = min.Errors.Mode
		}
		if min.Errors.HasEmbed {
			out.Meta.Errors.EmbedErrors = min.Errors.EmbedErrors
		}
	}

	// FileInfo
	if min.FileInfo.HasEnabled {
		if out.Meta.FileInfo == nil {
			out.Meta.FileInfo = &FileInfoMeta{}
		}
		out.Meta.FileInfo.Enabled = min.FileInfo.Enabled
	}

	// Git
	if min.Git.HasEnabled {
		if out.Meta.Git == nil {
			out.Meta.Git = &GitMeta{}
		}
		out.Meta.Git.Enabled = min.Git.Enabled
	}

	// Workers (only when present in CUE)
	if min.Workers.HasCount {
		out.Meta.Workers = sanitizeWorkers(min.Workers.Count)
	}

	// LocatorPolicy
	if (min.LocatorPolicy.HasAllowAbs || min.LocatorPolicy.HasAllowParent || min.LocatorPolicy.HasPosix || min.LocatorPolicy.HasAllowURLs) || out.Meta.LocatorPolicy != nil {
		if out.Meta.LocatorPolicy == nil {
			out.Meta.LocatorPolicy = &LocatorPolicy{
				AllowAbsolute:   false,
				AllowParentRefs: false,
				PosixStyle:      true,
				AllowURLs:       false,
			}
		}
		if min.LocatorPolicy.HasAllowAbs {
			out.Meta.LocatorPolicy.AllowAbsolute = min.LocatorPolicy.AllowAbsolute
		}
		if min.LocatorPolicy.HasAllowParent {
			out.Meta.LocatorPolicy.AllowParentRefs = min.LocatorPolicy.AllowParentRefs
		}
		if min.LocatorPolicy.HasPosix {
			out.Meta.LocatorPolicy.PosixStyle = min.LocatorPolicy.PosixStyle
		}
		if min.LocatorPolicy.HasAllowURLs {
			out.Meta.LocatorPolicy.AllowURLs = min.LocatorPolicy.AllowURLs
		}
	}
}
