// File Guide for dev/ai agents:
// Purpose: Parse the discovery, validation, and limits sections that shape broad run behavior.
// Responsibilities:
// - Decode discovery root and include/exclude controls.
// - Decode validation flags for config strictness.
// - Decode processing limits such as YAML size and in-memory record caps.
// Architecture notes:
// - These sections are grouped here because they are global execution controls rather than action-specific features.
// - Each parser records Has* flags so later validation can tell omitted values from explicit settings.
package config

import (
	"fmt"

	"cuelang.org/go/cue"
)

// parseDiscoverySection extracts optional discovery.* fields.
func parseDiscoverySection(v cue.Value) (Discovery, error) {
	var d Discovery
	dv := v.LookupPath(cue.ParsePath("discovery"))
	if !dv.Exists() {
		return d, nil
	}
	rv := dv.LookupPath(cue.ParsePath("root"))
	if rv.Exists() && rv.Kind() == cue.StringKind {
		if err := rv.Decode(&d.Root); err == nil {
			d.HasRoot = true
		}
	}
	iv := dv.LookupPath(cue.ParsePath("include"))
	if iv.Exists() {
		if iv.Kind() != cue.ListKind {
			return Discovery{}, fmt.Errorf("invalid discovery.include: must be list of strings")
		}
		if err := iv.Decode(&d.Include); err != nil {
			return Discovery{}, fmt.Errorf("invalid discovery.include: must be list of strings")
		}
		d.HasInclude = true
	}
	ev := dv.LookupPath(cue.ParsePath("exclude"))
	if ev.Exists() {
		if ev.Kind() != cue.ListKind {
			return Discovery{}, fmt.Errorf("invalid discovery.exclude: must be list of strings")
		}
		if err := ev.Decode(&d.Exclude); err != nil {
			return Discovery{}, fmt.Errorf("invalid discovery.exclude: must be list of strings")
		}
		d.HasExclude = true
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
	return d, nil
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
