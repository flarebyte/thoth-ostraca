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
	var meta map[string]any
	if rec.Error != nil {
		return rec, nil, nil
	}
	locator = rec.Locator
	meta = rec.Meta

	ret, violation, err := runLuaScriptWithSandbox(luaMapStage, metaCfg, locator, map[string]any{
		"locator": locator,
		"meta":    meta,
	}, code)
	if err != nil {
		if mode == "keep-going" {
			return Record{Locator: locator, Meta: meta, Error: &RecError{Stage: luaMapStage, Message: err.Error()}}, &Error{Stage: luaMapStage, Locator: locator, Message: err.Error()}, nil
		}
		return Record{}, nil, fmt.Errorf("lua-map: %v", err)
	}
	if violation != "" {
		if mode == "keep-going" {
			return Record{Locator: locator, Meta: meta, Error: &RecError{Stage: luaMapStage, Message: violation}}, &Error{Stage: luaMapStage, Locator: locator, Message: violation}, nil
		}
		return Record{}, nil, luaViolationFailFast(luaMapStage, violation)
	}
	return Record{Locator: locator, Meta: meta, Mapped: ret}, nil, nil
}
