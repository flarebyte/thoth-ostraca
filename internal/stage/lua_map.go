package stage

import (
	"context"
	"errors"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func luaMapRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Determine transform
	code := "return { locator = locator, meta = meta }"
	if in.Meta != nil && in.Meta.Lua != nil && in.Meta.Lua.MapInline != "" {
		// Allow expressions without explicit return
		c := in.Meta.Lua.MapInline
		if !containsReturn(c) {
			code = "return (" + c + ")"
		} else {
			code = c
		}
	}

	out := in
	mode, _ := errorMode(in.Meta)
	for i, r := range in.Records {
		var locator string
		var meta map[string]any
		if rec, ok := r.(Record); ok {
			if rec.Error != nil {
				out.Records[i] = r
				continue
			}
			locator = rec.Locator
			meta = rec.Meta
		} else if m, ok := r.(map[string]any); ok {
			locVal, ok := m["locator"]
			if !ok {
				if mode == "keep-going" {
					appendEnvelopeError(&out, "lua-map", "", "missing locator")
					out.Records[i] = r
					continue
				}
				return Envelope{}, errors.New("lua-map: missing locator")
			}
			s, ok := locVal.(string)
			if !ok {
				if mode == "keep-going" {
					appendEnvelopeError(&out, "lua-map", "", "invalid locator type")
					out.Records[i] = r
					continue
				}
				return Envelope{}, errors.New("lua-map: invalid locator type")
			}
			locator = s
			meta, _ = m["meta"].(map[string]any)
		} else {
			if mode == "keep-going" {
				appendEnvelopeError(&out, "lua-map", "", "invalid record type")
				out.Records[i] = r
				continue
			}
			return Envelope{}, errors.New("lua-map: invalid record type")
		}

		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		// Minimal libs
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

		L.SetGlobal("locator", lua.LString(locator))
		L.SetGlobal("meta", toLValue(L, meta))

		fn, err := L.LoadString(code)
		if err != nil {
			if mode == "keep-going" {
				appendEnvelopeError(&out, "lua-map", locator, err.Error())
				out.Records[i] = Record{Locator: locator, Meta: meta, Error: &RecError{Stage: "lua-map", Message: err.Error()}}
				continue
			}
			return Envelope{}, fmt.Errorf("lua-map: %v", err)
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
					appendEnvelopeError(&out, "lua-map", locator, callErr.Error())
					out.Records[i] = Record{Locator: locator, Meta: meta, Error: &RecError{Stage: "lua-map", Message: callErr.Error()}}
					continue
				}
				return Envelope{}, fmt.Errorf("lua-map: %v", callErr)
			}
		case <-time.After(200 * time.Millisecond):
			if mode == "keep-going" {
				appendEnvelopeError(&out, "lua-map", locator, "timeout")
				out.Records[i] = Record{Locator: locator, Meta: meta, Error: &RecError{Stage: "lua-map", Message: "timeout"}}
				continue
			}
			return Envelope{}, fmt.Errorf("lua-map: timeout")
		}
		ret := L.Get(-1)
		L.Pop(1)
		mapped := fromLValue(ret)

		// Build new record retaining original fields
		out.Records[i] = Record{Locator: locator, Meta: meta, Mapped: mapped}
	}
	return out, nil
}

func fromLValue(v lua.LValue) any {
	switch v.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return lua.LVAsBool(v)
	case lua.LTNumber:
		return float64(v.(lua.LNumber))
	case lua.LTString:
		return v.String()
	case lua.LTTable:
		t := v.(*lua.LTable)
		// Decide object vs array by checking numeric keys 1..n
		// We'll build a map if non-sequential keys exist
		// First try array
		arr := []any{}
		isArray := true
		t.ForEach(func(k, val lua.LValue) {
			if isArray {
				if lk, ok := k.(lua.LNumber); ok && int(lk) == len(arr)+1 {
					arr = append(arr, fromLValue(val))
				} else {
					isArray = false
				}
			}
		})
		if isArray {
			return arr
		}
		obj := map[string]any{}
		t.ForEach(func(k, val lua.LValue) {
			obj[k.String()] = fromLValue(val)
		})
		return obj
	default:
		return nil
	}
}

func init() { Register("lua-map", luaMapRunner) }
