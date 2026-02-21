package run

import (
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

func keepGoingMeta(action string) *stage.Meta {
	return &stage.Meta{
		Config: &stage.ConfigMeta{Action: action},
		Errors: &stage.ErrorsMeta{Mode: "keep-going"},
	}
}

func TestEvaluateRunExit_KeepGoing_SuccessRecord(t *testing.T) {
	env := stage.Envelope{
		Meta:    keepGoingMeta("pipeline"),
		Records: []stage.Record{{Locator: "a"}},
		Errors:  []stage.Error{{Stage: "lua-map", Locator: "b", Message: "boom"}},
	}
	if err := evaluateRunExit(env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvaluateRunExit_KeepGoing_AllFailed(t *testing.T) {
	env := stage.Envelope{
		Meta:    keepGoingMeta("pipeline"),
		Records: []stage.Record{{Locator: "a", Error: &stage.RecError{Stage: "x", Message: "m"}}},
		Errors:  []stage.Error{{Stage: "x", Locator: "a", Message: "m"}},
	}
	err := evaluateRunExit(env)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "keep-going: no successful records" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvaluateRunExit_KeepGoing_DiffAggregateOutput(t *testing.T) {
	env := stage.Envelope{
		Meta:   keepGoingMeta("diff-meta"),
		Errors: []stage.Error{{Stage: "parse-validate-yaml", Locator: "x", Message: "invalid YAML"}},
	}
	env.Meta.Diff = &stage.DiffReport{PairedCount: 1}
	if err := evaluateRunExit(env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvaluateRunExit_FailFastMode(t *testing.T) {
	env := stage.Envelope{
		Meta: &stage.Meta{
			Config: &stage.ConfigMeta{Action: "pipeline"},
			Errors: &stage.ErrorsMeta{Mode: "fail-fast"},
		},
		Errors: []stage.Error{{Stage: "x", Message: "m"}},
	}
	if err := evaluateRunExit(env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
