package stage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const parseValidateYAMLStage = "parse-validate-yaml"

type yamlKV struct {
	locator string
	meta    map[string]any
}

// determineRoot returns the discovery root, defaulting to ".".
func determineRoot(in Envelope) string {
	root := "."
	if in.Meta != nil && in.Meta.Discovery != nil && in.Meta.Discovery.Root != "" {
		root = in.Meta.Discovery.Root
	}
	return root
}

// processYAMLRecord reads, parses, and validates a single YAML record according to the
// stage rules, returning either a kv pair on success, or an env error (keep-going) or
// fatal error.
func processYAMLRecord(r any, root string, mode string) (yamlKV, *Error, error) {
	var locator string
	switch rec := r.(type) {
	case Record:
		locator = rec.Locator
	case map[string]any:
		locVal, ok := rec["locator"]
		if !ok {
			return yamlKV{}, nil, fmt.Errorf("invalid input record: missing locator")
		}
		s, ok := locVal.(string)
		if !ok || s == "" {
			return yamlKV{}, nil, fmt.Errorf("invalid input record: locator must be string")
		}
		locator = s
	default:
		return yamlKV{}, nil, fmt.Errorf("invalid input record: expected object")
	}

	p := filepath.Join(root, filepath.FromSlash(locator))
	b, err := os.ReadFile(p)
	if err != nil {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: fmt.Sprintf("read error: %v", err)}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("read error %s: %w", p, err)
	}
	var y any
	if err := yaml.Unmarshal(b, &y); err != nil {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: fmt.Sprintf("invalid YAML: %v", err)}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("invalid YAML %s: %v", p, err)
	}
	ym, ok := y.(map[string]any)
	if !ok {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: "top-level must be mapping"}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("invalid YAML %s: top-level must be mapping", p)
	}
	yloc, ok := ym["locator"]
	if !ok {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: "missing required field: locator"}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("invalid YAML %s: missing required field: locator", p)
	}
	ylocStr, ok := yloc.(string)
	if !ok || ylocStr == "" {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: "invalid type for field: locator"}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("invalid YAML %s: invalid type for field: locator", p)
	}
	ymeta, ok := ym["meta"]
	if !ok {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: "missing required field: meta"}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("invalid YAML %s: missing required field: meta", p)
	}
	ymetaMap, ok := ymeta.(map[string]any)
	if !ok {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: "invalid type for field: meta"}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("invalid YAML %s: invalid type for field: meta", p)
	}
	return yamlKV{locator: ylocStr, meta: ymetaMap}, nil, nil
}
