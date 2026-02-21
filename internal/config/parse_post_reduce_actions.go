package config

import (
	"fmt"

	"cuelang.org/go/cue"
)

// parsePostMapSection extracts optional postMap.inline.
func parsePostMapSection(v cue.Value) PostMap {
	var pm PostMap
	p := v.LookupPath(cue.ParsePath("postMap"))
	if !p.Exists() {
		return pm
	}
	iv := p.LookupPath(cue.ParsePath("inline"))
	if iv.Exists() && iv.Kind() == cue.StringKind {
		if err := iv.Decode(&pm.Inline); err == nil {
			pm.HasInline = true
		}
	}
	return pm
}

// parseReduceSection extracts optional reduce.inline.
func parseReduceSection(v cue.Value) Reduce {
	var r Reduce
	rv := v.LookupPath(cue.ParsePath("reduce"))
	if !rv.Exists() {
		return r
	}
	iv := rv.LookupPath(cue.ParsePath("inline"))
	if iv.Exists() && iv.Kind() == cue.StringKind {
		if err := iv.Decode(&r.Inline); err == nil {
			r.HasInline = true
		}
	}
	return r
}

// parseUpdateMetaSection extracts optional updateMeta.patch object.
func parseUpdateMetaSection(v cue.Value) (UpdateMeta, error) {
	var u UpdateMeta
	uv := v.LookupPath(cue.ParsePath("updateMeta"))
	if !uv.Exists() {
		return u, nil
	}
	u.HasSection = true
	lv := uv.LookupPath(cue.ParsePath("expectedLua.inline"))
	if lv.Exists() {
		if lv.Kind() != cue.StringKind {
			return UpdateMeta{}, fmt.Errorf("invalid updateMeta.expectedLua.inline: must be string")
		}
		if err := lv.Decode(&u.ExpectedLuaInline); err != nil {
			return UpdateMeta{}, fmt.Errorf("invalid updateMeta.expectedLua.inline: must be string")
		}
		u.HasExpectedLuaCode = true
	}
	pv := uv.LookupPath(cue.ParsePath("patch"))
	if !pv.Exists() {
		return u, nil
	}
	tmp := map[string]any{}
	if err := pv.Decode(&tmp); err != nil {
		return UpdateMeta{}, fmt.Errorf("invalid updateMeta.patch: must be object")
	}
	u.Patch = tmp
	u.HasPatch = true
	return u, nil
}

// parseDiffMetaSection extracts optional diffMeta.expectedPatch object.
func parseDiffMetaSection(v cue.Value) (DiffMeta, error) {
	var d DiffMeta
	dv := v.LookupPath(cue.ParsePath("diffMeta"))
	if !dv.Exists() {
		return d, nil
	}
	d.HasSection = true
	d.Format = "summary"
	fv := dv.LookupPath(cue.ParsePath("format"))
	if fv.Exists() {
		var f string
		if err := fv.Decode(&f); err != nil {
			return DiffMeta{}, fmt.Errorf("invalid diffMeta.format: must be 'summary', 'detailed', or 'json-patch'")
		}
		if f != "summary" && f != "detailed" && f != "json-patch" {
			return DiffMeta{}, fmt.Errorf("invalid diffMeta.format: must be 'summary', 'detailed', or 'json-patch'")
		}
		d.Format = f
		d.HasFormat = true
	}
	focv := dv.LookupPath(cue.ParsePath("failOnChange"))
	if focv.Exists() {
		var foc bool
		if err := focv.Decode(&foc); err != nil {
			return DiffMeta{}, fmt.Errorf("invalid diffMeta.failOnChange: must be boolean")
		}
		d.FailOnChange = foc
		d.HasFailOnChange = true
	}
	elv := dv.LookupPath(cue.ParsePath("expectedLua.inline"))
	if elv.Exists() {
		if elv.Kind() != cue.StringKind {
			return DiffMeta{}, fmt.Errorf("invalid diffMeta.expectedLua.inline: must be string")
		}
		if err := elv.Decode(&d.ExpectedLuaInline); err != nil {
			return DiffMeta{}, fmt.Errorf("invalid diffMeta.expectedLua.inline: must be string")
		}
		d.HasExpectedLuaCode = true
	}
	pv := dv.LookupPath(cue.ParsePath("expectedPatch"))
	if !pv.Exists() {
		return d, nil
	}
	tmp := map[string]any{}
	if err := pv.Decode(&tmp); err != nil {
		return DiffMeta{}, fmt.Errorf("invalid diffMeta.expectedPatch: must be object")
	}
	d.ExpectedPatch = tmp
	d.HasExpectedPatch = true
	return d, nil
}
