package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

// compileCUE loads and compiles a CUE file at the given path.
// It preserves the original error messages used by callers.
func compileCUE(path string) (cue.Value, error) {
	if filepath.Ext(path) != ".cue" {
		return cue.Value{}, errors.New("unsupported config format: expected .cue")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cue.Value{}, fmt.Errorf("failed to read config: %w", err)
	}
	ctx := cuecontext.New()
	v := ctx.CompileBytes(data)
	if err := v.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("invalid config: %v", err)
	}
	return v, nil
}

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

// parseFilterSection extracts optional filter.inline.
func parseFilterSection(v cue.Value) Filter {
	var f Filter
	fv := v.LookupPath(cue.ParsePath("filter"))
	if !fv.Exists() {
		return f
	}
	iv := fv.LookupPath(cue.ParsePath("inline"))
	if iv.Exists() && iv.Kind() == cue.StringKind {
		if err := iv.Decode(&f.Inline); err == nil {
			f.HasInline = true
		}
	}
	return f
}

// parseMapSection extracts optional map.inline.
func parseMapSection(v cue.Value) Map {
	var m Map
	mv := v.LookupPath(cue.ParsePath("map"))
	if !mv.Exists() {
		return m
	}
	iv := mv.LookupPath(cue.ParsePath("inline"))
	if iv.Exists() && iv.Kind() == cue.StringKind {
		if err := iv.Decode(&m.Inline); err == nil {
			m.HasInline = true
		}
	}
	return m
}

// parseShellSection extracts optional shell.* fields.
func parseShellSection(v cue.Value) Shell {
	var s Shell
	sv := v.LookupPath(cue.ParsePath("shell"))
	if !sv.Exists() {
		return s
	}
	ev := sv.LookupPath(cue.ParsePath("enabled"))
	if ev.Exists() && ev.Kind() == cue.BoolKind {
		_ = ev.Decode(&s.Enabled)
		s.HasEnabled = true
	}
	pv := sv.LookupPath(cue.ParsePath("program"))
	if pv.Exists() && pv.Kind() == cue.StringKind {
		_ = pv.Decode(&s.Program)
		s.HasProgram = true
	}
	av := sv.LookupPath(cue.ParsePath("argsTemplate"))
	if av.Exists() && av.Kind() == cue.ListKind {
		_ = av.Decode(&s.ArgsTemplate)
		if len(s.ArgsTemplate) > 0 {
			s.HasArgs = true
		}
	}
	tv := sv.LookupPath(cue.ParsePath("timeoutMs"))
	if tv.Exists() && tv.Kind() == cue.IntKind {
		_ = tv.Decode(&s.TimeoutMs)
		s.HasTimeout = true
	}
	return s
}

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

// parseOutputSection extracts optional output.* fields.
func parseOutputSection(v cue.Value) Output {
	var o Output
	ov := v.LookupPath(cue.ParsePath("output"))
	if !ov.Exists() {
		return o
	}
	ovOut := ov.LookupPath(cue.ParsePath("out"))
	if ovOut.Exists() && ovOut.Kind() == cue.StringKind {
		_ = ovOut.Decode(&o.Out)
		o.HasOut = true
	}
	ovPretty := ov.LookupPath(cue.ParsePath("pretty"))
	if ovPretty.Exists() && ovPretty.Kind() == cue.BoolKind {
		_ = ovPretty.Decode(&o.Pretty)
		o.HasPretty = true
	}
	lv := ov.LookupPath(cue.ParsePath("lines"))
	if lv.Exists() && lv.Kind() == cue.BoolKind {
		_ = lv.Decode(&o.Lines)
		o.HasLines = true
	}
	return o
}

// parseErrorsSection extracts optional errors.* fields.
func parseErrorsSection(v cue.Value) Errors {
	var e Errors
	ev := v.LookupPath(cue.ParsePath("errors"))
	if !ev.Exists() {
		return e
	}
	mv := ev.LookupPath(cue.ParsePath("mode"))
	if mv.Exists() && mv.Kind() == cue.StringKind {
		_ = mv.Decode(&e.Mode)
		e.HasMode = true
	}
	emb := ev.LookupPath(cue.ParsePath("embedErrors"))
	if emb.Exists() && emb.Kind() == cue.BoolKind {
		_ = emb.Decode(&e.EmbedErrors)
		e.HasEmbed = true
	}
	return e
}

// parseWorkersSection extracts optional workers count.
func parseWorkersSection(v cue.Value) Workers {
	var w Workers
	wv := v.LookupPath(cue.ParsePath("workers"))
	if wv.Exists() && wv.Kind() == cue.IntKind {
		_ = wv.Decode(&w.Count)
		w.HasCount = true
	}
	return w
}

// parseFileInfoSection extracts optional fileInfo.enabled.
func parseFileInfoSection(v cue.Value) FileInfo {
	var fi FileInfo
	fiv := v.LookupPath(cue.ParsePath("fileInfo"))
	if !fiv.Exists() {
		return fi
	}
	ev := fiv.LookupPath(cue.ParsePath("enabled"))
	if ev.Exists() && ev.Kind() == cue.BoolKind {
		_ = ev.Decode(&fi.Enabled)
		fi.HasEnabled = true
	}
	return fi
}

// parseGitSection extracts optional git.enabled.
func parseGitSection(v cue.Value) Git {
	var g Git
	gv := v.LookupPath(cue.ParsePath("git"))
	if !gv.Exists() {
		return g
	}
	ev := gv.LookupPath(cue.ParsePath("enabled"))
	if ev.Exists() && ev.Kind() == cue.BoolKind {
		_ = ev.Decode(&g.Enabled)
		g.HasEnabled = true
	}
	return g
}
