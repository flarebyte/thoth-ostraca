package stage

import (
	"strings"
	"time"
)

func normalizeRFC3339(t time.Time) string { return t.UTC().Truncate(time.Second).Format(time.RFC3339) }

func normalizeAuthor(name, email string) string {
	n := strings.TrimSpace(name)
	e := strings.TrimSpace(email)
	if n == "" {
		if e == "" {
			return ""
		}
		return "<" + e + ">"
	}
	if e == "" {
		return n
	}
	return n + " <" + e + ">"
}

type idxEntry struct {
	Hash [20]byte
}

type commitMeta struct {
	Hash   string
	Tree   string
	Parent string
	Author string
	When   time.Time
}

type blobLookup struct {
	hash string
	ok   bool
}

type gitContext struct {
	repoRoot      string
	gitDir        string
	absRoot       string
	idx           map[string]idxEntry
	head          string
	lastCommitFor map[string]*RecGitCommit
	commitMetaFor map[string]*commitMeta
	blobAtPathFor map[string]blobLookup
}
