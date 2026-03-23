package stage

import (
	"strings"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/config"
)

func TestValidateConfig_DiscoveryIncludeExcludeMustBeNonEmpty(t *testing.T) {
	content := "{\n  configVersion: \"" +
		config.CurrentConfigVersion +
		"\"\n  action: \"input-pipeline\"\n  discovery: { include: [\"\"], exclude: [\"ok/**\"] }\n}\n"
	_, err := runValidateConfigWithContent(
		t,
		"discovery_include_invalid_validate_test.cue",
		content,
	)
	if err == nil || !strings.Contains(err.Error(), "invalid discovery.include") {
		t.Fatalf("expected invalid discovery.include error, got: %v", err)
	}
}
