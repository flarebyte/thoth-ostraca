// File Guide for dev/ai agents:
// Purpose: Attach normalized filesystem stat information to each input record when file info enrichment is enabled.
// Responsibilities:
// - Read file metadata for each locator relative to the discovery root.
// - Normalize modTime, mode, and size into the RecFileInfo contract.
// - Surface stat failures through the standard stage error model.
// Architecture notes:
// - File timestamps are normalized to UTC and second precision so outputs remain deterministic across platforms and filesystems.
// - This stage enriches records only; it does not influence discovery, so missing files become stage errors rather than silently changing the input set.
package stage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const enrichFileInfoStage = "enrich-fileinfo"

func statToRecInfo(fi os.FileInfo) *RecFileInfo {
	// Normalize to UTC, seconds precision
	mt := fi.ModTime().UTC().Truncate(time.Second).Format(time.RFC3339)
	return &RecFileInfo{
		Size:    fi.Size(),
		Mode:    fi.Mode().String(),
		ModTime: mt,
		IsDir:   fi.IsDir(),
	}
}

func enrichFileInfoRunner(_ context.Context, in Envelope, _ Deps) (Envelope, error) {
	// If disabled, passthrough
	if in.Meta == nil || in.Meta.FileInfo == nil || !in.Meta.FileInfo.Enabled {
		return in, nil
	}
	root := determineRoot(in)
	mode, embed := errorMode(in.Meta)
	out := in
	var envErrs []Error
	for i, r := range in.Records {
		if r.Error != nil {
			continue
		}
		abs := filepath.Join(root, filepath.FromSlash(r.Locator))
		fi, err := os.Stat(abs)
		if err != nil {
			msg := sanitizeErrorMessage(err.Error())
			envErrs = append(envErrs, Error{Stage: enrichFileInfoStage, Locator: r.Locator, Message: msg})
			if mode == "keep-going" {
				if embed {
					rr := r
					rr.Error = &RecError{Stage: enrichFileInfoStage, Message: msg}
					out.Records[i] = rr
				}
				continue
			}
			return Envelope{}, fmt.Errorf("%s: %v", enrichFileInfoStage, err)
		}
		rr := r
		rr.FileInfo = statToRecInfo(fi)
		out.Records[i] = rr
	}
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
		SortEnvelopeErrors(&out)
	}
	return out, nil
}

func init() { Register(enrichFileInfoStage, enrichFileInfoRunner) }
