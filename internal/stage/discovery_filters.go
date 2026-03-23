// File Guide for dev/ai agents:
// Purpose: Centralize discovery pattern matching and default path exclusion rules for file-oriented discovery.
// Responsibilities:
// - Read include and exclude lists from envelope metadata.
// - Normalize discovery paths and evaluate glob-like match rules against files and directories.
// - Define which directory names are always or normally excluded during input discovery.
// Architecture notes:
// - Input discovery uses simple normalized pattern helpers here instead of reusing gitignore semantics so include/exclude behavior stays explicit and deterministic.
// - .git is treated as always excluded separately from the broader defaultExcludedDirNames set to prevent accidental opt-in through generic include patterns.
package stage

import (
	"path"
	"strings"
)

var defaultExcludedDirNames = map[string]bool{
	".e2e-bin":     true,
	".gocache":     true,
	".git":         true,
	".gomodcache":  true,
	"__fixtures__": true,
	"fixture":      true,
	"fixtures":     true,
	"node_modules": true,
	"temp":         true,
	"testdata":     true,
	"tmp":          true,
}

func discoveryIncludes(meta *Meta) []string {
	if meta == nil || meta.Discovery == nil {
		return nil
	}
	return meta.Discovery.Include
}

func discoveryExcludes(meta *Meta) []string {
	if meta == nil || meta.Discovery == nil {
		return nil
	}
	return meta.Discovery.Exclude
}

func normalizeDiscoveryPath(s string) string {
	return strings.Trim(strings.ReplaceAll(s, "\\", "/"), "/")
}

func relHasAlwaysExcludedDir(rel string) bool {
	return relHasDirName(rel, ".git")
}

func relHasDefaultExcludedDir(rel string) bool {
	for _, seg := range strings.Split(normalizeDiscoveryPath(rel), "/") {
		if defaultExcludedDirNames[seg] && seg != ".git" {
			return true
		}
	}
	return false
}

func relHasDirName(rel, name string) bool {
	for _, seg := range strings.Split(normalizeDiscoveryPath(rel), "/") {
		if seg == name {
			return true
		}
	}
	return false
}

func matchesDiscoveryPattern(pattern, rel string) bool {
	p := normalizeDiscoveryPath(pattern)
	r := normalizeDiscoveryPath(rel)
	if p == "" {
		return false
	}
	if p == "**" {
		return true
	}
	if strings.HasSuffix(p, "/**") {
		prefix := strings.TrimSuffix(p, "/**")
		return r == prefix || strings.HasPrefix(r, prefix+"/")
	}
	ok, err := path.Match(p, r)
	return err == nil && ok
}

func dirCouldMatchPattern(pattern, dirRel string) bool {
	p := normalizeDiscoveryPath(pattern)
	d := normalizeDiscoveryPath(dirRel)
	if p == "" || p == "**" {
		return true
	}
	if strings.HasSuffix(p, "/**") {
		prefix := strings.TrimSuffix(p, "/**")
		return d == prefix ||
			strings.HasPrefix(prefix, d+"/") ||
			strings.HasPrefix(d, prefix+"/")
	}
	if matchesDiscoveryPattern(p, d) {
		return true
	}
	prefix := staticPatternPrefix(p)
	if prefix == "" {
		return true
	}
	return d == prefix ||
		strings.HasPrefix(prefix, d+"/") ||
		strings.HasPrefix(d, prefix+"/")
}

func staticPatternPrefix(pattern string) string {
	p := normalizeDiscoveryPath(pattern)
	for i := 0; i < len(p); i++ {
		switch p[i] {
		case '*', '?', '[':
			return strings.Trim(p[:i], "/")
		}
	}
	return p
}

func matchesAnyPattern(patterns []string, rel string) bool {
	for _, pattern := range patterns {
		if matchesDiscoveryPattern(pattern, rel) {
			return true
		}
	}
	return false
}

func dirCouldMatchAnyPattern(patterns []string, rel string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		if dirCouldMatchPattern(pattern, rel) {
			return true
		}
	}
	return false
}
