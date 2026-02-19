package stage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateConfig_ExposesDiffMetaExpectedPatch_Default(t *testing.T) {
	_ = os.MkdirAll("temp", 0o755)
	cfg := filepath.Join("temp", "diff_meta_expected_patch_default_validate_test.cue")
	content := "{\n  configVersion: \"v0\"\n  action: \"diff-meta\"\n}\n"
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	in := Envelope{Records: []Record{}, Meta: &Meta{ConfigPath: cfg}}
	out, err := Run(context.Background(), "validate-config", in, Deps{})
	if err != nil {
		t.Fatalf("validate-config: %v", err)
	}
	if out.Meta == nil || out.Meta.DiffMeta == nil || out.Meta.DiffMeta.ExpectedPatch == nil {
		t.Fatalf("missing meta.diffMeta.expectedPatch")
	}
	if len(out.Meta.DiffMeta.ExpectedPatch) != 0 {
		t.Fatalf("expected empty default patch, got: %+v", out.Meta.DiffMeta.ExpectedPatch)
	}
}

func TestValidateConfig_ExposesDiffMetaExpectedPatch_Configured(t *testing.T) {
	_ = os.MkdirAll("temp", 0o755)
	cfg := filepath.Join("temp", "diff_meta_expected_patch_validate_test.cue")
	content := "{\n  configVersion: \"v0\"\n  action: \"diff-meta\"\n  diffMeta: { expectedPatch: { b: 2, obj: { y: 9 } } }\n}\n"
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	in := Envelope{Records: []Record{}, Meta: &Meta{ConfigPath: cfg}}
	out, err := Run(context.Background(), "validate-config", in, Deps{})
	if err != nil {
		t.Fatalf("validate-config: %v", err)
	}
	if out.Meta == nil || out.Meta.DiffMeta == nil || out.Meta.DiffMeta.ExpectedPatch == nil {
		t.Fatalf("missing meta.diffMeta.expectedPatch")
	}
	if out.Meta.DiffMeta.ExpectedPatch["b"] != int64(2) && out.Meta.DiffMeta.ExpectedPatch["b"] != 2 {
		t.Fatalf("unexpected patch value: %+v", out.Meta.DiffMeta.ExpectedPatch)
	}
}

func TestValidateConfig_DiffMetaExpectedPatchMustBeObject(t *testing.T) {
	_ = os.MkdirAll("temp", 0o755)
	cfg := filepath.Join("temp", "diff_meta_expected_patch_invalid_validate_test.cue")
	content := "{\n  configVersion: \"v0\"\n  action: \"diff-meta\"\n  diffMeta: { expectedPatch: 1 }\n}\n"
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	in := Envelope{Records: []Record{}, Meta: &Meta{ConfigPath: cfg}}
	_, err := Run(context.Background(), "validate-config", in, Deps{})
	if err == nil || !strings.Contains(err.Error(), "invalid diffMeta.expectedPatch") {
		t.Fatalf("expected invalid diffMeta.expectedPatch error, got: %v", err)
	}
}
