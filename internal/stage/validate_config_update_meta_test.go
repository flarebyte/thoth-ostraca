package stage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateConfig_ExposesUpdateMetaPatch(t *testing.T) {
	_ = os.MkdirAll("temp", 0o755)
	cfg := filepath.Join("temp", "update_meta_patch_validate_test.cue")
	content := "{\n  configVersion: \"v0\"\n  action: \"update-meta\"\n  updateMeta: { patch: { b: 2, obj: { y: 9 } } }\n}\n"
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	in := Envelope{Records: []Record{}, Meta: &Meta{ConfigPath: cfg}}
	out, err := Run(context.Background(), "validate-config", in, Deps{})
	if err != nil {
		t.Fatalf("validate-config: %v", err)
	}
	if out.Meta == nil || out.Meta.UpdateMeta == nil || out.Meta.UpdateMeta.Patch == nil {
		t.Fatalf("missing meta.updateMeta.patch")
	}
	if out.Meta.UpdateMeta.Patch["b"] != int64(2) && out.Meta.UpdateMeta.Patch["b"] != 2 {
		t.Fatalf("unexpected patch value: %+v", out.Meta.UpdateMeta.Patch)
	}
}

func TestValidateConfig_UpdateMetaPatchMustBeObject(t *testing.T) {
	_ = os.MkdirAll("temp", 0o755)
	cfg := filepath.Join("temp", "update_meta_patch_invalid_validate_test.cue")
	content := "{\n  configVersion: \"v0\"\n  action: \"update-meta\"\n  updateMeta: { patch: 1 }\n}\n"
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	in := Envelope{Records: []Record{}, Meta: &Meta{ConfigPath: cfg}}
	_, err := Run(context.Background(), "validate-config", in, Deps{})
	if err == nil || !strings.Contains(err.Error(), "invalid updateMeta.patch") {
		t.Fatalf("expected invalid updateMeta.patch error, got: %v", err)
	}
}
