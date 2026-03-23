// File Guide for dev/ai agents:
// Purpose: Resolve deterministic sidecar output roots and file paths for create, update, and file-pipeline persistence flows.
// Responsibilities:
// - Compute the effective persistence root from discovery root and optional outDir.
// - Derive the canonical sidecar relative path from a locator.
// - Return matching absolute and relative paths for downstream read/write stages.
// Architecture notes:
// - Path resolution is centralized here so alongside-source mode and dedicated outDir mode cannot drift across stages.
// - Relative sidecar paths stay locator-based even when outDir is used, which keeps output contracts stable across persistence modes.
package stage

import "path/filepath"

func persistMetaRoot(meta *Meta, discoveryRoot string) string {
	root := discoveryRoot
	if meta == nil || meta.PersistMeta == nil || meta.PersistMeta.OutDir == "" {
		return root
	}
	if filepath.IsAbs(meta.PersistMeta.OutDir) {
		return meta.PersistMeta.OutDir
	}
	return filepath.Join(root, meta.PersistMeta.OutDir)
}

func persistMetaFilePath(
	meta *Meta,
	discoveryRoot string,
	locator string,
) (abs, rel string) {
	rel = filepath.ToSlash(filepath.Join(locator + ".thoth.yaml"))
	abs = filepath.Join(
		persistMetaRoot(meta, discoveryRoot),
		filepath.FromSlash(rel),
	)
	return abs, rel
}
