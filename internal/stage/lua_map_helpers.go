// File Guide for dev/ai agents:
// Purpose: Execute Lua map transforms that derive per-record structured data from the current record context.
// Responsibilities:
// - Normalize map code into runnable Lua.
// - Run the map transform for one record.
// - Attach mapped output or stage errors back onto the record.
// Architecture notes:
// - Map stage behavior is intentionally minimal: it only computes `record.Mapped`; later stages decide how to consume it.
// - Error shaping mirrors the other Lua helpers so keep-going/fail-fast behavior stays consistent across stages.
package stage

import (
	"fmt"
)

const luaMapStage = "lua-map"

// buildLuaMapCode returns the Lua mapping code, wrapping expressions without explicit return.
func buildLuaMapCode(in Envelope) string {
	code := "return { locator = locator, meta = meta }"
	if in.Meta != nil && in.Meta.Lua != nil && in.Meta.Lua.MapInline != "" {
		c := in.Meta.Lua.MapInline
		if !containsReturn(c) {
			code = "return (" + c + ")"
		} else {
			code = c
		}
	}
	return code
}

// processLuaMapRecord runs the Lua map code for a single record.
func processLuaMapRecord(rec Record, code string, mode string, metaCfg *Meta) (Record, *Error, error) {
	var locator string
	if rec.Error != nil {
		return rec, nil, nil
	}
	locator = rec.Locator

	ret, violation, err := runLuaScriptWithSandbox(
		luaMapStage,
		metaCfg,
		locator,
		luaRecordContext(rec),
		code,
	)
	if err != nil {
		msg := formatLuaError(luaMapStage, locator, code, err.Error())
		if mode == "keep-going" {
			rec.Error = &RecError{Stage: luaMapStage, Message: msg}
			return rec, &Error{Stage: luaMapStage, Locator: locator, Message: msg}, nil
		}
		return Record{}, nil, fmt.Errorf("lua-map: %s", msg)
	}
	if violation != "" {
		msg := formatLuaError(luaMapStage, locator, code, violation)
		if mode == "keep-going" {
			rec.Error = &RecError{Stage: luaMapStage, Message: msg}
			return rec, &Error{Stage: luaMapStage, Locator: locator, Message: msg}, nil
		}
		return Record{}, nil, luaViolationFailFast(luaMapStage, msg)
	}
	rec.Mapped = ret
	return rec, nil, nil
}
