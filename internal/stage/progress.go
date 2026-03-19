package stage

import "context"

// ProgressEvent is a stable user-visible progress snapshot.
type ProgressEvent struct {
	Stage     string
	Event     string
	Completed int
	Total     int
	Rejected  int
	Errors    int
}

// ProgressReporter receives user-visible progress events.
type ProgressReporter interface {
	ReportProgress(ProgressEvent)
}

type progressReporterKey struct{}

// WithProgressReporter attaches a progress reporter to the context.
func WithProgressReporter(
	ctx context.Context,
	reporter ProgressReporter,
) context.Context {
	return context.WithValue(ctx, progressReporterKey{}, reporter)
}

// ProgressReporterFromContext returns the attached progress reporter, if any.
func ProgressReporterFromContext(ctx context.Context) ProgressReporter {
	if ctx == nil {
		return nil
	}
	reporter, _ := ctx.Value(progressReporterKey{}).(ProgressReporter)
	return reporter
}
