// File Guide for dev/ai agents:
// Purpose: Hold build-time CLI version variables injected by ldflags for release and diagnostic output.
// Responsibilities:
// - Define the exported Version and Date variables used by build and version-reporting code.
// - Keep the ldflags injection target path stable.
// - Avoid runtime logic in this package.
// Architecture notes:
// - This package is intentionally tiny because its main job is to provide a stable symbol path for build-time metadata injection.
// - The real user-visible version command reads from internal/buildinfo; this file exists for linker wiring compatibility.
package cli

// Version and Date should be set at build time using ldflags, e.g.:
//
//	-ldflags "-X 'github.com/flarebyte/thoth-ostraca/cli.Version=1.2.3' -X 'github.com/flarebyte/thoth-ostraca/cli.Date=2026-02-09'"
var (
	Version string
	Date    string
)
