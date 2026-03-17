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
