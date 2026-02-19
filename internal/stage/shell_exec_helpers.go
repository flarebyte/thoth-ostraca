package stage

import (
	"regexp"
)

const shellExecStage = "shell-exec"

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
