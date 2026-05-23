package root

import (
	"testing"
)

func TestNewRootCmdAndExecuteHelp(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Use != "thoth" {
		t.Fatalf("unexpected use: %s", cmd.Use)
	}
	if len(cmd.Commands()) < 4 {
		t.Fatalf("expected subcommands, got %d", len(cmd.Commands()))
	}
	if err := Execute([]string{}); err != nil {
		t.Fatalf("Execute empty args should show help, got err: %v", err)
	}
}
