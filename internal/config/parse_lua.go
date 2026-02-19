package config

import "cuelang.org/go/cue"

// parseLuaSandboxSection extracts optional lua sandbox settings.
func parseLuaSandboxSection(v cue.Value) LuaSandbox {
	var s LuaSandbox
	lv := v.LookupPath(cue.ParsePath("lua"))
	if !lv.Exists() {
		return s
	}
	s.HasSection = true

	tv := lv.LookupPath(cue.ParsePath("timeoutMs"))
	if tv.Exists() && tv.Kind() == cue.IntKind {
		if err := tv.Decode(&s.TimeoutMs); err == nil {
			s.HasTimeoutMs = true
		}
	}
	iv := lv.LookupPath(cue.ParsePath("instructionLimit"))
	if iv.Exists() && iv.Kind() == cue.IntKind {
		if err := iv.Decode(&s.InstructionLimit); err == nil {
			s.HasInstructionLimit = true
		}
	}
	mv := lv.LookupPath(cue.ParsePath("memoryLimitBytes"))
	if mv.Exists() && mv.Kind() == cue.IntKind {
		if err := mv.Decode(&s.MemoryLimitBytes); err == nil {
			s.HasMemoryLimitBytes = true
		}
	}
	dv := lv.LookupPath(cue.ParsePath("deterministicRandom"))
	if dv.Exists() && dv.Kind() == cue.BoolKind {
		if err := dv.Decode(&s.DeterministicRandom); err == nil {
			s.HasDeterministicRandom = true
		}
	}

	libs := lv.LookupPath(cue.ParsePath("libs"))
	if !libs.Exists() {
		return s
	}
	bv := libs.LookupPath(cue.ParsePath("base"))
	if bv.Exists() && bv.Kind() == cue.BoolKind {
		_ = bv.Decode(&s.Libs.Base)
		s.Libs.HasBase = true
	}
	tblv := libs.LookupPath(cue.ParsePath("table"))
	if tblv.Exists() && tblv.Kind() == cue.BoolKind {
		_ = tblv.Decode(&s.Libs.Table)
		s.Libs.HasTable = true
	}
	sv := libs.LookupPath(cue.ParsePath("string"))
	if sv.Exists() && sv.Kind() == cue.BoolKind {
		_ = sv.Decode(&s.Libs.String)
		s.Libs.HasString = true
	}
	mathv := libs.LookupPath(cue.ParsePath("math"))
	if mathv.Exists() && mathv.Kind() == cue.BoolKind {
		_ = mathv.Decode(&s.Libs.Math)
		s.Libs.HasMath = true
	}
	return s
}
