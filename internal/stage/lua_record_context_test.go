package stage

import "testing"

func TestLuaRecordContext(t *testing.T) {
	rec := Record{
		Locator: "sub/a.go",
		Meta:    map[string]any{"name": "A"},
		FileInfo: &RecFileInfo{
			Size:    12,
			Mode:    "-rw-r--r--",
			ModTime: "2026-01-01T00:00:00Z",
			IsDir:   false,
		},
		Git: &RecGit{
			Tracked: true,
			Ignored: false,
			Status:  "M",
			LastCommit: &RecGitCommit{
				Hash:   "abc",
				Author: "dev",
				Time:   "2026-01-01T00:00:00Z",
			},
		},
	}

	got := luaRecordContext(rec)
	if got["locator"] != "sub/a.go" {
		t.Fatalf("locator mismatch: %#v", got["locator"])
	}
	fileInfo, ok := got["fileInfo"].(map[string]any)
	if !ok || fileInfo["size"] != int64(12) {
		t.Fatalf("fileInfo mismatch: %#v", got["fileInfo"])
	}
	git, ok := got["git"].(map[string]any)
	if !ok || git["status"] != "M" {
		t.Fatalf("git mismatch: %#v", got["git"])
	}
	lastCommit, ok := git["lastCommit"].(map[string]any)
	if !ok || lastCommit["hash"] != "abc" {
		t.Fatalf("lastCommit mismatch: %#v", git["lastCommit"])
	}
}
