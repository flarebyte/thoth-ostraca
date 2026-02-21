package stage

import (
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

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
