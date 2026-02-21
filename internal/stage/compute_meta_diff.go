package stage

import (
	"context"
	"fmt"
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
				handled, fatalErr := handleDiffMetaExpectedLuaFailure(&in, recordIdxByLocator, &envErrs, loc, err.Error(), mode, embed)
				if handled {
					continue
				}
				return Envelope{}, fatalErr
			}
			if violation != "" {
				handled, fatalErr := handleDiffMetaExpectedLuaFailure(&in, recordIdxByLocator, &envErrs, loc, violation, mode, embed)
				if handled {
					continue
				}
				return Envelope{}, fatalErr
			}
			expectedPerLocator = next
		}
		format := "summary"
		if in.Meta != nil && in.Meta.DiffMeta != nil && in.Meta.DiffMeta.Format != "" {
			format = in.Meta.DiffMeta.Format
		}
		var s diffSummary
		switch format {
		case "detailed":
			s = diffMetaMapsV3Detailed(existing, expectedPerLocator)
		case "json-patch":
			s = diffMetaMapsV3JSONPatch(existing, expectedPerLocator)
		default:
			s = diffMetaMapsV3(existing, expectedPerLocator)
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

func handleDiffMetaExpectedLuaFailure(
	in *Envelope,
	recordIdxByLocator map[string]int,
	envErrs *[]Error,
	loc, message, mode string,
	embed bool,
) (bool, error) {
	msg := sanitizeErrorMessage(message)
	if mode == "keep-going" {
		if idx, ok := recordIdxByLocator[loc]; ok {
			rr, envE := recordFailure(outRecordFallback(in.Records[idx], loc), diffMetaExpectedLuaStage, msg, embed)
			in.Records[idx] = rr
			if envE != nil {
				*envErrs = append(*envErrs, *envE)
			}
		} else {
			*envErrs = append(*envErrs, Error{Stage: diffMetaExpectedLuaStage, Locator: loc, Message: msg})
		}
		return true, nil
	}
	return false, luaViolationFailFast(diffMetaExpectedLuaStage, msg)
}

func init() { Register(computeMetaDiffStage, computeMetaDiffRunner) }
