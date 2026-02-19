package run

import (
	"context"
	"fmt"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

// executePipeline runs the fixed Phase 1 pipeline for `thoth run`.
func executePipeline(ctx context.Context, cfgPath string) (stage.Envelope, error) {
	// Always start by validating config to determine action
	in := stage.Envelope{Records: []stage.Record{}, Meta: &stage.Meta{ConfigPath: cfgPath}}
	out, err := stage.Run(ctx, "validate-config", in, stage.Deps{})
	if err != nil {
		return stage.Envelope{}, err
	}
	action := "pipeline"
	if out.Meta != nil && out.Meta.Config != nil && out.Meta.Config.Action != "" {
		action = out.Meta.Config.Action
	}
	switch action {
	case "pipeline", "nop":
		return executeMetaPipeline(ctx, out)
	case "validate":
		stages := []string{
			"discover-meta-files",
			"parse-validate-yaml",
			"validate-locators",
			"write-output",
		}
		return runStages(ctx, out, stages)
	case "create-meta":
		stages := []string{
			"discover-input-files",
			"enrich-fileinfo",
			"enrich-git",
			"write-meta-files",
			"write-output",
		}
		return runStages(ctx, out, stages)
	case "update-meta":
		stages := []string{
			"discover-input-files",
			"enrich-fileinfo",
			"enrich-git",
			"load-existing-meta",
			"merge-meta",
			"write-updated-meta-files",
			"write-output",
		}
		return runStages(ctx, out, stages)
	case "diff-meta":
		stages := []string{
			"discover-input-files",
			"discover-meta-files",
			"parse-validate-yaml",
			"validate-locators",
			"compute-meta-diff",
			"write-output",
		}
		return runStages(ctx, out, stages)
	default:
		// Should not happen; validate-config already enforced
		return stage.Envelope{}, fmt.Errorf("invalid action")
	}
}

// output is handled by the write-output stage.

func executeMetaPipeline(ctx context.Context, in stage.Envelope) (stage.Envelope, error) {
	preStages := []string{
		"discover-meta-files",
		"parse-validate-yaml",
		"validate-locators",
	}
	out, err := runStages(ctx, in, preStages)
	if err != nil {
		return stage.Envelope{}, err
	}

	streamingRequested := outputLinesEnabled(out.Meta)
	reduceEnabled := reduceInlineEnabled(out.Meta)
	streamingAllowed := streamingRequested && !reduceEnabled

	if !streamingAllowed {
		if err := enforceBufferedRecordLimit(out); err != nil {
			return stage.Envelope{}, err
		}
		if streamingRequested && reduceEnabled {
			forceBufferedOutput(&out)
		}
		stages := []string{
			"lua-filter",
			"lua-map",
			"shell-exec",
			"lua-postmap",
			"lua-reduce",
			"write-output",
		}
		return runStages(ctx, out, stages)
	}

	return runStreamingNDJSONPipeline(ctx, out)
}

func runStreamingNDJSONPipeline(ctx context.Context, in stage.Envelope) (stage.Envelope, error) {
	streamIn := in
	streamIn.Records = []stage.Record{}
	stream := make(chan stage.Record, len(in.Records))

	type writeResult struct {
		out stage.Envelope
		err error
	}
	writeDone := make(chan writeResult, 1)
	go func() {
		out, err := stage.Run(ctx, "write-output", streamIn, stage.Deps{RecordStream: stream})
		writeDone <- writeResult{out: out, err: err}
	}()

	perRecordStages := []string{"lua-filter", "lua-map", "shell-exec", "lua-postmap"}
	cur := in
	for _, rec := range in.Records {
		recEnv := stage.Envelope{
			Records: []stage.Record{rec},
			Meta:    cur.Meta,
			Errors:  append([]stage.Error(nil), cur.Errors...),
		}
		var err error
		for _, name := range perRecordStages {
			recEnv, err = stage.Run(ctx, name, recEnv, stage.Deps{})
			if err != nil {
				close(stream)
				<-writeDone
				return stage.Envelope{}, err
			}
		}
		cur.Errors = recEnv.Errors
		if len(recEnv.Records) == 1 {
			stream <- recEnv.Records[0]
		}
	}
	close(stream)
	wr := <-writeDone
	if wr.err != nil {
		return stage.Envelope{}, wr.err
	}
	cur.Records = []stage.Record{}
	return cur, nil
}

func outputLinesEnabled(meta *stage.Meta) bool {
	return meta != nil && meta.Output != nil && meta.Output.Lines
}

func reduceInlineEnabled(meta *stage.Meta) bool {
	return meta != nil && meta.Lua != nil && meta.Lua.ReduceInline != ""
}

func forceBufferedOutput(env *stage.Envelope) {
	if env == nil || env.Meta == nil || env.Meta.Output == nil {
		return
	}
	env.Meta.Output.Lines = false
}

func maxRecordsInMemory(meta *stage.Meta) int {
	if meta != nil && meta.Limits != nil && meta.Limits.MaxRecordsInMemory > 0 {
		return meta.Limits.MaxRecordsInMemory
	}
	return 10000
}

func enforceBufferedRecordLimit(env stage.Envelope) error {
	limit := maxRecordsInMemory(env.Meta)
	if len(env.Records) <= limit {
		return nil
	}
	return fmt.Errorf("buffered mode exceeds maxRecordsInMemory=%d; set output.lines=true", limit)
}
