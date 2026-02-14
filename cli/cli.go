package cli

// Version and Date should be set at build time using ldflags, e.g.:
//
//	-ldflags "-X 'github.com/flarebyte/thoth-ostraca/cli.Version=1.2.3' -X 'github.com/flarebyte/thoth-ostraca/cli.Date=2026-02-09'"
var (
	Version string
	Date    string
)
