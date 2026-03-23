// File Guide for dev/ai agents:
// Purpose: Provide the real process entrypoint that executes the root command and normalizes top-level CLI exit behavior.
// Responsibilities:
// - Invoke the root command with process arguments.
// - Render one short single-line error to stderr on failure.
// - Exit with the mapped code when commands return an exitCoder.
// Architecture notes:
// - Error text is whitespace-normalized here so CLI failures stay compact and predictable across nested command errors.
// - Exit code mapping is handled only at the top level so subcommands can return typed errors without depending on os.Exit directly.
package main

import (
	"os"
	"strings"

	"github.com/flarebyte/thoth-ostraca/cmd/thoth/root"
)

type exitCoder interface {
	ExitCode() int
}

func main() {
	if err := root.Execute(os.Args[1:]); err != nil {
		// Print a short, single-line error to stderr on failures.
		// Do not print usage or stack traces.
		msg := strings.Join(strings.Fields(err.Error()), " ")
		if msg == "" {
			msg = "error"
		}
		_, _ = os.Stderr.WriteString(msg + "\n")
		code := 1
		if ec, ok := err.(exitCoder); ok {
			if c := ec.ExitCode(); c != 0 {
				code = c
			}
		}
		os.Exit(code)
	}
}
