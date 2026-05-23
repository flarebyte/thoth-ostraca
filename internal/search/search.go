package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

// Options configures thoth metadata search behavior.
type Options struct {
	Root   string
	Term   string
	Fields []string
	Out    string
}

// ResultItem is one JSON output record.
type ResultItem struct {
	Locator string         `json:"locator"`
	Meta    map[string]any `json:"meta"`
}

// Execute runs the metadata search and returns deterministic results.
func Execute(ctx context.Context, opts Options) ([]ResultItem, error) {
	root := opts.Root
	if strings.TrimSpace(root) == "" {
		root = "."
	}

	// Ensure the root exists and is a directory before running discovery.
	fi, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("invalid root %q: %w", root, err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("invalid root %q: not a directory", root)
	}

	env := stage.Envelope{
		Meta: &stage.Meta{
			Discovery: &stage.DiscoveryMeta{Root: root},
		},
	}

	out, err := stage.Run(ctx, "discover-meta-files", env, stage.Deps{})
	if err != nil {
		return nil, err
	}
	out, err = stage.Run(ctx, "parse-validate-yaml", out, stage.Deps{})
	if err != nil {
		return nil, err
	}
	out, err = stage.Run(ctx, "validate-locators", out, stage.Deps{})
	if err != nil {
		return nil, err
	}

	term := strings.ToLower(strings.TrimSpace(opts.Term))
	fields := normalizeFields(opts.Fields)

	results := make([]ResultItem, 0, len(out.Records))
	for _, rec := range out.Records {
		if rec.Meta == nil {
			continue
		}
		if term != "" {
			metaBlob, err := json.Marshal(rec.Meta)
			if err != nil {
				return nil, fmt.Errorf("marshal meta for locator %q: %w", rec.Locator, err)
			}
			if !strings.Contains(strings.ToLower(string(metaBlob)), term) {
				continue
			}
		}

		projected := projectMeta(rec.Meta, fields)
		results = append(results, ResultItem{
			Locator: rec.Locator,
			Meta:    projected,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Locator < results[j].Locator
	})
	return results, nil
}

// WriteJSON writes results to stdout when outPath is empty, otherwise to outPath.
func WriteJSON(results []ResultItem, outPath string, stdout io.Writer) error {
	blob, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal result JSON: %w", err)
	}
	blob = append(blob, '\n')

	if strings.TrimSpace(outPath) == "" {
		if _, err := stdout.Write(blob); err != nil {
			return fmt.Errorf("write stdout: %w", err)
		}
		return nil
	}

	dir := filepath.Dir(outPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	tmp := outPath + ".tmp"
	if err := os.WriteFile(tmp, blob, 0o644); err != nil {
		return fmt.Errorf("write temporary output file: %w", err)
	}
	if err := os.Rename(tmp, outPath); err != nil {
		return fmt.Errorf("finalize output file: %w", err)
	}
	return nil
}

func normalizeFields(fields []string) []string {
	if len(fields) == 0 {
		return nil
	}
	out := make([]string, 0, len(fields))
	seen := map[string]struct{}{}
	for _, f := range fields {
		for _, part := range strings.Split(f, ",") {
			k := strings.TrimSpace(part)
			if k == "" {
				continue
			}
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			out = append(out, k)
		}
	}
	return out
}

func projectMeta(meta map[string]any, fields []string) map[string]any {
	if len(fields) == 0 {
		return meta
	}
	out := make(map[string]any, len(fields))
	for _, f := range fields {
		if v, ok := meta[f]; ok {
			out[f] = v
		}
	}
	return out
}
