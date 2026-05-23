package run

import (
	"context"
	"errors"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

func TestRunStagesAndRunStage(t *testing.T) {
	stage.Register("runh-pass", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
		out := in
		out.Records = append(out.Records, stage.Record{Locator: "x"})
		return out, nil
	})
	stage.Register("runh-fail", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
		return stage.Envelope{}, errors.New("boom")
	})

	out, err := runStages(context.Background(), stage.Envelope{Records: []stage.Record{}}, []string{"runh-pass"})
	if err != nil || len(out.Records) != 1 {
		t.Fatalf("runStages pass got err=%v out=%+v", err, out)
	}
	if _, err := runStages(context.Background(), stage.Envelope{Records: []stage.Record{}}, []string{"runh-fail"}); err == nil {
		t.Fatalf("expected runStages failure")
	}
	if _, err := runStage(context.Background(), "runh-pass", stage.Envelope{Records: []stage.Record{}}, stage.Deps{}); err != nil {
		t.Fatalf("runStage err: %v", err)
	}
}
