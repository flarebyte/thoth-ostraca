package stage

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

const computeMetaDiffStage = "compute-meta-diff"

func computeMetaDiffRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	if in.Meta == nil {
		return Envelope{}, fmt.Errorf("compute-meta-diff: missing meta")
	}
	inputs := append([]string(nil), in.Meta.Inputs...)
	metas := append([]string(nil), in.Meta.MetaFiles...)
	// Sort for determinism
	sort.Strings(inputs)
	sort.Strings(metas)
	// Build set of input base paths
	inputSet := map[string]struct{}{}
	for _, s := range inputs {
		inputSet[s] = struct{}{}
	}
	var orphans []string
	presentCount := 0
	for _, m := range metas {
		base := strings.TrimSuffix(m, ".thoth.yaml")
		if _, ok := inputSet[base]; ok {
			presentCount++
		} else {
			orphans = append(orphans, m)
		}
	}
	sort.Strings(orphans)
	out := in
	out.Meta.Diff = &DiffReport{Orphans: orphans, PresentCount: presentCount, OrphanCount: len(orphans)}
	return out, nil
}

func init() { Register(computeMetaDiffStage, computeMetaDiffRunner) }
