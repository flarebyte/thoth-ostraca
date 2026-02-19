package stage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateConfig_ExposesShellDefaultsWhenSectionPresent(t *testing.T) {
	_ = os.MkdirAll("temp", 0o755)
	cfg := filepath.Join("temp", "shell_validate_test.cue")
	content := "{\n  configVersion: \"v0\"\n  action: \"nop\"\n  shell: { enabled: true, argsTemplate: [\"-c\", \"echo ok\"] }\n}\n"
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	in := Envelope{Records: []Record{}, Meta: &Meta{ConfigPath: cfg}}
	out, err := Run(context.Background(), "validate-config", in, Deps{})
	if err != nil {
		t.Fatalf("validate-config: %v", err)
	}
	if out.Meta == nil || out.Meta.Shell == nil {
		t.Fatalf("missing meta.shell")
	}
	s := out.Meta.Shell
	if !s.Enabled {
		t.Fatalf("expected enabled=true")
	}
	if s.Program != "bash" || s.WorkingDir != "." || s.TimeoutMs != 60000 || s.TermGraceMs != 2000 {
		t.Fatalf("unexpected shell defaults: %+v", s)
	}
	if !s.Capture.Stdout || !s.Capture.Stderr || s.Capture.MaxBytes != 1048576 {
		t.Fatalf("unexpected capture defaults: %+v", s.Capture)
	}
	if !s.StrictTemplating || !s.KillProcessGroup {
		t.Fatalf("unexpected strict/kill defaults: %+v", s)
	}
	if len(s.ArgsTemplate) != 2 {
		t.Fatalf("expected argsTemplate copied")
	}
	if s.Env == nil || len(s.Env) != 0 {
		t.Fatalf("expected empty env map")
	}
}
