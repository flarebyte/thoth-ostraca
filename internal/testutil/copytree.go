// File Guide for dev/ai agents:
// Purpose: Provide a tiny recursive fixture-copy helper for tests that need isolated writable repos.
// Responsibilities:
// - Remove any existing destination tree before copying.
// - Recreate directories and files under a new root.
// - Preserve simple test fixture contents without extra metadata handling.
// Architecture notes:
// - This helper is intentionally minimal and test-only; it does not try to preserve permissions, symlinks, or special files.
// - The destructive destination cleanup is expected in tests and should not be copied into production code paths.
package testutil

import (
	"io/fs"
	"os"
	"path/filepath"
)

func CopyTree(src, dst string) error {
	_ = os.RemoveAll(dst)
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		out := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	})
}
