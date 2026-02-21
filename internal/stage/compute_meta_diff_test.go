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

func TestDiffMetaMapsV3Detailed_ChangesKinds(t *testing.T) {
	existing := map[string]any{
		"same": 1,
		"chg":  1,
		"type": 1,
		"gone": true,
	}
	expected := map[string]any{
		"same": 1,
		"chg":  2,
		"type": "1",
		"new":  "x",
	}
	s := diffMetaMapsV3Detailed(existing, expected)
	got := s.changes
	want := []DiffChange{
		{Path: "chg", Kind: "changed", OldValue: 1, NewValue: 2},
		{Path: "gone", Kind: "removed", OldValue: true},
		{Path: "new", Kind: "added", NewValue: "x"},
		{Path: "type", Kind: "type-changed", OldValue: 1, NewValue: "1"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("changes mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestDiffMetaMapsV3Detailed_ArrayIndexChanged(t *testing.T) {
	existing := map[string]any{"arr": []any{1, 2, 3}}
	expected := map[string]any{"arr": []any{1, 9}}
	s := diffMetaMapsV3Detailed(existing, expected)
	want := []DiffChange{
		{Path: "arr[1]", Kind: "array-index-changed", OldValue: 2, NewValue: 9},
		{Path: "arr[2]", Kind: "removed", OldValue: 3},
	}
	if !reflect.DeepEqual(s.changes, want) {
		t.Fatalf("changes mismatch:\n got: %#v\nwant: %#v", s.changes, want)
	}
}

func TestDiffMetaMapsV3Detailed_ChangesSorted(t *testing.T) {
	existing := map[string]any{"b": 1, "a": 1}
	expected := map[string]any{"b": "1", "a": 2}
	s := diffMetaMapsV3Detailed(existing, expected)
	if len(s.changes) != 2 {
		t.Fatalf("changes len mismatch: %#v", s.changes)
	}
	if s.changes[0].Path != "a" || s.changes[0].Kind != "changed" {
		t.Fatalf("first change not sorted: %#v", s.changes)
	}
	if s.changes[1].Path != "b" || s.changes[1].Kind != "type-changed" {
		t.Fatalf("second change not sorted: %#v", s.changes)
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

func TestComputeMetaDiffRunner_DetailedIncludesChanges(t *testing.T) {
	in := Envelope{
		Records: []Record{
			{Locator: "a.txt", Meta: map[string]any{"arr": []any{1, 2}, "t": 1}},
		},
		Meta: &Meta{
			Inputs:    []string{"a.txt"},
			MetaFiles: []string{"a.txt.thoth.yaml"},
			DiffMeta: &DiffMetaMeta{
				Format:        "detailed",
				ExpectedPatch: map[string]any{"arr": []any{1, 9, 3}, "t": "1"},
			},
		},
	}
	out, err := computeMetaDiffRunner(context.Background(), in, Deps{})
	if err != nil {
		t.Fatalf("compute-meta-diff: %v", err)
	}
	if out.Meta == nil || out.Meta.Diff == nil || len(out.Meta.Diff.Details) != 1 {
		t.Fatalf("missing details")
	}
	changes := out.Meta.Diff.Details[0].Changes
	if len(changes) == 0 {
		t.Fatalf("expected detailed changes")
	}
	if changes[0].Path != "arr[1]" {
		t.Fatalf("unexpected changes order/path: %#v", changes)
	}
}

func TestEscapeJSONPointer(t *testing.T) {
	got := joinJSONPointer("/obj", "a~/b")
	want := "/obj/a~0~1b"
	if got != want {
		t.Fatalf("pointer mismatch: got %q want %q", got, want)
	}
}

func TestDiffMetaJSONPatch_OrderingStable(t *testing.T) {
	existing := map[string]any{
		"b": map[string]any{"x": 1},
		"a": 1,
	}
	expected := map[string]any{
		"b": map[string]any{"x": 2, "y": 3},
		"a": 2,
	}
	got := diffMetaJSONPatch(existing, expected)
	want := []DiffOp{
		{Op: "replace", Path: "/a", Value: 2},
		{Op: "replace", Path: "/b/x", Value: 2},
		{Op: "add", Path: "/b/y", Value: 3},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ops mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestDiffMetaJSONPatch_ArrayReplaceSingleOp(t *testing.T) {
	existing := map[string]any{"arr": []any{1, 2, 3}}
	expected := map[string]any{"arr": []any{1, 9}}
	got := diffMetaJSONPatch(existing, expected)
	want := []DiffOp{
		{Op: "replace", Path: "/arr", Value: []any{1, 9}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ops mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestComputeMetaDiffRunner_JSONPatchIncludesPatch(t *testing.T) {
	in := Envelope{
		Records: []Record{
			{Locator: "a.txt", Meta: map[string]any{"arr": []any{1, 2}, "t": 1}},
		},
		Meta: &Meta{
			Inputs:    []string{"a.txt"},
			MetaFiles: []string{"a.txt.thoth.yaml"},
			DiffMeta: &DiffMetaMeta{
				Format:        "json-patch",
				ExpectedPatch: map[string]any{"arr": []any{1, 9}, "t": "1"},
			},
		},
	}
	out, err := computeMetaDiffRunner(context.Background(), in, Deps{})
	if err != nil {
		t.Fatalf("compute-meta-diff: %v", err)
	}
	if out.Meta == nil || out.Meta.Diff == nil || len(out.Meta.Diff.Details) != 1 {
		t.Fatalf("missing details")
	}
	patch := out.Meta.Diff.Details[0].Patch
	if len(patch) == 0 {
		t.Fatalf("expected json patch ops")
	}
	if patch[0].Path != "/arr" || patch[0].Op != "replace" {
		t.Fatalf("unexpected patch ordering/content: %#v", patch)
	}
}
