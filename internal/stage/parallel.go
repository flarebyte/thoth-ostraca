package stage

import (
	"runtime"
	"sort"
)

// getWorkers returns the configured worker count or a sane default.
func getWorkers(meta *Meta) int {
	n := runtime.NumCPU()
	if meta != nil && meta.Workers > 0 {
		n = meta.Workers
	}
	if n < 1 {
		n = 1
	}
	return n
}

// SortEnvelopeErrors sorts errors by (stage, locator, message) deterministically.
func SortEnvelopeErrors(env *Envelope) {
	if env == nil || len(env.Errors) == 0 {
		return
	}
	sort.Slice(env.Errors, func(i, j int) bool {
		ei, ej := env.Errors[i], env.Errors[j]
		if ei.Stage != ej.Stage {
			return ei.Stage < ej.Stage
		}
		if ei.Locator != ej.Locator {
			return ei.Locator < ej.Locator
		}
		return ei.Message < ej.Message
	})
}
