package stage

import (
	"context"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func luaReduceRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	// Default: count records
	if in.Meta == nil || in.Meta.Lua == nil || in.Meta.Lua.ReduceInline == "" {
		out := in
		out.Meta.Reduced = len(in.Records)
		return out, nil
	}
	code := in.Meta.Lua.ReduceInline
	if !containsReturn(code) {
		code = "return (" + code + ")"
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

	var acc any
	for _, r := range in.Records {
		var item any
		if rec, ok := r.(Record); ok {
			if rec.Post != nil {
				item = rec.Post
			} else {
				// Convert record to map for Lua
				m := map[string]any{"locator": rec.Locator}
				if rec.Meta != nil {
					m["meta"] = rec.Meta
				}
				if rec.Mapped != nil {
					m["mapped"] = rec.Mapped
				}
				if rec.Shell != nil {
					m["shell"] = map[string]any{"exitCode": rec.Shell.ExitCode, "stdout": rec.Shell.Stdout, "stderr": rec.Shell.Stderr}
				}
				if rec.Post != nil {
					m["post"] = rec.Post
				}
				item = m
			}
		} else {
			item = r
		}

		L.SetGlobal("acc", toLValue(L, acc))
		L.SetGlobal("item", toLValue(L, item))
		fn, err := L.LoadString(code)
		if err != nil {
			return Envelope{}, fmt.Errorf("lua-reduce: %v", err)
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
				return Envelope{}, fmt.Errorf("lua-reduce: %v", callErr)
			}
		case <-time.After(200 * time.Millisecond):
			return Envelope{}, fmt.Errorf("lua-reduce: timeout")
		}
		ret := L.Get(-1)
		L.Pop(1)
		acc = fromLValue(ret)
	}

	out := in
	if out.Meta == nil {
		out.Meta = &Meta{}
	}
	out.Meta.Reduced = acc
	return out, nil
}

func init() { Register("lua-reduce", luaReduceRunner) }
