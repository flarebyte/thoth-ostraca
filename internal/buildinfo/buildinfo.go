package buildinfo

import (
	"strings"

	"github.com/flarebyte/thoth-ostraca/cli"
)

// Package buildinfo exposes version metadata for the CLI. Values can be
// overridden at build time via -ldflags. This package also honors values set in
// the cli package (cli.Version/cli.Date) for compatibility with external build scripts.

var (
	// Version is the semantic version or custom string. Defaults to cli.Version or "dev".
	Version = "dev"
	// Commit is the VCS commit hash (optional).
	Commit = ""
	// Date is the build time in RFC3339 or similar (optional). Falls back to cli.Date.
	Date = ""
	// BuiltBy is an optional builder identifier (optional).
	BuiltBy = ""
)

// Summary returns a concise single-line version string.
func Summary() string {
	v := Version
	if v == "" {
		v = cli.Version
	}
	if v == "" {
		v = "dev"
	}

	d := Date
	if d == "" {
		d = cli.Date
	}

	parts := make([]string, 0, 2)
	if Commit != "" {
		c := Commit
		if len(c) > 7 {
			c = c[:7]
		}
		parts = append(parts, "commit="+c)
	}
	if d != "" {
		parts = append(parts, "date="+d)
	}
	if len(parts) > 0 {
		v += " (" + strings.Join(parts, ", ") + ")"
	}
	return v
}
