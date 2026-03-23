// File Guide for dev/ai agents:
// Purpose: Discover arbitrary input files for file-oriented actions while applying root, ignore, and include/exclude policy.
// Responsibilities:
// - Walk the configured discovery root and collect eligible non-sidecar files.
// - Apply always-excluded, default-excluded, gitignore, include, and exclude rules deterministically.
// - Materialize sorted input records and diff-meta input metadata from the discovered locators.
// Architecture notes:
// - This file owns input discovery only; pattern matching helpers live in discovery_filters.go and gitignore matching is reused from meta discovery helpers.
// - Existing .thoth.yaml files and .gitignore files are always excluded here so file actions operate on source inputs, not metadata artifacts.
package stage

import (
	"context"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// discover-input-files: find regular files under root (gitignore respected), excluding *.thoth.yaml
func discoverInputFilesRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	root := determineRoot(in)
	noGitignore := false
	includes := discoveryIncludes(in.Meta)
	excludes := discoveryExcludes(in.Meta)
	if in.Meta != nil && in.Meta.Discovery != nil {
		noGitignore = in.Meta.Discovery.NoGitignore
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Envelope{}, err
	}
	var locators []string
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
		rel = filepath.ToSlash(rel)
		isDir := d.IsDir()
		if relHasAlwaysExcludedDir(rel) {
			if isDir {
				return fs.SkipDir
			}
			return nil
		}
		if !noGitignore {
			if matchIgnore(absRoot, rel, isDir) {
				if isDir {
					return fs.SkipDir
				}
				return nil
			}
		}
		if isDir {
			if matchesAnyPattern(excludes, rel) {
				return fs.SkipDir
			}
			explicitlyIncluded := dirCouldMatchAnyPattern(includes, rel)
			if relHasDefaultExcludedDir(rel) && !explicitlyIncluded {
				return fs.SkipDir
			}
			if len(includes) > 0 && !explicitlyIncluded {
				return fs.SkipDir
			}
			return nil
		}
		explicitlyIncluded := matchesAnyPattern(includes, rel)
		// Exclude existing meta files
		if strings.HasSuffix(d.Name(), ".thoth.yaml") {
			return nil
		}
		// Exclude .gitignore files themselves
		if d.Name() == ".gitignore" {
			return nil
		}
		if matchesAnyPattern(excludes, rel) {
			return nil
		}
		if relHasDefaultExcludedDir(rel) && !explicitlyIncluded {
			return nil
		}
		if len(includes) > 0 && !explicitlyIncluded {
			return nil
		}
		locators = append(locators, rel)
		return nil
	})
	if err != nil {
		return Envelope{}, err
	}
	sort.Strings(locators)
	out := in
	out.Records = make([]Record, 0, len(locators))
	for _, l := range locators {
		out.Records = append(out.Records, Record{Locator: l})
	}
	if out.Meta != nil && out.Meta.Config != nil && out.Meta.Config.Action == "diff-meta" {
		out.Meta.Inputs = append([]string(nil), locators...)
	}
	return out, nil
}

func init() { Register("discover-input-files", discoverInputFilesRunner) }
