// File Guide for dev/ai agents:
// Purpose: Persist nextMeta payloads back to sidecar files, including outDir and dry-run modes used by update and input-pipeline workflows.
// Responsibilities:
// - Resolve the target sidecar path from existingMetaPath or locator-based defaults.
// - Write merged metadata to disk or mark the write as skipped in dry-run mode.
// - Preserve metaPath in post-state and emit progress and envelope errors during persistence.
// Architecture notes:
// - Dry-run is implemented here, not in merge_meta.go, so the analysis and merge path is identical between preview and real writes.
// - The stage prefers an existingMetaPath when available so updates preserve the originally discovered sidecar location under outDir mode.
package stage

import (
	"context"
	"path/filepath"

	"github.com/flarebyte/thoth-ostraca/internal/metafile"
)

const writeUpdatedMetaFilesStage = "write-updated-meta-files"

func writeOneUpdatedWithMeta(meta *Meta, root string, r Record) (Record, *Error, error) {
	rel := ""
	if r.Post != nil {
		if pm, ok := r.Post.(map[string]any); ok {
			if s, ok2 := pm["existingMetaPath"].(string); ok2 && s != "" {
				rel = s
			}
		}
	}
	if rel == "" {
		_, rel = persistMetaFilePath(meta, root, r.Locator)
	}
	abs := filepath.Join(persistMetaRoot(meta, root), filepath.FromSlash(rel))
	dryRun := meta != nil &&
		meta.PersistMeta != nil &&
		meta.PersistMeta.DryRun
	next := map[string]any{}
	if r.Post != nil {
		if pm, ok := r.Post.(map[string]any); ok {
			if nm, ok2 := pm["nextMeta"].(map[string]any); ok2 {
				next = nm
			}
		}
	}
	if !dryRun {
		if err := metafile.Write(abs, r.Locator, next); err != nil {
			return r, &Error{Stage: writeUpdatedMetaFilesStage, Locator: r.Locator, Message: err.Error()}, err
		}
	}
	// Attach metaPath (for output symmetry)
	m := map[string]any{"metaPath": rel}
	if dryRun {
		m["writeSkipped"] = "dry-run"
	}
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
	reporter := ProgressReporterFromContext(ctx)
	total := len(in.Records)
	completed := 0
	var envErrs []Error
	for i, r := range in.Records {
		if r.Error != nil {
			continue
		}
		rr, envE, err := writeOneUpdatedWithMeta(in.Meta, root, r)
		if envE != nil {
			envErrs = append(envErrs, *envE)
		}
		if err != nil {
			if mode == "keep-going" {
				if embed {
					rr = r
					msg := "write updated meta failed"
					if envE != nil {
						msg = sanitizeErrorMessage(envE.Message)
					}
					rr.Error = &RecError{Stage: writeUpdatedMetaFilesStage, Message: msg}
				}
				out.Records[i] = rr
				continue
			}
			return Envelope{}, err
		}
		out.Records[i] = rr
		completed++
		if reporter != nil {
			reporter.ReportProgress(ProgressEvent{
				Stage:     writeUpdatedMetaFilesStage,
				Event:     "progress",
				Completed: completed,
				Total:     total,
				Rejected:  0,
				Errors:    len(envErrs),
			})
		}
	}
	if len(envErrs) > 0 {
		for _, e := range envErrs {
			out.Errors = append(out.Errors, sanitizedError(e))
		}
		SortEnvelopeErrors(&out)
	}
	return out, nil
}

func init() { Register(writeUpdatedMetaFilesStage, writeUpdatedMetaFilesRunner) }
