package stage

import (
	"context"
	"fmt"
	"reflect"
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
	sort.Strings(inputs)
	sort.Strings(metas)

	inputSet := map[string]struct{}{}
	for _, s := range inputs {
		inputSet[s] = struct{}{}
	}

	metaSet := map[string]struct{}{}
	for _, m := range metas {
		metaSet[m] = struct{}{}
	}

	existingByLocator := map[string]map[string]any{}
	for _, r := range in.Records {
		if r.Error != nil || r.Meta == nil {
			continue
		}
		if cp, ok := deepCopyAny(r.Meta).(map[string]any); ok {
			existingByLocator[r.Locator] = cp
		}
	}

	expected := map[string]any{}
	if in.Meta.DiffMeta != nil && in.Meta.DiffMeta.ExpectedPatch != nil {
		if cp, ok := deepCopyAny(in.Meta.DiffMeta.ExpectedPatch).(map[string]any); ok {
			expected = cp
		}
	}

	var orphans []string
	var details []DiffDetail
	sort.Strings(orphans)

	for _, m := range metas {
		base := strings.TrimSuffix(m, ".thoth.yaml")
		if _, ok := inputSet[base]; !ok {
			orphans = append(orphans, m)
		}
	}
	sort.Strings(orphans)

	for _, loc := range inputs {
		metaFile := loc + ".thoth.yaml"
		if _, ok := metaSet[metaFile]; !ok {
			continue
		}
		existing, ok := existingByLocator[loc]
		if !ok {
			continue
		}
		added, removed, changed := diffMetaMaps(existing, expected)
		details = append(details, DiffDetail{
			Locator:     loc,
			MetaFile:    metaFile,
			AddedKeys:   added,
			RemovedKeys: removed,
			ChangedKeys: changed,
		})
	}
	sort.Slice(details, func(i, j int) bool { return details[i].Locator < details[j].Locator })

	changedCount := 0
	for _, d := range details {
		if len(d.AddedKeys) > 0 || len(d.RemovedKeys) > 0 || len(d.ChangedKeys) > 0 {
			changedCount++
		}
	}

	out := in
	out.Meta.Diff = &DiffReport{
		OrphanMetaFiles: orphans,
		PairedCount:     len(details),
		OrphanCount:     len(orphans),
		ChangedCount:    changedCount,
		Details:         details,
		Orphans:         orphans,
		PresentCount:    len(details),
	}
	return out, nil
}

func init() { Register(computeMetaDiffStage, computeMetaDiffRunner) }

func diffMetaMaps(existing, expected map[string]any) ([]string, []string, []string) {
	added := make([]string, 0)
	removed := make([]string, 0)
	changed := make([]string, 0)
	diffMetaAtPath(existing, expected, "", &added, &removed, &changed)
	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)
	return added, removed, changed
}

func diffMetaAtPath(existing, expected map[string]any, prefix string, added, removed, changed *[]string) {
	keys := map[string]struct{}{}
	for k := range existing {
		keys[k] = struct{}{}
	}
	for k := range expected {
		keys[k] = struct{}{}
	}
	all := make([]string, 0, len(keys))
	for k := range keys {
		all = append(all, k)
	}
	sort.Strings(all)

	for _, k := range all {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		ev, inExisting := existing[k]
		pv, inExpected := expected[k]
		if !inExisting && inExpected {
			*added = append(*added, path)
			continue
		}
		if inExisting && !inExpected {
			*removed = append(*removed, path)
			continue
		}
		em, eok := asStringMap(ev)
		pm, pok := asStringMap(pv)
		if eok && pok {
			diffMetaAtPath(em, pm, path, added, removed, changed)
			continue
		}
		if !metaScalarEqual(ev, pv) {
			*changed = append(*changed, path)
		}
	}
}

func metaScalarEqual(a, b any) bool {
	if af, aok := toFloat64(a); aok {
		if bf, bok := toFloat64(b); bok {
			return af == bf
		}
	}
	as, aok := a.([]any)
	bs, bok := b.([]any)
	if aok && bok {
		if len(as) != len(bs) {
			return false
		}
		for i := range as {
			if !metaScalarEqual(as[i], bs[i]) {
				return false
			}
		}
		return true
	}
	return reflect.DeepEqual(a, b)
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}
