package config

import "cuelang.org/go/cue"

// parseDiscoverySection extracts optional discovery.* fields.
func parseDiscoverySection(v cue.Value) Discovery {
	var d Discovery
	dv := v.LookupPath(cue.ParsePath("discovery"))
	if !dv.Exists() {
		return d
	}
	rv := dv.LookupPath(cue.ParsePath("root"))
	if rv.Exists() && rv.Kind() == cue.StringKind {
		if err := rv.Decode(&d.Root); err == nil {
			d.HasRoot = true
		}
	}
	ngv := dv.LookupPath(cue.ParsePath("noGitignore"))
	if ngv.Exists() && (ngv.Kind() == cue.BoolKind) {
		if err := ngv.Decode(&d.NoGitignore); err == nil {
			d.HasNoGitignore = true
		}
	}
	fsv := dv.LookupPath(cue.ParsePath("followSymlinks"))
	if fsv.Exists() && (fsv.Kind() == cue.BoolKind) {
		if err := fsv.Decode(&d.FollowSymlinks); err == nil {
			d.HasFollowSymlink = true
		}
	}
	return d
}

// parseValidationSection extracts optional validation.allowUnknownTopLevel.
func parseValidationSection(v cue.Value) Validation {
	var val Validation
	vv := v.LookupPath(cue.ParsePath("validation"))
	if !vv.Exists() {
		return val
	}
	auv := vv.LookupPath(cue.ParsePath("allowUnknownTopLevel"))
	if auv.Exists() && (auv.Kind() == cue.BoolKind) {
		if err := auv.Decode(&val.AllowUnknownTopLevel); err == nil {
			val.HasAllowUnknownTop = true
		}
	}
	return val
}

// parseLimitsSection extracts optional limits.maxYAMLBytes and limits.maxRecordsInMemory.
func parseLimitsSection(v cue.Value) Limits {
	var l Limits
	lv := v.LookupPath(cue.ParsePath("limits"))
	if !lv.Exists() {
		return l
	}
	mv := lv.LookupPath(cue.ParsePath("maxYAMLBytes"))
	if mv.Exists() && mv.Kind() == cue.IntKind {
		if err := mv.Decode(&l.MaxYAMLBytes); err == nil {
			l.HasMaxYAMLBytes = true
		}
	}
	rv := lv.LookupPath(cue.ParsePath("maxRecordsInMemory"))
	if rv.Exists() && rv.Kind() == cue.IntKind {
		if err := rv.Decode(&l.MaxRecordsInMemory); err == nil {
			l.HasMaxRecordsInMemory = true
		}
	}
	return l
}
