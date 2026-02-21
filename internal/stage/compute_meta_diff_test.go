package stage

import (
	"context"
	"reflect"
	"testing"
)

func TestDiffMetaMaps_Rules(t *testing.T) {
	existing := map[string]any{
		"a":   1,
		"obj": map[string]any{"x": 1, "y": 2},
		"arr": []any{1, 2},
	}
	expected := map[string]any{
		"a":   1,
		"obj": map[string]any{"y": 9, "z": 3},
		"arr": []any{1, 2, 3},
	}
	added, removed, changed := diffMetaMaps(existing, expected)
	if !reflect.DeepEqual(added, []string{"obj.z"}) {
		t.Fatalf("added mismatch: %+v", added)
	}
	if !reflect.DeepEqual(removed, []string{"obj.x"}) {
		t.Fatalf("removed mismatch: %+v", removed)
	}
	if !reflect.DeepEqual(changed, []string{"obj.y"}) {
		t.Fatalf("changed mismatch: %+v", changed)
	}
}

func TestDiffMetaMapsV3_TypeChange(t *testing.T) {
	existing := map[string]any{"a": 1}
	expected := map[string]any{"a": "1"}
	s := diffMetaMapsV3(existing, expected)
	if !reflect.DeepEqual(s.typeChanged, []string{"a"}) {
		t.Fatalf("typeChanged mismatch: %+v", s.typeChanged)
	}
}

func TestDiffMetaMapsV3_ArrayIndexDiff(t *testing.T) {
	existing := map[string]any{"arr": []any{1, 2, 3}}
	expected := map[string]any{"arr": []any{1, 9}}
	s := diffMetaMapsV3(existing, expected)
	if len(s.arrays) != 1 || s.arrays[0].Path != "arr" {
		t.Fatalf("arrays mismatch: %+v", s.arrays)
	}
	if !reflect.DeepEqual(s.arrays[0].ChangedIndices, []int{1}) {
		t.Fatalf("changedIndices mismatch: %+v", s.arrays[0].ChangedIndices)
	}
	if !reflect.DeepEqual(s.arrays[0].RemovedIndices, []int{2}) {
		t.Fatalf("removedIndices mismatch: %+v", s.arrays[0].RemovedIndices)
	}
	if len(s.arrays[0].AddedIndices) != 0 {
		t.Fatalf("addedIndices expected empty: %+v", s.arrays[0].AddedIndices)
	}
}

func TestDiffMetaMapsV3_NestedArrayPath(t *testing.T) {
	existing := map[string]any{
		"obj": map[string]any{
			"arr": []any{
				map[string]any{"x": 1},
				map[string]any{"x": 2},
			},
		},
	}
	expected := map[string]any{
		"obj": map[string]any{
			"arr": []any{
				map[string]any{"x": 1},
				map[string]any{"x": 3},
				map[string]any{"x": 4},
			},
		},
	}
	s := diffMetaMapsV3(existing, expected)
	if len(s.arrays) != 1 || s.arrays[0].Path != "obj.arr" {
		t.Fatalf("arrays mismatch: %+v", s.arrays)
	}
	if !reflect.DeepEqual(s.arrays[0].ChangedIndices, []int{1}) {
		t.Fatalf("changedIndices mismatch: %+v", s.arrays[0].ChangedIndices)
	}
	if !reflect.DeepEqual(s.arrays[0].AddedIndices, []int{2}) {
		t.Fatalf("addedIndices mismatch: %+v", s.arrays[0].AddedIndices)
	}
}

func TestComputeMetaDiffRunner_V2Report(t *testing.T) {
	in := Envelope{
		Records: []Record{
			{Locator: "a.txt", Meta: map[string]any{"obj": map[string]any{"x": 1, "y": 2}, "arr": []any{1, 2}, "a": 1}},
			{Locator: "b.txt", Meta: map[string]any{"obj": map[string]any{"y": 9, "z": 3}, "arr": []any{1, 2, 3}, "a": 1}},
		},
		Meta: &Meta{
			Inputs:    []string{"a.txt", "b.txt"},
			MetaFiles: []string{"a.txt.thoth.yaml", "b.txt.thoth.yaml", "orphan.thoth.yaml"},
			DiffMeta: &DiffMetaMeta{
				ExpectedPatch: map[string]any{
					"a":   1,
					"obj": map[string]any{"y": 9, "z": 3},
					"arr": []any{1, 2, 3},
				},
			},
		},
	}
	out, err := computeMetaDiffRunner(context.Background(), in, Deps{})
	if err != nil {
		t.Fatalf("compute-meta-diff: %v", err)
	}
	if out.Meta == nil || out.Meta.Diff == nil {
		t.Fatalf("missing meta.diff")
	}
	if out.Meta.Diff.PairedCount != 2 || out.Meta.Diff.OrphanCount != 1 || out.Meta.Diff.ChangedCount != 1 {
		t.Fatalf("unexpected counts: %+v", out.Meta.Diff)
	}
	if !reflect.DeepEqual(out.Meta.Diff.OrphanMetaFiles, []string{"orphan.thoth.yaml"}) {
		t.Fatalf("unexpected orphans: %+v", out.Meta.Diff.OrphanMetaFiles)
	}
	if len(out.Meta.Diff.Details) != 2 {
		t.Fatalf("unexpected details: %+v", out.Meta.Diff.Details)
	}
	if out.Meta.Diff.Details[0].Locator != "a.txt" || out.Meta.Diff.Details[1].Locator != "b.txt" {
		t.Fatalf("details not sorted: %+v", out.Meta.Diff.Details)
	}
	d0 := out.Meta.Diff.Details[0]
	if len(d0.Arrays) != 1 || d0.Arrays[0].Path != "arr" || !reflect.DeepEqual(d0.Arrays[0].AddedIndices, []int{2}) {
		t.Fatalf("unexpected array diff: %+v", d0.Arrays)
	}
}
