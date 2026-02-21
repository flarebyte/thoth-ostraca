package run

import (
	"fmt"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

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

func hasMeaningfulAggregateOutput(env stage.Envelope) bool {
	if actionName(env.Meta) == "diff-meta" {
		return env.Meta != nil && env.Meta.Diff != nil
	}
	return false
}

func evaluateRunExit(env stage.Envelope) error {
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
	return fmt.Errorf("keep-going: no successful records")
}
