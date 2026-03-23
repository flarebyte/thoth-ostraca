// File Guide for dev/ai agents:
// Purpose: Define the minimal progress-reporting contract used to emit user-visible stage progress without contaminating JSON output.
// Responsibilities:
// - Define the stable ProgressEvent payload shared by run-time progress emitters.
// - Define the ProgressReporter interface consumed by stages.
// - Attach and retrieve reporters via context.
// Architecture notes:
// - Progress is carried through context so stages can remain decoupled from the CLI implementation and still report user-visible milestones.
// - The event struct is intentionally small and stable because acceptance tests assert these lines directly.
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
