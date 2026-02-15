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
		isDir := d.IsDir()
		if !noGitignore {
			if matchIgnore(absRoot, rel, isDir) {
				if isDir {
					return fs.SkipDir
				}
				return nil
			}
		}
		if isDir {
			return nil
		}
		// Exclude existing meta files
		if strings.HasSuffix(d.Name(), ".thoth.yaml") {
			return nil
		}
		// Exclude .gitignore files themselves
		if d.Name() == ".gitignore" {
			return nil
		}
		locators = append(locators, filepath.ToSlash(rel))
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
