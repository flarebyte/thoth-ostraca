package run

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

type progressReporter struct {
	enabled  bool
	interval time.Duration
	w        io.Writer

	mu        sync.Mutex
	stageName string
	processed int
	errors    int
}

func newProgressReporter(meta *stage.Meta, w io.Writer) *progressReporter {
	if meta == nil || meta.UI == nil || !meta.UI.Progress {
		return &progressReporter{enabled: false}
	}
	interval := meta.UI.ProgressIntervalMs
	if interval <= 0 {
		interval = 500
	}
	return &progressReporter{
		enabled:  true,
		interval: time.Duration(interval) * time.Millisecond,
		w:        w,
	}
}

func (p *progressReporter) runStage(ctx context.Context, name string, in stage.Envelope, deps stage.Deps) (stage.Envelope, error) {
	if p == nil || !p.enabled {
		return stage.Run(ctx, name, in, deps)
	}

	p.setSnapshot(name, len(in.Records), len(in.Errors))
	p.emit()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				p.emit()
			case <-done:
				return
			}
		}
	}()

	out, err := stage.Run(ctx, name, in, deps)
	close(done)
	if err == nil {
		p.setSnapshot(name, len(out.Records), len(out.Errors))
		p.emit()
	}
	return out, err
}

func (p *progressReporter) setSnapshot(stageName string, processed int, errs int) {
	p.mu.Lock()
	p.stageName = stageName
	p.processed = processed
	p.errors = errs
	p.mu.Unlock()
}

func (p *progressReporter) emit() {
	if p == nil || !p.enabled || p.w == nil {
		return
	}
	p.mu.Lock()
	stageName := p.stageName
	processed := p.processed
	errs := p.errors
	_, _ = fmt.Fprintf(p.w, "progress stage=%s processed=%d errors=%d\n", stageName, processed, errs)
	p.mu.Unlock()
}

type progressReporterKey struct{}

func withProgressReporter(ctx context.Context, reporter *progressReporter) context.Context {
	return context.WithValue(ctx, progressReporterKey{}, reporter)
}

func progressReporterFromContext(ctx context.Context) *progressReporter {
	if ctx == nil {
		return nil
	}
	r, _ := ctx.Value(progressReporterKey{}).(*progressReporter)
	return r
}
