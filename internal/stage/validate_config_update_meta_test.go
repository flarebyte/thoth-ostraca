package stage

import (
	"strings"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/config"
)

func TestValidateConfig_ExposesUpdateMetaPatch(t *testing.T) {
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"update-meta\"\n  updateMeta: { patch: { b: 2, obj: { y: 9 } }, expectedLua: { inline: \"return function(locator, existingMeta) return {} end\" } }\n}\n"
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
	if out.Meta.UpdateMeta.ExpectedLuaInline == "" {
		t.Fatalf("expected updateMeta.expectedLuaInline")
	}
}

func TestValidateConfig_UpdateMetaPatchMustBeObject(t *testing.T) {
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"update-meta\"\n  updateMeta: { patch: 1 }\n}\n"
	_, err := runValidateConfigWithContent(t, "update_meta_patch_invalid_validate_test.cue", content)
	if err == nil || !strings.Contains(err.Error(), "invalid updateMeta.patch") {
		t.Fatalf("expected invalid updateMeta.patch error, got: %v", err)
	}
}

func TestValidateConfig_UpdateMetaExpectedLuaMustBeString(t *testing.T) {
	content := "{\n  configVersion: \"" + config.CurrentConfigVersion + "\"\n  action: \"update-meta\"\n  updateMeta: { expectedLua: { inline: 1 } }\n}\n"
	_, err := runValidateConfigWithContent(t, "update_meta_expected_lua_invalid_validate_test.cue", content)
	if err == nil || !strings.Contains(err.Error(), "invalid updateMeta.expectedLua.inline") {
		t.Fatalf("expected invalid updateMeta.expectedLua.inline error, got: %v", err)
	}
}

func TestValidateConfig_PersistMetaOnlyAllowedForInputPipeline(t *testing.T) {
	content := "{\n  configVersion: \"" +
		config.CurrentConfigVersion +
		"\"\n  action: \"update-meta\"\n  persistMeta: { enabled: true }\n}\n"
	_, err := runValidateConfigWithContent(
		t,
		"persist_meta_invalid_action_validate_test.cue",
		content,
	)
	if err == nil || !strings.Contains(err.Error(), "invalid persistMeta") {
		t.Fatalf("expected invalid persistMeta error, got: %v", err)
	}
}

func TestValidateConfig_PersistMetaOutDirRequiresEnabledAndNonEmpty(t *testing.T) {
	content := "{\n  configVersion: \"" +
		config.CurrentConfigVersion +
		"\"\n  action: \"input-pipeline\"\n  persistMeta: { outDir: \"   \" }\n}\n"
	_, err := runValidateConfigWithContent(
		t,
		"persist_meta_outdir_invalid_validate_test.cue",
		content,
	)
	if err == nil || !strings.Contains(err.Error(), "invalid persistMeta.outDir") {
		t.Fatalf("expected invalid persistMeta.outDir error, got: %v", err)
	}
}
