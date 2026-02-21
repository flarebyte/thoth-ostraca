package stage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/config"
)

func TestValidateConfig_ExposesLuaSandboxDefaultsWhenLuaSectionPresent(t *testing.T) {
	_ = os.MkdirAll("temp", 0o755)
	cfg := filepath.Join("temp", "lua_sandbox_validate_test.cue")
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"nop\"\n  lua: {}\n}\n"
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	in := Envelope{Records: []Record{}, Meta: &Meta{ConfigPath: cfg}}
	out, err := Run(context.Background(), "validate-config", in, Deps{})
	if err != nil {
		t.Fatalf("validate-config: %v", err)
	}
	if out.Meta == nil || out.Meta.LuaSandbox == nil {
		t.Fatalf("missing meta.luaSandbox")
	}
	sb := out.Meta.LuaSandbox
	if sb.TimeoutMs != 2000 || sb.InstructionLimit != 1000000 || sb.MemoryLimitBytes != 8388608 {
		t.Fatalf("unexpected defaults: %+v", sb)
	}
	if !sb.Libs.Base || !sb.Libs.Table || !sb.Libs.String || !sb.Libs.Math {
		t.Fatalf("unexpected libs defaults: %+v", sb.Libs)
	}
	if !sb.DeterministicRandom {
		t.Fatalf("expected deterministicRandom=true")
	}
}
