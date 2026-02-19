package config

import "cuelang.org/go/cue"

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
	s.HasSection = true
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
	wv := sv.LookupPath(cue.ParsePath("workingDir"))
	if wv.Exists() && wv.Kind() == cue.StringKind {
		_ = wv.Decode(&s.WorkingDir)
		s.HasWorkingDir = true
	}
	envv := sv.LookupPath(cue.ParsePath("env"))
	if envv.Exists() {
		tmp := map[string]string{}
		if err := envv.Decode(&tmp); err == nil {
			s.Env = tmp
			s.HasEnv = true
		}
	}
	tv := sv.LookupPath(cue.ParsePath("timeoutMs"))
	if tv.Exists() && tv.Kind() == cue.IntKind {
		_ = tv.Decode(&s.TimeoutMs)
		s.HasTimeout = true
	}
	cv := sv.LookupPath(cue.ParsePath("capture"))
	if cv.Exists() {
		s.HasCapture = true
		sov := cv.LookupPath(cue.ParsePath("stdout"))
		if sov.Exists() && sov.Kind() == cue.BoolKind {
			_ = sov.Decode(&s.CaptureStdout)
			s.HasCaptureStdout = true
		}
		sev := cv.LookupPath(cue.ParsePath("stderr"))
		if sev.Exists() && sev.Kind() == cue.BoolKind {
			_ = sev.Decode(&s.CaptureStderr)
			s.HasCaptureStderr = true
		}
		mbv := cv.LookupPath(cue.ParsePath("maxBytes"))
		if mbv.Exists() && mbv.Kind() == cue.IntKind {
			_ = mbv.Decode(&s.CaptureMaxBytes)
			s.HasCaptureMax = true
		}
	}
	stv := sv.LookupPath(cue.ParsePath("strictTemplating"))
	if stv.Exists() && stv.Kind() == cue.BoolKind {
		_ = stv.Decode(&s.StrictTemplating)
		s.HasStrictTpl = true
	}
	kv := sv.LookupPath(cue.ParsePath("killProcessGroup"))
	if kv.Exists() && kv.Kind() == cue.BoolKind {
		_ = kv.Decode(&s.KillProcessGroup)
		s.HasKillPG = true
	}
	tgv := sv.LookupPath(cue.ParsePath("termGraceMs"))
	if tgv.Exists() && tgv.Kind() == cue.IntKind {
		_ = tgv.Decode(&s.TermGraceMs)
		s.HasTermGrace = true
	}
	return s
}
