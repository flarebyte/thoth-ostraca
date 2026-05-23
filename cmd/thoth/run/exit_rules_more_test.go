package run

import (
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

func TestActionNameAndHasDiffChanges(t *testing.T) {
	if got := actionName(nil); got != "" {
		t.Fatalf("actionName(nil)=%q", got)
	}
	if got := actionName(&stage.Meta{}); got != "" {
		t.Fatalf("actionName(empty)=%q", got)
	}
	if got := actionName(&stage.Meta{Config: &stage.ConfigMeta{Action: "diff-meta"}}); got != "diff-meta" {
		t.Fatalf("actionName=%q", got)
	}

	if hasDiffChanges(stage.Envelope{}) {
		t.Fatalf("expected false")
	}
	if hasDiffChanges(stage.Envelope{Meta: &stage.Meta{}}) {
		t.Fatalf("expected false")
	}
	if hasDiffChanges(stage.Envelope{Meta: &stage.Meta{Diff: &stage.DiffReport{Details: nil}}}) {
		t.Fatalf("expected false")
	}
	if !hasDiffChanges(stage.Envelope{Meta: &stage.Meta{Diff: &stage.DiffReport{Details: []stage.DiffDetail{{AddedKeys: []string{"a"}}}}}}) {
		t.Fatalf("expected true on added keys")
	}
	if !hasDiffChanges(stage.Envelope{Meta: &stage.Meta{Diff: &stage.DiffReport{Details: []stage.DiffDetail{{Arrays: []stage.ArrayDiff{{AddedIndices: []int{1}}}}}}}}) {
		t.Fatalf("expected true on array index changes")
	}
	if !hasDiffChanges(stage.Envelope{Meta: &stage.Meta{Diff: &stage.DiffReport{Details: []stage.DiffDetail{{Patch: []stage.DiffOp{{Op: "replace", Value: []any{1}}}}}}}}) {
		t.Fatalf("expected true on replace array patch")
	}
	if hasDiffChanges(stage.Envelope{Meta: &stage.Meta{Diff: &stage.DiffReport{Details: []stage.DiffDetail{{Patch: []stage.DiffOp{{Op: "replace", Value: "x"}}}}}}}) {
		t.Fatalf("expected false on non-array replace patch")
	}
}
