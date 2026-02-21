package run

import "github.com/flarebyte/thoth-ostraca/internal/stage"

const (
	exitCodeSuccess = 0
	exitCodeExecErr = 1
	exitCodeDrift   = 2
)

type runExitError struct {
	code int
	msg  string
}

func (e runExitError) Error() string { return e.msg }
func (e runExitError) ExitCode() int { return e.code }

func keepGoingMode(meta *stage.Meta) bool {
	return meta != nil && meta.Errors != nil && meta.Errors.Mode == "keep-going"
}

func actionName(meta *stage.Meta) string {
	if meta != nil && meta.Config != nil {
		return meta.Config.Action
	}
	return ""
}

func countRecordResults(records []stage.Record) (successes int, failures int) {
	for _, r := range records {
		if r.Error != nil {
			failures++
		} else {
			successes++
		}
	}
	return
}

func hasActionFailures(env stage.Envelope) bool {
	_, failures := countRecordResults(env.Records)
	return failures > 0 || len(env.Errors) > 0
}

func hasExecutionErrors(env stage.Envelope) bool {
	return hasActionFailures(env)
}

func driftDetectionEnabled(meta *stage.Meta) bool {
	return actionName(meta) == "diff-meta" && meta != nil && meta.DiffMeta != nil && meta.DiffMeta.FailOnChange
}

func hasDiffChanges(env stage.Envelope) bool {
	if env.Meta == nil || env.Meta.Diff == nil {
		return false
	}
	for _, d := range env.Meta.Diff.Details {
		if len(d.AddedKeys) > 0 || len(d.RemovedKeys) > 0 || len(d.ChangedKeys) > 0 || len(d.TypeChangedKeys) > 0 || len(d.Changes) > 0 {
			return true
		}
		for _, a := range d.Arrays {
			if len(a.AddedIndices) > 0 || len(a.RemovedIndices) > 0 || len(a.ChangedIndices) > 0 {
				return true
			}
		}
		for _, op := range d.Patch {
			if op.Op == "replace" {
				if _, ok := op.Value.([]any); ok {
					return true
				}
			}
		}
	}
	return false
}

func hasMeaningfulAggregateOutput(env stage.Envelope) bool {
	if actionName(env.Meta) == "diff-meta" {
		return env.Meta != nil && env.Meta.Diff != nil
	}
	return false
}

func evaluateRunExit(env stage.Envelope) error {
	if driftDetectionEnabled(env.Meta) {
		if hasExecutionErrors(env) {
			return runExitError{code: exitCodeExecErr, msg: "execution errors"}
		}
		if hasDiffChanges(env) {
			return runExitError{code: exitCodeDrift, msg: "drift detected"}
		}
		return nil
	}

	if !keepGoingMode(env.Meta) {
		return nil
	}
	if !hasActionFailures(env) {
		return nil
	}
	successes, _ := countRecordResults(env.Records)
	if successes > 0 {
		return nil
	}
	if hasMeaningfulAggregateOutput(env) {
		return nil
	}
	return runExitError{code: exitCodeExecErr, msg: "keep-going: no successful records"}
}
