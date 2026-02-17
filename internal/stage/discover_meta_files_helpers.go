package stage

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	gitgitignore "github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// hasThothYAML reports whether a filename ends with .thoth.yaml
func hasThothYAML(name string) bool {
	return strings.HasSuffix(name, ".thoth.yaml")
}

// dirsForRel returns the list of directories from "." to the directory of rel.
func dirsForRel(rel string) []string {
	dir := filepath.Dir(rel)
	if rel == "." {
		dir = "."
	}
	parts := []string{}
	if dir != "." {
		parts = strings.Split(dir, string(os.PathSeparator))
	}
	cur := "."
	dirs := []string{"."}
	for _, part := range parts {
		if cur == "." {
			cur = part
		} else {
			cur = filepath.Join(cur, part)
		}
		dirs = append(dirs, cur)
	}
	return dirs
}

// readGitignorePatterns reads .gitignore patterns from the given directories under absRoot.
func readGitignorePatterns(absRoot string, dirs []string) []gitgitignore.Pattern {
	var patterns []gitgitignore.Pattern
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
	return patterns
}

// matchIgnore reports whether rel should be ignored according to .gitignore files under absRoot.
func matchIgnore(absRoot string, rel string, isDir bool) bool {
	patterns := readGitignorePatterns(absRoot, dirsForRel(rel))
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

func displayDiscoveryPath(absRoot string, p string) string {
	rel, err := filepath.Rel(absRoot, p)
	if err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(p)
}

// findThothYAMLs walks absRoot and returns sorted relative locators of *.thoth.yaml files,
// respecting .gitignore patterns unless noGitignore is true.
func findThothYAMLs(absRoot string, noGitignore bool, mode string) ([]string, []Error, error) {
	var locators []string
	var envErrs []Error
	err := filepath.WalkDir(absRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			pathLabel := displayDiscoveryPath(absRoot, p)
			if mode == "keep-going" {
				envErrs = append(envErrs, Error{Stage: "discover-meta-files", Locator: pathLabel, Message: err.Error()})
				if d != nil && d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			return fmt.Errorf("discover-meta-files: %s: %v", pathLabel, err)
		}
		if p == absRoot {
			return nil
		}
		rel, err := filepath.Rel(absRoot, p)
		if err != nil {
			pathLabel := displayDiscoveryPath(absRoot, p)
			if mode == "keep-going" {
				envErrs = append(envErrs, Error{Stage: "discover-meta-files", Locator: pathLabel, Message: err.Error()})
				return nil
			}
			return fmt.Errorf("discover-meta-files: %s: %v", pathLabel, err)
		}
		isDir := d.IsDir()
		if !noGitignore {
			if matchIgnore(absRoot, rel, isDir) {
				if isDir {
					return fs.SkipDir
				}
				return nil
			}
		}
		if !isDir && hasThothYAML(d.Name()) {
			locators = append(locators, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(locators)
	return locators, envErrs, nil
}
