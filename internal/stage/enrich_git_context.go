package stage

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func newGitContext(root, repoRoot string) (*gitContext, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, errGitRepoOpenFailed
	}
	gitDir, err := gitDirForRepo(repoRoot)
	if err != nil {
		return nil, errGitRepoOpenFailed
	}
	idx, err := parseIndex(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			idx = map[string]idxEntry{}
		} else {
			return nil, errGitStatusFailed
		}
	}
	head, err := readHEAD(gitDir)
	if err != nil {
		return nil, errGitStatusFailed
	}
	return &gitContext{
		repoRoot:      repoRoot,
		gitDir:        gitDir,
		absRoot:       absRoot,
		idx:           idx,
		head:          head,
		lastCommitFor: map[string]*RecGitCommit{},
		commitMetaFor: map[string]*commitMeta{},
		blobAtPathFor: map[string]blobLookup{},
	}, nil
}

func (g *gitContext) recGitFor(locator string) (*RecGit, error) {
	absPath := filepath.Join(g.absRoot, filepath.FromSlash(locator))
	relRepo, err := filepath.Rel(g.repoRoot, absPath)
	if err != nil {
		return nil, errGitStatusFailed
	}
	if relRepo == "." || relRepo == ".." || strings.HasPrefix(relRepo, ".."+string(filepath.Separator)) {
		return nil, errGitStatusFailed
	}
	repoPath := filepath.ToSlash(filepath.Clean(relRepo))
	_, tracked := g.idx[repoPath]
	status, err := g.statusForPath(repoPath, absPath, tracked)
	if err != nil {
		return nil, err
	}
	ignored := matchIgnore(g.repoRoot, filepath.FromSlash(repoPath), false)

	var lc *RecGitCommit
	if tracked {
		lc, err = g.commitForPath(repoPath)
		if err != nil {
			return nil, err
		}
	}
	return &RecGit{Tracked: tracked, Ignored: ignored, Status: status, LastCommit: lc}, nil
}

func (g *gitContext) statusForPath(repoPath, absPath string, tracked bool) (string, error) {
	if !tracked {
		return "untracked", nil
	}
	ent := g.idx[repoPath]
	b, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "deleted", nil
		}
		return "", errGitStatusFailed
	}
	h := blobHashBytes(b)
	if !hashEq(ent.Hash, h) {
		return "modified", nil
	}
	return "clean", nil
}

func hashEq(a, b [20]byte) bool {
	for i := 0; i < 20; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func blobHashBytes(content []byte) [20]byte {
	prefix := fmt.Sprintf("blob %d\x00", len(content))
	// nosemgrep: go.lang.security.audit.crypto.use_of_weak_crypto.use-of-sha1
	// Git object IDs are defined with SHA-1 for compatibility with repository data.
	h := sha1.New()
	_, _ = h.Write([]byte(prefix))
	_, _ = h.Write(content)
	var out [20]byte
	copy(out[:], h.Sum(nil))
	return out
}
