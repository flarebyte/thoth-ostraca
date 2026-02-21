package config

import "cuelang.org/go/cue"

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

// parseUISection extracts optional ui.* fields.
func parseUISection(v cue.Value) UI {
	var u UI
	uv := v.LookupPath(cue.ParsePath("ui"))
	if !uv.Exists() {
		return u
	}
	u.HasSection = true
	pv := uv.LookupPath(cue.ParsePath("progress"))
	if pv.Exists() && pv.Kind() == cue.BoolKind {
		_ = pv.Decode(&u.Progress)
		u.HasProgress = true
	}
	iv := uv.LookupPath(cue.ParsePath("progressIntervalMs"))
	if iv.Exists() && iv.Kind() == cue.IntKind {
		_ = iv.Decode(&u.ProgressIntervalMs)
		u.HasIntervalMs = true
	}
	return u
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
