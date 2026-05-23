package search

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecute_SearchProjectionAndOrdering(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	mustWrite(t, filepath.Join(root, "b.thoth.yaml"), `locator: src/b.go
meta:
  language: go
  purpose: parser
  score: 7
`)
	mustWrite(t, filepath.Join(root, "a.thoth.yaml"), `locator: src/a.go
meta:
  language: Go
  purpose: workflow
`)
	mustWrite(t, filepath.Join(root, "sub", "c.thoth.yaml"), `locator: src/c.go
meta:
  language: markdown
`)

	results, err := Execute(context.Background(), Options{
		Root:   root,
		Term:   "GO",
		Fields: []string{"language,purpose"},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Locator != "src/a.go" || results[1].Locator != "src/b.go" {
		t.Fatalf("unexpected ordering/locators: %+v", results)
	}
	if _, ok := results[0].Meta["language"]; !ok {
		t.Fatalf("expected projected language key in first result")
	}
	if _, ok := results[0].Meta["score"]; ok {
		t.Fatalf("did not expect non-projected score key")
	}
	if _, ok := results[0].Meta["purpose"]; !ok {
		t.Fatalf("expected projected purpose key in first result")
	}
	if _, ok := results[1].Meta["purpose"]; !ok {
		t.Fatalf("expected projected purpose key in second result")
	}
}

func TestExecute_NoTermReturnsAll(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "a.thoth.yaml"), `locator: x/a.go
meta:
  language: go
`)
	mustWrite(t, filepath.Join(root, "b.thoth.yaml"), `locator: x/b.go
meta:
  language: ts
`)

	results, err := Execute(context.Background(), Options{Root: root})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}

func TestExecute_InvalidYAMLFails(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "bad.thoth.yaml"), `locator: x/a.go
meta: [
`)

	_, err := Execute(context.Background(), Options{Root: root})
	if err == nil {
		t.Fatalf("expected error for invalid YAML")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "invalid yaml") {
		t.Fatalf("expected invalid YAML error, got %v", err)
	}
}

func TestWriteJSON_ToFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	out := filepath.Join(root, "out", "search.json")
	items := []ResultItem{{Locator: "a", Meta: map[string]any{"language": "go"}}}
	if err := WriteJSON(items, out, os.Stdout); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "\n  {") {
		t.Fatalf("expected pretty JSON output, got %q", s)
	}
	if !strings.Contains(s, `"locator": "a"`) {
		t.Fatalf("missing locator in output: %q", s)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
