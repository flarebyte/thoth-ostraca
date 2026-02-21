package stage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestEnrichGit_RepoNotFound_FailFast(t *testing.T) {
	root := filepath.Join(t.TempDir(), "no-repo")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	in := Envelope{
		Records: []Record{{Locator: "a.txt"}},
		Meta: &Meta{
			Discovery: &DiscoveryMeta{Root: root},
			Git:       &GitMeta{Enabled: true},
			Errors:    &ErrorsMeta{Mode: "fail-fast"},
		},
	}
	_, err := enrichGitRunner(context.Background(), in, Deps{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := err.Error(); got != "enrich-git: git repo not found" {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestEnrichGit_RepoNotFound_KeepGoing(t *testing.T) {
	root := filepath.Join(t.TempDir(), "no-repo")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	in := Envelope{
		Records: []Record{{Locator: "a.txt"}},
		Meta: &Meta{
			Discovery: &DiscoveryMeta{Root: root},
			Git:       &GitMeta{Enabled: true},
			Errors:    &ErrorsMeta{Mode: "keep-going", EmbedErrors: true},
		},
	}
	out, err := enrichGitRunner(context.Background(), in, Deps{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Errors) != 1 {
		t.Fatalf("expected one envelope error, got %d", len(out.Errors))
	}
	if out.Errors[0].Stage != enrichGitStage || out.Errors[0].Message != "git repo not found" {
		t.Fatalf("unexpected envelope error: %+v", out.Errors[0])
	}
	if len(out.Records) != 1 || out.Records[0].Error != nil || out.Records[0].Git != nil {
		t.Fatalf("record should be unchanged on keep-going repo-not-found: %+v", out.Records)
	}
}
