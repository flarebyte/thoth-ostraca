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
	mode, embed := errorMode(in.Meta)
	return runSequentialRecordStage(in, writeMetaFilesStage, mode, embed, func(r Record) (Record, *Error, error) {
		return writeSingleMeta(root, r)
	})
}

func init() { Register(writeMetaFilesStage, writeMetaFilesRunner) }
