package stage

import (
	"context"
	"fmt"
)

const mergeMetaStage = "merge-meta"
const updateMetaExpectedLuaStage = "update-meta-expectedLua"

func mergeMetaRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	out := in
	mode, embed := errorMode(in.Meta)
	var envErrs []Error
	persistEnabled := in.Meta != nil &&
		in.Meta.PersistMeta != nil &&
		in.Meta.PersistMeta.Enabled
	derived := map[string]any{}
	expectedLuaInline := ""
	if in.Meta != nil && in.Meta.UpdateMeta != nil {
		expectedLuaInline = in.Meta.UpdateMeta.ExpectedLuaInline
	}
	if expectedLuaInline == "" && in.Meta != nil && in.Meta.UpdateMeta != nil && in.Meta.UpdateMeta.Patch != nil {
		if cp, ok := deepCopyAny(in.Meta.UpdateMeta.Patch).(map[string]any); ok {
			derived = cp
		}
	}
	for i, r := range in.Records {
		if r.Error != nil {
			continue
		}
		var existing map[string]any
		if r.Post != nil {
			if pm, ok := r.Post.(map[string]any); ok {
				if em, ok2 := pm["existingMeta"].(map[string]any); ok2 {
					existing = em
				}
			}
		}
		if persistEnabled {
			derivedPerRecord, handled, outErr := mergeMetaFromPost(
				&out,
				&envErrs,
				i,
				r,
				mode,
				embed,
			)
			if outErr != nil {
				return Envelope{}, outErr
			}
			if handled {
				continue
			}
			next := deepMerge(existing, derivedPerRecord)
			r.Post = withNextMeta(next, r.Post)
			out.Records[i] = r
			continue
		}
		derivedPerRecord := derived
		if expectedLuaInline != "" {
			next, violation, err := runExpectedLuaInline(updateMetaExpectedLuaStage, in.Meta, r.Locator, existing, expectedLuaInline)
			if err != nil {
				handled, outErr := handleUpdateMetaLuaFailure(&out, &envErrs, i, r, mode, embed, err.Error())
				if outErr != nil {
					return Envelope{}, outErr
				}
				if handled {
					continue
				}
			}
			if violation != "" {
				handled, outErr := handleUpdateMetaLuaFailure(&out, &envErrs, i, r, mode, embed, violation)
				if outErr != nil {
					return Envelope{}, outErr
				}
				if handled {
					continue
				}
			}
			derivedPerRecord = next
		}
		next := deepMerge(existing, derivedPerRecord)
		r.Post = withNextMeta(next, r.Post)
		out.Records[i] = r
	}
	appendSanitizedErrors(&out, envErrs)
	return out, nil
}

func withNextMeta(next map[string]any, post any) map[string]any {
	m := map[string]any{"nextMeta": next}
	pm, ok := post.(map[string]any)
	if !ok {
		return m
	}
	for k, v := range pm {
		m[k] = v
	}
	return m
}

func mergeMetaFromPost(
	out *Envelope,
	envErrs *[]Error,
	idx int,
	rec Record,
	mode string,
	embed bool,
) (map[string]any, bool, error) {
	pm, ok := rec.Post.(map[string]any)
	if !ok {
		msg := "post.meta missing or invalid"
		return nil, true, handleMergeMetaFailure(
			out,
			envErrs,
			idx,
			rec,
			mode,
			embed,
			msg,
		)
	}
	rawMeta, ok := pm["meta"]
	if !ok {
		msg := "post.meta missing or invalid"
		return nil, true, handleMergeMetaFailure(
			out,
			envErrs,
			idx,
			rec,
			mode,
			embed,
			msg,
		)
	}
	patch, ok := asStringMap(rawMeta)
	if !ok {
		msg := "post.meta must be object"
		return nil, true, handleMergeMetaFailure(
			out,
			envErrs,
			idx,
			rec,
			mode,
			embed,
			msg,
		)
	}
	if cp, ok := deepCopyAny(patch).(map[string]any); ok {
		return cp, false, nil
	}
	return map[string]any{}, false, nil
}

func handleMergeMetaFailure(
	out *Envelope,
	envErrs *[]Error,
	idx int,
	rec Record,
	mode string,
	embed bool,
	rawMsg string,
) error {
	msg := sanitizeErrorMessage(rawMsg)
	if mode == "keep-going" {
		rr, envE := recordFailure(rec, mergeMetaStage, msg, embed)
		out.Records[idx] = rr
		if envE != nil {
			*envErrs = append(*envErrs, *envE)
		}
		return nil
	}
	return fmt.Errorf("%s: %s", mergeMetaStage, msg)
}

func handleUpdateMetaLuaFailure(out *Envelope, envErrs *[]Error, idx int, rec Record, mode string, embed bool, rawMsg string) (bool, error) {
	msg := sanitizeErrorMessage(rawMsg)
	if mode == "keep-going" {
		rr, envE := recordFailure(rec, updateMetaExpectedLuaStage, msg, embed)
		out.Records[idx] = rr
		if envE != nil {
			*envErrs = append(*envErrs, *envE)
		}
		return true, nil
	}
	return false, luaViolationFailFast(updateMetaExpectedLuaStage, msg)
}

func init() { Register(mergeMetaStage, mergeMetaRunner) }

func deepMerge(existing map[string]any, patch map[string]any) map[string]any {
	base := map[string]any{}
	if existing != nil {
		if cp, ok := deepCopyAny(existing).(map[string]any); ok {
			base = cp
		}
	}
	if patch == nil {
		return base
	}
	for k, pv := range patch {
		if ev, ok := base[k]; ok {
			em, eok := asStringMap(ev)
			pm, pok := asStringMap(pv)
			if eok && pok {
				base[k] = deepMerge(em, pm)
				continue
			}
		}
		base[k] = deepCopyAny(pv)
	}
	return base
}

func asStringMap(v any) (map[string]any, bool) {
	switch x := v.(type) {
	case map[string]any:
		return x, true
	case map[any]any:
		out := map[string]any{}
		for k, vv := range x {
			ks, ok := k.(string)
			if !ok {
				return nil, false
			}
			out[ks] = vv
		}
		return out, true
	default:
		return nil, false
	}
}
