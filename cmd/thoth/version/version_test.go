package version

import (
	"io"
	"os"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/buildinfo"
)

func TestVersionDefaultOutputStable(t *testing.T) {
	oldVersion, oldCommit, oldDate := buildinfo.Version, buildinfo.Commit, buildinfo.Date
	oldShort, oldJSON := flagShort, flagJSON
	defer func() {
		buildinfo.Version, buildinfo.Commit, buildinfo.Date = oldVersion, oldCommit, oldDate
		flagShort, flagJSON = oldShort, oldJSON
	}()

	buildinfo.Version = ""
	buildinfo.Commit = ""
	buildinfo.Date = ""
	flagShort = false
	flagJSON = false

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	if err := VersionCmd.RunE(VersionCmd, nil); err != nil {
		t.Fatalf("run: %v", err)
	}
	_ = w.Close()
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "thoth dev\n" {
		t.Fatalf("unexpected output: %q", string(got))
	}
}
