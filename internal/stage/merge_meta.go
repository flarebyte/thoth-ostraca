package stage

import "context"

const mergeMetaStage = "merge-meta"
const updateMetaExpectedLuaStage = "update-meta-expectedLua"

func mergeMetaRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	out := in
	mode, embed := errorMode(in.Meta)
	var envErrs []Error
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
		m := map[string]any{"nextMeta": next}
		if r.Post != nil {
			if pm, ok := r.Post.(map[string]any); ok {
				for k, v := range pm {
					m[k] = v
				}
			}
		}
		r.Post = m
		out.Records[i] = r
	}
	appendSanitizedErrors(&out, envErrs)
	return out, nil
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
