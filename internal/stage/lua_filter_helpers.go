package stage

import (
	"fmt"
	"time"

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

// newMinimalLua creates a Lua state with a minimal set of libraries opened.
func newMinimalLua() *lua.LState {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	L.Push(L.NewFunction(lua.OpenBase))
	L.Push(lua.LString("base"))
	L.Call(1, 0)
	L.Push(L.NewFunction(lua.OpenString))
	L.Push(lua.LString("string"))
	L.Call(1, 0)
	L.Push(L.NewFunction(lua.OpenTable))
	L.Push(lua.LString("table"))
	L.Call(1, 0)
	L.Push(L.NewFunction(lua.OpenMath))
	L.Push(lua.LString("math"))
	L.Call(1, 0)
	return L
}

// processLuaFilterRecord applies the lua predicate to a single record and
// returns the outcome mimicking the original behavior. The mode determines
// whether to keep going or fail fast.
func processLuaFilterRecord(rec Record, pred string, mode string) (keep bool, out Record, envE *Error, fatal error) {
	locator := rec.Locator
	meta := rec.Meta
	recErr := rec.Error

	if recErr != nil {
		return true, rec, nil, nil
	}

	L := newMinimalLua()
	// Set globals
	L.SetGlobal("locator", lua.LString(locator))
	L.SetGlobal("meta", toLValue(L, meta))

	fn, err := L.LoadString(pred)
	if err != nil {
		if mode == "keep-going" {
			L.Close()
			return true, Record{Locator: locator, Meta: meta, Error: &RecError{Stage: luaFilterStage, Message: err.Error()}}, &Error{Stage: luaFilterStage, Locator: locator, Message: err.Error()}, nil
		}
		L.Close()
		return false, Record{}, nil, fmt.Errorf("lua-filter: %v", err)
	}
	L.Push(fn)
	done := make(chan struct{})
	var callErr error
	go func() {
		callErr = L.PCall(0, 1, nil)
		close(done)
	}()
	select {
	case <-done:
		if callErr != nil {
			if mode == "keep-going" {
				L.Close()
				return true, Record{Locator: locator, Meta: meta, Error: &RecError{Stage: luaFilterStage, Message: callErr.Error()}}, &Error{Stage: luaFilterStage, Locator: locator, Message: callErr.Error()}, nil
			}
			L.Close()
			return false, Record{}, nil, fmt.Errorf("lua-filter: %v", callErr)
		}
	case <-time.After(200 * time.Millisecond):
		if mode == "keep-going" {
			L.Close()
			return true, Record{Locator: locator, Meta: meta, Error: &RecError{Stage: luaFilterStage, Message: "timeout"}}, &Error{Stage: luaFilterStage, Locator: locator, Message: "timeout"}, nil
		}
		L.Close()
		return false, Record{}, nil, fmt.Errorf("lua-filter: timeout")
	}

	ret := L.Get(-1)
	L.Pop(1)
	keep = lua.LVAsBool(ret)
	L.Close()
	if keep {
		return true, Record{Locator: locator, Meta: meta}, nil, nil
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
