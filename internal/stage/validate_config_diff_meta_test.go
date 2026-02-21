package stage

import (
	"strings"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/config"
)

func TestValidateConfig_ExposesDiffMetaExpectedPatch_Default(t *testing.T) {
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"diff-meta\"\n}\n"
	out, err := runValidateConfigWithContent(t, "diff_meta_expected_patch_default_validate_test.cue", content)
	if err != nil {
		t.Fatalf("validate-config: %v", err)
	}
	if out.Meta == nil || out.Meta.DiffMeta == nil || out.Meta.DiffMeta.ExpectedPatch == nil {
		t.Fatalf("missing meta.diffMeta.expectedPatch")
	}
	if out.Meta.DiffMeta.Format != "summary" {
		t.Fatalf("expected default format=summary, got: %q", out.Meta.DiffMeta.Format)
	}
	if out.Meta.DiffMeta.FailOnChange {
		t.Fatalf("expected default failOnChange=false")
	}
	if len(out.Meta.DiffMeta.ExpectedPatch) != 0 {
		t.Fatalf("expected empty default patch, got: %+v", out.Meta.DiffMeta.ExpectedPatch)
	}
}

func TestValidateConfig_ExposesDiffMetaExpectedPatch_Configured(t *testing.T) {
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"diff-meta\"\n  diffMeta: { format: \"detailed\", failOnChange: true, expectedPatch: { b: 2, obj: { y: 9 } } }\n}\n"
	out, err := runValidateConfigWithContent(t, "diff_meta_expected_patch_validate_test.cue", content)
	if err != nil {
		t.Fatalf("validate-config: %v", err)
	}
	if out.Meta == nil || out.Meta.DiffMeta == nil || out.Meta.DiffMeta.ExpectedPatch == nil {
		t.Fatalf("missing meta.diffMeta.expectedPatch")
	}
	if out.Meta.DiffMeta.Format != "detailed" {
		t.Fatalf("expected format=detailed, got: %q", out.Meta.DiffMeta.Format)
	}
	if !out.Meta.DiffMeta.FailOnChange {
		t.Fatalf("expected failOnChange=true")
	}
	if out.Meta.DiffMeta.ExpectedPatch["b"] != int64(2) && out.Meta.DiffMeta.ExpectedPatch["b"] != 2 {
		t.Fatalf("unexpected patch value: %+v", out.Meta.DiffMeta.ExpectedPatch)
	}
}

func TestValidateConfig_DiffMetaExpectedPatchMustBeObject(t *testing.T) {
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"diff-meta\"\n  diffMeta: { expectedPatch: 1 }\n}\n"
	_, err := runValidateConfigWithContent(t, "diff_meta_expected_patch_invalid_validate_test.cue", content)
	if err == nil || !strings.Contains(err.Error(), "invalid diffMeta.expectedPatch") {
		t.Fatalf("expected invalid diffMeta.expectedPatch error, got: %v", err)
	}
}

func TestValidateConfig_DiffMetaFormatMustBeKnown(t *testing.T) {
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"diff-meta\"\n  diffMeta: { format: \"full\" }\n}\n"
	_, err := runValidateConfigWithContent(t, "diff_meta_format_invalid_validate_test.cue", content)
	if err == nil || !strings.Contains(err.Error(), "invalid diffMeta.format") {
		t.Fatalf("expected invalid diffMeta.format error, got: %v", err)
	}
}

func TestValidateConfig_DiffMetaFormatJSONPatch(t *testing.T) {
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"diff-meta\"\n  diffMeta: { format: \"json-patch\" }\n}\n"
	out, err := runValidateConfigWithContent(t, "diff_meta_format_jsonpatch_validate_test.cue", content)
	if err != nil {
		t.Fatalf("validate-config: %v", err)
	}
	if out.Meta == nil || out.Meta.DiffMeta == nil {
		t.Fatalf("missing meta.diffMeta")
	}
	if out.Meta.DiffMeta.Format != "json-patch" {
		t.Fatalf("expected format=json-patch, got: %q", out.Meta.DiffMeta.Format)
	}
}

func TestValidateConfig_DiffMetaFailOnChangeMustBeBoolean(t *testing.T) {
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"diff-meta\"\n  diffMeta: { failOnChange: \"yes\" }\n}\n"
	_, err := runValidateConfigWithContent(t, "diff_meta_fail_on_change_invalid_validate_test.cue", content)
	if err == nil || !strings.Contains(err.Error(), "invalid diffMeta.failOnChange") {
		t.Fatalf("expected invalid diffMeta.failOnChange error, got: %v", err)
	}
}
