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
