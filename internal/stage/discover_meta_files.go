package stage

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	gitgitignore "github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

func discoverRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Defaults
	root := "."
	noGitignore := false
	if in.Meta != nil && in.Meta.Discovery != nil {
		if in.Meta.Discovery.Root != "" {
			root = in.Meta.Discovery.Root
		}
		noGitignore = in.Meta.Discovery.NoGitignore
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Envelope{}, err
	}

	var locators []string

	// Walk the filesystem
	err = filepath.WalkDir(absRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == absRoot {
			return nil
		}
		rel, err := filepath.Rel(absRoot, p)
		if err != nil {
			return err
		}

		isDir := d.IsDir()
		if !noGitignore {
			if ignored(absRoot, rel, isDir) {
				if isDir {
					return fs.SkipDir
				}
				return nil
			}
		}

		if !isDir && strings.HasSuffix(d.Name(), ".thoth.yaml") {
			locators = append(locators, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return Envelope{}, err
	}

	sort.Strings(locators)
	out := in
	out.Records = make([]any, 0, len(locators))
	for _, l := range locators {
		out.Records = append(out.Records, map[string]any{"locator": l})
	}
	return out, nil
}

// ignored reports whether the relative path should be ignored per .gitignore files.
func ignored(absRoot string, rel string, isDir bool) bool {
	// Build patterns from all .gitignore files along the path from root to the containing directory
	// of rel.
	var patterns []gitgitignore.Pattern
	dir := filepath.Dir(rel)
	if rel == "." {
		dir = "."
	}
	// Iterate over each prefix directory
	// e.g., for a/b/c -> [".", "a", "a/b", "a/b/c" (if dir)]
	parts := []string{}
	if dir != "." {
		parts = strings.Split(dir, string(os.PathSeparator))
	}
	cur := "."
	// include root directory first
	dirs := []string{"."}
	for _, part := range parts {
		if cur == "." {
			cur = part
		} else {
			cur = filepath.Join(cur, part)
		}
		dirs = append(dirs, cur)
	}
	for _, d := range dirs {
		giPath := filepath.Join(absRoot, d, ".gitignore")
		b, err := os.ReadFile(giPath)
		if err != nil {
			continue
		}
		lines := strings.Split(string(b), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			base := []string{}
			if d != "." && d != "" {
				base = strings.Split(filepath.ToSlash(d), "/")
			}
			patterns = append(patterns, gitgitignore.ParsePattern(line, base))
		}
	}
	if len(patterns) == 0 {
		return false
	}
	m := gitgitignore.NewMatcher(patterns)
	comps := []string{}
	if rel != "." && rel != "" {
		comps = strings.Split(rel, string(os.PathSeparator))
	}
	return m.Match(comps, isDir)
}

func init() { Register("discover-meta-files", discoverRunner) }
