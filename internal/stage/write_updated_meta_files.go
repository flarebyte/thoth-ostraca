package stage

import (
	"context"
	"path/filepath"

	"github.com/flarebyte/thoth-ostraca/internal/metafile"
)

const writeUpdatedMetaFilesStage = "write-updated-meta-files"

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
	// nextMeta
	next := map[string]any{}
	if r.Post != nil {
		if pm, ok := r.Post.(map[string]any); ok {
			if nm, ok2 := pm["nextMeta"].(map[string]any); ok2 {
				next = nm
			}
		}
	}
	if err := metafile.Write(abs, r.Locator, next); err != nil {
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
