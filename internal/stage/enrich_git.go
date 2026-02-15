package stage

import (
	"bufio"
	"compress/zlib"
	"context"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const enrichGitStage = "enrich-git"

func normalizeRFC3339(t time.Time) string { return t.UTC().Truncate(time.Second).Format(time.RFC3339) }

// readHEAD returns HEAD commit hash (string) or "" if missing.
func readHEAD(repo string) (string, error) {
	headPath := filepath.Join(repo, ".git", "HEAD")
	b, err := os.ReadFile(headPath)
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(string(b))
	if strings.HasPrefix(s, "ref: ") {
		ref := strings.TrimSpace(strings.TrimPrefix(s, "ref: "))
		refPath := filepath.Join(repo, ".git", filepath.FromSlash(ref))
		hb, err := os.ReadFile(refPath)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(hb)), nil
	}
	// detached head contains hash directly
	return s, nil
}

// readLooseObject reads and returns the raw decompressed data of an object hash if present as loose object.
func readLooseObject(repo, hash string) ([]byte, error) {
	if len(hash) < 40 {
		return nil, fmt.Errorf("invalid hash")
	}
	dir := filepath.Join(repo, ".git", "objects", hash[:2])
	file := filepath.Join(dir, hash[2:])
	f, err := os.Open(file)
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

// parseCommitAuthor parses commit content for author line and returns author string and time.
func parseCommitAuthor(commit []byte) (author string, when time.Time, err error) {
	// format: "commit <size>\x00<headers>\n\n..."
	// find header end
	nul := bytesIndex(commit, '\x00')
	if nul < 0 {
		return "", time.Time{}, fmt.Errorf("bad commit object")
	}
	headers := string(commit[nul+1:])
	// find author line
	for _, line := range strings.Split(headers, "\n") {
		if strings.HasPrefix(line, "author ") {
			rest := strings.TrimPrefix(line, "author ")
			// rest ends with " <email> <secs> <tz>"
			parts := strings.Fields(rest)
			if len(parts) >= 3 {
				secs := parts[len(parts)-2]
				tsec, _ := parseInt64(secs)
				when = time.Unix(tsec, 0).UTC()
				// author string as "Name <email>"
				// reconstruct from rest without the last two fields (secs, tz)
				base := strings.Join(parts[:len(parts)-2], " ")
				author = base
				return author, when, nil
			}
		}
	}
	return "", time.Time{}, fmt.Errorf("author not found")
}

func bytesIndex(b []byte, ch byte) int {
	for i, c := range b {
		if c == ch {
			return i
		}
	}
	return -1
}
func parseInt64(s string) (int64, error) { var x int64; _, err := fmt.Sscan(s, &x); return x, err }

// parseIndex returns a map of posix paths -> mtime seconds from .git/index (v2+ simplified).
type idxEntry struct {
	Mtime uint32
	Hash  [20]byte
}

func parseIndex(repo string) (map[string]idxEntry, error) {
	idxPath := filepath.Join(repo, ".git", "index")
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
		return nil, fmt.Errorf("bad index")
	}
	// version := binary.BigEndian.Uint32(hdr[4:8])
	count := binary.BigEndian.Uint32(hdr[8:12])
	entries := make(map[string]idxEntry, count)
	for i := uint32(0); i < count; i++ {
		// fixed 62 bytes before path (v2)
		fixed := make([]byte, 62)
		if _, err := io.ReadFull(r, fixed); err != nil {
			return nil, err
		}
		mtimeSec := binary.BigEndian.Uint32(fixed[8:12])
		var h [20]byte
		copy(h[:], fixed[40:60])
		// flags at bytes 60-61 contain name length (lower 12 bits), but if 0x0fff then longer
		// Read path until NUL
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
		// pad to multiple of 8 from beginning of entry
		consumed := 62 + 1 + len(nameBytes)
		pad := (8 - (consumed % 8)) % 8
		if pad > 0 {
			if _, err := io.CopyN(io.Discard, r, int64(pad)); err != nil {
				return nil, err
			}
		}
		entries[string(nameBytes)] = idxEntry{Mtime: mtimeSec, Hash: h}
	}
	return entries, nil
}

func computeBlobHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	// size
	st, err := f.Stat()
	if err != nil {
		return "", err
	}
	hdr := fmt.Sprintf("blob %d\x00", st.Size())
	h := sha1.New()
	if _, err := io.WriteString(h, hdr); err != nil {
		return "", err
	}
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum), nil
}

func enrichGitRunner(_ context.Context, in Envelope, _ Deps) (Envelope, error) {
	if in.Meta == nil || in.Meta.Git == nil || !in.Meta.Git.Enabled {
		return in, nil
	}
	root := determineRoot(in)
	absRoot, _ := filepath.Abs(root)
	head, err := readHEAD(root)
	if err != nil {
		if in.Meta != nil && in.Meta.Errors != nil && in.Meta.Errors.Mode == "keep-going" {
			out := in
			out.Errors = append(out.Errors, Error{Stage: enrichGitStage, Message: err.Error()})
			return out, nil
		}
		return Envelope{}, fmt.Errorf("%s: %v", enrichGitStage, err)
	}
	authorStr, authorTime := authorFromHead(root, head)
	idx, _ := parseIndex(root)

	out := in
	for i, r := range in.Records {
		if r.Error != nil {
			continue
		}
		rr := r
		rr.Git = recGitFor(r.Locator, root, absRoot, idx, head, authorStr, authorTime)
		out.Records[i] = rr
	}
	return out, nil
}

func init() { Register(enrichGitStage, enrichGitRunner) }
