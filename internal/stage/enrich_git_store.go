package stage

import (
	"bufio"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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
