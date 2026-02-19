package stage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flarebyte/thoth-ostraca/internal/metafile"
)

const writeMetaFilesStage = "write-meta-files"

func metaFilePath(root, locator string) (abs, rel string) {
	rel = filepath.ToSlash(filepath.Join(locator + ".thoth.yaml"))
	abs = filepath.Join(root, filepath.FromSlash(rel))
	return abs, rel
}

func writeSingleMeta(root string, rec Record) (Record, *Error, error) {
	abs, rel := metaFilePath(root, rec.Locator)
	if _, err := os.Stat(abs); err == nil {
		return rec, &Error{Stage: writeMetaFilesStage, Locator: rec.Locator, Message: fmt.Sprintf("meta exists: %s", rel)}, fmt.Errorf("meta exists: %s", rel)
	}
	if err := metafile.Write(abs, rec.Locator, map[string]any{}); err != nil {
		return rec, &Error{Stage: writeMetaFilesStage, Locator: rec.Locator, Message: err.Error()}, err
	}
	rec.Post = map[string]any{"metaPath": rel}
	return rec, nil, nil
}

func writeMetaFilesRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	root := determineRoot(in)
	out := in
	mode, embed := errorMode(in.Meta)
	var envErrs []Error
	for i, r := range in.Records {
		if r.Error != nil {
			continue
		}
		rec, envE, err := writeSingleMeta(root, r)
		if envE != nil {
			envErrs = append(envErrs, *envE)
		}
		if err != nil {
			if mode == "keep-going" {
				if embed {
					rec = r
					rec.Error = &RecError{Stage: writeMetaFilesStage, Message: envE.Message}
				}
				out.Records[i] = rec
				continue
			}
			return Envelope{}, err
		}
		out.Records[i] = rec
	}
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
	}
	return out, nil
}

func init() { Register(writeMetaFilesStage, writeMetaFilesRunner) }
