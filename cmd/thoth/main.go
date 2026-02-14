package main

import (
	"os"

	"github.com/flarebyte/thoth-ostraca/cmd/thoth/root"
)

func main() {
	if err := root.Execute(os.Args[1:]); err != nil {
		// Print a short, single-line error to stderr on failures.
		// Do not print usage or stack traces.
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
