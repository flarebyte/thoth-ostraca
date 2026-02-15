package stage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const writeUpdatedMetaFilesStage = "write-updated-meta-files"

func yamlInlineFromMap(m map[string]any) string {
	if len(m) == 0 {
		return "{}"
	}
	// stable order
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := "{ "
	first := true
	for _, k := range keys {
		if !first {
			out += ", "
		} else {
			first = false
		}
		v := m[k]
		switch vv := v.(type) {
		case string:
			out += fmt.Sprintf("%s: %s", k, vv)
		default:
			out += fmt.Sprintf("%s: %v", k, vv)
		}
	}
	out += " }"
	return out
}

func writeOneUpdated(root string, r Record) (Record, *Error, error) {
	// Determine path
	rel := ""
	if r.Post != nil {
		if pm, ok := r.Post.(map[string]any); ok {
			if s, ok2 := pm["existingMetaPath"].(string); ok2 && s != "" {
				rel = s
			}
		}
	}
	if rel == "" {
		_, rel = metaFilePath(root, r.Locator)
	}
	abs := filepath.Join(root, filepath.FromSlash(rel))
	// Ensure dir
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return r, &Error{Stage: writeUpdatedMetaFilesStage, Locator: r.Locator, Message: err.Error()}, err
	}
	// nextMeta
	next := map[string]any{}
	if r.Post != nil {
		if pm, ok := r.Post.(map[string]any); ok {
			if nm, ok2 := pm["nextMeta"].(map[string]any); ok2 {
				next = nm
			}
		}
	}
	content := fmt.Sprintf("locator: %s\nmeta: %s\n", r.Locator, yamlInlineFromMap(next))
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return r, &Error{Stage: writeUpdatedMetaFilesStage, Locator: r.Locator, Message: err.Error()}, err
	}
	// Attach metaPath (for output symmetry)
	m := map[string]any{"metaPath": rel}
	if r.Post != nil {
		if pm, ok := r.Post.(map[string]any); ok {
			for k, v := range pm {
				m[k] = v
			}
		}
	}
	r.Post = m
	return r, nil, nil
}

func writeUpdatedMetaFilesRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	root := determineRoot(in)
	out := in
	mode, embed := errorMode(in.Meta)
	var envErrs []Error
	for i, r := range in.Records {
		if r.Error != nil {
			continue
		}
		rr, envE, err := writeOneUpdated(root, r)
		if envE != nil {
			envErrs = append(envErrs, *envE)
		}
		if err != nil {
			if mode == "keep-going" {
				if embed {
					rr = r
					rr.Error = &RecError{Stage: writeUpdatedMetaFilesStage, Message: envE.Message}
				}
				out.Records[i] = rr
				continue
			}
			return Envelope{}, err
		}
		out.Records[i] = rr
	}
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
	}
	return out, nil
}

func init() { Register(writeUpdatedMetaFilesStage, writeUpdatedMetaFilesRunner) }
