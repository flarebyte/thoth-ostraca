package stage

import (
	"regexp"
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
