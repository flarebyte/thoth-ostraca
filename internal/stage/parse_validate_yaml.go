package stage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

func parseValidateYAMLRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Determine root
	root := "."
	if in.Meta != nil && in.Meta.Discovery != nil && in.Meta.Discovery.Root != "" {
		root = in.Meta.Discovery.Root
	}

	// Collect output
	type kv struct {
		locator string
		meta    map[string]any
	}
	outs := make([]kv, 0, len(in.Records))

	for _, r := range in.Records {
		// Expect each r to be map[...], with locator string
		m, ok := r.(map[string]any)
		if !ok {
			return Envelope{}, errors.New("invalid input record: expected object")
		}
		locVal, ok := m["locator"]
		if !ok {
			return Envelope{}, errors.New("invalid input record: missing locator")
		}
		locator, ok := locVal.(string)
		if !ok || locator == "" {
			return Envelope{}, errors.New("invalid input record: locator must be string")
		}

		// Read + parse YAML
		p := filepath.Join(root, filepath.FromSlash(locator))
		b, err := os.ReadFile(p)
		if err != nil {
			return Envelope{}, fmt.Errorf("read error %s: %w", p, err)
		}
		var y any
		if err := yaml.Unmarshal(b, &y); err != nil {
			return Envelope{}, fmt.Errorf("invalid YAML %s: %v", p, err)
		}
		ym, ok := y.(map[string]any)
		if !ok {
			return Envelope{}, fmt.Errorf("invalid YAML %s: top-level must be mapping", p)
		}
		// Validate required fields
		yloc, ok := ym["locator"]
		if !ok {
			return Envelope{}, fmt.Errorf("invalid YAML %s: missing required field: locator", p)
		}
		ylocStr, ok := yloc.(string)
		if !ok || ylocStr == "" {
			return Envelope{}, fmt.Errorf("invalid YAML %s: invalid type for field: locator", p)
		}
		ymeta, ok := ym["meta"]
		if !ok {
			return Envelope{}, fmt.Errorf("invalid YAML %s: missing required field: meta", p)
		}
		ymetaMap, ok := ymeta.(map[string]any)
		if !ok {
			return Envelope{}, fmt.Errorf("invalid YAML %s: invalid type for field: meta", p)
		}

		outs = append(outs, kv{locator: ylocStr, meta: ymetaMap})
	}

	sort.Slice(outs, func(i, j int) bool { return outs[i].locator < outs[j].locator })

	out := in
	out.Records = make([]any, 0, len(outs))
	for _, pr := range outs {
		out.Records = append(out.Records, map[string]any{
			"locator": pr.locator,
			"meta":    pr.meta,
		})
	}
	return out, nil
}

func init() { Register("parse-validate-yaml", parseValidateYAMLRunner) }
