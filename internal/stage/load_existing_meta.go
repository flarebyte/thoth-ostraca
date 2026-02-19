package stage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const loadExistingStage = "load-existing-meta"

func loadOneExisting(root string, rec Record) (Record, *Error, error) {
	_, rel := metaFilePath(root, rec.Locator)
	abs := filepath.Join(root, filepath.FromSlash(rel))
	b, err := os.ReadFile(abs)
	if err != nil {
		// Not found → expose path only
		if os.IsNotExist(err) {
			// attach existingMetaPath only
			m := map[string]any{"existingMetaPath": rel}
			if rec.Post != nil {
				if pm, ok := rec.Post.(map[string]any); ok {
					for k, v := range pm {
						m[k] = v
					}
				}
			}
			rec.Post = m
			return rec, nil, nil
		}
		return rec, &Error{Stage: loadExistingStage, Locator: rec.Locator, Message: err.Error()}, err
	}
	var y any
	if err := yaml.Unmarshal(b, &y); err != nil {
		return rec, &Error{Stage: loadExistingStage, Locator: rec.Locator, Message: fmt.Sprintf("invalid YAML: %v", err)}, err
	}
	ym, ok := y.(map[string]any)
	if !ok {
		return rec, &Error{Stage: loadExistingStage, Locator: rec.Locator, Message: "top-level must be mapping"}, fmt.Errorf("invalid meta YAML: %s", rel)
	}
	yloc, ok := ym["locator"].(string)
	if !ok || yloc == "" {
		return rec, &Error{Stage: loadExistingStage, Locator: rec.Locator, Message: "missing or invalid locator"}, fmt.Errorf("invalid meta YAML: %s", rel)
	}
	ymeta, ok := ym["meta"].(map[string]any)
	if !ok {
		return rec, &Error{Stage: loadExistingStage, Locator: rec.Locator, Message: "missing or invalid meta"}, fmt.Errorf("invalid meta YAML: %s", rel)
	}
	m := map[string]any{"existingMetaPath": rel, "existingMeta": ymeta}
	if rec.Post != nil {
		if pm, ok := rec.Post.(map[string]any); ok {
			for k, v := range pm {
				m[k] = v
			}
		}
	}
	rec.Post = m
	return rec, nil, nil
}

func loadExistingMetaRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	root := determineRoot(in)
	mode, embed := errorMode(in.Meta)
	return runSequentialRecordStage(in, loadExistingStage, mode, embed, func(r Record) (Record, *Error, error) {
		return loadOneExisting(root, r)
	})
}

func init() { Register(loadExistingStage, loadExistingMetaRunner) }
