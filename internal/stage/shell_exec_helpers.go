// File Guide for dev/ai agents:
// Purpose: Hold the compact shared shell stage option struct and stage constant.
// Responsibilities:
// - Define the internal shellOptions shape used across shell helper files.
// - Provide the shell stage name constant shared in diagnostics.
// - Keep shell helper coupling explicit without exporting stage internals broadly.
// Architecture notes:
// - This file is intentionally small; it exists to avoid circular drift in option shape across multiple shell helper files.
package stage

const shellExecStage = "shell-exec"

type shellOptions struct {
	enabled          bool
	decodeJSONStdout bool
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
