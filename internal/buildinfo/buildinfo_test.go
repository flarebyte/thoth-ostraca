package buildinfo

import (
	"testing"

	"github.com/flarebyte/thoth-ostraca/cli"
)

func TestSummary(t *testing.T) {
	oldV, oldC, oldD := Version, Commit, Date
	oldCliV, oldCliD := cli.Version, cli.Date
	defer func() {
		Version, Commit, Date = oldV, oldC, oldD
		cli.Version, cli.Date = oldCliV, oldCliD
	}()

	Version = ""
	Commit = "abcdef123456"
	Date = ""
	cli.Version = "1.2.3"
	cli.Date = "2026-01-01"

	got := Summary()
	if got == "" || got[:5] != "1.2.3" {
		t.Fatalf("unexpected summary: %q", got)
	}
	if got == "1.2.3" {
		t.Fatalf("expected commit/date suffix, got %q", got)
	}
}
