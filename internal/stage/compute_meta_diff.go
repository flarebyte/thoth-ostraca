package stage

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

const computeMetaDiffStage = "compute-meta-diff"
const diffMetaExpectedLuaStage = "diff-meta-expectedLua"

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
	recordIdxByLocator := map[string]int{}
	for _, r := range in.Records {
		if r.Error != nil || r.Meta == nil {
			continue
		}
		if cp, ok := deepCopyAny(r.Meta).(map[string]any); ok {
			existingByLocator[r.Locator] = cp
		}
	}
	for i := range in.Records {
		if in.Records[i].Error == nil {
			recordIdxByLocator[in.Records[i].Locator] = i
		}
	}

	expected := map[string]any{}
	expectedLuaInline := ""
	mode, embed := errorMode(in.Meta)
	var envErrs []Error
	if in.Meta.DiffMeta != nil {
		expectedLuaInline = in.Meta.DiffMeta.ExpectedLuaInline
	}
	if expectedLuaInline == "" && in.Meta.DiffMeta != nil && in.Meta.DiffMeta.ExpectedPatch != nil {
		if cp, ok := deepCopyAny(in.Meta.DiffMeta.ExpectedPatch).(map[string]any); ok {
			expected = cp
		}
	}

	orphans := make([]string, 0)
	details := make([]DiffDetail, 0)
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
		expectedPerLocator := expected
		if expectedLuaInline != "" {
			next, violation, err := runExpectedLuaInline(diffMetaExpectedLuaStage, in.Meta, loc, existing, expectedLuaInline)
			if err != nil {
				msg := sanitizeErrorMessage(err.Error())
				if mode == "keep-going" {
					if idx, ok := recordIdxByLocator[loc]; ok {
						rr, envE := recordFailure(outRecordFallback(in.Records[idx], loc), diffMetaExpectedLuaStage, msg, embed)
						in.Records[idx] = rr
						if envE != nil {
							envErrs = append(envErrs, *envE)
						}
					} else {
						envErrs = append(envErrs, Error{Stage: diffMetaExpectedLuaStage, Locator: loc, Message: msg})
					}
					continue
				}
				return Envelope{}, luaViolationFailFast(diffMetaExpectedLuaStage, msg)
			}
			if violation != "" {
				msg := sanitizeErrorMessage(violation)
				if mode == "keep-going" {
					if idx, ok := recordIdxByLocator[loc]; ok {
						rr, envE := recordFailure(outRecordFallback(in.Records[idx], loc), diffMetaExpectedLuaStage, msg, embed)
						in.Records[idx] = rr
						if envE != nil {
							envErrs = append(envErrs, *envE)
						}
					} else {
						envErrs = append(envErrs, Error{Stage: diffMetaExpectedLuaStage, Locator: loc, Message: msg})
					}
					continue
				}
				return Envelope{}, luaViolationFailFast(diffMetaExpectedLuaStage, msg)
			}
			expectedPerLocator = next
		}
		format := "summary"
		if in.Meta != nil && in.Meta.DiffMeta != nil && in.Meta.DiffMeta.Format != "" {
			format = in.Meta.DiffMeta.Format
		}
		s := diffMetaMapsV3(existing, expectedPerLocator)
		if format == "detailed" {
			s = diffMetaMapsV3Detailed(existing, expectedPerLocator)
		} else if format == "json-patch" {
			s = diffMetaMapsV3JSONPatch(existing, expectedPerLocator)
		}
		details = append(details, DiffDetail{
			Locator:         loc,
			MetaFile:        metaFile,
			AddedKeys:       s.added,
			RemovedKeys:     s.removed,
			ChangedKeys:     s.changed,
			TypeChangedKeys: s.typeChanged,
			Arrays:          s.arrays,
			Changes:         s.changes,
			Patch:           s.patch,
		})
	}
	sort.Slice(details, func(i, j int) bool { return details[i].Locator < details[j].Locator })

	only := "all"
	if in.Meta != nil && in.Meta.DiffMeta != nil && in.Meta.DiffMeta.Only != "" {
		only = in.Meta.DiffMeta.Only
	}
	details = filterDiffDetails(details, only)

	changedCount := 0
	for _, d := range details {
		if detailHasChanges(d) {
			changedCount++
		}
	}

	out := in
	var outDetails []DiffDetail
	if only != "orphans" {
		outDetails = details
	}
	out.Meta.Diff = &DiffReport{
		OrphanMetaFiles: orphans,
		PairedCount:     len(details),
		OrphanCount:     len(orphans),
		ChangedCount:    changedCount,
		Details:         outDetails,
		Orphans:         orphans,
		PresentCount:    len(details),
	}
	if in.Meta != nil && in.Meta.DiffMeta != nil && in.Meta.DiffMeta.Summary && deps.Stderr != nil {
		emitDiffSummary(deps.Stderr, out.Meta.Diff)
	}
	appendSanitizedErrors(&out, envErrs)
	return out, nil
}

func emitDiffSummary(w interface{ Write([]byte) (int, error) }, report *DiffReport) {
	if report == nil {
		return
	}
	_, _ = fmt.Fprintf(w, "diff-summary paired=%d changed=%d orphans=%d\n", report.PairedCount, report.ChangedCount, report.OrphanCount)
	changed := make([]DiffDetail, 0, len(report.Details))
	for _, d := range report.Details {
		if detailHasChanges(d) {
			changed = append(changed, d)
		}
	}
	sort.Slice(changed, func(i, j int) bool { return changed[i].Locator < changed[j].Locator })
	for _, d := range changed {
		arraysCount := 0
		for _, ad := range d.Arrays {
			if arrayDiffHasChanges(ad) {
				arraysCount++
			}
		}
		_, _ = fmt.Fprintf(
			w,
			"changed locator=%s added=%d removed=%d changed=%d typeChanged=%d arrays=%d\n",
			d.Locator,
			len(d.AddedKeys),
			len(d.RemovedKeys),
			len(d.ChangedKeys),
			len(d.TypeChangedKeys),
			arraysCount,
		)
	}
	orphans := append([]string(nil), report.OrphanMetaFiles...)
	sort.Strings(orphans)
	for _, orphan := range orphans {
		_, _ = fmt.Fprintf(w, "orphan metaFile=%s\n", orphan)
	}
}

func filterDiffDetails(details []DiffDetail, only string) []DiffDetail {
	switch only {
	case "changed":
		out := make([]DiffDetail, 0, len(details))
		for _, d := range details {
			if detailHasChanges(d) {
				out = append(out, d)
			}
		}
		return out
	case "unchanged":
		out := make([]DiffDetail, 0, len(details))
		for _, d := range details {
			if !detailHasChanges(d) {
				out = append(out, d)
			}
		}
		return out
	case "orphans":
		return nil
	default:
		return details
	}
}

func detailHasChanges(d DiffDetail) bool {
	if len(d.AddedKeys) > 0 || len(d.RemovedKeys) > 0 || len(d.ChangedKeys) > 0 || len(d.TypeChangedKeys) > 0 {
		return true
	}
	if len(d.Changes) > 0 || len(d.Patch) > 0 {
		return true
	}
	for _, ad := range d.Arrays {
		if arrayDiffHasChanges(ad) {
			return true
		}
	}
	return false
}

func arrayDiffHasChanges(d ArrayDiff) bool {
	return len(d.AddedIndices) > 0 || len(d.RemovedIndices) > 0 || len(d.ChangedIndices) > 0
}

func outRecordFallback(r Record, locator string) Record {
	if r.Locator == "" {
		r.Locator = locator
	}
	return r
}

func init() { Register(computeMetaDiffStage, computeMetaDiffRunner) }

func diffMetaMaps(existing, expected map[string]any) ([]string, []string, []string) {
	s := diffMetaMapsV3(existing, expected)
	return s.added, s.removed, s.changed
}

type diffSummary struct {
	added       []string
	removed     []string
	changed     []string
	typeChanged []string
	arrays      []ArrayDiff
	changes     []DiffChange
	patch       []DiffOp
}

type diffCollector struct {
	added       []string
	removed     []string
	changed     []string
	typeChanged []string
	arrays      []ArrayDiff
	changes     []DiffChange
	format      string
}

func diffMetaMapsV3(existing, expected map[string]any) diffSummary {
	c := &diffCollector{}
	c.compareMaps("", existing, expected)
	return diffSummary{
		added:       uniqueSortedStrings(c.added),
		removed:     uniqueSortedStrings(c.removed),
		changed:     uniqueSortedStrings(c.changed),
		typeChanged: uniqueSortedStrings(c.typeChanged),
		arrays:      normalizeArrayDiffs(c.arrays),
	}
}

func diffMetaMapsV3Detailed(existing, expected map[string]any) diffSummary {
	c := &diffCollector{format: "detailed"}
	c.compareMaps("", existing, expected)
	return diffSummary{
		added:       uniqueSortedStrings(c.added),
		removed:     uniqueSortedStrings(c.removed),
		changed:     uniqueSortedStrings(c.changed),
		typeChanged: uniqueSortedStrings(c.typeChanged),
		arrays:      normalizeArrayDiffs(c.arrays),
		changes:     sortDiffChanges(c.changes),
	}
}

func diffMetaMapsV3JSONPatch(existing, expected map[string]any) diffSummary {
	s := diffMetaMapsV3(existing, expected)
	s.patch = diffMetaJSONPatch(existing, expected)
	return s
}

func (c *diffCollector) compareMaps(prefix string, existing, expected map[string]any) {
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
			c.added = append(c.added, path)
			c.addChange(path, "added", nil, pv)
			continue
		}
		if inExisting && !inExpected {
			c.removed = append(c.removed, path)
			c.addChange(path, "removed", ev, nil)
			continue
		}
		c.compareValues(path, ev, pv)
	}
}

func (c *diffCollector) compareValues(path string, existing, expected any) {
	em, eok := asStringMap(existing)
	pm, pok := asStringMap(expected)
	if eok && pok {
		c.compareMaps(path, em, pm)
		return
	}
	ea, eok := existing.([]any)
	pa, pok := expected.([]any)
	if eok && pok {
		c.compareArrays(path, ea, pa)
		return
	}
	if jsonType(existing) != jsonType(expected) {
		c.typeChanged = append(c.typeChanged, path)
		c.changed = append(c.changed, path)
		c.addChange(path, "type-changed", existing, expected)
		return
	}
	if !metaScalarEqual(existing, expected) {
		c.changed = append(c.changed, path)
		c.addChange(path, "changed", existing, expected)
	}
}

func (c *diffCollector) compareArrays(path string, existing, expected []any) {
	diff := ArrayDiff{Path: path}
	n := len(existing)
	if len(expected) < n {
		n = len(expected)
	}
	for i := 0; i < n; i++ {
		if metaScalarEqual(existing[i], expected[i]) {
			continue
		}
		diff.ChangedIndices = append(diff.ChangedIndices, i)
		idxPath := fmt.Sprintf("%s[%d]", path, i)
		if c.format == "detailed" {
			c.addChange(idxPath, "array-index-changed", existing[i], expected[i])
			continue
		}
		c.compareValues(idxPath, existing[i], expected[i])
	}
	for i := n; i < len(expected); i++ {
		diff.AddedIndices = append(diff.AddedIndices, i)
		if c.format == "detailed" {
			c.addChange(fmt.Sprintf("%s[%d]", path, i), "added", nil, expected[i])
		}
	}
	for i := n; i < len(existing); i++ {
		diff.RemovedIndices = append(diff.RemovedIndices, i)
		if c.format == "detailed" {
			c.addChange(fmt.Sprintf("%s[%d]", path, i), "removed", existing[i], nil)
		}
	}
	diff.AddedIndices = uniqueSortedInts(diff.AddedIndices)
	diff.RemovedIndices = uniqueSortedInts(diff.RemovedIndices)
	diff.ChangedIndices = uniqueSortedInts(diff.ChangedIndices)
	if len(diff.AddedIndices) > 0 || len(diff.RemovedIndices) > 0 || len(diff.ChangedIndices) > 0 {
		c.arrays = append(c.arrays, diff)
	}
}

func (c *diffCollector) addChange(path, kind string, oldValue, newValue any) {
	if c.format != "detailed" {
		return
	}
	ch := DiffChange{Path: path, Kind: kind}
	switch kind {
	case "added":
		ch.NewValue = deepCopyAny(newValue)
	case "removed":
		ch.OldValue = deepCopyAny(oldValue)
	default:
		ch.OldValue = deepCopyAny(oldValue)
		ch.NewValue = deepCopyAny(newValue)
	}
	c.changes = append(c.changes, ch)
}

func sortDiffChanges(in []DiffChange) []DiffChange {
	if len(in) == 0 {
		return nil
	}
	out := append([]DiffChange(nil), in...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}

func diffMetaJSONPatch(existing, expected map[string]any) []DiffOp {
	ops := make([]DiffOp, 0)
	collectDiffMetaJSONPatchOps("", existing, expected, &ops)
	if len(ops) == 0 {
		return nil
	}
	sort.Slice(ops, func(i, j int) bool {
		if ops[i].Path != ops[j].Path {
			return ops[i].Path < ops[j].Path
		}
		return ops[i].Op < ops[j].Op
	})
	return ops
}

func collectDiffMetaJSONPatchOps(base string, existing, expected map[string]any, ops *[]DiffOp) {
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
		path := joinJSONPointer(base, k)
		ev, inExisting := existing[k]
		pv, inExpected := expected[k]
		if !inExisting && inExpected {
			*ops = append(*ops, DiffOp{Op: "add", Path: path, Value: deepCopyAny(pv)})
			continue
		}
		if inExisting && !inExpected {
			*ops = append(*ops, DiffOp{Op: "remove", Path: path})
			continue
		}

		em, emOK := asStringMap(ev)
		pm, pmOK := asStringMap(pv)
		if emOK && pmOK {
			collectDiffMetaJSONPatchOps(path, em, pm, ops)
			continue
		}

		ea, eaOK := ev.([]any)
		pa, paOK := pv.([]any)
		if eaOK && paOK {
			if !metaScalarEqual(ea, pa) {
				*ops = append(*ops, DiffOp{Op: "replace", Path: path, Value: deepCopyAny(pv)})
			}
			continue
		}

		if !metaScalarEqual(ev, pv) {
			*ops = append(*ops, DiffOp{Op: "replace", Path: path, Value: deepCopyAny(pv)})
		}
	}
}

func joinJSONPointer(base, segment string) string {
	if base == "" {
		return "/" + escapeJSONPointer(segment)
	}
	return base + "/" + escapeJSONPointer(segment)
}

func escapeJSONPointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	return strings.ReplaceAll(s, "/", "~1")
}

func jsonType(v any) string {
	if v == nil {
		return "null"
	}
	if _, ok := asStringMap(v); ok {
		return "object"
	}
	if _, ok := v.([]any); ok {
		return "array"
	}
	if _, ok := toFloat64(v); ok {
		return "number"
	}
	switch v.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	default:
		return fmt.Sprintf("%T", v)
	}
}

func uniqueSortedStrings(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	sort.Strings(in)
	out := make([]string, 0, len(in))
	for _, s := range in {
		if len(out) == 0 || out[len(out)-1] != s {
			out = append(out, s)
		}
	}
	return out
}

func uniqueSortedInts(in []int) []int {
	if len(in) == 0 {
		return []int{}
	}
	sort.Ints(in)
	out := make([]int, 0, len(in))
	for _, v := range in {
		if len(out) == 0 || out[len(out)-1] != v {
			out = append(out, v)
		}
	}
	return out
}

func normalizeArrayDiffs(in []ArrayDiff) []ArrayDiff {
	if len(in) == 0 {
		return nil
	}
	sort.Slice(in, func(i, j int) bool { return in[i].Path < in[j].Path })
	out := make([]ArrayDiff, 0, len(in))
	for _, d := range in {
		d.AddedIndices = uniqueSortedInts(d.AddedIndices)
		d.RemovedIndices = uniqueSortedInts(d.RemovedIndices)
		d.ChangedIndices = uniqueSortedInts(d.ChangedIndices)
		if len(out) == 0 || out[len(out)-1].Path != d.Path {
			out = append(out, d)
			continue
		}
		last := &out[len(out)-1]
		last.AddedIndices = uniqueSortedInts(append(last.AddedIndices, d.AddedIndices...))
		last.RemovedIndices = uniqueSortedInts(append(last.RemovedIndices, d.RemovedIndices...))
		last.ChangedIndices = uniqueSortedInts(append(last.ChangedIndices, d.ChangedIndices...))
	}
	for i := range out {
		if len(out[i].AddedIndices) == 0 {
			out[i].AddedIndices = nil
		}
		if len(out[i].RemovedIndices) == 0 {
			out[i].RemovedIndices = nil
		}
		if len(out[i].ChangedIndices) == 0 {
			out[i].ChangedIndices = nil
		}
	}
	return out
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
