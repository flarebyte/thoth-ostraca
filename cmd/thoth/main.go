package main

import (
	"os"
	"strings"

	"github.com/flarebyte/thoth-ostraca/cmd/thoth/root"
)

func main() {
	if err := root.Execute(os.Args[1:]); err != nil {
		// Print a short, single-line error to stderr on failures.
		// Do not print usage or stack traces.
		msg := strings.Join(strings.Fields(err.Error()), " ")
		if msg == "" {
			msg = "error"
		}
		_, _ = os.Stderr.WriteString(msg + "\n")
		os.Exit(1)
	}
}
