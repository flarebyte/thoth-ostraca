// File Guide for dev/ai agents:
// Purpose: Turn config-supplied Lua predicates into per-record filter behavior for record-driven actions.
// Responsibilities:
// - Normalize filter code into a runnable predicate.
// - Execute the predicate against one record with the current sandbox context.
// - Translate predicate failures into fail-fast or keep-going stage errors.
// Architecture notes:
// - The code-wrapping behavior for bare expressions is intentional and shared with other Lua stage helpers.
// - This file also hosts shared Lua conversion helpers used by the sandbox entrypoint; keep that coupling in mind before moving them.
package stage

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

const luaFilterStage = "lua-filter"

type luaFilterRes struct {
	idx   int
	keep  bool
	out   Record
	envE  *Error
	fatal error
}

// buildLuaPredicate returns the predicate string from envelope meta, wrapping
// expressions without an explicit return.
func buildLuaPredicate(in Envelope) string {
	pred := "return true"
	if in.Meta != nil && in.Meta.Lua != nil && in.Meta.Lua.FilterInline != "" {
		code := in.Meta.Lua.FilterInline
		if !containsReturn(code) {
			pred = "return (" + code + ")"
		} else {
			pred = code
		}
	}
	return pred
}

// processLuaFilterRecord applies the lua predicate to a single record and
// returns the outcome mimicking the original behavior. The mode determines
// whether to keep going or fail fast.
func processLuaFilterRecord(rec Record, pred string, mode string, metaCfg *Meta) (keep bool, out Record, envE *Error, fatal error) {
	locator := rec.Locator
	meta := rec.Meta
	recErr := rec.Error

	if recErr != nil {
		return true, rec, nil, nil
	}

	ret, violation, err := runLuaScriptWithSandbox(
		luaFilterStage,
		metaCfg,
		locator,
		luaRecordContext(rec),
		pred,
	)
	if err != nil {
		msg := formatLuaError(luaFilterStage, locator, pred, err.Error())
		if mode == "keep-going" {
			rr, envErr := recordFailure(rec, luaFilterStage, msg, true)
			return true, rr, envErr, nil
		}
		return false, Record{}, nil, fmt.Errorf("lua-filter: %s", msg)
	}
	if violation != "" {
		msg := formatLuaError(luaFilterStage, locator, pred, violation)
		if mode == "keep-going" {
			rr, envErr := recordFailure(Record{Locator: locator, Meta: meta}, luaFilterStage, msg, true)
			return true, rr, envErr, nil
		}
		return false, Record{}, nil, luaViolationFailFast(luaFilterStage, msg)
	}
	keep, _ = ret.(bool)
	if keep {
		return true, rec, nil, nil
	}
	return false, Record{}, nil, nil
}

// containsReturn reports whether the code string contains the token "return".
func containsReturn(s string) bool {
	for i := 0; i+5 <= len(s); i++ {
		if s[i] == 'r' && i+6 <= len(s) && s[i:i+6] == "return" {
			return true
		}
	}
	return false
}

// toLValue converts a Go value to a Lua value.
func toLValue(L *lua.LState, v any) lua.LValue {
	switch x := v.(type) {
	case nil:
		return lua.LNil
	case string:
		return lua.LString(x)
	case bool:
		if x {
			return lua.LTrue
		}
		return lua.LFalse
	case int:
		return lua.LNumber(float64(x))
	case int64:
		return lua.LNumber(float64(x))
	case float64:
		return lua.LNumber(x)
	case map[string]any:
		tbl := L.NewTable()
		for k, v2 := range x {
			tbl.RawSetString(k, toLValue(L, v2))
		}
		return tbl
	case []any:
		tbl := L.NewTable()
		for i, v2 := range x {
			tbl.RawSetInt(i+1, toLValue(L, v2))
		}
		return tbl
	default:
		return lua.LNil
	}
}
