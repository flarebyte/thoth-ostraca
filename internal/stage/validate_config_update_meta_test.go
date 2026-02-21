package stage

import (
	"strings"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/config"
)

func TestValidateConfig_ExposesUpdateMetaPatch(t *testing.T) {
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"update-meta\"\n  updateMeta: { patch: { b: 2, obj: { y: 9 } } }\n}\n"
	out, err := runValidateConfigWithContent(t, "update_meta_patch_validate_test.cue", content)
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
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"update-meta\"\n  updateMeta: { patch: 1 }\n}\n"
	_, err := runValidateConfigWithContent(t, "update_meta_patch_invalid_validate_test.cue", content)
	if err == nil || !strings.Contains(err.Error(), "invalid updateMeta.patch") {
		t.Fatalf("expected invalid updateMeta.patch error, got: %v", err)
	}
}
