package metafile

import (
	"bytes"
	"testing"
)

func TestMarshal_RewriteStable(t *testing.T) {
	meta := map[string]any{
		"z": 1,
		"a": map[string]any{
			"d": 4,
			"b": 2,
			"c": map[string]any{
				"y": 2,
				"x": 1,
			},
		},
	}
	b1, err := Marshal("z.txt", meta)
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	b2, err := Marshal("z.txt", meta)
	if err != nil {
		t.Fatalf("marshal second: %v", err)
	}
	if !bytes.Equal(b1, b2) {
		t.Fatalf("not rewrite-stable\nfirst:\n%s\nsecond:\n%s", string(b1), string(b2))
	}
	want := "locator: z.txt\nmeta:\n  a:\n    b: 2\n    c:\n      x: 1\n      " +
		"y: 2\n    d: 4\n  z: 1\n"
	if string(b1) != want {
		t.Fatalf("unexpected canonical output\nwant:\n%s\ngot:\n%s", want, string(b1))
	}
}
