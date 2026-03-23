// File Guide for dev/ai agents:
// Purpose: Load existing sidecar YAML into record post-state so later stages can merge or compare metadata against what is already on disk.
// Responsibilities:
// - Resolve the expected sidecar path for each locator, including configured outDir mode.
// - Read and validate the locator/meta structure of existing sidecar YAML files.
// - Attach existing metadata and path information into rec.Post for later merge or persistence stages.
// Architecture notes:
// - Missing sidecars are not errors here; the stage records only the expected path so create/update flows can decide what to do next.
// - Existing metadata is attached in post-state rather than replacing rec.Meta so pipeline metadata and file metadata stay distinct.
package stage

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const loadExistingStage = "load-existing-meta"

func loadOneExistingWithMeta(meta *Meta, root string, rec Record) (Record, *Error, error) {
	abs, rel := persistMetaFilePath(meta, root, rec.Locator)
	b, err := os.ReadFile(abs)
	if err != nil {
		// Not found → expose path only
		if os.IsNotExist(err) {
			// attach existingMetaPath only
			rec.Post = mergePostMap(rec, map[string]any{"existingMetaPath": rel})
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
	rec.Post = mergePostMap(rec, map[string]any{"existingMetaPath": rel, "existingMeta": ymeta})
	return rec, nil, nil
}

func mergePostMap(rec Record, base map[string]any) map[string]any {
	m := make(map[string]any, len(base))
	for k, v := range base {
		m[k] = v
	}
	if pm, ok := rec.Post.(map[string]any); ok {
		for k, v := range pm {
			m[k] = v
		}
	}
	return m
}

func loadExistingMetaRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	root := determineRoot(in)
	mode, embed := errorMode(in.Meta)
	return runSequentialRecordStage(in, loadExistingStage, mode, embed, func(r Record) (Record, *Error, error) {
		return loadOneExistingWithMeta(in.Meta, root, r)
	})
}

func init() { Register(loadExistingStage, loadExistingMetaRunner) }
