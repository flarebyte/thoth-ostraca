package stage

import "testing"

func TestSortEnvelopeErrors_ByStageLocatorMessage(t *testing.T) {
	env := Envelope{
		Errors: []Error{
			{Stage: "z", Locator: "b", Message: "m2"},
			{Stage: "a", Locator: "z", Message: "m2"},
			{Stage: "a", Locator: "a", Message: "m3"},
			{Stage: "a", Locator: "a", Message: "m1"},
		},
	}
	SortEnvelopeErrors(&env)
	got := env.Errors
	want := []Error{
		{Stage: "a", Locator: "a", Message: "m1"},
		{Stage: "a", Locator: "a", Message: "m3"},
		{Stage: "a", Locator: "z", Message: "m2"},
		{Stage: "z", Locator: "b", Message: "m2"},
	}
	if len(got) != len(want) {
		t.Fatalf("unexpected count: %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d mismatch: got=%+v want=%+v", i, got[i], want[i])
		}
	}
}
