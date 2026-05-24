package search

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/testutil"
)

func TestSearchHelpShowsFlags(t *testing.T) {
	cmd := NewCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	out := buf.String()
	for _, want := range []string{"--root", "--term", "--fields", "--out"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help output missing %q:\n%s", want, out)
		}
	}
}

func TestSearchCommandOutputsJSON(t *testing.T) {
	root := t.TempDir()
	testutil.MustWriteFile(t, filepath.Join(root, "a.thoth.yaml"), `locator: src/a.go
meta:
  language: go
  purpose: parser
`)

	cmd := NewCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--root", root, "--term", "PARSER", "--fields", "language"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"locator": "src/a.go"`) {
		t.Fatalf("missing locator in output: %s", out)
	}
	if !strings.Contains(out, `"language": "go"`) {
		t.Fatalf("missing projected field in output: %s", out)
	}
	if strings.Contains(out, `"purpose"`) {
		t.Fatalf("unexpected unprojected field in output: %s", out)
	}
}

func TestSearchCommandOutFile(t *testing.T) {
	root := t.TempDir()
	outPath := filepath.Join(root, "out", "search.json")
	testutil.MustWriteFile(t, filepath.Join(root, "a.thoth.yaml"), `locator: src/a.go
meta:
  language: go
`)

	cmd := NewCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--root", root, "--out", outPath})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "" {
		t.Fatalf("expected no stdout when --out is used, got %q", got)
	}
	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(b), `"locator": "src/a.go"`) {
		t.Fatalf("output file missing locator: %s", string(b))
	}
}
