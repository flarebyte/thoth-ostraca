package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMinimal_UnknownConfigVersion(t *testing.T) {
	d := t.TempDir()
	cfg := filepath.Join(d, "unknown_version.cue")
	content := "{\n  configVersion: \"2\"\n  action: \"nop\"\n}\n"
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	_, err := ParseMinimal(cfg)
	if err == nil {
		t.Fatalf("expected error")
	}
	want := "unsupported configVersion: \"2\" (supported: 1)"
	if err.Error() != want {
		t.Fatalf("unexpected error\nwant: %s\n got: %s", want, err.Error())
	}
}
