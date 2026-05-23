package search

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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

func TestExecute_DeterministicAcrossRuns(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for i := 0; i < 40; i++ {
		name := fmt.Sprintf("f%02d.thoth.yaml", i)
		locator := fmt.Sprintf("src/f%02d.go", 39-i)
		body := fmt.Sprintf("locator: %s\nmeta:\n  language: go\n  purpose: item-%02d\n", locator, i)
		mustWrite(t, filepath.Join(root, "nested", name), body)
	}

	first, err := Execute(context.Background(), Options{Root: root, Term: "GO"})
	if err != nil {
		t.Fatalf("Execute() first run error = %v", err)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first result: %v", err)
	}

	for i := 0; i < 5; i++ {
		got, err := Execute(context.Background(), Options{Root: root, Term: "GO"})
		if err != nil {
			t.Fatalf("Execute() run %d error = %v", i+2, err)
		}
		if !reflect.DeepEqual(first, got) {
			t.Fatalf("run %d produced non-deterministic result", i+2)
		}
		gotJSON, err := json.Marshal(got)
		if err != nil {
			t.Fatalf("marshal run %d result: %v", i+2, err)
		}
		if string(firstJSON) != string(gotJSON) {
			t.Fatalf("run %d produced non-deterministic JSON", i+2)
		}
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
