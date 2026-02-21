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
