package stage

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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
	h := sha1.New()
	_, _ = h.Write([]byte(prefix))
	_, _ = h.Write(content)
	var out [20]byte
	copy(out[:], h.Sum(nil))
	return out
}

func (g *gitContext) commitForPath(path string) (*RecGitCommit, error) {
	if c, ok := g.lastCommitFor[path]; ok {
		return c, nil
	}
	if g.head == "" {
		g.lastCommitFor[path] = nil
		return nil, nil
	}
	cur := g.head
	for cur != "" {
		meta, err := g.commitMeta(cur)
		if err != nil {
			return nil, errGitCommitLookupFailed
		}
		touched, err := g.commitTouchesPath(meta, path)
		if err != nil {
			return nil, errGitCommitLookupFailed
		}
		if touched {
			rec := &RecGitCommit{Hash: meta.Hash, Author: meta.Author, Time: normalizeRFC3339(meta.When)}
			g.lastCommitFor[path] = rec
			return rec, nil
		}
		cur = meta.Parent
	}
	g.lastCommitFor[path] = nil
	return nil, nil
}

func (g *gitContext) commitTouchesPath(meta *commitMeta, path string) (bool, error) {
	curHash, curOK, err := g.blobHashAtPath(meta.Tree, path)
	if err != nil {
		return false, err
	}
	if meta.Parent == "" {
		return curOK, nil
	}
	parentMeta, err := g.commitMeta(meta.Parent)
	if err != nil {
		return false, err
	}
	parentHash, parentOK, err := g.blobHashAtPath(parentMeta.Tree, path)
	if err != nil {
		return false, err
	}
	if curOK != parentOK {
		return true, nil
	}
	if !curOK && !parentOK {
		return false, nil
	}
	return curHash != parentHash, nil
}

func (g *gitContext) blobHashAtPath(treeHash, path string) (string, bool, error) {
	key := treeHash + "|" + path
	if got, ok := g.blobAtPathFor[key]; ok {
		return got.hash, got.ok, nil
	}
	parts := strings.Split(path, "/")
	h, ok, err := lookupBlobInTree(g.gitDir, treeHash, parts)
	if err != nil {
		return "", false, err
	}
	g.blobAtPathFor[key] = blobLookup{hash: h, ok: ok}
	return h, ok, nil
}

func (g *gitContext) commitMeta(hash string) (*commitMeta, error) {
	if got, ok := g.commitMetaFor[hash]; ok {
		return got, nil
	}
	obj, err := readLooseObject(g.gitDir, hash)
	if err != nil {
		return nil, err
	}
	payload, err := objectPayload(obj, "commit")
	if err != nil {
		return nil, err
	}
	meta, err := parseCommitMeta(hash, payload)
	if err != nil {
		return nil, err
	}
	g.commitMetaFor[hash] = meta
	return meta, nil
}

func parseCommitMeta(hash string, payload []byte) (*commitMeta, error) {
	meta := &commitMeta{Hash: hash}
	for _, line := range strings.Split(string(payload), "\n") {
		if line == "" {
			break
		}
		switch {
		case strings.HasPrefix(line, "tree "):
			meta.Tree = strings.TrimSpace(strings.TrimPrefix(line, "tree "))
		case strings.HasPrefix(line, "parent "):
			if meta.Parent == "" {
				meta.Parent = strings.TrimSpace(strings.TrimPrefix(line, "parent "))
			}
		case strings.HasPrefix(line, "author "):
			author, when := parseAuthorLine(strings.TrimPrefix(line, "author "))
			meta.Author = author
			meta.When = when
		}
	}
	if meta.Tree == "" {
		return nil, errors.New("missing tree")
	}
	return meta, nil
}

func parseAuthorLine(s string) (string, time.Time) {
	line := strings.TrimSpace(s)
	lt := strings.LastIndex(line, "<")
	gt := strings.LastIndex(line, ">")
	if lt >= 0 && gt > lt {
		name := strings.TrimSpace(line[:lt])
		email := strings.TrimSpace(line[lt+1 : gt])
		rest := strings.TrimSpace(line[gt+1:])
		parts := strings.Fields(rest)
		if len(parts) >= 1 {
			if sec, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				return normalizeAuthor(name, email), time.Unix(sec, 0).UTC()
			}
		}
		return normalizeAuthor(name, email), time.Time{}
	}
	return strings.TrimSpace(line), time.Time{}
}

func lookupBlobInTree(gitDir, treeHash string, parts []string) (string, bool, error) {
	if len(parts) == 0 {
		return "", false, nil
	}
	obj, err := readLooseObject(gitDir, treeHash)
	if err != nil {
		return "", false, err
	}
	payload, err := objectPayload(obj, "tree")
	if err != nil {
		return "", false, err
	}
	i := 0
	for i < len(payload) {
		modeStart := i
		for i < len(payload) && payload[i] != ' ' {
			i++
		}
		if i >= len(payload) {
			return "", false, errors.New("bad tree mode")
		}
		mode := string(payload[modeStart:i])
		i++
		nameStart := i
		for i < len(payload) && payload[i] != 0 {
			i++
		}
		if i >= len(payload) {
			return "", false, errors.New("bad tree name")
		}
		name := string(payload[nameStart:i])
		i++
		if i+20 > len(payload) {
			return "", false, errors.New("bad tree hash")
		}
		h := hex.EncodeToString(payload[i : i+20])
		i += 20
		if name != parts[0] {
			continue
		}
		if len(parts) == 1 {
			if mode == "40000" || mode == "040000" {
				return "", false, nil
			}
			return h, true, nil
		}
		if mode != "40000" && mode != "040000" {
			return "", false, nil
		}
		return lookupBlobInTree(gitDir, h, parts[1:])
	}
	return "", false, nil
}

func objectPayload(obj []byte, typ string) ([]byte, error) {
	n := bytesIndex(obj, 0)
	if n < 0 {
		return nil, errors.New("invalid object")
	}
	head := string(obj[:n])
	if !strings.HasPrefix(head, typ+" ") {
		return nil, errors.New("object type mismatch")
	}
	return obj[n+1:], nil
}

func bytesIndex(b []byte, ch byte) int {
	for i, c := range b {
		if c == ch {
			return i
		}
	}
	return -1
}

func gitDirForRepo(repoRoot string) (string, error) {
	p := filepath.Join(repoRoot, ".git")
	st, err := os.Stat(p)
	if err != nil {
		return "", err
	}
	if st.IsDir() {
		return p, nil
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(string(b))
	if !strings.HasPrefix(s, "gitdir:") {
		return "", errors.New("invalid .git file")
	}
	d := strings.TrimSpace(strings.TrimPrefix(s, "gitdir:"))
	if !filepath.IsAbs(d) {
		d = filepath.Clean(filepath.Join(repoRoot, d))
	}
	return d, nil
}

func readHEAD(gitDir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(string(b))
	if strings.HasPrefix(s, "ref: ") {
		ref := strings.TrimSpace(strings.TrimPrefix(s, "ref: "))
		return resolveRef(gitDir, ref)
	}
	return s, nil
}

func resolveRef(gitDir, ref string) (string, error) {
	refPath := filepath.Join(gitDir, filepath.FromSlash(ref))
	if b, err := os.ReadFile(refPath); err == nil {
		return strings.TrimSpace(string(b)), nil
	}
	packed := filepath.Join(gitDir, "packed-refs")
	f, err := os.Open(packed)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer func() { _ = f.Close() }()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "^") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == ref {
			return parts[0], nil
		}
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return "", nil
}

func parseIndex(gitDir string) (map[string]idxEntry, error) {
	idxPath := filepath.Join(gitDir, "index")
	f, err := os.Open(idxPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReader(f)
	hdr := make([]byte, 12)
	if _, err := io.ReadFull(r, hdr); err != nil {
		return nil, err
	}
	if string(hdr[:4]) != "DIRC" {
		return nil, errors.New("bad index")
	}
	count := binary.BigEndian.Uint32(hdr[8:12])
	entries := make(map[string]idxEntry, count)
	for i := uint32(0); i < count; i++ {
		fixed := make([]byte, 62)
		if _, err := io.ReadFull(r, fixed); err != nil {
			return nil, err
		}
		var h [20]byte
		copy(h[:], fixed[40:60])
		nameBytes := []byte{}
		for {
			b, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			if b == 0 {
				break
			}
			nameBytes = append(nameBytes, b)
		}
		consumed := 62 + 1 + len(nameBytes)
		pad := (8 - (consumed % 8)) % 8
		if pad > 0 {
			if _, err := io.CopyN(io.Discard, r, int64(pad)); err != nil {
				return nil, err
			}
		}
		entries[string(nameBytes)] = idxEntry{Hash: h}
	}
	return entries, nil
}

func readLooseObject(gitDir, hash string) ([]byte, error) {
	if len(hash) < 40 {
		return nil, errors.New("invalid hash")
	}
	f, err := os.Open(filepath.Join(gitDir, "objects", hash[:2], hash[2:]))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	zr, err := zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()
	return io.ReadAll(zr)
}
