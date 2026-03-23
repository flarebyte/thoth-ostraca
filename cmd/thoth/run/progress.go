// File Guide for dev/ai agents:
// Purpose: Emit user-visible progress lines for `thoth run` without altering the JSON output contract.
// Responsibilities:
// - Decide whether progress reporting is enabled from UI metadata.
// - Wrap stage execution with start, done, and failed progress events.
// - Render stable progress lines to stderr and compute rejected counts.
// Architecture notes:
// - Progress formatting lives in the CLI layer rather than the stage package so runtime stages only depend on the abstract reporter interface.
// - The emitted line format is intentionally stable and simple because acceptance tests assert it directly.
package run

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

type progressReporter struct {
	enabled bool
	w       io.Writer
	mu      sync.Mutex
}

func newProgressReporter(meta *stage.Meta, w io.Writer) *progressReporter {
	if meta == nil || meta.UI == nil || !meta.UI.Progress {
		return &progressReporter{enabled: false}
	}
	return &progressReporter{
		enabled: true,
		w:       w,
	}
}

func (p *progressReporter) runStage(
	ctx context.Context,
	name string,
	in stage.Envelope,
	deps stage.Deps,
) (stage.Envelope, error) {
	if p == nil || !p.enabled {
		return stage.Run(ctx, name, in, deps)
	}
	ctx = stage.WithProgressReporter(ctx, p)
	p.ReportProgress(stage.ProgressEvent{
		Stage:     name,
		Event:     "start",
		Completed: 0,
		Total:     len(in.Records),
		Rejected:  0,
		Errors:    len(in.Errors),
	})

	out, err := stage.Run(ctx, name, in, deps)
	if err != nil {
		p.ReportProgress(stage.ProgressEvent{
			Stage:     name,
			Event:     "failed",
			Completed: 0,
			Total:     len(in.Records),
			Rejected:  0,
			Errors:    len(in.Errors) + 1,
		})
		return stage.Envelope{}, err
	}
	p.ReportProgress(stage.ProgressEvent{
		Stage:     name,
		Event:     "done",
		Completed: len(out.Records),
		Total:     stageTotal(len(in.Records), len(out.Records)),
		Rejected:  rejectedCount(stageTotal(len(in.Records), len(out.Records)), len(out.Records)),
		Errors:    len(out.Errors),
	})
	return out, nil
}

func (p *progressReporter) ReportProgress(ev stage.ProgressEvent) {
	if p == nil || !p.enabled || p.w == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	_, _ = fmt.Fprintf(
		p.w,
		"progress stage=%s event=%s completed=%d total=%d rejected=%d errors=%d\n",
		ev.Stage,
		ev.Event,
		ev.Completed,
		ev.Total,
		ev.Rejected,
		ev.Errors,
	)
}

func rejectedCount(total, completed int) int {
	if total <= completed {
		return 0
	}
	return total - completed
}

func stageTotal(inCount, outCount int) int {
	if inCount == 0 && outCount > 0 {
		return outCount
	}
	return inCount
}
