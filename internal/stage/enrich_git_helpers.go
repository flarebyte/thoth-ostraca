package stage

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"time"
)

// authorFromHead attempts to read commit author/time from a loose HEAD object.
// Returns empty values on any failure to keep behavior lenient.
func authorFromHead(repo, head string) (string, time.Time) {
	if head == "" {
		return "", time.Time{}
	}
	data, err := readLooseObject(repo, head)
	if err != nil {
		return "", time.Time{}
	}
	a, t, err := parseCommitAuthor(data)
	if err != nil {
		return "", time.Time{}
	}
	return a, t
}

// recGitFor computes the RecGit for a locator using index info and file state.
func recGitFor(locator, root, absRoot string, idx map[string]idxEntry, head, authorStr string, authorTime time.Time) *RecGit {
	abs := filepath.Join(root, filepath.FromSlash(locator))
	ignored := matchIgnore(absRoot, locator, false)

	_, err := os.Stat(abs)
	exists := err == nil

	posix := filepath.ToSlash(locator)
	ent, tracked := idx[posix]

	status := "untracked"
	if tracked {
		if !exists {
			status = "deleted"
		} else {
			if h, err := computeBlobHash(abs); err == nil {
				if h != hex.EncodeToString(ent.Hash[:]) {
					status = "modified"
				} else {
					status = "clean"
				}
			} else {
				status = "modified"
			}
		}
	}

	var lc *RecGitCommit
	if tracked && !authorTime.IsZero() {
		lc = &RecGitCommit{Hash: head, Author: authorStr, Time: normalizeRFC3339(authorTime)}
	}
	return &RecGit{Tracked: tracked, Ignored: ignored, Status: status, LastCommit: lc}
}
