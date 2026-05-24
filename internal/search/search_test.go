package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
	"github.com/flarebyte/thoth-ostraca/internal/testutil"
)

func TestExecute_SearchProjectionAndOrdering(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	testutil.MustWriteFile(t, filepath.Join(root, "b.thoth.yaml"), `locator: src/b.go
meta:
  language: go
  purpose: parser
  score: 7
`)
	testutil.MustWriteFile(t, filepath.Join(root, "a.thoth.yaml"), `locator: src/a.go
meta:
  language: Go
  purpose: workflow
`)
	testutil.MustWriteFile(t, filepath.Join(root, "sub", "c.thoth.yaml"), `locator: src/c.go
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
	testutil.MustWriteFile(t, filepath.Join(root, "a.thoth.yaml"), `locator: x/a.go
meta:
  language: go
`)
	testutil.MustWriteFile(t, filepath.Join(root, "b.thoth.yaml"), `locator: x/b.go
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
	testutil.MustWriteFile(t, filepath.Join(root, "bad.thoth.yaml"), `locator: x/a.go
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

func TestExecute_InvalidRootErrors(t *testing.T) {
	t.Parallel()

	if _, err := Execute(context.Background(), Options{Root: "/definitely/not/existing/path"}); err == nil {
		t.Fatalf("expected invalid root error")
	}

	root := t.TempDir()
	filePath := filepath.Join(root, "notdir")
	testutil.MustWriteFile(t, filePath, "x")
	if _, err := Execute(context.Background(), Options{Root: filePath}); err == nil {
		t.Fatalf("expected not-a-directory error")
	}
}

func TestSearchRecordBranches(t *testing.T) {
	t.Parallel()

	if item, err := searchRecord(stage.Record{Locator: "a"}, "", nil); err != nil || item != nil {
		t.Fatalf("expected nil result for nil meta, got item=%+v err=%v", item, err)
	}

	if item, err := searchRecord(stage.Record{Locator: "a", Meta: map[string]any{"k": "v"}}, "zzz", nil); err != nil || item != nil {
		t.Fatalf("expected term miss, got item=%+v err=%v", item, err)
	}

	if item, err := searchRecord(stage.Record{Locator: "a", Meta: map[string]any{"k": "v"}}, "v", []string{"k"}); err != nil || item == nil || item.Meta["k"] != "v" {
		t.Fatalf("expected projected match item, got item=%+v err=%v", item, err)
	}

	// json.Marshal should fail on function values.
	if _, err := searchRecord(stage.Record{Locator: "a", Meta: map[string]any{"bad": func() {}}}, "x", nil); err == nil {
		t.Fatalf("expected marshal error")
	}
}

type errWriter struct{}

func (errWriter) Write(_ []byte) (int, error) { return 0, io.ErrClosedPipe }

func TestWriteJSON_ErrorBranches(t *testing.T) {
	t.Parallel()

	// stdout writer error
	if err := WriteJSON([]ResultItem{{Locator: "a", Meta: map[string]any{"k": "v"}}}, "", errWriter{}); err == nil {
		t.Fatalf("expected stdout write error")
	}

	// marshal error for unsupported value
	if err := WriteJSON([]ResultItem{{Locator: "a", Meta: map[string]any{"bad": func() {}}}}, "", os.Stdout); err == nil {
		t.Fatalf("expected marshal error")
	}

	// mkdir failure: parent is a file
	root := t.TempDir()
	blocker := filepath.Join(root, "blocker")
	testutil.MustWriteFile(t, blocker, "x")
	if err := WriteJSON([]ResultItem{{Locator: "a", Meta: map[string]any{"k": "v"}}}, filepath.Join(blocker, "out.json"), os.Stdout); err == nil {
		t.Fatalf("expected mkdir failure")
	}

	// rename failure: destination path is an existing directory
	outDir := filepath.Join(root, "dirdest")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := WriteJSON([]ResultItem{{Locator: "a", Meta: map[string]any{"k": "v"}}}, outDir, os.Stdout); err == nil {
		t.Fatalf("expected rename failure")
	}
}

func TestExecute_DeterministicAcrossRuns(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for i := 0; i < 40; i++ {
		name := fmt.Sprintf("f%02d.thoth.yaml", i)
		locator := fmt.Sprintf("src/f%02d.go", 39-i)
		body := fmt.Sprintf("locator: %s\nmeta:\n  language: go\n  purpose: item-%02d\n", locator, i)
		testutil.MustWriteFile(t, filepath.Join(root, "nested", name), body)
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
