package run

import (
	"context"
	"errors"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

func registerPipelineStubs() {
	stage.Register("validate-config", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
		out := in
		if out.Meta == nil {
			out.Meta = &stage.Meta{}
		}
		if out.Meta.Config == nil {
			out.Meta.Config = &stage.ConfigMeta{}
		}
		if out.Meta.Config.Action == "" {
			out.Meta.Config.Action = "pipeline"
		}
		return out, nil
	})
	stage.Register("discover-meta-files", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
		out := in
		out.Records = []stage.Record{{Locator: "a", Meta: map[string]any{"k": "v"}}}
		return out, nil
	})
	stage.Register("parse-validate-yaml", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("validate-locators", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("lua-filter", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("lua-map", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("shell-exec", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("lua-postmap", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("lua-reduce", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("write-output", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("discover-input-files", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
		out := in
		out.Records = []stage.Record{{Locator: "f"}}
		return out, nil
	})
	stage.Register("compute-meta-diff", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("write-meta-files", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("load-existing-meta", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("merge-meta", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
	stage.Register("write-updated-meta-files", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) { return in, nil })
}

func TestPipelineHelpers(t *testing.T) {
	if outputLinesEnabled(nil) || reduceInlineEnabled(nil) {
		t.Fatalf("expected disabled by default")
	}
	env := stage.Envelope{Meta: &stage.Meta{Output: &stage.OutputMeta{Lines: true}}}
	if !outputLinesEnabled(env.Meta) {
		t.Fatalf("expected output lines enabled")
	}
	env.Meta.Lua = &stage.LuaMeta{ReduceInline: "x"}
	if !reduceInlineEnabled(env.Meta) {
		t.Fatalf("expected reduce enabled")
	}
	forceBufferedOutput(&env)
	if env.Meta.Output.Lines {
		t.Fatalf("expected lines disabled")
	}
	forceBufferedOutput(nil)
	forceBufferedOutput(&stage.Envelope{})
	forceBufferedOutput(&stage.Envelope{Meta: &stage.Meta{}})
	forceBufferedOutput(&stage.Envelope{Meta: &stage.Meta{Output: nil}})
	if maxRecordsInMemory(nil) != 10000 {
		t.Fatalf("unexpected default max records")
	}
	if err := enforceBufferedRecordLimit(stage.Envelope{Meta: &stage.Meta{Limits: &stage.LimitsMeta{MaxRecordsInMemory: 1}}, Records: []stage.Record{{}, {}}}); err == nil {
		t.Fatalf("expected buffered limit error")
	}
}

func TestExecutePipelineBranches(t *testing.T) {
	registerPipelineStubs()
	out, err := executePipeline(context.Background(), "cfg.cue")
	if err != nil {
		t.Fatalf("executePipeline pipeline err: %v", err)
	}
	if len(out.Records) == 0 {
		t.Fatalf("expected records in pipeline output")
	}

	stage.Register("validate-config", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
		out := in
		out.Meta = &stage.Meta{Config: &stage.ConfigMeta{Action: "input-pipeline"}}
		return out, nil
	})
	if _, err := executePipeline(context.Background(), "cfg.cue"); err != nil {
		t.Fatalf("executePipeline input-pipeline err: %v", err)
	}

	stage.Register("validate-config", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
		out := in
		out.Meta = &stage.Meta{Config: &stage.ConfigMeta{Action: "bad"}}
		return out, nil
	})
	if _, err := executePipeline(context.Background(), "cfg.cue"); err == nil {
		t.Fatalf("expected invalid action error")
	}

	stage.Register("validate-config", func(ctx context.Context, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
		return stage.Envelope{}, errors.New("boom")
	})
	if _, err := executePipeline(context.Background(), "cfg.cue"); err == nil {
		t.Fatalf("expected validate-config error")
	}
}

func TestExecuteMetaPipelineAndStreaming(t *testing.T) {
	registerPipelineStubs()
	in := stage.Envelope{Records: []stage.Record{}, Meta: &stage.Meta{}}
	if _, err := executeMetaPipeline(context.Background(), in); err != nil {
		t.Fatalf("executeMetaPipeline buffered err: %v", err)
	}

	in.Meta.Output = &stage.OutputMeta{Lines: true}
	if _, err := executeMetaPipeline(context.Background(), in); err != nil {
		t.Fatalf("executeMetaPipeline streaming err: %v", err)
	}

	// direct streaming function branch
	if _, err := runStreamingNDJSONPipeline(context.Background(), stage.Envelope{Records: []stage.Record{{Locator: "a"}}, Meta: &stage.Meta{}}); err != nil {
		t.Fatalf("runStreamingNDJSONPipeline err: %v", err)
	}
}
