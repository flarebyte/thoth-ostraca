package stage

import (
	"context"
	"path/filepath"
)

func discoverRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// If no discovery root configured, passthrough
	if in.Meta == nil || in.Meta.Discovery == nil || in.Meta.Discovery.Root == "" {
		return in, nil
	}
	root := in.Meta.Discovery.Root
	noGitignore := in.Meta.Discovery.NoGitignore
	followSymlinks := in.Meta.Discovery.FollowSymlinks
	mode, _ := errorMode(in.Meta)

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Envelope{}, err
	}

	locators, envErrs, err := findThothYAMLs(absRoot, noGitignore, followSymlinks, mode)
	if err != nil {
		return Envelope{}, err
	}
	out := in
	if len(envErrs) > 0 {
		out.Errors = append(out.Errors, envErrs...)
		SortEnvelopeErrors(&out)
	}
	out.Records = make([]Record, 0, len(locators))
	for _, l := range locators {
		out.Records = append(out.Records, Record{Locator: l})
	}
	if out.Meta != nil && out.Meta.Config != nil && out.Meta.Config.Action == "diff-meta" {
		out.Meta.MetaFiles = append([]string(nil), locators...)
	}
	return out, nil
}

// ignored reports whether the relative path should be ignored per .gitignore files.

func init() { Register("discover-meta-files", discoverRunner) }
