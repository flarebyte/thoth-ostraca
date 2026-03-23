// File Guide for dev/ai agents:
// Purpose: Create new empty .thoth.yaml sidecars for discovered input files during create-meta workflows.
// Responsibilities:
// - Resolve the target sidecar path for each locator.
// - Refuse to overwrite an existing sidecar when bootstrapping metadata.
// - Write the initial sidecar file and expose its path in record post-state.
// Architecture notes:
// - This stage writes only baseline empty metadata on purpose; richer metadata derivation belongs to programmable input-pipeline flows.
// - Create-meta keeps fail-on-existing behavior so accidental re-bootstrap runs are visible instead of silently mutating existing sidecars.
package stage

import (
	"context"
	"fmt"
	"os"

	"github.com/flarebyte/thoth-ostraca/internal/metafile"
)

const writeMetaFilesStage = "write-meta-files"

func writeSingleMeta(meta *Meta, root string, rec Record) (Record, *Error, error) {
	abs, rel := persistMetaFilePath(meta, root, rec.Locator)
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
		return writeSingleMeta(in.Meta, root, r)
	})
}

func init() { Register(writeMetaFilesStage, writeMetaFilesRunner) }
