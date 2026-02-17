package stage

import (
	"fmt"
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

func addDiscoveryError(envErrs *[]Error, absRoot string, p string, err error) {
	*envErrs = append(*envErrs, Error{
		Stage:   "discover-meta-files",
		Locator: displayDiscoveryPath(absRoot, p),
		Message: err.Error(),
	})
}

func discoveryFatal(absRoot string, p string, err error) error {
	return fmt.Errorf("discover-meta-files: %s: %v", displayDiscoveryPath(absRoot, p), err)
}

func shouldIgnore(absRoot, rel string, isDir bool, noGitignore bool) bool {
	if noGitignore {
		return false
	}
	return matchIgnore(absRoot, rel, isDir)
}

// findThothYAMLs walks absRoot and returns sorted relative locators of *.thoth.yaml files,
// respecting .gitignore patterns unless noGitignore is true.
func findThothYAMLs(absRoot string, noGitignore bool, followSymlinks bool, mode string) ([]string, []Error, error) {
	var envErrs []Error
	locatorSet := map[string]struct{}{}
	visitedDirs := map[string]struct{}{}

	var walkDir func(string) error
	walkDir = func(dirPath string) error {
		relDir := displayDiscoveryPath(absRoot, dirPath)
		if relDir != "." && shouldIgnore(absRoot, filepath.FromSlash(relDir), true, noGitignore) {
			return nil
		}

		canonDir, err := filepath.EvalSymlinks(dirPath)
		if err != nil {
			if mode == "keep-going" {
				addDiscoveryError(&envErrs, absRoot, dirPath, err)
				return nil
			}
			return discoveryFatal(absRoot, dirPath, err)
		}
		if _, ok := visitedDirs[canonDir]; ok {
			return nil
		}
		visitedDirs[canonDir] = struct{}{}

		entries, err := os.ReadDir(dirPath)
		if err != nil {
			if mode == "keep-going" {
				addDiscoveryError(&envErrs, absRoot, dirPath, err)
				return nil
			}
			return discoveryFatal(absRoot, dirPath, err)
		}

		var symlinkDirs []string
		for _, ent := range entries {
			name := ent.Name()
			childPath := filepath.Join(dirPath, name)
			relChild, err := filepath.Rel(absRoot, childPath)
			if err != nil {
				if mode == "keep-going" {
					addDiscoveryError(&envErrs, absRoot, childPath, err)
					continue
				}
				return discoveryFatal(absRoot, childPath, err)
			}

			info, err := os.Lstat(childPath)
			if err != nil {
				if mode == "keep-going" {
					addDiscoveryError(&envErrs, absRoot, childPath, err)
					continue
				}
				return discoveryFatal(absRoot, childPath, err)
			}
			isSymlink := (info.Mode() & os.ModeSymlink) != 0

			if isSymlink {
				targetInfo, err := os.Stat(childPath)
				if err != nil {
					if mode == "keep-going" {
						addDiscoveryError(&envErrs, absRoot, childPath, err)
						continue
					}
					return discoveryFatal(absRoot, childPath, err)
				}
				if targetInfo.IsDir() {
					if followSymlinks {
						symlinkDirs = append(symlinkDirs, childPath)
					}
					continue
				}
			}

			if info.IsDir() {
				if err := walkDir(childPath); err != nil {
					return err
				}
				continue
			}

			if shouldIgnore(absRoot, relChild, false, noGitignore) {
				continue
			}
			if hasThothYAML(name) {
				locatorSet[filepath.ToSlash(relChild)] = struct{}{}
			}
		}

		sort.Strings(symlinkDirs)
		for _, symlinkDir := range symlinkDirs {
			relChild, err := filepath.Rel(absRoot, symlinkDir)
			if err != nil {
				if mode == "keep-going" {
					addDiscoveryError(&envErrs, absRoot, symlinkDir, err)
					continue
				}
				return discoveryFatal(absRoot, symlinkDir, err)
			}
			if shouldIgnore(absRoot, relChild, true, noGitignore) {
				continue
			}
			if err := walkDir(symlinkDir); err != nil {
				return err
			}
		}
		return nil
	}

	if err := walkDir(absRoot); err != nil {
		return nil, nil, err
	}

	locators := make([]string, 0, len(locatorSet))
	for l := range locatorSet {
		locators = append(locators, l)
	}
	sort.Strings(locators)
	return locators, envErrs, nil
}
