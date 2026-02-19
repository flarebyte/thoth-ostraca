package stage

import (
	"context"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/metafile"
)

func TestDeepMerge_Rules(t *testing.T) {
	existing := map[string]any{
		"a":   1,
		"obj": map[string]any{"x": 1, "y": 2},
		"arr": []any{1, 2},
	}
	patch := map[string]any{
		"b":   2,
		"obj": map[string]any{"y": 9, "z": 3},
		"arr": []any{7},
	}
	got := deepMerge(existing, patch)
	obj, _ := got["obj"].(map[string]any)
	arr, _ := got["arr"].([]any)
	if got["a"] != 1 || got["b"] != 2 {
		t.Fatalf("unexpected top-level merge: %+v", got)
	}
	if obj["x"] != 1 || obj["y"] != 9 || obj["z"] != 3 {
		t.Fatalf("unexpected object merge: %+v", obj)
	}
	if len(arr) != 1 || arr[0] != 7 {
		t.Fatalf("unexpected array merge: %+v", arr)
	}

	b, err := metafile.Marshal("a.txt", got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := "locator: a.txt\nmeta:\n  a: 1\n  arr:\n    - 7\n  b: 2\n  obj:\n    x: 1\n    y: 9\n    z: 3\n"
	if string(b) != want {
		t.Fatalf("unexpected canonical output\nwant:\n%s\ngot:\n%s", want, string(b))
	}
}

func TestMergeMetaRunner_UsesPatchWhenNoExisting(t *testing.T) {
	in := Envelope{
		Records: []Record{{Locator: "x"}},
		Meta:    &Meta{UpdateMeta: &UpdateMetaMeta{Patch: map[string]any{"k": "v"}}},
	}
	out, err := mergeMetaRunner(context.Background(), in, Deps{})
	if err != nil {
		t.Fatalf("merge-meta: %v", err)
	}
	pm, _ := out.Records[0].Post.(map[string]any)
	next, _ := pm["nextMeta"].(map[string]any)
	if next["k"] != "v" {
		t.Fatalf("expected patch-only nextMeta, got: %+v", next)
	}
}
