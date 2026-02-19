package config

import "cuelang.org/go/cue"

// LocatorPolicy holds optional locator policy booleans and presence flags.
type LocatorPolicy struct {
	AllowAbsolute   bool
	AllowParentRefs bool
	PosixStyle      bool
	AllowURLs       bool
	HasAllowAbs     bool
	HasAllowParent  bool
	HasPosix        bool
	HasAllowURLs    bool
}

// parseLocatorPolicySection extracts optional locatorPolicy.* fields.
func parseLocatorPolicySection(v cue.Value) LocatorPolicy {
	var lp LocatorPolicy
	pv := v.LookupPath(cue.ParsePath("locatorPolicy"))
	if !pv.Exists() {
		return lp
	}
	aav := pv.LookupPath(cue.ParsePath("allowAbsolute"))
	if aav.Exists() && aav.Kind() == cue.BoolKind {
		_ = aav.Decode(&lp.AllowAbsolute)
		lp.HasAllowAbs = true
	}
	apv := pv.LookupPath(cue.ParsePath("allowParentRefs"))
	if apv.Exists() && apv.Kind() == cue.BoolKind {
		_ = apv.Decode(&lp.AllowParentRefs)
		lp.HasAllowParent = true
	}
	psv := pv.LookupPath(cue.ParsePath("posixStyle"))
	if psv.Exists() && psv.Kind() == cue.BoolKind {
		_ = psv.Decode(&lp.PosixStyle)
		lp.HasPosix = true
	}
	auv := pv.LookupPath(cue.ParsePath("allowURLs"))
	if auv.Exists() && auv.Kind() == cue.BoolKind {
		_ = auv.Decode(&lp.AllowURLs)
		lp.HasAllowURLs = true
	}
	return lp
}
