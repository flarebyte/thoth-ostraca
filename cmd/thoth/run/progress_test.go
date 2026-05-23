package run

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

func TestProgressReporterHelpers(t *testing.T) {
	if rejectedCount(3, 5) != 0 || rejectedCount(5, 3) != 2 {
		t.Fatalf("rejectedCount unexpected")
	}
	if stageTotal(0, 2) != 2 || stageTotal(3, 1) != 3 {
		t.Fatalf("stageTotal unexpected")
	}

	buf := &bytes.Buffer{}
	r := newProgressReporter(&stage.Meta{UI: &stage.UIMeta{Progress: true}}, buf)
	r.ReportProgress(stage.ProgressEvent{Stage: "s", Event: "start", Completed: 0, Total: 1})
	if !strings.Contains(buf.String(), "progress stage=s") {
		t.Fatalf("missing progress line: %q", buf.String())
	}
	if newProgressReporter(nil, buf).enabled {
		t.Fatalf("expected reporter disabled")
	}
}

func TestProgressReporterRunStage(t *testing.T) {
	stage.Register("prog-pass", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
		return stage.Envelope{Records: []stage.Record{{Locator: "a"}}}, nil
	})
	stage.Register("prog-fail", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
		return stage.Envelope{}, errors.New("boom")
	})
	buf := &bytes.Buffer{}
	r := newProgressReporter(&stage.Meta{UI: &stage.UIMeta{Progress: true}}, buf)
	out, err := r.runStage(context.Background(), "prog-pass", stage.Envelope{Records: []stage.Record{}}, stage.Deps{})
	if err != nil || len(out.Records) != 1 {
		t.Fatalf("runStage pass err=%v out=%+v", err, out)
	}
	if _, err := r.runStage(context.Background(), "prog-fail", stage.Envelope{Records: []stage.Record{}}, stage.Deps{}); err == nil {
		t.Fatalf("expected failure")
	}
}
