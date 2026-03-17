package stage

import (
	"path/filepath"
	"testing"
)

func TestPersistMetaRootAndPath(t *testing.T) {
	root := filepath.FromSlash("repo/src")
	meta := &Meta{
		PersistMeta: &PersistMetaMeta{
			Enabled: true,
			OutDir:  "../sidecars",
		},
	}
	gotRoot := persistMetaRoot(meta, root)
	wantRoot := filepath.Join(root, "..", "sidecars")
	if gotRoot != wantRoot {
		t.Fatalf("root mismatch: want=%q got=%q", wantRoot, gotRoot)
	}
	abs, rel := persistMetaFilePath(meta, root, "sub/a.go")
	if rel != "sub/a.go.thoth.yaml" {
		t.Fatalf("rel mismatch: %q", rel)
	}
	wantAbs := filepath.Join(wantRoot, "sub", "a.go.thoth.yaml")
	if abs != wantAbs {
		t.Fatalf("abs mismatch: want=%q got=%q", wantAbs, abs)
	}
}
