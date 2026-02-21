package stage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const parseValidateYAMLStage = "parse-validate-yaml"
const defaultMaxYAMLBytes = 1048576

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

func allowUnknownTopLevel(in Envelope) bool {
	if in.Meta != nil && in.Meta.Validation != nil {
		return in.Meta.Validation.AllowUnknownTopLevel
	}
	return false
}

func maxYAMLBytes(in Envelope) int {
	if in.Meta != nil && in.Meta.Limits != nil && in.Meta.Limits.MaxYAMLBytes > 0 {
		return in.Meta.Limits.MaxYAMLBytes
	}
	return defaultMaxYAMLBytes
}

// processYAMLRecord reads, parses, and validates a single YAML record according to the
// stage rules, returning either a kv pair on success, or an env error (keep-going) or
// fatal error.
func processYAMLRecord(rec Record, root string, mode string, allowUnknownTop bool, maxBytes int) (yamlKV, *Error, error) {
	locator := rec.Locator

	p := filepath.Join(root, filepath.FromSlash(locator))
	info, statErr := os.Stat(p)
	if statErr != nil {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: sanitizeErrorMessage(fmt.Sprintf("read error: %v", statErr))}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("read error %s: %s", p, sanitizeErrorMessage(statErr.Error()))
	}
	if info.Size() > int64(maxBytes) {
		msg := fmt.Sprintf("file exceeds maxYAMLBytes limit: %d", maxBytes)
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: sanitizeErrorMessage(msg)}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("yaml too large %s: exceeds maxYAMLBytes %d", p, maxBytes)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: sanitizeErrorMessage(fmt.Sprintf("read error: %v", err))}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("read error %s: %s", p, sanitizeErrorMessage(err.Error()))
	}
	var y any
	if err := yaml.Unmarshal(b, &y); err != nil {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: sanitizeErrorMessage(fmt.Sprintf("invalid YAML: %v", err))}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("invalid YAML %s: %s", p, sanitizeErrorMessage(err.Error()))
	}
	ym, ok := y.(map[string]any)
	if !ok {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: "top-level must be mapping"}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("invalid YAML %s: top-level must be mapping", p)
	}
	if !allowUnknownTop {
		for k := range ym {
			if k != "locator" && k != "meta" {
				if mode == "keep-going" {
					return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: fmt.Sprintf("unknown top-level field: %s", k)}, nil
				}
				return yamlKV{}, nil, fmt.Errorf("invalid YAML %s: unknown top-level field: %s", p, k)
			}
		}
	}
	yloc, ok := ym["locator"]
	if !ok {
		if mode == "keep-going" {
			return yamlKV{locator: locator, meta: nil}, &Error{Stage: parseValidateYAMLStage, Locator: locator, Message: "missing required field: locator"}, nil
		}
		return yamlKV{}, nil, fmt.Errorf("invalid YAML %s: missing required field: locator", p)
	}
	ylocStr, ok := yloc.(string)
	if !ok {
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
