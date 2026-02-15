package stage

import (
	"errors"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
)

const luaMapStage = "lua-map"

type luaMapRes struct {
	idx   int
	rec   any
	envE  *Error
	fatal error
}

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
func processLuaMapRecord(r any, code string, mode string) (any, *Error, error) {
	var locator string
	var meta map[string]any
	if rec, ok := r.(Record); ok {
		if rec.Error != nil {
			return rec, nil, nil
		}
		locator = rec.Locator
		meta = rec.Meta
	} else if m, ok := r.(map[string]any); ok {
		locVal, ok := m["locator"]
		if !ok {
			if mode == "keep-going" {
				return r, &Error{Stage: luaMapStage, Message: "missing locator"}, nil
			}
			return nil, nil, errors.New("lua-map: missing locator")
		}
		s, ok := locVal.(string)
		if !ok {
			if mode == "keep-going" {
				return r, &Error{Stage: luaMapStage, Message: "invalid locator type"}, nil
			}
			return nil, nil, errors.New("lua-map: invalid locator type")
		}
		locator = s
		meta, _ = m["meta"].(map[string]any)
	} else {
		if mode == "keep-going" {
			return r, &Error{Stage: luaMapStage, Message: "invalid record type"}, nil
		}
		return nil, nil, errors.New("lua-map: invalid record type")
	}

	L := newMinimalLua()
	L.SetGlobal("locator", lua.LString(locator))
	L.SetGlobal("meta", toLValue(L, meta))

	fn, err := L.LoadString(code)
	if err != nil {
		if mode == "keep-going" {
			L.Close()
			return Record{Locator: locator, Meta: meta, Error: &RecError{Stage: luaMapStage, Message: err.Error()}}, &Error{Stage: luaMapStage, Locator: locator, Message: err.Error()}, nil
		}
		L.Close()
		return nil, nil, fmt.Errorf("lua-map: %v", err)
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
				return Record{Locator: locator, Meta: meta, Error: &RecError{Stage: luaMapStage, Message: callErr.Error()}}, &Error{Stage: luaMapStage, Locator: locator, Message: callErr.Error()}, nil
			}
			L.Close()
			return nil, nil, fmt.Errorf("lua-map: %v", callErr)
		}
	case <-time.After(200 * time.Millisecond):
		if mode == "keep-going" {
			L.Close()
			return Record{Locator: locator, Meta: meta, Error: &RecError{Stage: luaMapStage, Message: "timeout"}}, &Error{Stage: luaMapStage, Locator: locator, Message: "timeout"}, nil
		}
		L.Close()
		return nil, nil, fmt.Errorf("lua-map: timeout")
	}
	ret := L.Get(-1)
	L.Pop(1)
	mapped := fromLValue(ret)
	L.Close()
	return Record{Locator: locator, Meta: meta, Mapped: mapped}, nil, nil
}
