package stage

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLParser_DuplicateTopLevelKeyBehavior(t *testing.T) {
	const doc = "locator: first\nlocator: second\nmeta: {}\n"
	var out map[string]any
	err := yaml.Unmarshal([]byte(doc), &out)
	if err == nil {
		t.Fatalf("expected duplicate key parse error, got nil")
	}
	msg := err.Error()
	if msg != "yaml: unmarshal errors:\n  line 2: mapping key \"locator\" already defined at line 1" {
		t.Fatalf("unexpected duplicate-key error: %q", msg)
	}
}
