package stage

import (
	"context"
	"errors"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func luaFilterRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Determine predicate
	pred := "return true"
	if in.Meta != nil && in.Meta.Lua != nil && in.Meta.Lua.FilterInline != "" {
		code := in.Meta.Lua.FilterInline
		// If code doesn't contain a return, wrap it
		if !containsReturn(code) {
			pred = "return (" + code + ")"
		} else {
			pred = code
		}
	}

	out := in
	out.Records = out.Records[:0]

	for _, r := range in.Records {
		m, ok := r.(map[string]any)
		if !ok {
			return Envelope{}, errors.New("lua-filter: invalid record type")
		}
		locVal, ok := m["locator"]
		if !ok {
			return Envelope{}, errors.New("lua-filter: missing locator")
		}
		locator, ok := locVal.(string)
		if !ok {
			return Envelope{}, errors.New("lua-filter: invalid locator type")
		}
		meta, _ := m["meta"].(map[string]any)

		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		defer L.Close()
		// Open minimal libs
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

		// Set globals
		L.SetGlobal("locator", lua.LString(locator))
		L.SetGlobal("meta", toLValue(L, meta))

		fn, err := L.LoadString(pred)
		if err != nil {
			return Envelope{}, fmt.Errorf("lua-filter: %v", err)
		}
		L.Push(fn)
		// Run with a simple timeout guard
		done := make(chan struct{})
		var callErr error
		go func() {
			callErr = L.PCall(0, 1, nil)
			close(done)
		}()
		select {
		case <-done:
			if callErr != nil {
				return Envelope{}, fmt.Errorf("lua-filter: %v", callErr)
			}
		case <-time.After(200 * time.Millisecond):
			return Envelope{}, fmt.Errorf("lua-filter: timeout")
		}
		ret := L.Get(-1)
		L.Pop(1)
		keep := lua.LVAsBool(ret)
		if keep {
			out.Records = append(out.Records, r)
		}
	}
	return out, nil
}

func containsReturn(s string) bool {
	for i := 0; i+5 <= len(s); i++ {
		if s[i] == 'r' && i+6 <= len(s) && s[i:i+6] == "return" {
			return true
		}
	}
	return false
}

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

func init() { Register("lua-filter", luaFilterRunner) }
